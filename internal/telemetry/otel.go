package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Shutdown func(context.Context) error

// Init configures OpenTelemetry with simple stdout exporters for both traces and metrics.
// For production you'd swap stdout exporters with OTLP (Prometheus/Tempo/etc.),
func Init(ctx context.Context) (Shutdown, error) {
	res, _ := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName("acai-chat")),
	)
	// traces -> stdout
	tExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(tExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// metrics -> stdout
	mExp, err := stdoutmetric.New(stdoutmetric.WithWriter(os.Stdout),
		stdoutmetric.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}
	mp := metric.NewMeterProvider(metric.WithReader(
		metric.NewPeriodicReader(mExp)), metric.WithResource(res))
	otel.SetMeterProvider(mp)

	return func(ctx context.Context) error {
		_ = mp.Shutdown(ctx)
		return tp.Shutdown(ctx)
	}, nil
}
