package telemetry

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	"go.opentelemetry.io/otel/sdk/resource"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Telemetry struct {
	TracerShutdown func(context.Context) error
	MeterProvider  *sdkmetric.MeterProvider
}

func Init(ctx context.Context) (*Telemetry, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	log.Printf("OTEL_EXPORTER_OTLP_ENDPOINT='%s'", endpoint)

	if endpoint == "" {
		endpoint = "adot.internal:4317"
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "go-otel-demo"
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("environment", "poc"),
		),
	)

	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)

	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter),
		),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)

	meter := meterProvider.Meter("pix")

	PixDuration, err = meter.Int64Histogram(
		"pix_confirmation_duration_ms",
		metric.WithDescription(
			"PIX confirmation duration in milliseconds",
		),
	)

	if err != nil {
		return nil, err
	}

	meter = meterProvider.Meter("vsync")

	VsyncExecutions, _ = meter.Int64Counter(
		"vsync_job_executions_total",
	)

	VsyncPartnerOK, _ = meter.Int64Counter(
		"vsync_partner_success_total",
	)

	VsyncPartnerFailed, _ = meter.Int64Counter(
		"vsync_partner_failed_total",
	)

	VsyncInvalid, _ = meter.Int64Counter(
		"vsync_invalid_total",
	)

	return &Telemetry{
		TracerShutdown: tracerProvider.Shutdown,
		MeterProvider:  meterProvider,
	}, nil
}
