package worker

import (
	"context"
	"fmt"
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
	TraceDepth  int
	NumberSpans int
	SpanLength  time.Duration
	MaxCoolDown time.Duration

	tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
	FinishTrace    chan struct{} // Manager notifies the worker on this channel that it can stop recording the current trace
	Logger         Logger

	Timeout time.Duration

	startT      time.Time
	sendT       time.Time
	finishT     time.Time
	sentFinishD time.Duration
}

func (w *Worker) initTracer() {
	exporter := otlptracegrpc.NewUnstarted(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("otel-collector:4317"),
	)
	log.Println("created exporter")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := exporter.Start(ctx)
	if err != nil {
		log.Fatalf("setup exporter: %v", err.Error())
	}
	log.Println("started exporter")
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
		w.generateSpans()
		w.tracerProvider.ForceFlush(context.Background())
		w.sendT = time.Now()
		tracesSent.WithLabelValues(w.managerName).Inc()
		//	timeoutTimer.Reset()
		select {
		case <-ctx.Done():
			w.log(stopped)
			return ctx.Err()

		case <-w.FinishTrace:
			w.finishT = time.Now()
			w.sentFinishD = w.finishT.Sub(w.sendT)
			//	timeoutTimer.Stop()
			w.log(success)
			traceRoundtrip.WithLabelValues(w.managerName).Observe(w.sentFinishD.Seconds())
			time.Sleep(time.Duration(rand.Int63n(w.MaxCoolDown.Milliseconds())))

			//case <-timeoutTimer.C:
			//	w.log(timeout)
		}
	}
}

// log sends a log message of the recorded timings into the (*Worker).Log channel.
func (w *Worker) log(s status) {
	w.Logger.Println(fmt.Sprintf("%s %d %d %d %d %d %d %d %d %d %d",
		w.managerName,
		w.ID,
		int(s),
		w.TraceDepth,
		w.NumberSpans,
		w.SpanLength.Milliseconds(),
		w.MaxCoolDown.Milliseconds(),
		w.startT.UnixMilli(),
		w.sendT.UnixMilli(),
		w.finishT.UnixMilli(),
		w.sentFinishD.Milliseconds(),
	))
}

func (w *Worker) generateSpans() {
	_, span := w.tracer.Start(context.Background(), "generate")
	time.Sleep(1 * time.Second)
	span.End()

}
