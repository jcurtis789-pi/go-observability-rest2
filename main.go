package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var tracer = otel.Tracer("gor2")
var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "processed_requests",
		Help: "The total number of requests processed",
	})
)

func handler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(context.Background(), "handle")
	defer span.End()

	opsProcessed.Inc()
	fmt.Fprintf(w, "bar")
}

func main() {
	headers := map[string]string{
		"content-type": "application/json",
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint("jaeger-inmemory-instance-collector.default.svc.cluster.local:4318"),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		log.Fatalf("failed...")
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-observability-rest2"),
		)),
	)
	otel.SetTracerProvider(tp)

	fmt.Println("Tracing enabled...")
	http.HandleFunc("/", handler)
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
