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
	statusInitialized status = iota
	statusSuccess
	statusSendTimeout
	statusSendError
	statusReceiveTimeout
	statusStopped
)

type Logger interface {
	Println(m ...interface{})
	Printf(format string, v ...interface{})
}

type Worker struct {
	managerName string
	ID          int

	Config config.WorkerConfig

	tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
	FinishTrace    chan struct{} // Manager notifies the worker on this channel that it can stop recording the current trace
	Logger         Logger

	// recorded Values
	traceDepth    int
	spanLength    time.Duration
	coolDown      time.Duration
	startT        time.Time
	sendT         time.Time
	receiveT      time.Time
	sentReceivedD time.Duration
}

func (w *Worker) initTracer(target string) {
	exporter := otlptracegrpc.NewUnstarted(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(target),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
	w.tracer = tp.Tracer(fmt.Sprintf("M:%s-W:%d", w.managerName, w.ID))
	w.tracerProvider = tp
	w.log(statusInitialized)
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		// w.run should not be inlined here as to avoid a defer loop.
		err := w.run(ctx)
		if err != nil {
			if err != context.Canceled {
				workerErrors.WithLabelValues(w.managerName, err.Error()).Inc()
			}
			return err
		}
	}
}

func (w *Worker) run(ctx context.Context) error {
	w.startT = time.Now()
	w.generateTrace()
	defer w.reset()
	sendTimeout, cancelSend := context.WithTimeout(context.Background(), w.Config.SendTimeout.Duration)
	defer cancelSend()
	w.sendT = time.Now()
	if err := w.tracerProvider.ForceFlush(sendTimeout); err != nil {
		w.receiveT = time.Now()
		w.sentReceivedD = w.receiveT.Sub(w.sendT)
		if err != context.DeadlineExceeded {
			w.log(statusSendError)
			return err
		}
		w.log(statusSendTimeout)
		return fmt.Errorf("send timeout: %w", sendTimeout.Err())
	}
	tracesSent.WithLabelValues(w.managerName).Inc()
	receiveTimeout, cancelReceive := context.WithTimeout(context.Background(), w.Config.ReceiveTimeout.Duration)
	defer cancelReceive()
	select {
	case <-ctx.Done():
		activeWorkers.WithLabelValues(w.managerName).Dec()
		w.log(statusStopped)
		return ctx.Err()

	case <-receiveTimeout.Done():
		w.receiveT = time.Now()
		w.sentReceivedD = w.receiveT.Sub(w.sendT)
		w.log(statusReceiveTimeout)
		return fmt.Errorf("receive timeout: %w", sendTimeout.Err())

	case <-w.FinishTrace:
		w.receiveT = time.Now()
		w.sentReceivedD = w.receiveT.Sub(w.sendT)
		cooldown := time.Duration(rand.Int63n(w.Config.MaxCoolDown.Milliseconds())) * time.Millisecond
		w.coolDown = cooldown
		w.log(statusSuccess)
		traceRoundtrip.WithLabelValues(w.managerName).Observe(w.sentReceivedD.Seconds())
		time.Sleep(cooldown)
	}
	return nil
}

// log logs the last request to w.Logger.
func (w *Worker) log(s status) {
	w.Logger.Println(fmt.Sprintf("%d %d %d %d %d %d %d %d %d",
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
