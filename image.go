package conoha

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// ------------------------------------------------------------
// Image Types
// ------------------------------------------------------------

// Image represents an OS image.
type Image struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Status          string   `json:"status"`
	Visibility      string   `json:"visibility"`
	OSType          string   `json:"os_type,omitempty"`
	Size            int64    `json:"size"`
	DiskFormat      string   `json:"disk_format"`
	ContainerFormat string   `json:"container_format"`
	MinDisk         int      `json:"min_disk"`
	MinRAM          int      `json:"min_ram"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	Checksum        string   `json:"checksum,omitempty"`
	Owner           string   `json:"owner"`
	Protected       bool     `json:"protected"`
	Architecture    string   `json:"architecture,omitempty"`
	Tags            []string `json:"tags"`
	OSHashAlgo      string   `json:"os_hash_algo,omitempty"`
	OSHashValue     string   `json:"os_hash_value,omitempty"`
	OSHidden        bool     `json:"os_hidden"`
	VirtualSize     *int64   `json:"virtual_size"`
	// ISO-specific fields
	HWRescueBus    string `json:"hw_rescue_bus,omitempty"`
	HWRescueDevice string `json:"hw_rescue_device,omitempty"`
	// Additional detail fields
	Bootable            string `json:"bootable,omitempty"`
	HWVideoModel        string `json:"hw_video_model,omitempty"`
	HWVifMultiqueue     string `json:"hw_vif_multiqueue_enabled,omitempty"`
	HWQemuGuestAgent    string `json:"hw_qemu_guest_agent,omitempty"`
}

// ImageQuota represents image storage quota.
type ImageQuota struct {
	ImageSize string `json:"image_size"`
}

// ImageUsage represents image storage usage.
type ImageUsage struct {
	Size int64 `json:"size"`
}

// CreateISOImageRequest is the request to create an ISO image entry.
type CreateISOImageRequest struct {
	Name            string `json:"name"`
	DiskFormat      string `json:"disk_format"`
	HWRescueBus     string `json:"hw_rescue_bus"`
	HWRescueDevice  string `json:"hw_rescue_device"`
	ContainerFormat string `json:"container_format"`
}

type imageListResponse struct {
	Images []Image `json:"images"`
}

type imageQuotaResponse struct {
	Quota ImageQuota `json:"quota"`
}

type imageUsageResponse struct {
	Images ImageUsage `json:"images"`
}

// ListImagesOptions are options for listing images.
type ListImagesOptions struct {
	Limit      int
	Marker     string
	Visibility string // "public" or "shared"
	OSType     string // "linux" or "windows"
	Sort       string
	SortKey    string
	SortDir    string
	Name       string
	Status     string
}

// ListImages lists available images.
func (c *Client) ListImages(ctx context.Context, opts *ListImagesOptions) ([]Image, error) {
	url := c.ImageServiceURL + "/images"
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Marker != "" {
			params["marker"] = opts.Marker
		}
		if opts.Visibility != "" {
			params["visibility"] = opts.Visibility
		}
		if opts.OSType != "" {
			params["os_type"] = opts.OSType
		}
		if opts.Sort != "" {
			params["sort"] = opts.Sort
		}
		if opts.SortKey != "" {
			params["sort_key"] = opts.SortKey
		}
		if opts.SortDir != "" {
			params["sort_dir"] = opts.SortDir
		}
		if opts.Name != "" {
			params["name"] = opts.Name
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result imageListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Images, nil
}

// GetImage gets an image's details.
func (c *Client) GetImage(ctx context.Context, imageID string) (*Image, error) {
	url := fmt.Sprintf("%s/images/%s", c.ImageServiceURL, imageID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result Image
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteImage deletes an image.
func (c *Client) DeleteImage(ctx context.Context, imageID string) error {
	url := fmt.Sprintf("%s/images/%s", c.ImageServiceURL, imageID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// GetImageQuota gets the image storage quota.
func (c *Client) GetImageQuota(ctx context.Context) (*ImageQuota, error) {
	url := c.ImageServiceURL + "/quota"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result imageQuotaResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Quota, nil
}

// GetImageUsage gets the current image storage usage.
func (c *Client) GetImageUsage(ctx context.Context) (*ImageUsage, error) {
	url := c.ImageServiceURL + "/images/total"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result imageUsageResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Images, nil
}

// SetImageQuota changes the image storage quota.
// imageSize format: "50GB", "550GB", etc. Minimum 50GB, additions in 500GB increments.
func (c *Client) SetImageQuota(ctx context.Context, imageSize string) (*ImageQuota, error) {
	url := c.ImageServiceURL + "/quota"
	body := map[string]interface{}{
		"quota": map[string]string{"image_size": imageSize},
	}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result imageQuotaResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Quota, nil
}

// CreateISOImage creates an ISO image metadata entry.
func (c *Client) CreateISOImage(ctx context.Context, name string) (*Image, error) {
	url := c.ImageServiceURL + "/images"
	body := CreateISOImageRequest{
		Name:            name,
		DiskFormat:      "iso",
		HWRescueBus:     "ide",
		HWRescueDevice:  "cdrom",
		ContainerFormat: "bare",
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result Image
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UploadISOImage uploads ISO file data to a previously created image entry.
func (c *Client) UploadISOImage(ctx context.Context, imageID string, data io.Reader) error {
	url := fmt.Sprintf("%s/images/%s/file", c.ImageServiceURL, imageID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Auth-Token", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(respBody),
		}
	}
	return nil
}
