# A Go-based microservice project designed with an SRE (Site Reliability Engineering) focus: Prometheus Monitoring, Grafana, Jaeger Tracing, EFK, Docker, and Kubernetes

A set of Go-based microservices with built-in health checks, Prometheus metrics, and Kubernetes deployment.
 - Monitoring: Custom metrics instrumented and monitored using Prometheus and Grafana
 - Tracing: Distributed tracing implemented using Jaeger for end-to-end request tracking
 - Logging: Centralized logging stack leveraging Fluentbit(FEK) for log aggregration and analysis 
 - Containerization: Distroless Docker build
 - Kubernetes: Deployed using Kustomize and Helm charts on Amazon EKS
 - High Availablility: Configured with Alertmanager for proactive issue detection.