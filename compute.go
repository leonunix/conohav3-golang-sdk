package conoha

import (
	"context"
	"fmt"
	"net/http"
)

// ------------------------------------------------------------
// Server Types
// ------------------------------------------------------------

// Server represents a server (basic).
type Server struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Links []Link `json:"links"`
}

// ServerDetail represents a server with full details.
type ServerDetail struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Status              string                 `json:"status"`
	TenantID            string                 `json:"tenant_id"`
	UserID              string                 `json:"user_id"`
	Metadata            map[string]string      `json:"metadata"`
	HostID              string                 `json:"hostId"`
	Image               interface{}            `json:"image"`
	Flavor              FlavorRef              `json:"flavor"`
	Created             string                 `json:"created"`
	Updated             string                 `json:"updated"`
	Addresses           map[string][]Address   `json:"addresses"`
	AccessIPv4          string                 `json:"accessIPv4"`
	AccessIPv6          string                 `json:"accessIPv6"`
	Links               []Link                 `json:"links"`
	DiskConfig          string                 `json:"OS-DCF:diskConfig"`
	AvailabilityZone    string                 `json:"OS-EXT-AZ:availability_zone"`
	ConfigDrive         string                 `json:"config_drive"`
	KeyName             *string                `json:"key_name"`
	LaunchedAt          string                 `json:"OS-SRV-USG:launched_at"`
	TerminatedAt        *string                `json:"OS-SRV-USG:terminated_at"`
	Host                string                 `json:"OS-EXT-SRV-ATTR:host"`
	InstanceName        string                 `json:"OS-EXT-SRV-ATTR:instance_name"`
	HypervisorHostname  string                 `json:"OS-EXT-SRV-ATTR:hypervisor_hostname"`
	TaskState           *string                `json:"OS-EXT-STS:task_state"`
	VMState             string                 `json:"OS-EXT-STS:vm_state"`
	PowerState          int                    `json:"OS-EXT-STS:power_state"`
	VolumesAttached     []VolumeAttachmentRef  `json:"os-extended-volumes:volumes_attached"`
	SecurityGroups      []SecurityGroupRef     `json:"security_groups"`
	Progress            int                    `json:"progress"`
}

// FlavorRef references a flavor.
type FlavorRef struct {
	ID    string `json:"id"`
	Links []Link `json:"links,omitempty"`
	VCPUs int    `json:"vcpus,omitempty"`
	RAM   int    `json:"ram,omitempty"`
	Disk  int    `json:"disk,omitempty"`
}

// Address represents a network address.
type Address struct {
	Version int    `json:"version"`
	Addr    string `json:"addr"`
	Type    string `json:"OS-EXT-IPS:type"`
	MACAddr string `json:"OS-EXT-IPS-MAC:mac_addr"`
}

// VolumeAttachmentRef references an attached volume.
type VolumeAttachmentRef struct {
	ID string `json:"id"`
}

// SecurityGroupRef references a security group.
type SecurityGroupRef struct {
	Name string `json:"name"`
}

// CreateServerRequest is the request to create a server.
type CreateServerRequest struct {
	FlavorRef           string              `json:"flavorRef"`
	AdminPass           string              `json:"adminPass"`
	BlockDeviceMapping  []BlockDeviceMap    `json:"block_device_mapping_v2"`
	Metadata            map[string]string   `json:"metadata,omitempty"`
	SecurityGroups      []SecurityGroupRef  `json:"security_groups,omitempty"`
	KeyName             string              `json:"key_name,omitempty"`
	UserData            string              `json:"user_data,omitempty"`
}

// BlockDeviceMap represents a block device mapping.
type BlockDeviceMap struct {
	UUID string `json:"uuid"`
}

// CreateServerResponse is the response from creating a server.
type CreateServerResponse struct {
	ID             string             `json:"id"`
	Links          []Link             `json:"links"`
	DiskConfig     string             `json:"OS-DCF:diskConfig"`
	SecurityGroups []SecurityGroupRef `json:"security_groups"`
	AdminPass      string             `json:"adminPass"`
}

// RebuildServerRequest is the request to rebuild a server OS.
type RebuildServerRequest struct {
	ImageRef  string `json:"imageRef"`
	AdminPass string `json:"adminPass"`
	KeyName   string `json:"key_name,omitempty"`
}

// RemoteConsoleRequest is the request for a console URL.
type RemoteConsoleRequest struct {
	Protocol string `json:"protocol"`
	Type     string `json:"type"`
}

// RemoteConsole represents a remote console.
type RemoteConsole struct {
	Protocol string `json:"protocol"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

// ListServersOptions are options for listing servers.
type ListServersOptions struct {
	Limit  int
	Marker string
	Status string
	Name   string
}

type serverListResponse struct {
	Servers []Server `json:"servers"`
}

type serverDetailListResponse struct {
	Servers []ServerDetail `json:"servers"`
}

type serverDetailResponse struct {
	Server ServerDetail `json:"server"`
}

type createServerWrapper struct {
	Server CreateServerResponse `json:"server"`
}

type remoteConsoleResponse struct {
	RemoteConsole RemoteConsole `json:"remote_console"`
}

// ------------------------------------------------------------
// Server CRUD
// ------------------------------------------------------------

// ListServers lists servers (basic).
func (c *Client) ListServers(ctx context.Context, opts *ListServersOptions) ([]Server, error) {
	url := c.ComputeURL + "/servers"
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Marker != "" {
			params["marker"] = opts.Marker
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		if opts.Name != "" {
			params["name"] = opts.Name
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result serverListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Servers, nil
}

// ListServersDetail lists servers with full details.
func (c *Client) ListServersDetail(ctx context.Context, opts *ListServersOptions) ([]ServerDetail, error) {
	url := c.ComputeURL + "/servers/detail"
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Marker != "" {
			params["marker"] = opts.Marker
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		if opts.Name != "" {
			params["name"] = opts.Name
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result serverDetailListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Servers, nil
}

// GetServer gets a server's details.
func (c *Client) GetServer(ctx context.Context, serverID string) (*ServerDetail, error) {
	url := fmt.Sprintf("%s/servers/%s", c.ComputeURL, serverID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result serverDetailResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Server, nil
}

// CreateServer creates a new server.
func (c *Client) CreateServer(ctx context.Context, opts CreateServerRequest) (*CreateServerResponse, error) {
	url := c.ComputeURL + "/servers"
	body := map[string]interface{}{"server": opts}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result createServerWrapper
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Server, nil
}

// DeleteServer deletes a server.
func (c *Client) DeleteServer(ctx context.Context, serverID string) error {
	url := fmt.Sprintf("%s/servers/%s", c.ComputeURL, serverID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Server Actions
// ------------------------------------------------------------

func (c *Client) serverAction(ctx context.Context, serverID string, body interface{}) error {
	url := fmt.Sprintf("%s/servers/%s/action", c.ComputeURL, serverID)
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// StartServer starts a server.
func (c *Client) StartServer(ctx context.Context, serverID string) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{"os-start": nil})
}

// StopServer stops a server.
func (c *Client) StopServer(ctx context.Context, serverID string) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{"os-stop": nil})
}

// RebootServer soft-reboots a server.
func (c *Client) RebootServer(ctx context.Context, serverID string) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{
		"reboot": map[string]string{"type": "SOFT"},
	})
}

// ForceStopServer forces a server to stop.
func (c *Client) ForceStopServer(ctx context.Context, serverID string) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{
		"os-stop": map[string]bool{"force_shutdown": true},
	})
}

// RebuildServer reinstalls the server OS.
func (c *Client) RebuildServer(ctx context.Context, serverID string, opts RebuildServerRequest) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{"rebuild": opts})
}

// ResizeServer initiates a plan change.
func (c *Client) ResizeServer(ctx context.Context, serverID, flavorRef string) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{
		"resize": map[string]string{"flavorRef": flavorRef},
	})
}

// ConfirmResize confirms a resize operation.
func (c *Client) ConfirmResize(ctx context.Context, serverID string) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{"confirmResize": nil})
}

// RevertResize reverts a resize operation.
func (c *Client) RevertResize(ctx context.Context, serverID string) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{"revertResize": nil})
}

// SetVideoDevice sets the video device model (vga, qxl, cirrus).
func (c *Client) SetVideoDevice(ctx context.Context, serverID, model string) error {
	return c.serverAction(ctx, serverID, map[string]string{"hwVideoModel": model})
}

// SetNetworkAdapter sets the network adapter model (virtio, e1000).
func (c *Client) SetNetworkAdapter(ctx context.Context, serverID, model string) error {
	return c.serverAction(ctx, serverID, map[string]string{"hwVifModel": model})
}

// SetStorageController sets the storage controller (virtio, ide).
func (c *Client) SetStorageController(ctx context.Context, serverID, bus string) error {
	return c.serverAction(ctx, serverID, map[string]string{"hwDiskBus": bus})
}

// MountISO mounts an ISO image (enters rescue mode).
func (c *Client) MountISO(ctx context.Context, serverID, imageRef string) (string, error) {
	url := fmt.Sprintf("%s/servers/%s/action", c.ComputeURL, serverID)
	body := map[string]interface{}{
		"rescue": map[string]string{"rescue_image_ref": imageRef},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return "", err
	}
	var result map[string]string
	if _, err := c.do(req, &result); err != nil {
		return "", err
	}
	return result["adminPass"], nil
}

// UnmountISO unmounts an ISO image (exits rescue mode).
func (c *Client) UnmountISO(ctx context.Context, serverID string) error {
	return c.serverAction(ctx, serverID, map[string]interface{}{"unrescue": nil})
}

// ------------------------------------------------------------
// Server Network Info
// ------------------------------------------------------------

type addressesResponse struct {
	Addresses map[string][]Address `json:"addresses"`
}

// GetServerAddresses gets all IP addresses of a server.
func (c *Client) GetServerAddresses(ctx context.Context, serverID string) (map[string][]Address, error) {
	url := fmt.Sprintf("%s/servers/%s/ips", c.ComputeURL, serverID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result addressesResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Addresses, nil
}

// GetServerAddressesByNetwork gets IP addresses of a server for a specific network.
func (c *Client) GetServerAddressesByNetwork(ctx context.Context, serverID, networkName string) ([]Address, error) {
	url := fmt.Sprintf("%s/servers/%s/ips/%s", c.ComputeURL, serverID, networkName)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result map[string][]Address
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result[networkName], nil
}

// ------------------------------------------------------------
// Server Security Groups
// ------------------------------------------------------------

// ServerSecurityGroup represents a security group attached to a server.
type ServerSecurityGroup struct {
	ID          string                    `json:"id"`
	Description string                   `json:"description"`
	Name        string                   `json:"name"`
	TenantID    string                   `json:"tenant_id"`
	Rules       []ServerSecurityGroupRule `json:"rules"`
}

// ServerSecurityGroupRule represents a rule in a server security group.
type ServerSecurityGroupRule struct {
	ID            string      `json:"id"`
	ParentGroupID string      `json:"parent_group_id"`
	IPProtocol    *string     `json:"ip_protocol"`
	FromPort      *int        `json:"from_port"`
	ToPort        *int        `json:"to_port"`
	Group         interface{} `json:"group"`
	IPRange       interface{} `json:"ip_range"`
}

type serverSecurityGroupsResponse struct {
	SecurityGroups []ServerSecurityGroup `json:"security_groups"`
}

// GetServerSecurityGroups gets security groups of a server.
func (c *Client) GetServerSecurityGroups(ctx context.Context, serverID string) ([]ServerSecurityGroup, error) {
	url := fmt.Sprintf("%s/servers/%s/os-security-groups", c.ComputeURL, serverID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result serverSecurityGroupsResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.SecurityGroups, nil
}

// ------------------------------------------------------------
// Server Console
// ------------------------------------------------------------

// GetConsoleURL gets a remote console URL for a server.
func (c *Client) GetConsoleURL(ctx context.Context, serverID string, opts RemoteConsoleRequest) (*RemoteConsole, error) {
	url := fmt.Sprintf("%s/servers/%s/remote-consoles", c.ComputeURL, serverID)
	body := map[string]interface{}{"remote_console": opts}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result remoteConsoleResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.RemoteConsole, nil
}

// GetVNCConsoleURL is a convenience method to get a VNC console URL.
func (c *Client) GetVNCConsoleURL(ctx context.Context, serverID string) (string, error) {
	console, err := c.GetConsoleURL(ctx, serverID, RemoteConsoleRequest{
		Protocol: "vnc",
		Type:     "novnc",
	})
	if err != nil {
		return "", err
	}
	return console.URL, nil
}

// ------------------------------------------------------------
// Server Metadata
// ------------------------------------------------------------

type metadataResponse struct {
	Metadata map[string]string `json:"metadata"`
}

// GetServerMetadata gets a server's metadata.
func (c *Client) GetServerMetadata(ctx context.Context, serverID string) (map[string]string, error) {
	url := fmt.Sprintf("%s/servers/%s/metadata", c.ComputeURL, serverID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result metadataResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Metadata, nil
}

// UpdateServerMetadata updates a server's metadata.
func (c *Client) UpdateServerMetadata(ctx context.Context, serverID string, metadata map[string]string) (map[string]string, error) {
	url := fmt.Sprintf("%s/servers/%s/metadata", c.ComputeURL, serverID)
	body := map[string]interface{}{"metadata": metadata}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result metadataResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Metadata, nil
}

// ------------------------------------------------------------
// Flavors
// ------------------------------------------------------------

// Flavor represents a server flavor (basic).
type Flavor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Links []Link `json:"links"`
}

// FlavorDetail represents a flavor with full details.
type FlavorDetail struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RAM        int    `json:"ram"`
	Disk       int    `json:"disk"`
	Swap       string `json:"swap"`
	VCPUs      int    `json:"vcpus"`
	RxTxFactor float64 `json:"rxtx_factor"`
	Links      []Link `json:"links"`
	Ephemeral  int    `json:"OS-FLV-EXT-DATA:ephemeral"`
	Disabled   bool   `json:"OS-FLV-DISABLED:disabled"`
	IsPublic   bool   `json:"os-flavor-access:is_public"`
}

type flavorListResponse struct {
	Flavors []Flavor `json:"flavors"`
}

type flavorDetailListResponse struct {
	Flavors []FlavorDetail `json:"flavors"`
}

type flavorDetailResponse struct {
	Flavor FlavorDetail `json:"flavor"`
}

// ListFlavors lists available flavors (basic).
func (c *Client) ListFlavors(ctx context.Context) ([]Flavor, error) {
	url := c.ComputeURL + "/flavors"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result flavorListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Flavors, nil
}

// ListFlavorsDetail lists available flavors with full details.
func (c *Client) ListFlavorsDetail(ctx context.Context) ([]FlavorDetail, error) {
	url := c.ComputeURL + "/flavors/detail"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result flavorDetailListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Flavors, nil
}

// GetFlavor gets a flavor's details.
func (c *Client) GetFlavor(ctx context.Context, flavorID string) (*FlavorDetail, error) {
	url := fmt.Sprintf("%s/flavors/%s", c.ComputeURL, flavorID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result flavorDetailResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Flavor, nil
}

// ------------------------------------------------------------
// SSH Keypairs
// ------------------------------------------------------------

// Keypair represents an SSH keypair.
type Keypair struct {
	Name        string  `json:"name"`
	PublicKey   string  `json:"public_key"`
	PrivateKey  string  `json:"private_key,omitempty"`
	Fingerprint string  `json:"fingerprint"`
	UserID      string  `json:"user_id,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	Deleted     bool    `json:"deleted,omitempty"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
	ID          int     `json:"id,omitempty"`
	UpdatedAt   *string `json:"updated_at,omitempty"`
}

type keypairWrapper struct {
	Keypair Keypair `json:"keypair"`
}

type keypairListResponse struct {
	Keypairs []keypairWrapper `json:"keypairs"`
}

type keypairResponse struct {
	Keypair Keypair `json:"keypair"`
}

// ListKeypairsOptions are options for listing keypairs.
type ListKeypairsOptions struct {
	Limit  int
	Marker string
}

// ListKeypairs lists SSH keypairs.
func (c *Client) ListKeypairs(ctx context.Context, opts *ListKeypairsOptions) ([]Keypair, error) {
	url := c.ComputeURL + "/os-keypairs"
	if opts != nil {
		params := map[string]string{}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Marker != "" {
			params["marker"] = opts.Marker
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result keypairListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	keypairs := make([]Keypair, len(result.Keypairs))
	for i, kw := range result.Keypairs {
		keypairs[i] = kw.Keypair
	}
	return keypairs, nil
}

// CreateKeypair generates a new SSH keypair.
func (c *Client) CreateKeypair(ctx context.Context, name string) (*Keypair, error) {
	url := c.ComputeURL + "/os-keypairs"
	body := map[string]interface{}{
		"keypair": map[string]string{"name": name},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result keypairResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Keypair, nil
}

// ImportKeypair imports an existing public key.
func (c *Client) ImportKeypair(ctx context.Context, name, publicKey string) (*Keypair, error) {
	url := c.ComputeURL + "/os-keypairs"
	body := map[string]interface{}{
		"keypair": map[string]string{
			"name":       name,
			"public_key": publicKey,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result keypairResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Keypair, nil
}

// GetKeypair gets a keypair detail.
func (c *Client) GetKeypair(ctx context.Context, name string) (*Keypair, error) {
	url := fmt.Sprintf("%s/os-keypairs/%s", c.ComputeURL, name)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result keypairResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Keypair, nil
}

// DeleteKeypair deletes an SSH keypair.
func (c *Client) DeleteKeypair(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/os-keypairs/%s", c.ComputeURL, name)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Port Attachments (Server Interfaces)
// ------------------------------------------------------------

// InterfaceAttachment represents a port attached to a server.
type InterfaceAttachment struct {
	NetID     string    `json:"net_id"`
	PortID    string    `json:"port_id"`
	MACAddr   string    `json:"mac_addr"`
	PortState string    `json:"port_state"`
	FixedIPs  []FixedIP `json:"fixed_ips"`
}

// FixedIP represents a fixed IP address.
type FixedIP struct {
	SubnetID  string `json:"subnet_id"`
	IPAddress string `json:"ip_address"`
}

type interfaceListResponse struct {
	InterfaceAttachments []InterfaceAttachment `json:"interfaceAttachments"`
}

type interfaceResponse struct {
	InterfaceAttachment InterfaceAttachment `json:"interfaceAttachment"`
}

// ListServerInterfaces lists ports attached to a server.
func (c *Client) ListServerInterfaces(ctx context.Context, serverID string) ([]InterfaceAttachment, error) {
	url := fmt.Sprintf("%s/servers/%s/os-interface", c.ComputeURL, serverID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result interfaceListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.InterfaceAttachments, nil
}

// GetServerInterface gets a specific port attachment.
func (c *Client) GetServerInterface(ctx context.Context, serverID, portID string) (*InterfaceAttachment, error) {
	url := fmt.Sprintf("%s/servers/%s/os-interface/%s", c.ComputeURL, serverID, portID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result interfaceResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.InterfaceAttachment, nil
}

// AttachPort attaches a port to a server.
func (c *Client) AttachPort(ctx context.Context, serverID, portID string) (*InterfaceAttachment, error) {
	url := fmt.Sprintf("%s/servers/%s/os-interface", c.ComputeURL, serverID)
	body := map[string]interface{}{
		"interfaceAttachment": map[string]string{"port_id": portID},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result interfaceResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.InterfaceAttachment, nil
}

// DetachPort detaches a port from a server.
func (c *Client) DetachPort(ctx context.Context, serverID, portID string) error {
	url := fmt.Sprintf("%s/servers/%s/os-interface/%s", c.ComputeURL, serverID, portID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Volume Attachments (Server Volumes)
// ------------------------------------------------------------

// ServerVolumeAttachment represents a volume attached to a server.
type ServerVolumeAttachment struct {
	ID       string `json:"id"`
	VolumeID string `json:"volumeId"`
	ServerID string `json:"serverId"`
	Device   string `json:"device"`
}

type volumeAttachmentListResponse struct {
	VolumeAttachments []ServerVolumeAttachment `json:"volumeAttachments"`
}

type volumeAttachmentResponse struct {
	VolumeAttachment ServerVolumeAttachment `json:"volumeAttachment"`
}

// ListServerVolumes lists volumes attached to a server.
func (c *Client) ListServerVolumes(ctx context.Context, serverID string) ([]ServerVolumeAttachment, error) {
	url := fmt.Sprintf("%s/servers/%s/os-volume_attachments", c.ComputeURL, serverID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result volumeAttachmentListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.VolumeAttachments, nil
}

// GetServerVolume gets a specific volume attachment.
func (c *Client) GetServerVolume(ctx context.Context, serverID, volumeID string) (*ServerVolumeAttachment, error) {
	url := fmt.Sprintf("%s/servers/%s/os-volume_attachments/%s", c.ComputeURL, serverID, volumeID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result volumeAttachmentResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.VolumeAttachment, nil
}

// AttachVolume attaches a volume to a server.
func (c *Client) AttachVolume(ctx context.Context, serverID, volumeID string) (*ServerVolumeAttachment, error) {
	url := fmt.Sprintf("%s/servers/%s/os-volume_attachments", c.ComputeURL, serverID)
	body := map[string]interface{}{
		"volumeAttachment": map[string]string{"volumeId": volumeID},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result volumeAttachmentResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.VolumeAttachment, nil
}

// DetachVolume detaches a volume from a server.
func (c *Client) DetachVolume(ctx context.Context, serverID, volumeID string) error {
	url := fmt.Sprintf("%s/servers/%s/os-volume_attachments/%s", c.ComputeURL, serverID, volumeID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Monitoring (RRD Graphs)
// ------------------------------------------------------------

// RRDData represents monitoring data points.
type RRDData struct {
	Schema []string        `json:"schema"`
	Data   [][]interface{} `json:"data"`
}

// MonitoringOptions are options for monitoring queries.
type MonitoringOptions struct {
	StartDateRaw string // UTC datetime
	EndDateRaw   string // UTC datetime
	Mode         string // average, max, min
}

// GetCPUUsage gets CPU usage data for a server.
func (c *Client) GetCPUUsage(ctx context.Context, serverID string, opts *MonitoringOptions) (*RRDData, error) {
	url := fmt.Sprintf("%s/servers/%s/rrd/cpu", c.ComputeURL, serverID)
	if opts != nil {
		params := map[string]string{}
		if opts.StartDateRaw != "" {
			params["start_date_raw"] = opts.StartDateRaw
		}
		if opts.EndDateRaw != "" {
			params["end_date_raw"] = opts.EndDateRaw
		}
		if opts.Mode != "" {
			params["mode"] = opts.Mode
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]RRDData
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	data := result["cpu"]
	return &data, nil
}

// DiskMonitoringOptions extends MonitoringOptions with device selection.
type DiskMonitoringOptions struct {
	MonitoringOptions
	Device string // vda or vdb
}

// GetDiskIO gets disk I/O data for a server.
func (c *Client) GetDiskIO(ctx context.Context, serverID string, opts *DiskMonitoringOptions) (*RRDData, error) {
	url := fmt.Sprintf("%s/servers/%s/rrd/disk", c.ComputeURL, serverID)
	if opts != nil {
		params := map[string]string{}
		if opts.Device != "" {
			params["device"] = opts.Device
		}
		if opts.StartDateRaw != "" {
			params["start_date_raw"] = opts.StartDateRaw
		}
		if opts.EndDateRaw != "" {
			params["end_date_raw"] = opts.EndDateRaw
		}
		if opts.Mode != "" {
			params["mode"] = opts.Mode
		}
		url += buildQueryString(params)
	}
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]RRDData
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	data := result["disk"]
	return &data, nil
}

// NetworkMonitoringOptions extends MonitoringOptions with port selection.
type NetworkMonitoringOptions struct {
	MonitoringOptions
	PortID string // Required
}

// GetNetworkTraffic gets network traffic data for a server.
func (c *Client) GetNetworkTraffic(ctx context.Context, serverID string, opts NetworkMonitoringOptions) (*RRDData, error) {
	if opts.PortID == "" {
		return nil, fmt.Errorf("conoha: PortID is required for GetNetworkTraffic")
	}
	url := fmt.Sprintf("%s/servers/%s/rrd/interface", c.ComputeURL, serverID)
	params := map[string]string{"port_id": opts.PortID}
	if opts.StartDateRaw != "" {
		params["start_date_raw"] = opts.StartDateRaw
	}
	if opts.EndDateRaw != "" {
		params["end_date_raw"] = opts.EndDateRaw
	}
	if opts.Mode != "" {
		params["mode"] = opts.Mode
	}
	url += buildQueryString(params)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]RRDData
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	data := result["interface"]
	return &data, nil
}
