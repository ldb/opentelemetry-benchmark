package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ldb/openetelemtry-benchmark/config"
	"github.com/ldb/openetelemtry-benchmark/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// start listening for connections on port for ctl
// start listening for connections from SUT
// start OTEL receiver
// receive config
// dump config into log file
// initialize thread pool
// initialize n workers
// run benchmark according to config
// record timing for each sent / received request

func main() {
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithInsecure(),
		//otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlptracegrpc.WithEndpoint("localhost:4318"),
	)
	if err != nil {
		log.Fatalf("setup exporter: %v", err.Error())
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("solotraced-poc"),
		),
	)
	if err != nil {
		log.Fatalf("create resource: %v", err.Error())
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
	)
	otel.SetTracerProvider(tp)

	cfg := config.WorkerConfig{
		TraceDepth:  1,
		NumberSpans: 1,
		SpanLength:  1 * time.Second,
		MaxCoolDown: 1 * time.Second,
	}

	workerManager := worker.NewManager("test", cfg)
	workerManager.TracerProvider = tp
	if err := workerManager.AddWorkers(2); err != nil {
		log.Fatalf("error adding workers: %v", err)
	}

	workerManager.Start()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	go func() {
		for {
			time.Sleep(3 * time.Second)
			for i := 1; i <= 2; i++ {
				if err := workerManager.FinishTrace(i); err != nil {
					log.Fatalf("error finishing trace for worker %d", i)
				}

			}
		}
	}()

	workerManager2 := worker.NewManager("test2", cfg)
	workerManager2.TracerProvider = tp
	if err := workerManager2.AddWorkers(2); err != nil {
		log.Fatalf("error adding workers: %v", err)
	}

	workerManager2.Start()
	for {
		time.Sleep(3 * time.Second)
		for i := 1; i <= 2; i++ {
			if err := workerManager2.FinishTrace(i); err != nil {
				log.Fatalf("error finishing trace for worker %d", i)
			}

		}
	}

}
