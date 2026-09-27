package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeGoogle stands in for both the GKE metadata server and the Secret Manager API.
func fakeGoogle(t *testing.T, secretStatus int, payload []byte) *secretClient {
	t.Helper()
	const token = "test-token"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if r.Header.Get("Metadata-Flavor") != "Google" {
				http.Error(w, "missing Metadata-Flavor header", http.StatusForbidden)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": token})
		case "/v1/projects/demo-project/secrets/demo-api-key/versions/latest:access":
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "bad token", http.StatusUnauthorized)
				return
			}
			if secretStatus != http.StatusOK {
				http.Error(w, "denied", secretStatus)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"payload": map[string]string{"data": base64.StdEncoding.EncodeToString(payload)},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return &secretClient{http: srv.Client(), metadataURL: srv.URL + "/token", secretManager: srv.URL + "/v1"}
}

func TestAccessLatestReturnsPayload(t *testing.T) {
	c := fakeGoogle(t, http.StatusOK, []byte("s3cr3t"))
	got, err := c.accessLatest(context.Background(), "demo-project", "demo-api-key")
	if err != nil {
		t.Fatalf("accessLatest: %v", err)
	}
	if string(got) != "s3cr3t" {
		t.Fatalf("payload = %q, want s3cr3t", got)
	}
}

func TestAccessLatestSurfacesPermissionDenied(t *testing.T) {
	c := fakeGoogle(t, http.StatusForbidden, nil)
	_, err := c.accessLatest(context.Background(), "demo-project", "demo-api-key")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("err = %v, want a 403 error", err)
	}
}

func TestFingerprintDoesNotRevealValue(t *testing.T) {
	fp := fingerprint([]byte("s3cr3t"))
	if len(fp) != 12 || strings.Contains(fp, "s3cr3t") {
		t.Fatalf("fingerprint = %q", fp)
	}
	if fp == fingerprint([]byte("other")) {
		t.Fatal("different values produced the same fingerprint")
	}
}
