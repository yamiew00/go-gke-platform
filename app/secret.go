package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// secretClient reads a Secret Manager secret using only the standard library.
//
// On GKE with Workload Identity Federation the pod never holds a key file: the GKE
// metadata server issues a short-lived access token for the pod's Kubernetes service
// account, and IAM on the secret decides whether that identity may read it.
type secretClient struct {
	http          *http.Client
	metadataURL   string // token endpoint of the GKE metadata server
	secretManager string // Secret Manager REST base URL
}

func newSecretClient() *secretClient {
	return &secretClient{
		http:          &http.Client{Timeout: 10 * time.Second},
		metadataURL:   "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token",
		secretManager: "https://secretmanager.googleapis.com/v1",
	}
}

// accessLatest returns the payload of the latest enabled version of project/secret.
func (c *secretClient) accessLatest(ctx context.Context, project, secret string) ([]byte, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	url := fmt.Sprintf("%s/projects/%s/secrets/%s/versions/latest:access", c.secretManager, project, secret)
	var out struct {
		Payload struct {
			Data string `json:"data"`
		} `json:"payload"`
	}
	if err := c.getJSON(ctx, url, map[string]string{"Authorization": "Bearer " + token}, &out); err != nil {
		return nil, fmt.Errorf("access secret %s: %w", secret, err)
	}
	return base64.StdEncoding.DecodeString(out.Payload.Data)
}

func (c *secretClient) token(ctx context.Context) (string, error) {
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := c.getJSON(ctx, c.metadataURL, map[string]string{"Metadata-Flavor": "Google"}, &out); err != nil {
		return "", err
	}
	return out.AccessToken, nil
}

func (c *secretClient) getJSON(ctx context.Context, url string, headers map[string]string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	for k, val := range headers {
		req.Header.Set(k, val)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %s", req.URL.Host, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// fingerprint identifies a secret value without revealing it: the first 12 hex chars of its SHA-256.
func fingerprint(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:12]
}
