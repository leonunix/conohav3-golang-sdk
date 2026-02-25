package conoha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	// DefaultRegion is the default ConoHa region.
	DefaultRegion = "c3j1"
)

// Service type constants used in the Service Catalog.
const (
	ServiceTypeIdentity     = "identity"
	ServiceTypeCompute      = "compute"
	ServiceTypeBlockStorage = "block-storage"
	ServiceTypeImage        = "image"
	ServiceTypeNetwork      = "network"
	ServiceTypeLBaaS        = "load-balancer"
	ServiceTypeObjectStore  = "object-store"
	ServiceTypeDNS          = "dns"
	ServiceTypeAccount      = "account"
)

// Client is the ConoHa VPS v3 API client.
//
// The Client is safe for concurrent use across goroutines. Internally it uses
// a sync.RWMutex to protect Token, TenantID, and endpoint URL fields from
// data races when Authenticate() is called concurrently with other API methods.
type Client struct {
	HTTPClient *http.Client
	Token      string
	TenantID   string
	Region     string

	IdentityURL      string
	ComputeURL       string
	BlockStorageURL  string
	ImageServiceURL  string
	NetworkingURL    string
	LBaaSURL         string
	ObjectStorageURL string
	DNSServiceURL    string

	// mu protects Token, TenantID, and endpoint URL fields from concurrent
	// read/write access (e.g. Authenticate writing while API methods read).
	mu sync.RWMutex

	// explicitURLs tracks which URLs were explicitly set by the user.
	// These will NOT be overridden by Service Catalog auto-discovery.
	explicitURLs map[string]bool

	// explicitRegion is true when the user explicitly called WithRegion().
	explicitRegion bool
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithRegion sets the ConoHa region and generates all endpoint URLs
// from the pattern https://{service}.{region}.conoha.io.
// URLs set via other options (e.g. WithComputeURL) take priority.
func WithRegion(region string) ClientOption {
	return func(c *Client) {
		c.Region = region
		c.explicitRegion = true
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithIdentityURL sets only the Identity API endpoint.
// Other endpoints will be auto-discovered from the Service Catalog after authentication.
func WithIdentityURL(url string) ClientOption {
	return func(c *Client) {
		c.IdentityURL = url
		c.explicitURLs["identity"] = true
	}
}

// WithComputeURL sets the Compute API endpoint.
func WithComputeURL(url string) ClientOption {
	return func(c *Client) {
		c.ComputeURL = url
		c.explicitURLs["compute"] = true
	}
}

// WithBlockStorageURL sets the Block Storage API endpoint.
func WithBlockStorageURL(url string) ClientOption {
	return func(c *Client) {
		c.BlockStorageURL = url
		c.explicitURLs["block-storage"] = true
	}
}

// WithImageServiceURL sets the Image Service API endpoint.
func WithImageServiceURL(url string) ClientOption {
	return func(c *Client) {
		c.ImageServiceURL = url
		c.explicitURLs["image"] = true
	}
}

// WithNetworkingURL sets the Networking API endpoint.
func WithNetworkingURL(url string) ClientOption {
	return func(c *Client) {
		c.NetworkingURL = url
		c.explicitURLs["network"] = true
	}
}

// WithLBaaSURL sets the Load Balancer API endpoint.
func WithLBaaSURL(url string) ClientOption {
	return func(c *Client) {
		c.LBaaSURL = url
		c.explicitURLs["load-balancer"] = true
	}
}

// WithObjectStorageURL sets the Object Storage API endpoint.
func WithObjectStorageURL(url string) ClientOption {
	return func(c *Client) {
		c.ObjectStorageURL = url
		c.explicitURLs["object-store"] = true
	}
}

// WithDNSServiceURL sets the DNS Service API endpoint.
func WithDNSServiceURL(url string) ClientOption {
	return func(c *Client) {
		c.DNSServiceURL = url
		c.explicitURLs["dns"] = true
	}
}

// Endpoints holds all service endpoint URLs.
// Use with WithEndpoints to set multiple URLs at once.
type Endpoints struct {
	Identity     string
	Compute      string
	BlockStorage string
	ImageService string
	Networking   string
	LBaaS        string
	ObjectStore  string
	DNS          string
}

// WithEndpoints sets all endpoint URLs at once.
// Empty strings are ignored (not treated as explicit).
func WithEndpoints(ep Endpoints) ClientOption {
	return func(c *Client) {
		if ep.Identity != "" {
			c.IdentityURL = ep.Identity
			c.explicitURLs["identity"] = true
		}
		if ep.Compute != "" {
			c.ComputeURL = ep.Compute
			c.explicitURLs["compute"] = true
		}
		if ep.BlockStorage != "" {
			c.BlockStorageURL = ep.BlockStorage
			c.explicitURLs["block-storage"] = true
		}
		if ep.ImageService != "" {
			c.ImageServiceURL = ep.ImageService
			c.explicitURLs["image"] = true
		}
		if ep.Networking != "" {
			c.NetworkingURL = ep.Networking
			c.explicitURLs["network"] = true
		}
		if ep.LBaaS != "" {
			c.LBaaSURL = ep.LBaaS
			c.explicitURLs["load-balancer"] = true
		}
		if ep.ObjectStore != "" {
			c.ObjectStorageURL = ep.ObjectStore
			c.explicitURLs["object-store"] = true
		}
		if ep.DNS != "" {
			c.DNSServiceURL = ep.DNS
			c.explicitURLs["dns"] = true
		}
	}
}

// NewClient creates a new ConoHa API client.
//
// Endpoint resolution order (highest priority first):
//  1. Explicitly set via With*URL() or WithEndpoints() — never overridden
//  2. Auto-discovered from Service Catalog after Authenticate()
//  3. Generated from Region pattern https://{service}.{region}.conoha.io
//
// Examples:
//
//	// Use default region c3j1, all URLs from pattern
//	client := conoha.NewClient()
//
//	// Use a different region
//	client := conoha.NewClient(conoha.WithRegion("c3j2"))
//
//	// Only set Identity URL, auto-discover rest after auth
//	client := conoha.NewClient(conoha.WithIdentityURL("https://identity.c3j2.conoha.io"))
//
//	// Set all URLs manually
//	client := conoha.NewClient(conoha.WithEndpoints(conoha.Endpoints{
//	    Identity: "https://identity.c3j1.conoha.io",
//	    Compute:  "https://compute.c3j1.conoha.io",
//	    // ...
//	}))
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		HTTPClient:   &http.Client{},
		Region:       DefaultRegion,
		explicitURLs: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(c)
	}

	// If the user set an explicit Identity URL but did not set the region,
	// try to infer the region from the URL (e.g. "https://identity.c3j2.conoha.io"
	// → region "c3j2"). This ensures that all other auto-generated URLs and
	// Service Catalog filtering use the correct region.
	if !c.explicitRegion && c.explicitURLs["identity"] {
		if r := extractRegionFromConoHaURL(c.IdentityURL); r != "" {
			c.Region = r
		}
	}

	// Ensure version paths are present on explicitly set URLs so that
	// SDK methods can append resource paths directly (e.g. "/auth/tokens").
	c.normalizeExplicitURLs()

	c.fillURLsFromRegion()
	return c
}

// fillURLsFromRegion fills any unset URLs using the region pattern.
// Only fills URLs that were NOT explicitly set by the user.
// Generated URLs include the API version path so that service methods
// can append resource paths directly (e.g. c.ComputeURL + "/servers").
func (c *Client) fillURLsFromRegion() {
	r := c.Region
	if c.IdentityURL == "" {
		c.IdentityURL = fmt.Sprintf("https://identity.%s.conoha.io/v3", r)
	}
	if c.ComputeURL == "" {
		c.ComputeURL = fmt.Sprintf("https://compute.%s.conoha.io/v2.1", r)
	}
	if c.BlockStorageURL == "" {
		c.BlockStorageURL = fmt.Sprintf("https://block-storage.%s.conoha.io/v3", r)
	}
	if c.ImageServiceURL == "" {
		c.ImageServiceURL = fmt.Sprintf("https://image-service.%s.conoha.io/v2", r)
	}
	if c.NetworkingURL == "" {
		c.NetworkingURL = fmt.Sprintf("https://networking.%s.conoha.io/v2.0", r)
	}
	if c.LBaaSURL == "" {
		c.LBaaSURL = fmt.Sprintf("https://lbaas.%s.conoha.io/v2.0", r)
	}
	if c.ObjectStorageURL == "" {
		c.ObjectStorageURL = fmt.Sprintf("https://object-storage.%s.conoha.io/v1", r)
	}
	if c.DNSServiceURL == "" {
		c.DNSServiceURL = fmt.Sprintf("https://dns-service.%s.conoha.io/v1", r)
	}
}

// updateEndpointsFromCatalog updates endpoint URLs from the Service Catalog
// returned by authentication. Only updates URLs that were NOT explicitly set.
// When the client has a Region set, only endpoints matching that region are used.
//
// The ConoHa service catalog may return base URLs without version paths
// (e.g. "https://networking.c3j1.conoha.io" instead of ".../v2.0").
// This method normalizes catalog URLs by ensuring the required API version
// path is present, since SDK methods append only resource paths (e.g. "/servers").
func (c *Client) updateEndpointsFromCatalog(catalog []ServiceCatalog) {
	for _, svc := range catalog {
		// Find the public endpoint URL, preferring one that matches the client's region.
		var publicURL string
		for _, ep := range svc.Endpoints {
			if ep.Interface != "public" {
				continue
			}
			if c.Region != "" && ep.Region != c.Region && ep.RegionID != c.Region {
				continue
			}
			publicURL = strings.TrimRight(ep.URL, "/")
			break
		}
		if publicURL == "" {
			continue
		}

		switch svc.Type {
		case ServiceTypeIdentity:
			if !c.explicitURLs["identity"] {
				c.IdentityURL = ensureVersionPath(publicURL, "/v3")
			}
		case ServiceTypeCompute:
			if !c.explicitURLs["compute"] {
				c.ComputeURL = ensureVersionPath(publicURL, "/v2.1")
			}
		case ServiceTypeBlockStorage, "volumev3":
			if !c.explicitURLs["block-storage"] {
				// Catalog may include tenant ID in path (e.g. /v3/{tenantID}).
				// Strip everything after the version path since SDK methods add tenant ID.
				u := ensureVersionPath(publicURL, "/v3")
				if idx := strings.Index(u, "/v3/"); idx >= 0 {
					u = u[:idx+3]
				}
				c.BlockStorageURL = u
			}
		case ServiceTypeImage:
			if !c.explicitURLs["image"] {
				c.ImageServiceURL = ensureVersionPath(publicURL, "/v2")
			}
		case ServiceTypeNetwork:
			if !c.explicitURLs["network"] {
				c.NetworkingURL = ensureVersionPath(publicURL, "/v2.0")
			}
		case ServiceTypeLBaaS:
			if !c.explicitURLs["load-balancer"] {
				c.LBaaSURL = ensureVersionPath(publicURL, "/v2.0")
			}
		case ServiceTypeObjectStore:
			if !c.explicitURLs["object-store"] {
				// Catalog may include /v1/AUTH_{tenantID}. Strip the AUTH_ portion
				// since objectStoragePath() appends it.
				u := publicURL
				if idx := strings.Index(u, "/AUTH_"); idx >= 0 {
					u = u[:idx]
				}
				c.ObjectStorageURL = ensureVersionPath(u, "/v1")
			}
		case ServiceTypeDNS:
			if !c.explicitURLs["dns"] {
				c.DNSServiceURL = ensureVersionPath(publicURL, "/v1")
			}
		}
	}
}

// extractRegionFromConoHaURL extracts the region from a ConoHa-style URL.
// For example, "https://identity.c3j2.conoha.io/v3" returns "c3j2".
// Returns "" if the URL does not match the expected pattern.
func extractRegionFromConoHaURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	// Expected host format: {service}.{region}.conoha.io
	parts := strings.Split(u.Hostname(), ".")
	if len(parts) >= 4 && parts[len(parts)-2] == "conoha" && parts[len(parts)-1] == "io" {
		return parts[1]
	}
	return ""
}

// normalizeExplicitURLs ensures that explicitly set URLs contain the required
// API version path. Users commonly set URLs like "https://identity.c3j2.conoha.io"
// without the "/v3" suffix; this method appends it so SDK methods work correctly.
func (c *Client) normalizeExplicitURLs() {
	if c.explicitURLs["identity"] {
		c.IdentityURL = ensureVersionPath(strings.TrimRight(c.IdentityURL, "/"), "/v3")
	}
	if c.explicitURLs["compute"] {
		c.ComputeURL = ensureVersionPath(strings.TrimRight(c.ComputeURL, "/"), "/v2.1")
	}
	if c.explicitURLs["block-storage"] {
		c.BlockStorageURL = ensureVersionPath(strings.TrimRight(c.BlockStorageURL, "/"), "/v3")
	}
	if c.explicitURLs["image"] {
		c.ImageServiceURL = ensureVersionPath(strings.TrimRight(c.ImageServiceURL, "/"), "/v2")
	}
	if c.explicitURLs["network"] {
		c.NetworkingURL = ensureVersionPath(strings.TrimRight(c.NetworkingURL, "/"), "/v2.0")
	}
	if c.explicitURLs["load-balancer"] {
		c.LBaaSURL = ensureVersionPath(strings.TrimRight(c.LBaaSURL, "/"), "/v2.0")
	}
	if c.explicitURLs["object-store"] {
		c.ObjectStorageURL = ensureVersionPath(strings.TrimRight(c.ObjectStorageURL, "/"), "/v1")
	}
	if c.explicitURLs["dns"] {
		c.DNSServiceURL = ensureVersionPath(strings.TrimRight(c.DNSServiceURL, "/"), "/v1")
	}
}

// ensureVersionPath appends the version path suffix to the URL if not already present.
func ensureVersionPath(rawURL, versionPath string) string {
	if strings.Contains(rawURL, versionPath) {
		return rawURL
	}
	return rawURL + versionPath
}

// APIError represents an error response from the ConoHa API.
//
// The Body field always contains the raw response body string.
// If the response body is a standard OpenStack JSON error (e.g.
// {"badRequest": {"message": "Invalid input", "code": 400}}),
// the Message and Code fields are populated with the parsed values.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
	Message    string // Parsed error message from JSON body, if available.
	Code       int    // Parsed error code from JSON body, if available.
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("conoha api error: %s: %s", e.Status, e.Message)
	}
	return fmt.Sprintf("conoha api error: %s (body: %s)", e.Status, e.Body)
}

// newAPIError creates an APIError and attempts to parse the body as a
// standard OpenStack JSON error to extract a structured message and code.
func newAPIError(statusCode int, status, body string) *APIError {
	e := &APIError{
		StatusCode: statusCode,
		Status:     status,
		Body:       body,
	}

	// Try to parse OpenStack-style error: {"errorType": {"message": "...", "code": N}}
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return e
	}
	for _, raw := range parsed {
		var inner struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}
		if err := json.Unmarshal(raw, &inner); err != nil {
			continue
		}
		if inner.Message != "" {
			e.Message = inner.Message
			e.Code = inner.Code
			break
		}
	}
	return e
}

// Link represents a resource link.
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

// tenantID returns the TenantID under a read lock.
func (c *Client) tenantID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TenantID
}

func (c *Client) newRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.mu.RLock()
	token := c.Token
	c.mu.RUnlock()
	if token != "" {
		req.Header.Set("X-Auth-Token", token)
	}
	return req, nil
}

func (c *Client) do(req *http.Request, result interface{}) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return resp, newAPIError(resp.StatusCode, resp.Status, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return resp, fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return resp, nil
}

func (c *Client) doRaw(req *http.Request) (*http.Response, []byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return resp, respBody, newAPIError(resp.StatusCode, resp.Status, string(respBody))
	}
	return resp, respBody, nil
}

func buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	vals := make(url.Values)
	for k, v := range params {
		if v != "" {
			vals.Set(k, v)
		}
	}
	if len(vals) == 0 {
		return ""
	}
	return "?" + vals.Encode()
}
