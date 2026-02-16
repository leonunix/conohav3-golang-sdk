package conoha

import (
	"context"
	"fmt"
	"net/http"
)

// ------------------------------------------------------------
// Load Balancer Types
// ------------------------------------------------------------

// LoadBalancer represents a load balancer.
type LoadBalancer struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	ProvisioningStatus string          `json:"provisioning_status"`
	OperatingStatus    string          `json:"operating_status"`
	AdminStateUp       bool            `json:"admin_state_up"`
	ProjectID          string          `json:"project_id"`
	VIPAddress         string          `json:"vip_address"`
	VIPPortID          string          `json:"vip_port_id"`
	VIPSubnetID        string          `json:"vip_subnet_id"`
	VIPNetworkID       string          `json:"vip_network_id"`
	Listeners          []IDRef         `json:"listeners"`
	Pools              []IDRef         `json:"pools"`
	TenantID           string          `json:"tenant_id"`
}

// IDRef is a simple ID reference.
type IDRef struct {
	ID string `json:"id"`
}

// Listener represents a load balancer listener.
type Listener struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	ProvisioningStatus string  `json:"provisioning_status"`
	OperatingStatus    string  `json:"operating_status"`
	AdminStateUp       bool    `json:"admin_state_up"`
	Protocol           string  `json:"protocol"`
	ProtocolPort       int     `json:"protocol_port"`
	ConnectionLimit    int     `json:"connection_limit"`
	ProjectID          string  `json:"project_id"`
	DefaultPoolID      *string `json:"default_pool_id"`
	LoadBalancers      []IDRef `json:"loadbalancers"`
	TenantID           string  `json:"tenant_id"`
}

// Pool represents a load balancer pool.
type Pool struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	ProvisioningStatus string  `json:"provisioning_status"`
	OperatingStatus    string  `json:"operating_status"`
	AdminStateUp       bool    `json:"admin_state_up"`
	Protocol           string  `json:"protocol"`
	LBAlgorithm        string  `json:"lb_algorithm"`
	ProjectID          string  `json:"project_id"`
	LoadBalancers      []IDRef `json:"loadbalancers"`
	Listeners          []IDRef `json:"listeners"`
	Members            []IDRef `json:"members"`
	TenantID           string  `json:"tenant_id"`
}

// Member represents a pool member.
type Member struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	OperatingStatus    string `json:"operating_status"`
	ProvisioningStatus string `json:"provisioning_status"`
	AdminStateUp       bool   `json:"admin_state_up"`
	Address            string `json:"address"`
	ProtocolPort       int    `json:"protocol_port"`
	Weight             int    `json:"weight"`
	ProjectID          string `json:"project_id"`
	TenantID           string `json:"tenant_id"`
}

// HealthMonitor represents a health monitor.
type HealthMonitor struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Type               string  `json:"type"`
	Delay              int     `json:"delay"`
	Timeout            int     `json:"timeout"`
	MaxRetries         int     `json:"max_retries"`
	URLPath            *string `json:"url_path"`
	ExpectedCodes      *string `json:"expected_codes"`
	AdminStateUp       bool    `json:"admin_state_up"`
	ProjectID          string  `json:"project_id"`
	Pools              []IDRef `json:"pools"`
	ProvisioningStatus string  `json:"provisioning_status"`
	OperatingStatus    string  `json:"operating_status"`
	TenantID           string  `json:"tenant_id"`
}

// CreateHealthMonitorRequest is the request to create a health monitor.
type CreateHealthMonitorRequest struct {
	Name          string  `json:"name"`
	PoolID        string  `json:"pool_id"`
	Delay         int     `json:"delay"`
	MaxRetries    int     `json:"max_retries"`
	Timeout       int     `json:"timeout"`
	Type          string  `json:"type"`
	URLPath       string  `json:"url_path,omitempty"`
	ExpectedCodes string  `json:"expected_codes,omitempty"`
}

type lbListResponse struct {
	LoadBalancers []LoadBalancer `json:"loadbalancers"`
}

type lbResponse struct {
	LoadBalancer LoadBalancer `json:"loadbalancer"`
}

type listenerListResponse struct {
	Listeners []Listener `json:"listeners"`
}

type listenerResponse struct {
	Listener Listener `json:"listener"`
}

type poolListResponse struct {
	Pools []Pool `json:"pools"`
}

type poolResponse struct {
	Pool Pool `json:"pool"`
}

type memberListResponse struct {
	Members []Member `json:"members"`
}

type memberResponse struct {
	Member Member `json:"member"`
}

type healthMonitorListResponse struct {
	HealthMonitors []HealthMonitor `json:"healthmonitors"`
}

type healthMonitorResponse struct {
	HealthMonitor HealthMonitor `json:"healthmonitor"`
}

// ------------------------------------------------------------
// Load Balancers
// ------------------------------------------------------------

// ListLoadBalancers lists all load balancers.
func (c *Client) ListLoadBalancers(ctx context.Context) ([]LoadBalancer, error) {
	url := c.LBaaSURL + "/lbaas/loadbalancers"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result lbListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.LoadBalancers, nil
}

// GetLoadBalancer gets a load balancer's details.
func (c *Client) GetLoadBalancer(ctx context.Context, lbID string) (*LoadBalancer, error) {
	url := fmt.Sprintf("%s/lbaas/loadbalancers/%s", c.LBaaSURL, lbID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result lbResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.LoadBalancer, nil
}

// CreateLoadBalancer creates a load balancer.
func (c *Client) CreateLoadBalancer(ctx context.Context, name string) (*LoadBalancer, error) {
	url := c.LBaaSURL + "/lbaas/loadbalancers"
	body := map[string]interface{}{
		"loadbalancer": map[string]string{"name": name},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result lbResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.LoadBalancer, nil
}

// UpdateLoadBalancer updates a load balancer's name.
func (c *Client) UpdateLoadBalancer(ctx context.Context, lbID, name string) (*LoadBalancer, error) {
	url := fmt.Sprintf("%s/lbaas/loadbalancers/%s", c.LBaaSURL, lbID)
	body := map[string]interface{}{
		"loadbalancer": map[string]string{"name": name},
	}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result lbResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.LoadBalancer, nil
}

// DeleteLoadBalancer deletes a load balancer.
func (c *Client) DeleteLoadBalancer(ctx context.Context, lbID string) error {
	url := fmt.Sprintf("%s/lbaas/loadbalancers/%s", c.LBaaSURL, lbID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Listeners
// ------------------------------------------------------------

// ListListeners lists all listeners.
func (c *Client) ListListeners(ctx context.Context) ([]Listener, error) {
	url := c.LBaaSURL + "/lbaas/listeners"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result listenerListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Listeners, nil
}

// GetListener gets a listener's details.
func (c *Client) GetListener(ctx context.Context, listenerID string) (*Listener, error) {
	url := fmt.Sprintf("%s/lbaas/listeners/%s", c.LBaaSURL, listenerID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result listenerResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Listener, nil
}

// CreateListener creates a listener.
func (c *Client) CreateListener(ctx context.Context, name, protocol string, port int, lbID string) (*Listener, error) {
	url := c.LBaaSURL + "/lbaas/listeners"
	body := map[string]interface{}{
		"listener": map[string]interface{}{
			"name":            name,
			"protocol":        protocol,
			"protocol_port":   port,
			"loadbalancer_id": lbID,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result listenerResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Listener, nil
}

// UpdateListener updates a listener's name.
func (c *Client) UpdateListener(ctx context.Context, listenerID, name string) (*Listener, error) {
	url := fmt.Sprintf("%s/lbaas/listeners/%s", c.LBaaSURL, listenerID)
	body := map[string]interface{}{
		"listener": map[string]string{"name": name},
	}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result listenerResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Listener, nil
}

// DeleteListener deletes a listener.
func (c *Client) DeleteListener(ctx context.Context, listenerID string) error {
	url := fmt.Sprintf("%s/lbaas/listeners/%s", c.LBaaSURL, listenerID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Pools
// ------------------------------------------------------------

// ListPools lists all pools.
func (c *Client) ListPools(ctx context.Context) ([]Pool, error) {
	url := c.LBaaSURL + "/lbaas/pools"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result poolListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Pools, nil
}

// GetPool gets a pool's details.
func (c *Client) GetPool(ctx context.Context, poolID string) (*Pool, error) {
	url := fmt.Sprintf("%s/lbaas/pools/%s", c.LBaaSURL, poolID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result poolResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Pool, nil
}

// CreatePool creates a pool.
func (c *Client) CreatePool(ctx context.Context, name, protocol, lbAlgorithm, listenerID string) (*Pool, error) {
	url := c.LBaaSURL + "/lbaas/pools"
	body := map[string]interface{}{
		"pool": map[string]string{
			"name":         name,
			"protocol":     protocol,
			"lb_algorithm": lbAlgorithm,
			"listener_id":  listenerID,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result poolResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Pool, nil
}

// UpdatePool updates a pool.
func (c *Client) UpdatePool(ctx context.Context, poolID string, name, lbAlgorithm string) (*Pool, error) {
	url := fmt.Sprintf("%s/lbaas/pools/%s", c.LBaaSURL, poolID)
	poolBody := map[string]string{}
	if name != "" {
		poolBody["name"] = name
	}
	if lbAlgorithm != "" {
		poolBody["lb_algorithm"] = lbAlgorithm
	}
	body := map[string]interface{}{"pool": poolBody}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result poolResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Pool, nil
}

// DeletePool deletes a pool.
func (c *Client) DeletePool(ctx context.Context, poolID string) error {
	url := fmt.Sprintf("%s/lbaas/pools/%s", c.LBaaSURL, poolID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Members
// ------------------------------------------------------------

// ListMembers lists all members of a pool.
func (c *Client) ListMembers(ctx context.Context, poolID string) ([]Member, error) {
	url := fmt.Sprintf("%s/lbaas/pools/%s/members", c.LBaaSURL, poolID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result memberListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Members, nil
}

// GetMember gets a member's details.
func (c *Client) GetMember(ctx context.Context, poolID, memberID string) (*Member, error) {
	url := fmt.Sprintf("%s/lbaas/pools/%s/members/%s", c.LBaaSURL, poolID, memberID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result memberResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Member, nil
}

// AddMember adds a member to a pool.
func (c *Client) AddMember(ctx context.Context, poolID, name, address string, port int) (*Member, error) {
	url := fmt.Sprintf("%s/lbaas/pools/%s/members", c.LBaaSURL, poolID)
	body := map[string]interface{}{
		"member": map[string]interface{}{
			"name":          name,
			"address":       address,
			"protocol_port": port,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result memberResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Member, nil
}

// UpdateMember updates a member (enable/disable).
func (c *Client) UpdateMember(ctx context.Context, poolID, memberID string, adminStateUp bool) (*Member, error) {
	url := fmt.Sprintf("%s/lbaas/pools/%s/members/%s", c.LBaaSURL, poolID, memberID)
	body := map[string]interface{}{
		"member": map[string]bool{"admin_state_up": adminStateUp},
	}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result memberResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Member, nil
}

// DeleteMember removes a member from a pool.
func (c *Client) DeleteMember(ctx context.Context, poolID, memberID string) error {
	url := fmt.Sprintf("%s/lbaas/pools/%s/members/%s", c.LBaaSURL, poolID, memberID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Health Monitors
// ------------------------------------------------------------

// ListHealthMonitors lists all health monitors.
func (c *Client) ListHealthMonitors(ctx context.Context) ([]HealthMonitor, error) {
	url := c.LBaaSURL + "/lbaas/healthmonitors"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result healthMonitorListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.HealthMonitors, nil
}

// GetHealthMonitor gets a health monitor's details.
func (c *Client) GetHealthMonitor(ctx context.Context, hmID string) (*HealthMonitor, error) {
	url := fmt.Sprintf("%s/lbaas/healthmonitors/%s", c.LBaaSURL, hmID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result healthMonitorResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.HealthMonitor, nil
}

// CreateHealthMonitor creates a health monitor.
func (c *Client) CreateHealthMonitor(ctx context.Context, opts CreateHealthMonitorRequest) (*HealthMonitor, error) {
	url := c.LBaaSURL + "/lbaas/healthmonitors"
	body := map[string]interface{}{"healthmonitor": opts}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result healthMonitorResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.HealthMonitor, nil
}

// UpdateHealthMonitor updates a health monitor's name.
func (c *Client) UpdateHealthMonitor(ctx context.Context, hmID, name string) (*HealthMonitor, error) {
	url := fmt.Sprintf("%s/lbaas/healthmonitors/%s", c.LBaaSURL, hmID)
	body := map[string]interface{}{
		"healthmonitor": map[string]string{"name": name},
	}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result healthMonitorResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.HealthMonitor, nil
}

// DeleteHealthMonitor deletes a health monitor.
func (c *Client) DeleteHealthMonitor(ctx context.Context, hmID string) error {
	url := fmt.Sprintf("%s/lbaas/healthmonitors/%s", c.LBaaSURL, hmID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}
