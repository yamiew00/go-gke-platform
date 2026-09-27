// Command server is a small HTTP service that demonstrates a production-style
// deployment on GKE Autopilot: liveness/readiness probes, graceful shutdown,
// structured logs for Cloud Logging, and reading a Secret Manager secret through
// Workload Identity instead of a key file.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// version is stamped at build time: go build -ldflags "-X main.version=<commit sha>".
var version = "dev"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: cloudLoggingAttrs})))
	cfg := loadConfig()

	srv := newServer(version, loadSecret(cfg))
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	go func() {
		slog.Info("listening", "port", cfg.Port, "version", version)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	// Kubernetes sends SIGTERM and removes the pod from Service endpoints at the same time.
	// Fail readiness first and keep serving for a short delay, so no request lands on a closed socket.
	slog.Info("shutdown started", "drain_delay", cfg.ShutdownDelay.String())
	srv.ready.Store(false)
	time.Sleep(cfg.ShutdownDelay)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown incomplete", "error", err)
		os.Exit(1)
	}
	slog.Info("shutdown complete")
}

// loadSecret reads the demo secret once at startup. A failure is reported rather than fatal,
// so the service still runs locally without any GCP credentials.
func loadSecret(cfg config) secretStatus {
	if cfg.SecretName == "" || cfg.ProjectID == "" {
		return secretStatus{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	value, err := newSecretClient().accessLatest(ctx, cfg.ProjectID, cfg.SecretName)
	if err != nil {
		slog.Warn("secret not loaded", "secret", cfg.SecretName, "error", err)
		return secretStatus{Name: cfg.SecretName, Error: err.Error()}
	}
	slog.Info("secret loaded", "secret", cfg.SecretName, "fingerprint", fingerprint(value))
	return secretStatus{Name: cfg.SecretName, Loaded: true, Fingerprint: fingerprint(value)}
}

// cloudLoggingAttrs renames slog's keys to the fields Cloud Logging understands,
// so log levels show up as severities in Logs Explorer.
func cloudLoggingAttrs(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.LevelKey:
		a.Key = "severity"
	case slog.MessageKey:
		a.Key = "message"
	}
	return a
}
