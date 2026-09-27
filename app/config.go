package main

import (
	"os"
	"time"
)

// config is read once from the environment, which is how both the Helm chart
// and a local `go run .` configure the service.
type config struct {
	Port            string
	ProjectID       string        // GCP project that owns the secret
	SecretName      string        // Secret Manager secret read at startup; empty disables it
	ShutdownDelay   time.Duration // keep serving after readiness turns off, so endpoints can drain
	ShutdownTimeout time.Duration // upper bound for in-flight requests to finish
}

func loadConfig() config {
	return config{
		Port:            getenv("PORT", "8080"),
		ProjectID:       os.Getenv("GOOGLE_CLOUD_PROJECT"),
		SecretName:      os.Getenv("DEMO_SECRET_NAME"),
		ShutdownDelay:   durationEnv("SHUTDOWN_DELAY", 5*time.Second),
		ShutdownTimeout: durationEnv("SHUTDOWN_TIMEOUT", 15*time.Second),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return d
	}
	return fallback
}
