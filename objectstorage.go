package conoha

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ------------------------------------------------------------
// Object Storage Types
// ------------------------------------------------------------

// AccountInfo represents object storage account information (from HEAD response headers).
type AccountInfo struct {
	ContainerCount  int64
	ObjectCount     int64
	BytesUsed       int64
	BytesUsedActual int64
	QuotaBytes      int64
}

// ContainerInfo represents container details from response headers.
type ContainerInfo struct {
	ObjectCount      int64
	BytesUsed        int64
	ReadACL          string
	WriteACL         string
	VersionsLocation string
	Metadata         map[string]string
}

// Container represents an object storage container.
type Container struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Bytes int64  `json:"bytes"`
}

// ObjectInfo represents object details from response headers.
type ObjectInfo struct {
	ContentLength int64
	ContentType   string
	ETag          string
	LastModified  string
	DeleteAt      int64
	Metadata      map[string]string
}

// Object represents an object in a container.
type Object struct {
	Name         string `json:"name"`
	Hash         string `json:"hash"`
	Bytes        int64  `json:"bytes"`
	ContentType  string `json:"content_type"`
	LastModified string `json:"last_modified"`
}

// SLOSegment represents a segment for Static Large Object upload.
type SLOSegment struct {
	Path      string `json:"path"`
	Etag      string `json:"etag"`
	SizeBytes int64  `json:"size_bytes"`
}

// ListObjectsOptions are options for listing objects.
type ListObjectsOptions struct {
	Reverse   bool
	Limit     int
	Marker    string
	EndMarker string
	Prefix    string
	Delimiter string
}

func (c *Client) objectStoragePath(parts ...string) string {
	path := fmt.Sprintf("%s/AUTH_%s", c.ObjectStorageURL, c.tenantID())
	for _, p := range parts {
		// Encode each segment but preserve "/" within object names.
		encoded := url.PathEscape(p)
		encoded = strings.ReplaceAll(encoded, "%2F", "/")
		path += "/" + encoded
	}
	return path
}

// ------------------------------------------------------------
// Account Operations
// ------------------------------------------------------------

// GetAccountInfo gets object storage account information.
func (c *Client) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
	url := c.objectStoragePath()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	info := &AccountInfo{}
	parseHeaderInt64(resp.Header, "X-Account-Container-Count", &info.ContainerCount)
	parseHeaderInt64(resp.Header, "X-Account-Object-Count", &info.ObjectCount)
	parseHeaderInt64(resp.Header, "X-Account-Bytes-Used", &info.BytesUsed)
	parseHeaderInt64(resp.Header, "X-Account-Bytes-Used-Actual", &info.BytesUsedActual)
	parseHeaderInt64(resp.Header, "X-Account-Meta-Quota-Bytes", &info.QuotaBytes)
	return info, nil
}

// SetAccountQuota sets the object storage account quota in GB.
// Must be in 100GB increments (e.g., 100, 200, 300...).
func (c *Client) SetAccountQuota(ctx context.Context, gigaBytes string) error {
	url := c.objectStoragePath()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Account-Meta-Quota-Giga-Bytes", gigaBytes)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// ------------------------------------------------------------
// Container Operations
// ------------------------------------------------------------

// ListContainers lists all containers.
func (c *Client) ListContainers(ctx context.Context) ([]Container, error) {
	url := c.objectStoragePath() + "?format=json"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result []Container
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateContainer creates a container.
func (c *Client) CreateContainer(ctx context.Context, name string) error {
	url := c.objectStoragePath(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// DeleteContainer deletes an empty container.
func (c *Client) DeleteContainer(ctx context.Context, name string) error {
	url := c.objectStoragePath(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// GetContainerInfo gets container details from response headers.
func (c *Client) GetContainerInfo(ctx context.Context, name string) (*ContainerInfo, error) {
	url := c.objectStoragePath(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, newAPIError(resp.StatusCode, resp.Status, string(body))
	}

	info := &ContainerInfo{
		ReadACL:          resp.Header.Get("X-Container-Read"),
		WriteACL:         resp.Header.Get("X-Container-Write"),
		VersionsLocation: resp.Header.Get("X-Versions-Location"),
		Metadata:         map[string]string{},
	}
	parseHeaderInt64(resp.Header, "X-Container-Object-Count", &info.ObjectCount)
	parseHeaderInt64(resp.Header, "X-Container-Bytes-Used", &info.BytesUsed)
	parsePrefixedMetadata(resp.Header, "x-container-meta-", info.Metadata)
	return info, nil
}

// ------------------------------------------------------------
// Object Operations
// ------------------------------------------------------------

// ListObjects lists objects in a container.
func (c *Client) ListObjects(ctx context.Context, container string, opts *ListObjectsOptions) ([]Object, error) {
	url := c.objectStoragePath(container)
	params := map[string]string{"format": "json"}
	if opts != nil {
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Marker != "" {
			params["marker"] = opts.Marker
		}
		if opts.EndMarker != "" {
			params["end_marker"] = opts.EndMarker
		}
		if opts.Prefix != "" {
			params["prefix"] = opts.Prefix
		}
		if opts.Delimiter != "" {
			params["delimiter"] = opts.Delimiter
		}
		if opts.Reverse {
			params["reverse"] = "true"
		}
	}
	url += buildQueryString(params)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result []Object
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UploadObject uploads an object to a container.
func (c *Client) UploadObject(ctx context.Context, container, objectName string, data io.Reader) error {
	url := c.objectStoragePath(container, objectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, data)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// DownloadObject downloads an object from a container.
func (c *Client) DownloadObject(ctx context.Context, container, objectName string) (io.ReadCloser, error) {
	url := c.objectStoragePath(container, objectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return resp.Body, nil
}

// DeleteObject deletes an object from a container.
func (c *Client) DeleteObject(ctx context.Context, container, objectName string) error {
	url := c.objectStoragePath(container, objectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// GetObjectInfo gets object details from response headers.
func (c *Client) GetObjectInfo(ctx context.Context, container, objectName string) (*ObjectInfo, error) {
	url := c.objectStoragePath(container, objectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, newAPIError(resp.StatusCode, resp.Status, string(body))
	}

	info := &ObjectInfo{
		ContentType:  resp.Header.Get("Content-Type"),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Metadata:     map[string]string{},
	}
	parseHeaderInt64(resp.Header, "Content-Length", &info.ContentLength)
	parseHeaderInt64(resp.Header, "X-Delete-At", &info.DeleteAt)
	parsePrefixedMetadata(resp.Header, "x-object-meta-", info.Metadata)
	return info, nil
}

// CopyObject copies an object to another container/name.
func (c *Client) CopyObject(ctx context.Context, srcContainer, srcObject, dstContainer, dstObject string) error {
	url := c.objectStoragePath(srcContainer, srcObject)
	req, err := http.NewRequestWithContext(ctx, "COPY", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("Destination", fmt.Sprintf("%s/%s", dstContainer, dstObject))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// ScheduleObjectDeletion schedules an object for deletion at a specific Unix timestamp.
func (c *Client) ScheduleObjectDeletion(ctx context.Context, container, objectName string, deleteAt int64) error {
	url := c.objectStoragePath(container, objectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Delete-At", fmt.Sprintf("%d", deleteAt))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// ScheduleObjectDeletionAfter schedules an object for deletion after a duration in seconds.
func (c *Client) ScheduleObjectDeletionAfter(ctx context.Context, container, objectName string, deleteAfterSeconds int64) error {
	url := c.objectStoragePath(container, objectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Delete-After", fmt.Sprintf("%d", deleteAfterSeconds))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// ------------------------------------------------------------
// Container Configuration
// ------------------------------------------------------------

// EnableVersioning enables object versioning on a container.
func (c *Client) EnableVersioning(ctx context.Context, container, versionsContainer string) error {
	url := c.objectStoragePath(container)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Versions-Location", versionsContainer)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// DisableVersioning disables object versioning on a container.
func (c *Client) DisableVersioning(ctx context.Context, container string) error {
	url := c.objectStoragePath(container)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Remove-Versions-Location", "")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// EnableWebPublishing makes a container publicly accessible.
func (c *Client) EnableWebPublishing(ctx context.Context, container string) error {
	url := c.objectStoragePath(container)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Container-Read", ".r:*")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// DisableWebPublishing disables public access on a container.
func (c *Client) DisableWebPublishing(ctx context.Context, container string) error {
	url := c.objectStoragePath(container)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Container-Read", "")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// SetTempURLKey registers a key for temporary URL generation.
func (c *Client) SetTempURLKey(ctx context.Context, key string) error {
	url := c.objectStoragePath()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Account-Meta-Temp-URL-Key", key)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// RemoveTempURLKey removes the temporary URL key from the account metadata.
func (c *Client) RemoveTempURLKey(ctx context.Context) error {
	url := c.objectStoragePath()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Remove-Account-Meta-Temp-URL-Key", "")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// GenerateTempURL generates a signed temporary URL for object access.
func (c *Client) GenerateTempURL(method, container, objectName, key string, expires int64) (string, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return "", fmt.Errorf("method is required")
	}
	if container == "" {
		return "", fmt.Errorf("container is required")
	}
	if objectName == "" {
		return "", fmt.Errorf("object name is required")
	}
	if key == "" {
		return "", fmt.Errorf("temp url key is required")
	}
	if expires <= 0 {
		return "", fmt.Errorf("expires must be greater than 0")
	}

	u, err := url.Parse(c.objectStoragePath(container, objectName))
	if err != nil {
		return "", err
	}
	path := u.EscapedPath()
	payload := method + "\n" + strconv.FormatInt(expires, 10) + "\n" + path

	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	q := u.Query()
	q.Set("temp_url_sig", sig)
	q.Set("temp_url_expires", strconv.FormatInt(expires, 10))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// ------------------------------------------------------------
// Large Object Upload
// ------------------------------------------------------------

// CreateDLOManifest creates a Dynamic Large Object manifest.
func (c *Client) CreateDLOManifest(ctx context.Context, container, manifestName, segmentContainer, segmentPrefix string) error {
	url := c.objectStoragePath(container, manifestName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.Token)
	req.Header.Set("X-Object-Manifest", fmt.Sprintf("%s/%s", segmentContainer, segmentPrefix))
	req.ContentLength = 0

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// CreateSLOManifest creates a Static Large Object manifest.
func (c *Client) CreateSLOManifest(ctx context.Context, container, manifestName string, segments []SLOSegment) error {
	url := c.objectStoragePath(container, manifestName) + "?multipart-manifest=put"

	req, err := c.newRequest(ctx, http.MethodPut, url, segments)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, resp.Status, string(body))
	}
	return nil
}

// ------------------------------------------------------------
// Helpers
// ------------------------------------------------------------

func parseHeaderInt64(h http.Header, key string, dest *int64) {
	val := h.Get(key)
	if val == "" {
		return
	}
	var n int64
	fmt.Sscanf(val, "%d", &n)
	*dest = n
}

func parsePrefixedMetadata(h http.Header, prefix string, metadata map[string]string) {
	for k, vals := range h {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, prefix) && len(vals) > 0 {
			metadata[strings.TrimPrefix(lk, prefix)] = vals[0]
		}
	}
}
