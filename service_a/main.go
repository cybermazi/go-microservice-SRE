package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
)

var (
	// Counter for the number of HTTP requests
	httpRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	// Histogram for the duration of HTTP requests
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status_code"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(httpRequestCounter)
	prometheus.MustRegister(httpRequestDuration)
}

// initializeTracing initializes OpenTelemetry and the Jaeger exporter
func initializeTracing() (func(), error) {
	// Create a Jaeger exporter
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:5775" // Default endpoint if not set
	}

	exp, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(jaegerEndpoint),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "service-a",
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
			attribute.String("service.name", "service-a"),
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

// helloHandler is the handler for /hello endpoint
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Create a new trace span
	tracer := otel.Tracer("service-a")
	ctx, span := tracer.Start(r.Context(), "helloHandler")
	defer span.End()

	// Add some attributes to the span
	span.SetAttributes(attribute.String("method", r.Method))

	// Start tracking the request duration
	start := time.Now()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from Service A!"))

	// Record request duration
	duration := time.Since(start).Seconds()
	httpRequestCounter.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", http.StatusOK)).Inc()
	httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", http.StatusOK)).Observe(duration)

	// You can add custom events or further attributes here
	span.AddEvent("Hello response sent")
}

// metricsHandler exposes the metrics for Prometheus scraping
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type for Prometheus scraping
	w.Header().Set("Content-Type", prometheus.DefaultRegisterer.ContentType())
	// Write the metrics
	promhttp.Handler().ServeHTTP(w, r)
}

func main() {
	// Initialize tracing
	shutdown, err := initializeTracing()
	if err != nil {
		log.Fatal("Error initializing tracing: ", err)
	}
	defer shutdown()

	// Load environment variables from the .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the port from environment variable, default to 3001 if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001" // Default port if not set in .env
	}

	// Set up HTTP handlers
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/metrics", metricsHandler) // Endpoint to expose metrics

	// Create and start the server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      http.DefaultServeMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("Service A is running on port %s", port)
	log.Fatal(srv.ListenAndServe())
}
