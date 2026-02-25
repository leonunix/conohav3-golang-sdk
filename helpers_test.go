package conoha

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupTestServer creates an httptest.Server and a Client with all endpoint URLs
// pointing at the server. Token and TenantID are pre-set for convenience.
// URLs are set directly (without With*URL options) so that version path
// normalization is not applied — test handlers see raw resource paths.
func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient()
	client.IdentityURL = server.URL
	client.ComputeURL = server.URL
	client.BlockStorageURL = server.URL
	client.ImageServiceURL = server.URL
	client.NetworkingURL = server.URL
	client.LBaaSURL = server.URL
	client.ObjectStorageURL = server.URL
	client.DNSServiceURL = server.URL
	client.Token = "test-token"
	client.TenantID = "test-tenant-id"
	return server, client
}

// assertNoError fails the test if err is not nil.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// assertError fails the test if err is nil.
func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// assertAPIError checks that err is an *APIError with the expected status code.
func assertAPIError(t *testing.T, err error, expectedStatusCode int) {
	t.Helper()
	if err == nil {
		t.Fatal("expected APIError, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != expectedStatusCode {
		t.Fatalf("expected status code %d, got %d", expectedStatusCode, apiErr.StatusCode)
	}
}

// readJSONBody reads the request body and unmarshals it into v.
func readJSONBody(t *testing.T, r *http.Request, v interface{}) {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading request body: %v", err)
	}
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("unmarshaling request body: %v (body: %s)", err, string(body))
	}
}
