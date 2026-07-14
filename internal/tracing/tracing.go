// Package tracing provides distributed tracing support using OpenTelemetry.
package tracing

import (
	"context"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Tracer wraps the OpenTelemetry tracer provider.
type Tracer struct {
	tp *trace.TracerProvider
}

// NewTracer creates a new tracer instance based on configuration.
// If tracing is disabled, returns an empty tracer.
func NewTracer(cfg *config.Config) (*Tracer, error) {
	if !cfg.Tracing.Enabled {
		return &Tracer{}, nil
	}

	exporter, err := createExporter(cfg.Tracing.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.TraceIDRatioBased(cfg.Tracing.SampleRate)),
	)

	otel.SetTracerProvider(tp)

	return &Tracer{tp: tp}, nil
}

// Shutdown gracefully shuts down the tracer provider.
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.tp != nil {
		return t.tp.Shutdown(ctx)
	}
	return nil
}
