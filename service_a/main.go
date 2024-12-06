package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Define the custom metrics
	httpRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)
)

func init() {
	// Register the metrics with Prometheus
	prometheus.MustRegister(httpRequestCounter)
}

// Health endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Increment the counter metric
	httpRequestCounter.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", http.StatusOK)).Inc()

	response := map[string]string{
		"status": "healthy",
	}
	jsonResponse(w, response, http.StatusOK)
}

// Root endpoint
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Increment the counter metric
	httpRequestCounter.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", http.StatusOK)).Inc()

	response := map[string]string{
		"status": "running",
	}
	jsonResponse(w, response, http.StatusOK)
}

// Crash endpoint for testing error handling
func crashHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate a crash (panic)
	panic("This is a crash test!")
}

// Helper function to send JSON responses
func jsonResponse(w http.ResponseWriter, data map[string]string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Main function
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Register the HTTP handlers
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/healthy", healthHandler)
	http.HandleFunc("/crash", crashHandler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001" // Default port
	}

	log.Printf("Service is running on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
