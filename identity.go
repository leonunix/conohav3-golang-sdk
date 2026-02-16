package conoha

import (
	"context"
	"fmt"
	"net/http"
)

// ------------------------------------------------------------
// Token
// ------------------------------------------------------------

// AuthRequest is the request body for token issuance.
type AuthRequest struct {
	Auth AuthBody `json:"auth"`
}

// AuthBody contains identity and scope for authentication.
type AuthBody struct {
	Identity AuthIdentity `json:"identity"`
	Scope    *AuthScope   `json:"scope,omitempty"`
}

// AuthIdentity specifies the authentication method.
type AuthIdentity struct {
	Methods  []string     `json:"methods"`
	Password AuthPassword `json:"password"`
}

// AuthPassword contains user credentials.
type AuthPassword struct {
	User AuthUser `json:"user"`
}

// AuthUser contains user identification.
type AuthUser struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password"`
}

// AuthScope specifies the project scope.
type AuthScope struct {
	Project AuthProject `json:"project"`
}

// AuthProject identifies the project/tenant.
type AuthProject struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Token represents the token response.
type Token struct {
	AuditIDs  []string         `json:"audit_ids"`
	Catalog   []ServiceCatalog `json:"catalog"`
	ExpiresAt string           `json:"expires_at"`
	IssuedAt  string           `json:"issued_at"`
	Methods   []string         `json:"methods"`
	Project   TokenProject     `json:"project"`
	Roles     []Role           `json:"roles"`
	User      TokenUser        `json:"user"`
}

// ServiceCatalog represents a service in the catalog.
type ServiceCatalog struct {
	Endpoints []Endpoint `json:"endpoints"`
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
}

// Endpoint represents a service endpoint.
type Endpoint struct {
	ID        string `json:"id"`
	Interface string `json:"interface"`
	RegionID  string `json:"region_id"`
	URL       string `json:"url"`
	Region    string `json:"region"`
}

// TokenProject represents the project info in a token.
type TokenProject struct {
	Domain DomainRef `json:"domain"`
	ID     string    `json:"id"`
	Name   string    `json:"name"`
}

// DomainRef is a reference to a domain.
type DomainRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TokenUser represents user info in a token.
type TokenUser struct {
	Domain            DomainRef `json:"domain"`
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	PasswordExpiresAt *string   `json:"password_expires_at"`
}

type tokenResponse struct {
	Token Token `json:"token"`
}

// Authenticate authenticates using user ID and tenant ID, setting the token on the client.
func (c *Client) Authenticate(ctx context.Context, userID, password, tenantID string) (*Token, error) {
	req := &AuthRequest{
		Auth: AuthBody{
			Identity: AuthIdentity{
				Methods: []string{"password"},
				Password: AuthPassword{
					User: AuthUser{
						ID:       userID,
						Password: password,
					},
				},
			},
			Scope: &AuthScope{
				Project: AuthProject{ID: tenantID},
			},
		},
	}
	return c.authenticate(ctx, req, tenantID)
}

// AuthenticateByName authenticates using user name and tenant name.
func (c *Client) AuthenticateByName(ctx context.Context, userName, password, tenantName string) (*Token, error) {
	req := &AuthRequest{
		Auth: AuthBody{
			Identity: AuthIdentity{
				Methods: []string{"password"},
				Password: AuthPassword{
					User: AuthUser{
						Name:     userName,
						Password: password,
					},
				},
			},
			Scope: &AuthScope{
				Project: AuthProject{Name: tenantName},
			},
		},
	}
	return c.authenticate(ctx, req, "")
}

func (c *Client) authenticate(ctx context.Context, authReq *AuthRequest, tenantID string) (*Token, error) {
	url := c.IdentityURL + "/auth/tokens"
	httpReq, err := c.newRequest(ctx, http.MethodPost, url, authReq)
	if err != nil {
		return nil, err
	}
	// Don't send X-Auth-Token for authentication
	httpReq.Header.Del("X-Auth-Token")

	var result tokenResponse
	resp, err := c.do(httpReq, &result)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.Token = resp.Header.Get("X-Subject-Token")
	if tenantID != "" {
		c.TenantID = tenantID
	} else if result.Token.Project.ID != "" {
		c.TenantID = result.Token.Project.ID
	}

	// Auto-discover endpoint URLs from Service Catalog.
	// Only overrides URLs that were NOT explicitly set by the user.
	if len(result.Token.Catalog) > 0 {
		c.updateEndpointsFromCatalog(result.Token.Catalog)
	}
	c.mu.Unlock()

	return &result.Token, nil
}

// ------------------------------------------------------------
// Credentials
// ------------------------------------------------------------

// Credential represents an API credential (access/secret key pair).
type Credential struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id,omitempty"`
	TenantID  string `json:"tenant_id,omitempty"`
	Access    string `json:"access"`
	Secret    string `json:"secret"`
	TrustID   *string `json:"trust_id"`
}

type credentialListResponse struct {
	Credentials []Credential `json:"credentials"`
}

type credentialResponse struct {
	Credential Credential `json:"credential"`
}

// ListCredentials lists all credentials for a user.
func (c *Client) ListCredentials(ctx context.Context, userID string) ([]Credential, error) {
	url := fmt.Sprintf("%s/users/%s/credentials/OS-EC2", c.IdentityURL, userID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result credentialListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Credentials, nil
}

// CreateCredential creates a new credential for a user.
func (c *Client) CreateCredential(ctx context.Context, userID, tenantID string) (*Credential, error) {
	url := fmt.Sprintf("%s/users/%s/credentials/OS-EC2", c.IdentityURL, userID)
	body := map[string]string{"tenant_id": tenantID}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result credentialResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Credential, nil
}

// GetCredential gets a credential detail.
func (c *Client) GetCredential(ctx context.Context, userID, credentialID string) (*Credential, error) {
	url := fmt.Sprintf("%s/users/%s/credentials/OS-EC2/%s", c.IdentityURL, userID, credentialID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result credentialResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Credential, nil
}

// DeleteCredential deletes a credential.
func (c *Client) DeleteCredential(ctx context.Context, userID, credentialID string) error {
	url := fmt.Sprintf("%s/users/%s/credentials/OS-EC2/%s", c.IdentityURL, userID, credentialID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Sub-Users
// ------------------------------------------------------------

// SubUser represents a sub-user.
type SubUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Roles []Role `json:"roles"`
}

// Role represents a role.
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type subUserListResponse struct {
	Users []SubUser `json:"users"`
}

type subUserResponse struct {
	User SubUser `json:"user"`
}

// CreateSubUserRequest is the request to create a sub-user.
type CreateSubUserRequest struct {
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

// ListSubUsers lists all sub-users.
func (c *Client) ListSubUsers(ctx context.Context) ([]SubUser, error) {
	url := c.IdentityURL + "/sub-users"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result subUserListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Users, nil
}

// CreateSubUser creates a new sub-user.
func (c *Client) CreateSubUser(ctx context.Context, password string, roles []string) (*SubUser, error) {
	url := c.IdentityURL + "/sub-users"
	body := map[string]interface{}{
		"user": CreateSubUserRequest{
			Password: password,
			Roles:    roles,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result subUserResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

// GetSubUser gets a sub-user detail.
func (c *Client) GetSubUser(ctx context.Context, subUserID string) (*SubUser, error) {
	url := fmt.Sprintf("%s/sub-users/%s", c.IdentityURL, subUserID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result subUserResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

// UpdateSubUser updates a sub-user's password.
func (c *Client) UpdateSubUser(ctx context.Context, subUserID, password string) (*SubUser, error) {
	url := fmt.Sprintf("%s/sub-users/%s", c.IdentityURL, subUserID)
	body := map[string]interface{}{
		"user": map[string]string{"password": password},
	}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result subUserResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

// DeleteSubUser deletes a sub-user.
func (c *Client) DeleteSubUser(ctx context.Context, subUserID string) error {
	url := fmt.Sprintf("%s/sub-users/%s", c.IdentityURL, subUserID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// AssignRolesToSubUser assigns roles to a sub-user.
func (c *Client) AssignRolesToSubUser(ctx context.Context, subUserID string, roleIDs []string) (*SubUser, error) {
	url := fmt.Sprintf("%s/sub-users/%s/assign", c.IdentityURL, subUserID)
	body := map[string]interface{}{"roles": roleIDs}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result subUserResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

// UnassignRolesFromSubUser removes roles from a sub-user.
func (c *Client) UnassignRolesFromSubUser(ctx context.Context, subUserID string, roleIDs []string) (*SubUser, error) {
	url := fmt.Sprintf("%s/sub-users/%s/unassign", c.IdentityURL, subUserID)
	body := map[string]interface{}{"roles": roleIDs}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result subUserResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

// ------------------------------------------------------------
// Roles
// ------------------------------------------------------------

// RoleDetail represents a role with permissions.
type RoleDetail struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Visibility  string   `json:"visibility"`
	Permissions []string `json:"permissions,omitempty"`
}

type roleListResponse struct {
	Roles []RoleDetail `json:"roles"`
}

type roleDetailResponse struct {
	Role RoleDetail `json:"role"`
}

// ListRoles lists all roles.
func (c *Client) ListRoles(ctx context.Context) ([]RoleDetail, error) {
	url := c.IdentityURL + "/sub-users/roles"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result roleListResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Roles, nil
}

// CreateRole creates a new role with permissions.
func (c *Client) CreateRole(ctx context.Context, name string, permissions []string) (*RoleDetail, error) {
	url := c.IdentityURL + "/sub-users/roles"
	body := map[string]interface{}{
		"role": map[string]interface{}{
			"name":        name,
			"permissions": permissions,
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result roleDetailResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Role, nil
}

// GetRole gets a role detail.
func (c *Client) GetRole(ctx context.Context, roleID string) (*RoleDetail, error) {
	url := fmt.Sprintf("%s/sub-users/roles/%s", c.IdentityURL, roleID)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result roleDetailResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Role, nil
}

// UpdateRole updates a role's name.
func (c *Client) UpdateRole(ctx context.Context, roleID, name string) (*RoleDetail, error) {
	url := fmt.Sprintf("%s/sub-users/roles/%s", c.IdentityURL, roleID)
	body := map[string]interface{}{
		"role": map[string]string{"name": name},
	}
	req, err := c.newRequest(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	var result roleDetailResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Role, nil
}

// DeleteRole deletes a role.
func (c *Client) DeleteRole(ctx context.Context, roleID string) error {
	url := fmt.Sprintf("%s/sub-users/roles/%s", c.IdentityURL, roleID)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

// ------------------------------------------------------------
// Permissions
// ------------------------------------------------------------

type permissionsResponse struct {
	Permissions []string `json:"permissions"`
}

// ListPermissions lists all available permissions.
func (c *Client) ListPermissions(ctx context.Context) ([]string, error) {
	url := c.IdentityURL + "/permissions"
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result permissionsResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return result.Permissions, nil
}

// AssignPermissionsToRole assigns permissions to a role.
func (c *Client) AssignPermissionsToRole(ctx context.Context, roleID string, permissions []string) (*RoleDetail, error) {
	url := fmt.Sprintf("%s/sub-users/roles/%s/assign", c.IdentityURL, roleID)
	body := map[string]interface{}{"permissions": permissions}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result roleDetailResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Role, nil
}

// UnassignPermissionsFromRole removes permissions from a role.
func (c *Client) UnassignPermissionsFromRole(ctx context.Context, roleID string, permissions []string) (*RoleDetail, error) {
	url := fmt.Sprintf("%s/sub-users/roles/%s/unassign", c.IdentityURL, roleID)
	body := map[string]interface{}{"permissions": permissions}
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	var result roleDetailResponse
	if _, err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result.Role, nil
}
