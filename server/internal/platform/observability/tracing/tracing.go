package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type Config struct {
	Env                string
	ServiceName        string
	ServiceVersion     string
	OTLPTracesEndpoint string
}

type Provider struct {
	tracer *sdktrace.TracerProvider
}

func NewProvider(ctx context.Context, cfg Config) (*Provider, error) {
	options := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment.name", cfg.Env),
		)),
	}

	if cfg.OTLPTracesEndpoint != "" {
		exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cfg.OTLPTracesEndpoint))
		if err != nil {
			return nil, err
		}
		options = append(options, sdktrace.WithBatcher(exporter))
	}

	tracer := sdktrace.NewTracerProvider(options...)

	otel.SetTracerProvider(tracer)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{tracer: tracer}, nil
}

func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.tracer == nil {
		return nil
	}

	return p.tracer.Shutdown(ctx)
}
