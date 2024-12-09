package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/resource"
)

func initTracer() (func(), error) {
	// Create a Jaeger exporter
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:5775" // Default endpoint
	}

	exp, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(jaegerEndpoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "service-b",
			Tags: []attribute.KeyValue{
				attribute.String("env", "production"),
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Set up the tracer provider with the Jaeger exporter
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			attribute.String("service.name", "service-b"),
		)),
	)

	// Register the tracer provider globally
	otel.SetTracerProvider(tp)

	// Return a function to stop the tracer provider when the application exits
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to stop tracer provider: %v", err)
		}
	}, nil
}

// Add tracing to your HTTP handlers, for example, for the /hello endpoint
func traceHandler(w http.ResponseWriter, r *http.Request) {
	// Create a new trace span
	tracer := otel.Tracer("service-b")
	ctx, span := tracer.Start(r.Context(), "helloHandler")
	defer span.End()

	// Add some attributes to the span
	span.SetAttributes(attribute.String("method", r.Method))

	// Simulate the handler logic
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from Service B!"))

	// You can add custom events or further attributes here
	span.AddEvent("Hello response sent")

	// Record duration if needed, you can use `time.Since(start)` as well
}
