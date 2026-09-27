package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadinessTurnsOffDuringShutdownButLivenessStaysUp(t *testing.T) {
	srv := newServer("test", secretStatus{})
	h := srv.routes()

	if rec := get(h, "/readyz"); rec.Code != http.StatusOK {
		t.Fatalf("readyz before shutdown = %d, want 200", rec.Code)
	}

	srv.ready.Store(false)

	if rec := get(h, "/readyz"); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz during shutdown = %d, want 503", rec.Code)
	}
	// A failing liveness probe would make the kubelet kill a pod that is only draining.
	if rec := get(h, "/healthz"); rec.Code != http.StatusOK {
		t.Fatalf("healthz during shutdown = %d, want 200", rec.Code)
	}
}

func TestInfoReportsVersionAndSecretStatusWithoutValue(t *testing.T) {
	status := secretStatus{Name: "demo-api-key", Loaded: true, Fingerprint: fingerprint([]byte("s3cr3t"))}
	rec := get(newServer("abc123", status).routes(), "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rec.Code)
	}

	var body struct {
		Version string       `json:"version"`
		Secret  secretStatus `json:"secret"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Version != "abc123" {
		t.Errorf("version = %q, want abc123", body.Version)
	}
	if !body.Secret.Loaded || body.Secret.Fingerprint != status.Fingerprint {
		t.Errorf("secret = %+v, want %+v", body.Secret, status)
	}
}

func TestUnknownPathIsNotFound(t *testing.T) {
	if rec := get(newServer("test", secretStatus{}).routes(), "/nope"); rec.Code != http.StatusNotFound {
		t.Fatalf("GET /nope = %d, want 404", rec.Code)
	}
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}
