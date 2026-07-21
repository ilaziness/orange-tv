package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/spf13/cobra"
)

const healthRequestTimeout = 5 * time.Second

var (
	healthURL   string
	healthProbe string
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check application health via HTTP probe",
	Long: `Probe the running HTTP server for health status.
Defaults to /readiness (includes dependency checks such as database).
Uses https when http.tls.enabled is true (skips TLS verify for local/self-signed certs).
Exits 0 on HTTP 2xx, 1 otherwise.`,
	RunE: runHealth,
}

func init() {
	healthCmd.Flags().StringVar(&healthURL, "url", "", "Full probe URL (string URL; overrides --probe and config)")
	healthCmd.Flags().StringVar(&healthProbe, "probe", "readiness", "Probe name (string: health, readiness, or liveness)")
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	targetURL, insecureTLS, err := resolveHealthProbe()
	if err != nil {
		return err
	}

	client := newHealthClient(insecureTLS)
	resp, err := client.Get(targetURL)
	if err != nil {
		return fmt.Errorf("health probe failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("health probe returned status %d for %s", resp.StatusCode, targetURL)
	}

	return nil
}

func resolveHealthProbe() (targetURL string, insecureTLS bool, err error) {
	if healthURL != "" {
		return healthURL, strings.HasPrefix(healthURL, "https://"), nil
	}

	cfg, err := config.LoadWithEnv(cfgFile, env)
	if err != nil {
		return "", false, fmt.Errorf("failed to load config: %w", err)
	}

	url, err := buildHealthURL(cfg, healthProbe)
	if err != nil {
		return "", false, err
	}

	return url, cfg.HTTP.TLS.Enabled, nil
}

func buildHealthURL(cfg *config.Config, probe string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is required")
	}
	if !cfg.HTTP.Enabled {
		return "", fmt.Errorf("http server is disabled; enable http.enabled or pass --url")
	}

	path, err := probePath(probe)
	if err != nil {
		return "", err
	}

	host := cfg.HTTP.Host
	if host == "0.0.0.0" || host == "" {
		host = "127.0.0.1"
	}

	scheme := "http"
	if cfg.HTTP.TLS.Enabled {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s:%d%s", scheme, host, cfg.HTTP.Port, path), nil
}

func newHealthClient(insecureTLS bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // local health probe against self-signed certs
	}
	return &http.Client{
		Timeout:   healthRequestTimeout,
		Transport: transport,
	}
}

func probePath(probe string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(probe)) {
	case "health":
		return router.PathHealth, nil
	case "readiness", "ready":
		return router.PathReadiness, nil
	case "liveness", "live":
		return router.PathLiveness, nil
	default:
		return "", fmt.Errorf("invalid probe %q: use health, readiness, or liveness", probe)
	}
}
