package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/joho/godotenv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// helloHandler is the handler for /hello endpoint
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Start tracking the request duration
	start := time.Now()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from Service B!"))

	// Record request duration
	duration := time.Since(start).Seconds()
	httpRequestCounter.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", http.StatusOK)).Inc()
	httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", http.StatusOK)).Observe(duration)
}

// metricsHandler exposes the metrics for Prometheus scraping
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type for Prometheus scraping
	w.Header().Set("Content-Type", prometheus.DefaultRegisterer.ContentType())
	// Write the metrics
	promhttp.Handler().ServeHTTP(w, r)
}

func main() {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the port from environment variable, default to 3002 if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002" // Default port if not set in .env
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

	log.Printf("Service B is running on port %s", port)
	log.Fatal(srv.ListenAndServe())
}
