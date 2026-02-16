package conoha

import (
	"context"
	"fmt"
	"net/http"
)

// ------------------------------------------------------------
// Volume Types
// ------------------------------------------------------------

// Volume represents a block storage volume.
type Volume struct {
	ID               string                 `json:"id"`
	Status           string                 `json:"status"`
	Size             int                    `json:"size"`
	AvailabilityZone string                 `json:"availability_zone"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	Name             string                 `json:"name"`
	Description      *string                `json:"description"`
	VolumeType       string                 `json:"volume_type"`
	SnapshotID       *string                `json:"snapshot_id"`
	SourceVolID      *string                `json:"source_volid"`
	Metadata         map[string]string      `json:"metadata"`
	UserID           string                 `json:"user_id"`
	Bootable         string                 `json:"bootable"`
	Encrypted        bool                   `json:"encrypted"`
	Multiattach      bool                   `json:"multiattach"`
	Attachments      []interface{}          `json:"attachments"`
	Links            []Link                 `json:"links,omitempty"`
}

// VolumeType represents a volume type.
type VolumeType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsPublic    bool   `json:"is_public"`
	Description string `json:"description"`
}

// CreateVolumeRequest is the request to create a volume.
type CreateVolumeRequest struct {
	Size       int    `json:"size"`
	Name       string `json:"name"`
	VolumeType string `json:"volume_type"`
	Description *string `json:"description,omitempty"`
	ImageRef   string `json:"imageRef,omitempty"`
	SourceVolID string `json:"source_volid,omitempty"`
	BackupID   string `json:"backup_id,omitempty"`
}

// VolumeImageSaveResponse is the response from saving a volume as an image.
type VolumeImageSaveResponse struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Size            int    `json:"size"`
	ImageID         string `json:"image_id"`
	ContainerFormat string `json:"container_format"`
	DiskFormat      string `json:"disk_format"`
	ImageName       string `json:"image_name"`
}

type volumeListResponse struct {
	Volumes []Volume `json:"volumes"`
}

type volumeResponse struct {
	Volume Volume `json:"volume"`
}

type volumeTypeListResponse struct {
	VolumeTypes []VolumeType `json:"volume_types"`
}

type volumeTypeResponse struct {
	VolumeType VolumeType `json:"volume_type"`
}

// ListVolumesOptions are options for listing volumes.
type ListVolumesOptions struct {
	Limit     int
	Offset    int
	Marker    string
	Sort      string
	WithCount bool
}

// ListVolumes lists volumes (basic).
func (c *Client) ListVolumes(ctx context.Context, opts *ListVolumesOptions) ([]Volume, error) {
	url := fmt.Sprintf("%s/%s/volumes", c.BlockStorageURL, c.TenantID)
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", opts.Offset)
		}
		if opts.Marker != "" {
			params["marker"] = opts.Marker
		}
		if opts.Sort != "" {
			params["sort"] = opts.Sort
		}
		if opts.WithCount {
			params["with_count"] = "true"
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result volumeListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Volumes, nil
}

// ListVolumesDetail lists volumes with full details.
func (c *Client) ListVolumesDetail(ctx context.Context, opts *ListVolumesOptions) ([]Volume, error) {
	url := fmt.Sprintf("%s/%s/volumes/detail", c.BlockStorageURL, c.TenantID)
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", opts.Offset)
		}
		if opts.Marker != "" {
			params["marker"] = opts.Marker
		}
		if opts.Sort != "" {
			params["sort"] = opts.Sort
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result volumeListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Volumes, nil
}

// GetVolume gets a volume's details.
func (c *Client) GetVolume(ctx context.Context, volumeID string) (*Volume, error) {
	url := fmt.Sprintf("%s/%s/volumes/%s", c.BlockStorageURL, c.TenantID, volumeID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result volumeResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Volume, nil
}

// CreateVolume creates a new volume.
func (c *Client) CreateVolume(ctx context.Context, opts CreateVolumeRequest) (*Volume, error) {
	url := fmt.Sprintf("%s/%s/volumes", c.BlockStorageURL, c.TenantID)
	body := map[string]interface{}{"volume": opts}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result volumeResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Volume, nil
}

// DeleteVolume deletes a volume.
func (c *Client) DeleteVolume(ctx context.Context, volumeID string, force bool) error {
	url := fmt.Sprintf("%s/%s/volumes/%s", c.BlockStorageURL, c.TenantID, volumeID)
	if force {
		url += "?force=true"
	}
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// UpdateVolume updates a volume's name and description.
func (c *Client) UpdateVolume(ctx context.Context, volumeID, name string, description *string) (*Volume, error) {
	url := fmt.Sprintf("%s/%s/volumes/%s", c.BlockStorageURL, c.TenantID, volumeID)
	volumeBody := map[string]interface{}{"name": name}
	if description != nil {
		volumeBody["description"] = *description
	}
	body := map[string]interface{}{"volume": volumeBody}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result volumeResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Volume, nil
}

// SaveVolumeAsImage saves a volume as an image.
func (c *Client) SaveVolumeAsImage(ctx context.Context, volumeID, imageName string) (*VolumeImageSaveResponse, error) {
	url := fmt.Sprintf("%s/%s/volumes/%s/action", c.BlockStorageURL, c.TenantID, volumeID)
	body := map[string]interface{}{
		"os-volume_upload_image": map[string]string{"image_name": imageName},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result map[string]VolumeImageSaveResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	data := result["os-volume_upload_image"]
	return &data, nil
}

// ------------------------------------------------------------
// Volume Types
// ------------------------------------------------------------

// ListVolumeTypes lists available volume types.
func (c *Client) ListVolumeTypes(ctx context.Context) ([]VolumeType, error) {
	url := fmt.Sprintf("%s/%s/types", c.BlockStorageURL, c.TenantID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result volumeTypeListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.VolumeTypes, nil
}

// GetVolumeType gets a volume type's details.
func (c *Client) GetVolumeType(ctx context.Context, volumeTypeID string) (*VolumeType, error) {
	url := fmt.Sprintf("%s/%s/types/%s", c.BlockStorageURL, c.TenantID, volumeTypeID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result volumeTypeResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.VolumeType, nil
}

// ------------------------------------------------------------
// Backups
// ------------------------------------------------------------

// Backup represents a volume backup.
type Backup struct {
	ID                  string            `json:"id"`
	Status              string            `json:"status"`
	Size                int               `json:"size"`
	ObjectCount         int               `json:"object_count"`
	AvailabilityZone    *string           `json:"availability_zone"`
	Container           string            `json:"container"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
	Name                string            `json:"name"`
	Description         *string           `json:"description"`
	FailReason          *string           `json:"fail_reason"`
	VolumeID            string            `json:"volume_id"`
	Links               []Link            `json:"links,omitempty"`
	IsIncremental       bool              `json:"is_incremental"`
	HasDependentBackups bool              `json:"has_dependent_backups"`
	SnapshotID          *string           `json:"snapshot_id"`
	DataTimestamp       string            `json:"data_timestamp,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

// BackupRestoreResponse is the response from restoring a backup.
type BackupRestoreResponse struct {
	BackupID   string `json:"backup_id"`
	VolumeID   string `json:"volume_id"`
	VolumeName string `json:"volume_name"`
}

type backupListResponse struct {
	Backups []Backup `json:"backups"`
}

type backupResponse struct {
	Backup Backup `json:"backup"`
}

// ListBackupsOptions are options for listing backups.
type ListBackupsOptions struct {
	Limit  int
	Offset int
	Sort   string
}

// ListBackups lists backups (basic).
func (c *Client) ListBackups(ctx context.Context, opts *ListBackupsOptions) ([]Backup, error) {
	url := fmt.Sprintf("%s/%s/backups", c.BlockStorageURL, c.TenantID)
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", opts.Offset)
		}
		if opts.Sort != "" {
			params["sort"] = opts.Sort
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result backupListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Backups, nil
}

// ListBackupsDetail lists backups with full details.
func (c *Client) ListBackupsDetail(ctx context.Context, opts *ListBackupsOptions) ([]Backup, error) {
	url := fmt.Sprintf("%s/%s/backups/detail", c.BlockStorageURL, c.TenantID)
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", opts.Offset)
		}
		if opts.Sort != "" {
			params["sort"] = opts.Sort
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result backupListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Backups, nil
}

// GetBackup gets a backup's details.
func (c *Client) GetBackup(ctx context.Context, backupID string) (*Backup, error) {
	url := fmt.Sprintf("%s/%s/backups/%s", c.BlockStorageURL, c.TenantID, backupID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result backupResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Backup, nil
}

// EnableAutoBackup enables auto-backup for a server.
func (c *Client) EnableAutoBackup(ctx context.Context, serverID string) (*Backup, error) {
	url := fmt.Sprintf("%s/%s/backups", c.BlockStorageURL, c.TenantID)
	body := map[string]interface{}{
		"backup": map[string]string{"instance_uuid": serverID},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result backupResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Backup, nil
}

// DisableAutoBackup disables auto-backup for a server.
func (c *Client) DisableAutoBackup(ctx context.Context, serverID string) error {
	url := fmt.Sprintf("%s/%s/backups/%s", c.BlockStorageURL, c.TenantID, serverID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// RestoreBackup restores a backup to a volume.
func (c *Client) RestoreBackup(ctx context.Context, backupID, volumeID string) (*BackupRestoreResponse, error) {
	url := fmt.Sprintf("%s/%s/backups/%s/restore", c.BlockStorageURL, c.TenantID, backupID)
	body := map[string]interface{}{
		"restore": map[string]string{"volume_id": volumeID},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result map[string]BackupRestoreResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	data := result["restore"]
	return &data, nil
}
