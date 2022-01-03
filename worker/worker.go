package worker

import (
	"context"
	"fmt"
	"github.com/ldb/openetelemtry-benchmark/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"log"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type status int

const (
	success status = iota
	timeout
	started
	stopped
)

type Logger interface {
	Println(m ...interface{})
}

type Worker struct {
	managerName string
	ID          int

	Config config.WorkerConfig

	tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
	FinishTrace    chan struct{} // Manager notifies the worker on this channel that it can stop recording the current trace
	Logger         Logger

	Timeout time.Duration

	// recorded Values
	traceDepth    int
	spanLength    time.Duration
	coolDown      time.Duration
	startT        time.Time
	sendT         time.Time
	receiveT      time.Time
	sentReceivedD time.Duration
}

func (w *Worker) initTracer() {
	exporter := otlptracegrpc.NewUnstarted(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("otel-collector:4317"),
	)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := exporter.Start(ctx)
	if err != nil {
		log.Fatalf("setup exporter: %v", err.Error())
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(fmt.Sprintf("benchd-worker.%s.%d", w.managerName, w.ID)),
		),
	)
	if err != nil {
		log.Fatalf("create resource: %v", err.Error())
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)
	w.tracer = tp.Tracer("")
	w.tracerProvider = tp
}

func (w *Worker) Run(ctx context.Context) error {

	w.log(started)
	//timeoutTimer := time.NewTimer(w.Timeout)
	//defer timeoutTimer.Stop()
	for {
		w.startT = time.Now()
		w.generateTrace()
		w.tracerProvider.ForceFlush(context.Background())
		w.sendT = time.Now()
		tracesSent.WithLabelValues(w.managerName).Inc()
		//	timeoutTimer.Reset()
		select {
		case <-ctx.Done():
			w.log(stopped)
			return ctx.Err()

		case <-w.FinishTrace:
			w.receiveT = time.Now()
			w.sentReceivedD = w.receiveT.Sub(w.sendT)
			//	timeoutTimer.Stop()
			cooldown := time.Duration(rand.Int63n(w.Config.MaxCoolDown.Milliseconds())) * time.Millisecond
			w.coolDown = cooldown
			w.log(success)
			traceRoundtrip.WithLabelValues(w.managerName).Observe(w.sentReceivedD.Seconds())
			time.Sleep(cooldown)
			w.reset()

			//case <-timeoutTimer.C:
			//	w.log(timeout)
		}
	}
}

// log sends a log message of the recorded timings into the (*Worker).Log channel.
func (w *Worker) log(s status) {
	w.Logger.Println(fmt.Sprintf("%s %d %d %d %d %d %d %d %d %d",
		w.managerName,
		w.ID,
		int(s),
		w.traceDepth,
		w.spanLength.Milliseconds(),
		w.coolDown.Milliseconds(),
		w.startT.UnixMilli(),
		w.sendT.UnixMilli(),
		w.receiveT.UnixMilli(),
		w.sentReceivedD.Milliseconds(),
	))
}

func (w *Worker) generateTrace() {
	d := rand.Intn(w.Config.MaxTraceDepth)
	w.traceDepth = d
	ctx, trace := w.tracer.Start(context.Background(), "parentTrace")
	w.child(ctx, d)
	trace.End()
}

func (w *Worker) child(ctx context.Context, depth int) {
	cctx, sp := w.tracer.Start(ctx, fmt.Sprintf("worker.%d.child.%d", w.ID, depth))
	sl := time.Duration(rand.Int63n(w.Config.MaxSpanLength.Milliseconds())) * time.Millisecond
	w.spanLength += sl
	time.Sleep(sl)
	defer sp.End()
	if depth > 1 {
		sp.SetAttributes(attribute.Bool("hasChildren", true))
		sp.AddEvent("spawning child", trace.WithAttributes(attribute.Int("depth", depth)))
		w.child(cctx, depth-1)
	}
	/*sp.SetStatus(codes.Ok, "all good")
	if rand.Intn(100) < 30 {
		sp.SetStatus(codes.Ok, "all good")
	} else {
		sp.SetStatus(codes.Error, "something went wrong")
	}
	sp.AddEvent("stopping")*/
}

func (w *Worker) reset() {
	w.traceDepth = 0
	w.spanLength = 0
	w.coolDown = 0
	w.startT = time.Time{}
	w.sendT = time.Time{}
	w.receiveT = time.Time{}
	w.sentReceivedD = 0
}
