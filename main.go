package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/mv-kan/opentelemetry-go-example/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	log.Printf("Waiting for connection...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer func() {
		cancel()
	}()
	target := "localhost:4317"
	serviceName := "test-service"
	shutdownTelemetry, err := telemetry.Init(ctx, target, serviceName, time.Second)
	// _, err := telemetry.Init(ctx, target, serviceName, time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdownTelemetry(ctx); err != nil {
			log.Fatalf("failed to shutdown telemetry: %v", err)
		}
	}()

	name := "github.com/mv-kan/opentelemetry-go-example"
	tracer := telemetry.Tracer(name)
	meter := telemetry.Meter(name)

	// Attributes represent additional key-value descriptors that can be bound
	// to a metric observer or recorder.
	commonAttrs := []attribute.KeyValue{
		attribute.String("attrA", "chocolate"),
		attribute.String("attrB", "raspberry"),
		attribute.String("attrC", "vanilla"),
	}

	runCount, err := meter.Int64Counter("run", metric.WithDescription("The number of times the iteration ran"))
	if err != nil {
		log.Fatal(err)
	}

	// Work begins
	ctx, span := tracer.Start(
		ctx,
		"CollectorExporter-Example",
		trace.WithAttributes(commonAttrs...))
	defer span.End()
	for i := 0; i < 10; i++ {
		_, iSpan := tracer.Start(ctx, fmt.Sprintf("Sample-%d", i))
		runCount.Add(ctx, 1, metric.WithAttributes(commonAttrs...))
		log.Printf("Doing really hard work (%d / 10)\n", i+1)

		<-time.After(time.Second)
		iSpan.End()
	}

	log.Printf("Done!")
}
