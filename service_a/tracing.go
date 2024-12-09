package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/correlation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Initialize tracing and set up Jaeger exporter
func InitTracerProvider() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Set up Jaeger exporter
	jaegerURL := os.Getenv("JAEGER_URL")
	if jaegerURL == "" {
		jaegerURL = "http://localhost:5775" // Default Jaeger URL
	}

	exporter, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(jaegerURL),
	)
	if err != nil {
		log.Fatalf("Failed to create Jaeger exporter: %v", err)
	}

	// Create tracer provider and register it globally
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
}

// Start a new trace or span
func StartNewSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	tracer := otel.Tracer("microservice-a")
	ctx, span := tracer.Start(ctx, name)
	span.SetAttributes(attribute.String("service", "microservice-a"))
	return ctx, span
}

// Add custom attributes to spans
func AddAttributesToSpan(span trace.Span, attributes map[string]string) {
	for key, value := range attributes {
		span.SetAttributes(attribute.String(key, value))
	}
}

// Record the error (if any) in the span
func RecordErrorInSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
	}
}

// Shutdown the tracer provider
func ShutdownTracerProvider() {
	if err := otel.GetTracerProvider().Shutdown(context.Background()); err != nil {
		log.Fatalf("Error shutting down TracerProvider: %v", err)
	}
}

