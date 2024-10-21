package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// shorthand for otel.Tracer(name, opts...)
func Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(name, opts...)
}

// shorthand for otel.Meter(name, opts...)
func Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return otel.Meter(name, opts...)
}

func Init(ctx context.Context, target string, serviceName string) (func(context.Context) error, error) {
	conn, err := initConn(target)
	if err != nil {
		return nil, err
	}
	res, err := initResource(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	closeTracer, err := initTracerProvider(ctx, res, conn)
	if err != nil {
		return closeTracer, err
	}
	closeMeter, err := initMeterProvider(ctx, res, conn)
	if err != nil {
		return closeMeter, err
	}
	return func(ctx context.Context) error {
		err1 := closeMeter(ctx)
		err2 := closeTracer(ctx)
		if err1 != nil && err2 != nil {
			return fmt.Errorf("err close meter %w, err close tracer %w", err1, err2)
		}
		if err1 != nil {
			return fmt.Errorf("err close meter %w", err1)
		}
		if err2 != nil {
			return fmt.Errorf("err close tracer %w", err2)
		}
		return nil
	}, nil
}

func initConn(target string) (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// You may replace `localhost:4317` with your endpoint.
	conn, err := grpc.NewClient(target,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

func initResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// The service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	return res, err
}

// Initializes an OTLP exporter, and configures the corresponding trace provider.
func initTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

// Initializes an OTLP exporter, and configures the corresponding meter provider.
func initMeterProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown, nil
}
