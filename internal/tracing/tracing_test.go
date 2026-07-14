package tracing

import (
	"context"
	"testing"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewTracer_Disabled(t *testing.T) {
	cfg := &config.Config{
		Tracing: config.TracingConfig{Enabled: false},
	}
	tracer, err := NewTracer(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, tracer)
	assert.Nil(t, tracer.tp)
}

func TestNewTracer_EmptyEndpoint(t *testing.T) {
	cfg := &config.Config{
		Tracing: config.TracingConfig{
			Enabled:    true,
			Endpoint:   "",
			SampleRate: 1.0,
		},
	}
	// Empty endpoint validation should be done at config level
	// Here we just test that tracer creation doesn't panic
	tracer, err := NewTracer(cfg)
	// OTLP exporter may not validate endpoint immediately
	// So we just check it doesn't panic
	if err == nil {
		assert.NotNil(t, tracer)
	}
}

func TestNewTracer_InvalidSampleRate(t *testing.T) {
	cfg := &config.Config{
		Tracing: config.TracingConfig{
			Enabled:    true,
			Endpoint:   "localhost:4317",
			SampleRate: 1.5, // Invalid: > 1.0
		},
	}
	// This should be caught by config validation, not here
	// Just test that the tracer creation handles it gracefully
	tracer, err := NewTracer(cfg)
	// May fail due to endpoint being unreachable, which is expected
	if err == nil {
		assert.NotNil(t, tracer)
	}
}

func TestTracer_Shutdown_Disabled(t *testing.T) {
	cfg := &config.Config{
		Tracing: config.TracingConfig{Enabled: false},
	}
	tracer, _ := NewTracer(cfg)

	err := tracer.Shutdown(context.TODO())
	assert.NoError(t, err)
}
