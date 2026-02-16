package conoha

import (
	"context"
	"fmt"
	"net/http"
)

// ------------------------------------------------------------
// DNS Types
// ------------------------------------------------------------

// Domain represents a DNS domain.
type Domain struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
	Serial    int64  `json:"serial"`
	TTL       int    `json:"ttl"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// DNSRecord represents a DNS record.
type DNSRecord struct {
	UUID       string  `json:"uuid"`
	DomainUUID string  `json:"domain_uuid"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Data       string  `json:"data"`
	Priority   *int    `json:"priority"`
	Weight     *int    `json:"weight"`
	Port       *int    `json:"port"`
	TTL        int     `json:"ttl"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// CreateDomainRequest is the request to create a domain.
type CreateDomainRequest struct {
	Name  string `json:"name"`
	TTL   int    `json:"ttl"`
	Email string `json:"email"`
}

// UpdateDomainRequest is the request to update a domain.
type UpdateDomainRequest struct {
	TTL   int    `json:"ttl"`
	Email string `json:"email"`
}

// CreateDNSRecordRequest is the request to create a DNS record.
type CreateDNSRecordRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Data     string `json:"data"`
	Priority *int   `json:"priority"`
	Weight   *int   `json:"weight"`
	Port     *int   `json:"port"`
}

// UpdateDNSRecordRequest is the request to update a DNS record.
type UpdateDNSRecordRequest struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Data     string `json:"data,omitempty"`
	Priority *int   `json:"priority"`
	Weight   *int   `json:"weight"`
	Port     *int   `json:"port"`
}

type domainListResponse struct {
	Domains    []Domain `json:"domains"`
	TotalCount int      `json:"total_count"`
}

type recordListResponse struct {
	Records    []DNSRecord `json:"records"`
	TotalCount int         `json:"total_count"`
}

// ListDomainsOptions are options for listing domains.
type ListDomainsOptions struct {
	Limit    int
	Offset   int
	SortType string // "asc" or "desc"
	SortKey  string
}

// ListDNSRecordsOptions are options for listing DNS records.
type ListDNSRecordsOptions struct {
	Limit    int
	Offset   int
	SortType string
	SortKey  string
}

// ------------------------------------------------------------
// Domain Operations
// ------------------------------------------------------------

// ListDomains lists all DNS domains.
func (c *Client) ListDomains(ctx context.Context, opts *ListDomainsOptions) ([]Domain, error) {
	url := c.DNSServiceURL + "/v1/domains"
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", opts.Offset)
		}
		if opts.SortType != "" {
			params["sort_type"] = opts.SortType
		}
		if opts.SortKey != "" {
			params["sort_key"] = opts.SortKey
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result domainListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Domains, nil
}

// GetDomain gets a domain's details.
func (c *Client) GetDomain(ctx context.Context, domainID string) (*Domain, error) {
	url := fmt.Sprintf("%s/v1/domains/%s", c.DNSServiceURL, domainID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result Domain
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateDomain creates a new DNS domain.
// Domain name must end with a trailing period (e.g., "example.com.").
func (c *Client) CreateDomain(ctx context.Context, opts CreateDomainRequest) (*Domain, error) {
	url := c.DNSServiceURL + "/v1/domains"
	req, err := c.newRequest(ctx, http.MethodPost, url, opts)
	if err != nil {
		return nil, err
	}
	var result Domain
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateDomain updates a domain's TTL and email.
func (c *Client) UpdateDomain(ctx context.Context, domainID string, opts UpdateDomainRequest) (*Domain, error) {
	url := fmt.Sprintf("%s/v1/domains/%s", c.DNSServiceURL, domainID)
	req, err := c.newRequest(ctx, http.MethodPut, url, opts)
	if err != nil {
		return nil, err
	}
	var result Domain
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDomain deletes a DNS domain.
func (c *Client) DeleteDomain(ctx context.Context, domainID string) error {
	url := fmt.Sprintf("%s/v1/domains/%s", c.DNSServiceURL, domainID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// DNS Record Operations
// ------------------------------------------------------------

// ListDNSRecords lists all DNS records for a domain.
func (c *Client) ListDNSRecords(ctx context.Context, domainID string, opts *ListDNSRecordsOptions) ([]DNSRecord, error) {
	url := fmt.Sprintf("%s/v1/domains/%s/records", c.DNSServiceURL, domainID)
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", opts.Offset)
		}
		if opts.SortType != "" {
			params["sort_type"] = opts.SortType
		}
		if opts.SortKey != "" {
			params["sort_key"] = opts.SortKey
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result recordListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Records, nil
}

// GetDNSRecord gets a DNS record's details.
func (c *Client) GetDNSRecord(ctx context.Context, domainID, recordID string) (*DNSRecord, error) {
	url := fmt.Sprintf("%s/v1/domains/%s/records/%s", c.DNSServiceURL, domainID, recordID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result DNSRecord
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateDNSRecord creates a new DNS record.
// Record name must end with a trailing period (e.g., "www.example.com.").
func (c *Client) CreateDNSRecord(ctx context.Context, domainID string, opts CreateDNSRecordRequest) (*DNSRecord, error) {
	url := fmt.Sprintf("%s/v1/domains/%s/records", c.DNSServiceURL, domainID)
	req, err := c.newRequest(ctx, http.MethodPost, url, opts)
	if err != nil {
		return nil, err
	}
	var result DNSRecord
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateDNSRecord updates a DNS record.
func (c *Client) UpdateDNSRecord(ctx context.Context, domainID, recordID string, opts UpdateDNSRecordRequest) (*DNSRecord, error) {
	url := fmt.Sprintf("%s/v1/domains/%s/records/%s", c.DNSServiceURL, domainID, recordID)
	req, err := c.newRequest(ctx, http.MethodPut, url, opts)
	if err != nil {
		return nil, err
	}
	var result DNSRecord
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDNSRecord deletes a DNS record.
func (c *Client) DeleteDNSRecord(ctx context.Context, domainID, recordID string) error {
	url := fmt.Sprintf("%s/v1/domains/%s/records/%s", c.DNSServiceURL, domainID, recordID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}
