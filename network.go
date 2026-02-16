package conoha

import (
	"context"
	"fmt"
	"net/http"
)

// ------------------------------------------------------------
// QoS Policies
// ------------------------------------------------------------

// QoSPolicy represents a QoS policy.
type QoSPolicy struct {
	ID             string          `json:"id"`
	ProjectID      string          `json:"project_id"`
	Name           string          `json:"name"`
	Shared         bool            `json:"shared"`
	Rules          []QoSRule       `json:"rules"`
	IsDefault      bool            `json:"is_default"`
	RevisionNumber int             `json:"revision_number"`
	Description    string          `json:"description"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
	TenantID       string          `json:"tenant_id"`
	Tags           []string        `json:"tags"`
}

// QoSRule represents a QoS bandwidth rule.
type QoSRule struct {
	MaxKbps      int    `json:"max_kbps"`
	MaxBurstKbps int    `json:"max_burst_kbps"`
	Direction    string `json:"direction"`
	ID           string `json:"id"`
	QoSPolicyID  string `json:"qos_policy_id"`
	Type         string `json:"type"`
}

type qosPolicyListResponse struct {
	Policies []QoSPolicy `json:"policies"`
}

type qosPolicyResponse struct {
	Policy QoSPolicy `json:"policy"`
}

// ListQoSPolicies lists all QoS policies.
func (c *Client) ListQoSPolicies(ctx context.Context) ([]QoSPolicy, error) {
	url := c.NetworkingURL + "/v2.0/qos/policies"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result qosPolicyListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Policies, nil
}

// GetQoSPolicy gets a QoS policy's details.
func (c *Client) GetQoSPolicy(ctx context.Context, policyID string) (*QoSPolicy, error) {
	url := fmt.Sprintf("%s/v2.0/qos/policies/%s", c.NetworkingURL, policyID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result qosPolicyResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Policy, nil
}

// ------------------------------------------------------------
// Subnets
// ------------------------------------------------------------

// Subnet represents a network subnet.
type Subnet struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	TenantID        string           `json:"tenant_id"`
	NetworkID       string           `json:"network_id"`
	IPVersion       int              `json:"ip_version"`
	EnableDHCP      bool             `json:"enable_dhcp"`
	IPv6RAMode      *string          `json:"ipv6_ra_mode"`
	IPv6AddressMode *string          `json:"ipv6_address_mode"`
	GatewayIP       *string          `json:"gateway_ip"`
	CIDR            string           `json:"cidr"`
	AllocationPools []AllocationPool `json:"allocation_pools"`
	HostRoutes      []interface{}    `json:"host_routes"`
	DNSNameservers  []string         `json:"dns_nameservers"`
	ProjectID       string           `json:"project_id"`
}

// AllocationPool represents an IP allocation pool range.
type AllocationPool struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type subnetListResponse struct {
	Subnets []Subnet `json:"subnets"`
}

type subnetResponse struct {
	Subnet Subnet `json:"subnet"`
}

// ListSubnets lists all subnets.
func (c *Client) ListSubnets(ctx context.Context) ([]Subnet, error) {
	url := c.NetworkingURL + "/v2.0/subnets"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result subnetListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Subnets, nil
}

// GetSubnet gets a subnet's details.
func (c *Client) GetSubnet(ctx context.Context, subnetID string) (*Subnet, error) {
	url := fmt.Sprintf("%s/v2.0/subnets/%s", c.NetworkingURL, subnetID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result subnetResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Subnet, nil
}

// CreateSubnet creates a subnet on a local network.
func (c *Client) CreateSubnet(ctx context.Context, networkID, cidr string) (*Subnet, error) {
	url := c.NetworkingURL + "/v2.0/subnets"
	body := map[string]interface{}{
		"subnet": map[string]string{
			"network_id": networkID,
			"cidr":       cidr,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result subnetResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Subnet, nil
}

// DeleteSubnet deletes a subnet.
func (c *Client) DeleteSubnet(ctx context.Context, subnetID string) error {
	url := fmt.Sprintf("%s/v2.0/subnets/%s", c.NetworkingURL, subnetID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Security Groups
// ------------------------------------------------------------

// SecurityGroup represents a security group.
type SecurityGroup struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	TenantID    string              `json:"tenant_id"`
	Description string              `json:"description"`
	Shared      bool                `json:"shared"`
	ProjectID   string              `json:"project_id"`
	Rules       []SecurityGroupRule `json:"security_group_rules"`
}

// SecurityGroupRule represents a security group rule.
type SecurityGroupRule struct {
	ID              string  `json:"id"`
	TenantID        string  `json:"tenant_id"`
	SecurityGroupID string  `json:"security_group_id"`
	EtherType       string  `json:"ethertype"`
	Direction       string  `json:"direction"`
	Protocol        *string `json:"protocol"`
	PortRangeMin    *int    `json:"port_range_min"`
	PortRangeMax    *int    `json:"port_range_max"`
	RemoteIPPrefix  *string `json:"remote_ip_prefix"`
	RemoteGroupID   *string `json:"remote_group_id"`
	ProjectID       string  `json:"project_id"`
}

// CreateSecurityGroupRuleRequest is the request to create a security group rule.
type CreateSecurityGroupRuleRequest struct {
	SecurityGroupID string  `json:"security_group_id"`
	Direction       string  `json:"direction"`
	EtherType       string  `json:"ethertype"`
	Protocol        *string `json:"protocol,omitempty"`
	PortRangeMin    *int    `json:"port_range_min,omitempty"`
	PortRangeMax    *int    `json:"port_range_max,omitempty"`
	RemoteIPPrefix  *string `json:"remote_ip_prefix,omitempty"`
	RemoteGroupID   *string `json:"remote_group_id,omitempty"`
}

type securityGroupListResponse struct {
	SecurityGroups []SecurityGroup `json:"security_groups"`
}

type securityGroupResponse struct {
	SecurityGroup SecurityGroup `json:"security_group"`
}

type securityGroupRuleListResponse struct {
	SecurityGroupRules []SecurityGroupRule `json:"security_group_rules"`
}

type securityGroupRuleResponse struct {
	SecurityGroupRule SecurityGroupRule `json:"security_group_rule"`
}

// ListSecurityGroups lists all security groups.
func (c *Client) ListSecurityGroups(ctx context.Context) ([]SecurityGroup, error) {
	url := c.NetworkingURL + "/v2.0/security-groups"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result securityGroupListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.SecurityGroups, nil
}

// GetSecurityGroup gets a security group's details.
func (c *Client) GetSecurityGroup(ctx context.Context, sgID string) (*SecurityGroup, error) {
	url := fmt.Sprintf("%s/v2.0/security-groups/%s", c.NetworkingURL, sgID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result securityGroupResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.SecurityGroup, nil
}

// CreateSecurityGroup creates a security group.
func (c *Client) CreateSecurityGroup(ctx context.Context, name, description string) (*SecurityGroup, error) {
	url := c.NetworkingURL + "/v2.0/security-groups"
	body := map[string]interface{}{
		"security_group": map[string]string{
			"name":        name,
			"description": description,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result securityGroupResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.SecurityGroup, nil
}

// UpdateSecurityGroup updates a security group.
func (c *Client) UpdateSecurityGroup(ctx context.Context, sgID, name, description string) (*SecurityGroup, error) {
	url := fmt.Sprintf("%s/v2.0/security-groups/%s", c.NetworkingURL, sgID)
	sgBody := map[string]string{}
	if name != "" {
		sgBody["name"] = name
	}
	if description != "" {
		sgBody["description"] = description
	}
	body := map[string]interface{}{"security_group": sgBody}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result securityGroupResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.SecurityGroup, nil
}

// DeleteSecurityGroup deletes a security group.
func (c *Client) DeleteSecurityGroup(ctx context.Context, sgID string) error {
	url := fmt.Sprintf("%s/v2.0/security-groups/%s", c.NetworkingURL, sgID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ListSecurityGroupRules lists all security group rules.
func (c *Client) ListSecurityGroupRules(ctx context.Context) ([]SecurityGroupRule, error) {
	url := c.NetworkingURL + "/v2.0/security-group-rules"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result securityGroupRuleListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.SecurityGroupRules, nil
}

// GetSecurityGroupRule gets a security group rule's details.
func (c *Client) GetSecurityGroupRule(ctx context.Context, ruleID string) (*SecurityGroupRule, error) {
	url := fmt.Sprintf("%s/v2.0/security-group-rules/%s", c.NetworkingURL, ruleID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result securityGroupRuleResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.SecurityGroupRule, nil
}

// CreateSecurityGroupRule creates a security group rule.
func (c *Client) CreateSecurityGroupRule(ctx context.Context, opts CreateSecurityGroupRuleRequest) (*SecurityGroupRule, error) {
	url := c.NetworkingURL + "/v2.0/security-group-rules"
	body := map[string]interface{}{"security_group_rule": opts}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result securityGroupRuleResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.SecurityGroupRule, nil
}

// DeleteSecurityGroupRule deletes a security group rule.
func (c *Client) DeleteSecurityGroupRule(ctx context.Context, ruleID string) error {
	url := fmt.Sprintf("%s/v2.0/security-group-rules/%s", c.NetworkingURL, ruleID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Networks
// ------------------------------------------------------------

// Network represents a network.
type Network struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	TenantID     string   `json:"tenant_id"`
	AdminStateUp bool     `json:"admin_state_up"`
	MTU          int      `json:"mtu"`
	Status       string   `json:"status"`
	Subnets      []string `json:"subnets"`
	Shared       bool     `json:"shared"`
	ProjectID    string   `json:"project_id"`
	External     bool     `json:"router:external"`
}

type networkListResponse struct {
	Networks []Network `json:"networks"`
}

type networkResponse struct {
	Network Network `json:"network"`
}

// ListNetworks lists all networks.
func (c *Client) ListNetworks(ctx context.Context) ([]Network, error) {
	url := c.NetworkingURL + "/v2.0/networks"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result networkListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Networks, nil
}

// GetNetwork gets a network's details.
func (c *Client) GetNetwork(ctx context.Context, networkID string) (*Network, error) {
	url := fmt.Sprintf("%s/v2.0/networks/%s", c.NetworkingURL, networkID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result networkResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Network, nil
}

// CreateNetwork creates a local network.
func (c *Client) CreateNetwork(ctx context.Context) (*Network, error) {
	url := c.NetworkingURL + "/v2.0/networks"
	req, err := c.newRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	var result networkResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Network, nil
}

// DeleteNetwork deletes a local network.
func (c *Client) DeleteNetwork(ctx context.Context, networkID string) error {
	url := fmt.Sprintf("%s/v2.0/networks/%s", c.NetworkingURL, networkID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Ports
// ------------------------------------------------------------

// Port represents a network port.
type Port struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	NetworkID           string             `json:"network_id"`
	TenantID            string             `json:"tenant_id"`
	MACAddress          string             `json:"mac_address"`
	AdminStateUp        bool               `json:"admin_state_up"`
	Status              string             `json:"status"`
	DeviceID            string             `json:"device_id"`
	DeviceOwner         string             `json:"device_owner"`
	FixedIPs            []FixedIP          `json:"fixed_ips"`
	ProjectID           string             `json:"project_id"`
	SecurityGroups      []string           `json:"security_groups"`
	AllowedAddressPairs []AddressPair      `json:"allowed_address_pairs"`
	ExtraDHCPOpts       []interface{}      `json:"extra_dhcp_opts"`
	BindingVNICType     string             `json:"binding:vnic_type"`
}

// AddressPair represents an allowed address pair.
type AddressPair struct {
	MACAddress string `json:"mac_address,omitempty"`
	IPAddress  string `json:"ip_address"`
}

// CreatePortRequest is the request to create a port on a local network.
type CreatePortRequest struct {
	NetworkID           string        `json:"network_id"`
	FixedIPs            []FixedIP     `json:"fixed_ips,omitempty"`
	SecurityGroups      []string      `json:"security_groups,omitempty"`
	AllowedAddressPairs []AddressPair `json:"allowed_address_pairs,omitempty"`
}

// UpdatePortRequest is the request to update a port.
type UpdatePortRequest struct {
	SecurityGroups      []string      `json:"security_groups,omitempty"`
	QoSPolicyID         *string       `json:"qos_policy_id,omitempty"`
	FixedIPs            []FixedIP     `json:"fixed_ips,omitempty"`
	AllowedAddressPairs []AddressPair `json:"allowed_address_pairs,omitempty"`
}

// AllocateIPRequest is the request to allocate additional IPs.
type AllocateIPRequest struct {
	Count          int      `json:"count"`
	SecurityGroups []string `json:"security_groups,omitempty"`
}

type portListResponse struct {
	Ports []Port `json:"ports"`
}

type portResponse struct {
	Port Port `json:"port"`
}

// ListPorts lists all ports.
func (c *Client) ListPorts(ctx context.Context) ([]Port, error) {
	url := c.NetworkingURL + "/v2.0/ports"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result portListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Ports, nil
}

// GetPort gets a port's details.
func (c *Client) GetPort(ctx context.Context, portID string) (*Port, error) {
	url := fmt.Sprintf("%s/v2.0/ports/%s", c.NetworkingURL, portID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result portResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Port, nil
}

// CreatePort creates a port on a local network.
func (c *Client) CreatePort(ctx context.Context, opts CreatePortRequest) (*Port, error) {
	url := c.NetworkingURL + "/v2.0/ports"
	body := map[string]interface{}{"port": opts}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result portResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Port, nil
}

// AllocateAdditionalIP allocates additional public IP addresses.
func (c *Client) AllocateAdditionalIP(ctx context.Context, count int, securityGroups []string) (*Port, error) {
	url := c.NetworkingURL + "/v2.0/allocateips"
	body := map[string]interface{}{
		"allocateip": AllocateIPRequest{
			Count:          count,
			SecurityGroups: securityGroups,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result portResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Port, nil
}

// UpdatePort updates a port.
func (c *Client) UpdatePort(ctx context.Context, portID string, opts UpdatePortRequest) (*Port, error) {
	url := fmt.Sprintf("%s/v2.0/ports/%s", c.NetworkingURL, portID)
	body := map[string]interface{}{"port": opts}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result portResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Port, nil
}

// DeletePort deletes a port.
func (c *Client) DeletePort(ctx context.Context, portID string) error {
	url := fmt.Sprintf("%s/v2.0/ports/%s", c.NetworkingURL, portID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}
