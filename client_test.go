package conoha

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// ============================================================
// NewClient & Options
// ============================================================

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient()

	if c.Region != DefaultRegion {
		t.Errorf("Region = %q, want %q", c.Region, DefaultRegion)
	}
	if c.HTTPClient == nil {
		t.Fatal("HTTPClient is nil")
	}
	if c.IdentityURL != "https://identity.c3j1.conoha.io/v3" {
		t.Errorf("IdentityURL = %q", c.IdentityURL)
	}
	if c.ComputeURL != "https://compute.c3j1.conoha.io/v2.1" {
		t.Errorf("ComputeURL = %q", c.ComputeURL)
	}
	if c.BlockStorageURL != "https://block-storage.c3j1.conoha.io/v3" {
		t.Errorf("BlockStorageURL = %q", c.BlockStorageURL)
	}
	if c.ImageServiceURL != "https://image-service.c3j1.conoha.io/v2" {
		t.Errorf("ImageServiceURL = %q", c.ImageServiceURL)
	}
	if c.NetworkingURL != "https://networking.c3j1.conoha.io/v2.0" {
		t.Errorf("NetworkingURL = %q", c.NetworkingURL)
	}
	if c.LBaaSURL != "https://lbaas.c3j1.conoha.io/v2.0" {
		t.Errorf("LBaaSURL = %q", c.LBaaSURL)
	}
	if c.ObjectStorageURL != "https://object-storage.c3j1.conoha.io/v1" {
		t.Errorf("ObjectStorageURL = %q", c.ObjectStorageURL)
	}
	if c.DNSServiceURL != "https://dns-service.c3j1.conoha.io/v1" {
		t.Errorf("DNSServiceURL = %q", c.DNSServiceURL)
	}
}

func TestNewClient_WithRegion(t *testing.T) {
	c := NewClient(WithRegion("c3j2"))

	if c.Region != "c3j2" {
		t.Errorf("Region = %q, want %q", c.Region, "c3j2")
	}
	if !strings.Contains(c.ComputeURL, "c3j2") {
		t.Errorf("ComputeURL should contain c3j2: %q", c.ComputeURL)
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c := NewClient(WithHTTPClient(custom))

	if c.HTTPClient != custom {
		t.Error("HTTPClient was not set")
	}
}

func TestNewClient_WithExplicitURL(t *testing.T) {
	c := NewClient(WithComputeURL("https://custom-compute.example.com"))

	if c.ComputeURL != "https://custom-compute.example.com" {
		t.Errorf("ComputeURL = %q", c.ComputeURL)
	}
	if !c.explicitURLs["compute"] {
		t.Error("compute should be marked as explicit")
	}
	// Other URLs should still be from region pattern
	if c.IdentityURL != "https://identity.c3j1.conoha.io/v3" {
		t.Errorf("IdentityURL should use default: %q", c.IdentityURL)
	}
}

func TestNewClient_WithEndpoints(t *testing.T) {
	c := NewClient(WithEndpoints(Endpoints{
		Identity: "https://id.example.com",
		Compute:  "https://compute.example.com",
	}))

	if c.IdentityURL != "https://id.example.com" {
		t.Errorf("IdentityURL = %q", c.IdentityURL)
	}
	if c.ComputeURL != "https://compute.example.com" {
		t.Errorf("ComputeURL = %q", c.ComputeURL)
	}
	if !c.explicitURLs["identity"] || !c.explicitURLs["compute"] {
		t.Error("endpoints should be marked as explicit")
	}
	// Unset endpoints should still have defaults
	if c.NetworkingURL != "https://networking.c3j1.conoha.io/v2.0" {
		t.Errorf("NetworkingURL should use default: %q", c.NetworkingURL)
	}
}

func TestWithEndpoints_EmptyStringsIgnored(t *testing.T) {
	c := NewClient(WithEndpoints(Endpoints{
		Identity: "https://id.example.com",
		Compute:  "", // empty = not explicit
	}))

	if !c.explicitURLs["identity"] {
		t.Error("identity should be marked as explicit")
	}
	if c.explicitURLs["compute"] {
		t.Error("empty compute should NOT be marked as explicit")
	}
}

// ============================================================
// fillURLsFromRegion
// ============================================================

func TestFillURLsFromRegion_DoesNotOverwriteExplicit(t *testing.T) {
	c := NewClient(WithComputeURL("https://custom.example.com"))

	// fillURLsFromRegion is called in NewClient; verify explicit URL was preserved
	if c.ComputeURL != "https://custom.example.com" {
		t.Errorf("explicit ComputeURL was overwritten: %q", c.ComputeURL)
	}
}

// ============================================================
// updateEndpointsFromCatalog
// ============================================================

func TestUpdateEndpointsFromCatalog_Success(t *testing.T) {
	c := NewClient()

	catalog := []ServiceCatalog{
		{
			Type: ServiceTypeCompute,
			Endpoints: []Endpoint{
				{Interface: "public", Region: "c3j1", URL: "https://catalog-compute.example.com/v2.1"},
			},
		},
		{
			Type: ServiceTypeBlockStorage,
			Endpoints: []Endpoint{
				{Interface: "public", Region: "c3j1", URL: "https://catalog-bs.example.com/v3/"},
			},
		},
	}

	c.updateEndpointsFromCatalog(catalog)

	if c.ComputeURL != "https://catalog-compute.example.com/v2.1" {
		t.Errorf("ComputeURL = %q", c.ComputeURL)
	}
	// Trailing slash should be stripped
	if c.BlockStorageURL != "https://catalog-bs.example.com/v3" {
		t.Errorf("BlockStorageURL = %q (trailing slash not stripped?)", c.BlockStorageURL)
	}
}

func TestUpdateEndpointsFromCatalog_RespectsExplicit(t *testing.T) {
	c := NewClient(WithComputeURL("https://explicit.example.com"))

	catalog := []ServiceCatalog{
		{
			Type: ServiceTypeCompute,
			Endpoints: []Endpoint{
				{Interface: "public", Region: "c3j1", URL: "https://catalog.example.com/v2.1"},
			},
		},
	}

	c.updateEndpointsFromCatalog(catalog)

	if c.ComputeURL != "https://explicit.example.com" {
		t.Errorf("explicit ComputeURL was overridden: %q", c.ComputeURL)
	}
}

func TestUpdateEndpointsFromCatalog_FiltersRegion(t *testing.T) {
	c := NewClient(WithRegion("c3j1"))

	catalog := []ServiceCatalog{
		{
			Type: ServiceTypeCompute,
			Endpoints: []Endpoint{
				{Interface: "public", Region: "c3j2", URL: "https://wrong-region.example.com"},
			},
		},
	}

	originalURL := c.ComputeURL
	c.updateEndpointsFromCatalog(catalog)

	if c.ComputeURL != originalURL {
		t.Errorf("ComputeURL should not be updated for wrong region: %q", c.ComputeURL)
	}
}

func TestUpdateEndpointsFromCatalog_IgnoresNonPublic(t *testing.T) {
	c := NewClient()

	catalog := []ServiceCatalog{
		{
			Type: ServiceTypeCompute,
			Endpoints: []Endpoint{
				{Interface: "internal", Region: "c3j1", URL: "https://internal.example.com"},
			},
		},
	}

	originalURL := c.ComputeURL
	c.updateEndpointsFromCatalog(catalog)

	if c.ComputeURL != originalURL {
		t.Errorf("ComputeURL should not be updated for non-public interface: %q", c.ComputeURL)
	}
}

func TestUpdateEndpointsFromCatalog_EmptyCatalog(t *testing.T) {
	c := NewClient()
	originalURL := c.ComputeURL

	c.updateEndpointsFromCatalog(nil)

	if c.ComputeURL != originalURL {
		t.Errorf("ComputeURL changed with nil catalog: %q", c.ComputeURL)
	}
}

func TestUpdateEndpointsFromCatalog_RegionID(t *testing.T) {
	c := NewClient(WithRegion("c3j1"))

	catalog := []ServiceCatalog{
		{
			Type: ServiceTypeCompute,
			Endpoints: []Endpoint{
				{Interface: "public", RegionID: "c3j1", URL: "https://by-region-id.example.com/v2.1"},
			},
		},
	}

	c.updateEndpointsFromCatalog(catalog)

	if c.ComputeURL != "https://by-region-id.example.com/v2.1" {
		t.Errorf("should match by RegionID: %q", c.ComputeURL)
	}
}

// ============================================================
// newRequest
// ============================================================

func TestNewRequest_GET(t *testing.T) {
	c := NewClient()
	req, err := c.newRequest(context.Background(), http.MethodGet, "https://example.com/test", nil)
	assertNoError(t, err)

	if req.Method != http.MethodGet {
		t.Errorf("Method = %q", req.Method)
	}
	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("Accept header = %q", req.Header.Get("Accept"))
	}
	if req.Header.Get("Content-Type") != "" {
		t.Errorf("Content-Type should be empty for GET: %q", req.Header.Get("Content-Type"))
	}
	if req.Body != nil {
		t.Error("Body should be nil for GET")
	}
}

func TestNewRequest_POST(t *testing.T) {
	c := NewClient()
	body := map[string]string{"key": "value"}
	req, err := c.newRequest(context.Background(), http.MethodPost, "https://example.com/test", body)
	assertNoError(t, err)

	if req.Method != http.MethodPost {
		t.Errorf("Method = %q", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q", req.Header.Get("Content-Type"))
	}
	if req.Body == nil {
		t.Fatal("Body should not be nil for POST")
	}
}

func TestNewRequest_WithToken(t *testing.T) {
	c := NewClient()
	c.Token = "my-auth-token"

	req, err := c.newRequest(context.Background(), http.MethodGet, "https://example.com/test", nil)
	assertNoError(t, err)

	if req.Header.Get("X-Auth-Token") != "my-auth-token" {
		t.Errorf("X-Auth-Token = %q", req.Header.Get("X-Auth-Token"))
	}
}

func TestNewRequest_NoToken(t *testing.T) {
	c := NewClient()

	req, err := c.newRequest(context.Background(), http.MethodGet, "https://example.com/test", nil)
	assertNoError(t, err)

	if req.Header.Get("X-Auth-Token") != "" {
		t.Errorf("X-Auth-Token should be empty: %q", req.Header.Get("X-Auth-Token"))
	}
}

// ============================================================
// do
// ============================================================

func TestDo_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"name":"test"}`))
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodGet, server.URL+"/test", nil)
	var result struct {
		Name string `json:"name"`
	}
	_, err := client.do(req, &result)
	assertNoError(t, err)

	if result.Name != "test" {
		t.Errorf("Name = %q, want %q", result.Name, "test")
	}
}

func TestDo_NoContent(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodDelete, server.URL+"/test", nil)
	_, err := client.do(req, nil)
	assertNoError(t, err)
}

func TestDo_4xxError(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"itemNotFound":{"message":"Server not found","code":404}}`))
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodGet, server.URL+"/test", nil)
	_, err := client.do(req, nil)

	assertAPIError(t, err, 404)
	apiErr := err.(*APIError)
	if apiErr.Message != "Server not found" {
		t.Errorf("Message = %q", apiErr.Message)
	}
	if apiErr.Code != 404 {
		t.Errorf("Code = %d", apiErr.Code)
	}
}

func TestDo_5xxError(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`Internal Server Error`))
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodGet, server.URL+"/test", nil)
	_, err := client.do(req, nil)

	assertAPIError(t, err, 500)
}

func TestDo_InvalidJSON(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`not json`))
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodGet, server.URL+"/test", nil)
	var result struct{ Name string }
	_, err := client.do(req, &result)

	assertError(t, err)
	if !strings.Contains(err.Error(), "unmarshal response") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

func TestDo_EmptyBodyWithResult(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		// empty body
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodGet, server.URL+"/test", nil)
	var result struct{ Name string }
	_, err := client.do(req, &result)
	assertNoError(t, err) // should not error when body is empty even with result
}

func TestDo_ValidatesAuthHeader(t *testing.T) {
	var capturedToken string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedToken = r.Header.Get("X-Auth-Token")
		w.WriteHeader(200)
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodGet, server.URL+"/test", nil)
	client.do(req, nil)

	if capturedToken != "test-token" {
		t.Errorf("X-Auth-Token = %q, want %q", capturedToken, "test-token")
	}
}

// ============================================================
// doRaw
// ============================================================

func TestDoRaw_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("raw response data"))
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodGet, server.URL+"/test", nil)
	_, body, err := client.doRaw(req)
	assertNoError(t, err)

	if string(body) != "raw response data" {
		t.Errorf("body = %q", string(body))
	}
}

func TestDoRaw_Error(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("bad request"))
	})
	defer server.Close()

	req, _ := client.newRequest(context.Background(), http.MethodGet, server.URL+"/test", nil)
	_, _, err := client.doRaw(req)

	assertAPIError(t, err, 400)
}

// ============================================================
// buildQueryString
// ============================================================

func TestBuildQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name:   "empty",
			params: map[string]string{},
			want:   "",
		},
		{
			name:   "single",
			params: map[string]string{"key": "value"},
			want:   "?key=value",
		},
		{
			name:   "skip empty values",
			params: map[string]string{"key": "", "other": "val"},
			want:   "?other=val",
		},
		{
			name:   "all empty values",
			params: map[string]string{"a": "", "b": ""},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildQueryString(tt.params)
			if tt.want == "" {
				if got != "" {
					t.Errorf("got %q, want empty", got)
				}
				return
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildQueryString_MultipleParams(t *testing.T) {
	params := map[string]string{"a": "1", "b": "2"}
	got := buildQueryString(params)

	if !strings.HasPrefix(got, "?") {
		t.Fatalf("should start with ?: %q", got)
	}
	// Order may vary, just check both are present
	if !strings.Contains(got, "a=1") || !strings.Contains(got, "b=2") {
		t.Errorf("got %q, expected both a=1 and b=2", got)
	}
}

// ============================================================
// newAPIError
// ============================================================

func TestNewAPIError_OpenStackFormat(t *testing.T) {
	body := `{"badRequest":{"message":"Invalid input","code":400}}`
	err := newAPIError(400, "400 Bad Request", body)

	if err.StatusCode != 400 {
		t.Errorf("StatusCode = %d", err.StatusCode)
	}
	if err.Message != "Invalid input" {
		t.Errorf("Message = %q", err.Message)
	}
	if err.Code != 400 {
		t.Errorf("Code = %d", err.Code)
	}
	if err.Body != body {
		t.Errorf("Body = %q", err.Body)
	}
}

func TestNewAPIError_PlainText(t *testing.T) {
	err := newAPIError(500, "500 Internal Server Error", "something went wrong")

	if err.StatusCode != 500 {
		t.Errorf("StatusCode = %d", err.StatusCode)
	}
	if err.Message != "" {
		t.Errorf("Message should be empty: %q", err.Message)
	}
	if err.Body != "something went wrong" {
		t.Errorf("Body = %q", err.Body)
	}
}

func TestNewAPIError_EmptyBody(t *testing.T) {
	err := newAPIError(401, "401 Unauthorized", "")

	if err.StatusCode != 401 {
		t.Errorf("StatusCode = %d", err.StatusCode)
	}
	if err.Message != "" {
		t.Errorf("Message should be empty: %q", err.Message)
	}
}

// ============================================================
// APIError.Error()
// ============================================================

func TestAPIError_ErrorWithMessage(t *testing.T) {
	e := &APIError{
		StatusCode: 400,
		Status:     "400 Bad Request",
		Message:    "Invalid input",
	}
	got := e.Error()
	if !strings.Contains(got, "Invalid input") {
		t.Errorf("Error() = %q, should contain message", got)
	}
}

func TestAPIError_ErrorWithoutMessage(t *testing.T) {
	e := &APIError{
		StatusCode: 500,
		Status:     "500 Internal Server Error",
		Body:       "raw body",
	}
	got := e.Error()
	if !strings.Contains(got, "raw body") {
		t.Errorf("Error() = %q, should contain body", got)
	}
}

// ============================================================
// tenantID()
// ============================================================

func TestTenantID(t *testing.T) {
	c := NewClient()
	c.TenantID = "my-tenant"

	got := c.tenantID()
	if got != "my-tenant" {
		t.Errorf("tenantID() = %q", got)
	}
}

// ============================================================
// Concurrency
// ============================================================

func TestClient_ConcurrentTokenAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Subject-Token", "new-token")
		w.WriteHeader(201)
		w.Write([]byte(`{"token":{"catalog":[],"project":{"id":"tenant-123"}}}`))
	}))
	defer server.Close()

	c := NewClient(WithIdentityURL(server.URL))

	var wg sync.WaitGroup

	// Writer: Authenticate sets Token
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			c.Authenticate(context.Background(), "user", "pass", "tenant")
		}
	}()

	// Reader: read Token concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			_ = c.tenantID()
			c.mu.RLock()
			_ = c.Token
			c.mu.RUnlock()
		}
	}()

	wg.Wait()
}

// ============================================================
// WithEndpoints - all service types
// ============================================================

func TestWithAllExplicitURLOptions(t *testing.T) {
	tests := []struct {
		name     string
		opt      ClientOption
		field    func(*Client) string
		key      string
		expected string
	}{
		{"Identity", WithIdentityURL("https://id"), func(c *Client) string { return c.IdentityURL }, "identity", "https://id"},
		{"Compute", WithComputeURL("https://comp"), func(c *Client) string { return c.ComputeURL }, "compute", "https://comp"},
		{"BlockStorage", WithBlockStorageURL("https://bs"), func(c *Client) string { return c.BlockStorageURL }, "block-storage", "https://bs"},
		{"ImageService", WithImageServiceURL("https://img"), func(c *Client) string { return c.ImageServiceURL }, "image", "https://img"},
		{"Networking", WithNetworkingURL("https://net"), func(c *Client) string { return c.NetworkingURL }, "network", "https://net"},
		{"LBaaS", WithLBaaSURL("https://lb"), func(c *Client) string { return c.LBaaSURL }, "load-balancer", "https://lb"},
		{"ObjectStorage", WithObjectStorageURL("https://obj"), func(c *Client) string { return c.ObjectStorageURL }, "object-store", "https://obj"},
		{"DNS", WithDNSServiceURL("https://dns"), func(c *Client) string { return c.DNSServiceURL }, "dns", "https://dns"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.opt)
			if tt.field(c) != tt.expected {
				t.Errorf("URL = %q, want %q", tt.field(c), tt.expected)
			}
			if !c.explicitURLs[tt.key] {
				t.Errorf("%s should be marked as explicit", tt.key)
			}
		})
	}
}

// ============================================================
// Integration: request body marshaling
// ============================================================

func TestNewRequest_MarshalBody(t *testing.T) {
	c := NewClient()
	body := map[string]interface{}{
		"server": map[string]string{"name": "test-server"},
	}

	req, err := c.newRequest(context.Background(), http.MethodPost, "https://example.com/test", body)
	assertNoError(t, err)

	var parsed map[string]json.RawMessage
	readJSONBody(t, req, &parsed)

	if _, ok := parsed["server"]; !ok {
		t.Error("body should contain 'server' key")
	}
}
