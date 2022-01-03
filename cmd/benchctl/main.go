package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.6.1"
	"log"
	"time"
)

func main() {
	fmt.Println("benchd")

	exporter := otlptracegrpc.NewUnstarted(
		otlptracegrpc.WithInsecure(),
		//otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlptracegrpc.WithEndpoint("localhost:4317"),
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
			semconv.ServiceNameKey.String("benchd"),
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
}
