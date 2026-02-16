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

	// explicitURLs tracks which URLs were explicitly set by the user.
	// These will NOT be overridden by Service Catalog auto-discovery.
	explicitURLs map[string]bool
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithRegion sets the ConoHa region and generates all endpoint URLs
// from the pattern https://{service}.{region}.conoha.io.
// URLs set via other options (e.g. WithComputeURL) take priority.
func WithRegion(region string) ClientOption {
	return func(c *Client) {
		c.Region = region
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
// Catalog URLs are used as-is (including any version path such as /v2.1 or /v3).
// Service methods append only the resource path (e.g. "/servers"), not the version.
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
				c.IdentityURL = publicURL
			}
		case ServiceTypeCompute:
			if !c.explicitURLs["compute"] {
				c.ComputeURL = publicURL
			}
		case ServiceTypeBlockStorage:
			if !c.explicitURLs["block-storage"] {
				c.BlockStorageURL = publicURL
			}
		case ServiceTypeImage:
			if !c.explicitURLs["image"] {
				c.ImageServiceURL = publicURL
			}
		case ServiceTypeNetwork:
			if !c.explicitURLs["network"] {
				c.NetworkingURL = publicURL
			}
		case ServiceTypeLBaaS:
			if !c.explicitURLs["load-balancer"] {
				c.LBaaSURL = publicURL
			}
		case ServiceTypeObjectStore:
			if !c.explicitURLs["object-store"] {
				c.ObjectStorageURL = publicURL
			}
		case ServiceTypeDNS:
			if !c.explicitURLs["dns"] {
				c.DNSServiceURL = publicURL
			}
		}
	}
}

// APIError represents an error response from the ConoHa API.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("conoha api error: %s (body: %s)", e.Status, e.Body)
}

// Link represents a resource link.
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
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
	if c.Token != "" {
		req.Header.Set("X-Auth-Token", c.Token)
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
		return resp, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(respBody),
		}
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
		return resp, respBody, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(respBody),
		}
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
