package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type server struct {
	version  string
	hostname string
	secret   secretStatus
	ready    atomic.Bool
}

// secretStatus is safe to expose: it never contains the secret value itself.
type secretStatus struct {
	Name        string `json:"name,omitempty"`
	Loaded      bool   `json:"loaded"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Error       string `json:"error,omitempty"`
}

func newServer(version string, secret secretStatus) *server {
	host, _ := os.Hostname() // the pod name on Kubernetes
	s := &server{version: version, hostname: host, secret: secret}
	s.ready.Store(true)
	return s
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleInfo)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /readyz", s.handleReady)
	return logRequests(mux)
}

// handleInfo reports which build and pod served the request, and whether the
// Secret Manager secret could be read through Workload Identity.
func (s *server) handleInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "go-gke-platform",
		"version": s.version,
		"pod":     s.hostname,
		"secret":  s.secret,
	})
}

// handleHealth is the liveness probe: the process is up and able to serve.
func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReady is the readiness probe. It returns 503 as soon as shutdown starts, so the pod
// leaves Service endpoints before the server stops accepting connections.
func (s *server) handleReady(w http.ResponseWriter, _ *http.Request) {
	if !s.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "shutting down"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logRequests writes one structured log line per request. Probe traffic is skipped to keep logs useful.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			return
		}
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds())
	})
}
