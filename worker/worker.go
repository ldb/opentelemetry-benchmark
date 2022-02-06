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
	traceDepth          int
	riskyAttributeDepth int
	extraAttributes     int
	spanLength          time.Duration
	coolDown            time.Duration
	startT              time.Time     // Start time of run
	sendT               time.Time     // Beginning to send
	sendET              time.Time     // Done sending
	receiveT            time.Time     // Received values back
	sentReceivedD       time.Duration // Delta between sendET and receiveT
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
				e := err.Error()
				if len(e) >= 100 {
					// Read only 100 first chars of error message because of label cardinality
					e = e[:100]
				}
				workerErrors.WithLabelValues(w.managerName, e).Inc()
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
			return fmt.Errorf("error flushing trace: %w", err)
		}
		w.log(statusSendTimeout)
		return fmt.Errorf("send timeout: %w", sendTimeout.Err())
	}
	w.sendET = time.Now()
	tracesSent.WithLabelValues(w.managerName).Inc()
	receiveTimeout, cancelReceive := context.WithTimeout(context.Background(), w.Config.ReceiveTimeout.Duration)
	defer cancelReceive()
	select {
	case <-ctx.Done():
		activeWorkers.WithLabelValues(w.managerName).Dec()
		w.log(statusStopped)
		return fmt.Errorf("worker cancelled: %v", ctx.Err())

	case <-receiveTimeout.Done():
		w.receiveT = time.Now()
		w.sentReceivedD = w.receiveT.Sub(w.sendET)
		w.log(statusReceiveTimeout)
		return fmt.Errorf("receive timeout: %w", sendTimeout.Err())

	case <-w.FinishTrace:
		w.receiveT = time.Now()
		w.sentReceivedD = w.receiveT.Sub(w.sendET)
		tracesReceived.WithLabelValues(w.managerName).Inc()
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
	w.Logger.Println(fmt.Sprintf("%d %d %d %d %d %d %d %d %d %d %d %d",
		w.ID,                           // worker ID
		int(s),                         // worker status code
		w.traceDepth,                   // trace depth
		w.riskyAttributeDepth,          // depth of risky attribute
		w.extraAttributes,              // number of extra attributes in trace
		w.spanLength.Milliseconds(),    // accumulated spanlength
		w.coolDown.Milliseconds(),      // cooldown
		w.startT.UnixMilli(),           // start worker
		w.sendT.UnixMilli(),            // start sending payload
		w.sendET.UnixMilli(),           // end sending
		w.receiveT.UnixMilli(),         // receive response
		w.sentReceivedD.Milliseconds(), // delta end sending and receive
	))
}

func (w *Worker) generateTrace() {
	d := rand.Intn(w.Config.MaxTraceDepth)
	w.traceDepth = d
	ctx, trace := w.tracer.Start(context.Background(), "parentTrace")
	riskyAtDepth := 0
	if w.Config.RiskyAttributeProbability > 0 && d > 0 && rand.Intn(100) <= w.Config.RiskyAttributeProbability {
		riskyAtDepth = rand.Intn(d)
	}
	w.child(ctx, d, riskyAtDepth)
	trace.End()
}

func (w *Worker) child(ctx context.Context, maxDepth, riskyAtDepth int) {
	cctx, sp := w.tracer.Start(ctx, fmt.Sprintf("worker.%d.child.%d", w.ID, maxDepth))
	sl := time.Duration(rand.Int63n(w.Config.MaxSpanLength.Milliseconds())) * time.Millisecond
	if w.Config.MaxExtraAttributes > 0 {
		a := rand.Intn(w.Config.MaxExtraAttributes)
		for i := 0; i <= a; i++ {
			sp.SetAttributes(attribute.Int(fmt.Sprintf("extraAttribute-%d", i), i))
		}
		w.extraAttributes += a
	}
	if riskyAtDepth == maxDepth {
		w.riskyAttributeDepth = riskyAtDepth
		sp.SetAttributes(attribute.Int("risky", w.ID))
	}
	w.spanLength += sl
	time.Sleep(sl)
	defer sp.End()
	if maxDepth > 1 {
		sp.SetAttributes(attribute.Bool("hasChildren", true))
		sp.AddEvent("spawning child", trace.WithAttributes(attribute.Int("maxDepth", maxDepth)))
		w.child(cctx, maxDepth-1, riskyAtDepth)
	}
}

func (w *Worker) reset() {
	w.traceDepth = 0
	w.riskyAttributeDepth = 0
	w.extraAttributes = 0
	w.spanLength = 0
	w.coolDown = 0
	w.startT = time.Time{}
	w.sendT = time.Time{}
	w.receiveT = time.Time{}
	w.sentReceivedD = 0
}
