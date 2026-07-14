package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildHealthURL_HTTP(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{Enabled: true, Host: "0.0.0.0", Port: 8080},
	}

	url, err := buildHealthURL(cfg, "readiness")
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:8080/readiness", url)
}

func TestBuildHealthURL_HTTPS(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Enabled: true,
			Host:    "127.0.0.1",
			Port:    8443,
			TLS:     config.TLSConfig{Enabled: true},
		},
	}

	url, err := buildHealthURL(cfg, "liveness")
	require.NoError(t, err)
	assert.Equal(t, "https://127.0.0.1:8443/liveness", url)
}

func TestBuildHealthURL_HTTPDisabled(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{Enabled: false},
	}

	_, err := buildHealthURL(cfg, "readiness")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http server is disabled")
}

func TestRunHealth_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	healthURL = server.URL + "/readiness"
	healthProbe = "readiness"
	t.Cleanup(func() {
		healthURL = ""
		healthProbe = "readiness"
	})

	err := runHealth(nil, nil)
	require.NoError(t, err)
}

func TestRunHealth_non2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	healthURL = server.URL + "/readiness"
	t.Cleanup(func() { healthURL = "" })

	err := runHealth(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "503")
}
