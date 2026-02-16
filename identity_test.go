package conoha

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// ============================================================
// Authenticate
// ============================================================

func TestAuthenticate_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	var capturedAuthHeader string
	var capturedBody AuthRequest
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		capturedAuthHeader = r.Header.Get("X-Auth-Token")
		readJSONBody(t, r, &capturedBody)
		w.Header().Set("X-Subject-Token", "new-auth-token")
		w.WriteHeader(201)
		w.Write([]byte(`{
			"token": {
				"catalog": [
					{
						"type": "compute",
						"endpoints": [
							{"interface": "public", "region": "c3j1", "url": "https://catalog-compute.example.com/v2.1"}
						]
					}
				],
				"project": {"id": "project-123"},
				"expires_at": "2025-01-01T00:00:00Z"
			}
		}`))
	})
	defer server.Close()

	// Reset token to verify it gets set
	client.Token = ""
	client.TenantID = ""

	token, err := client.Authenticate(context.Background(), "user-id", "password", "tenant-id")
	assertNoError(t, err)

	// Check request
	if capturedMethod != http.MethodPost {
		t.Errorf("Method = %q", capturedMethod)
	}
	if !strings.HasSuffix(capturedPath, "/auth/tokens") {
		t.Errorf("Path = %q", capturedPath)
	}
	// X-Auth-Token should be removed for auth requests
	if capturedAuthHeader != "" {
		t.Error("X-Auth-Token should be removed for auth requests")
	}

	// Check request body
	if capturedBody.Auth.Identity.Password.User.ID != "user-id" {
		t.Errorf("user ID = %q", capturedBody.Auth.Identity.Password.User.ID)
	}
	if capturedBody.Auth.Scope.Project.ID != "tenant-id" {
		t.Errorf("tenant ID = %q", capturedBody.Auth.Scope.Project.ID)
	}

	// Check client state
	if client.Token != "new-auth-token" {
		t.Errorf("Token = %q", client.Token)
	}
	if client.TenantID != "tenant-id" {
		t.Errorf("TenantID = %q", client.TenantID)
	}

	// Check returned token
	if token == nil {
		t.Fatal("token is nil")
	}
	if token.ExpiresAt != "2025-01-01T00:00:00Z" {
		t.Errorf("ExpiresAt = %q", token.ExpiresAt)
	}
}

func TestAuthenticateByName_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Subject-Token", "name-token")
		w.WriteHeader(201)
		w.Write([]byte(`{
			"token": {
				"catalog": [],
				"project": {"id": "project-from-response"}
			}
		}`))
	})
	defer server.Close()
	client.Token = ""
	client.TenantID = ""

	_, err := client.AuthenticateByName(context.Background(), "user-name", "password", "tenant-name")
	assertNoError(t, err)

	// TenantID should come from response since it's not passed in
	if client.TenantID != "project-from-response" {
		t.Errorf("TenantID = %q, want %q", client.TenantID, "project-from-response")
	}
	if client.Token != "name-token" {
		t.Errorf("Token = %q", client.Token)
	}
}

func TestAuthenticate_Error(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"unauthorized":{"message":"Invalid credentials","code":401}}`))
	})
	defer server.Close()
	client.Token = ""

	_, err := client.Authenticate(context.Background(), "user", "wrong-pass", "tenant")

	assertAPIError(t, err, 401)
	apiErr := err.(*APIError)
	if apiErr.Message != "Invalid credentials" {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

func TestAuthenticate_CatalogAutoDiscovery(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Subject-Token", "token")
		w.WriteHeader(201)
		w.Write([]byte(`{
			"token": {
				"catalog": [
					{
						"type": "compute",
						"endpoints": [{"interface": "public", "region": "c3j1", "url": "https://discovered-compute.example.com/v2.1"}]
					},
					{
						"type": "identity",
						"endpoints": [{"interface": "public", "region": "c3j1", "url": "https://discovered-identity.example.com/v3"}]
					}
				],
				"project": {"id": "tenant-123"}
			}
		}`))
	})
	defer server.Close()

	// IdentityURL and ComputeURL are explicit (set by setupTestServer)
	// so they should NOT be overridden by catalog
	origIdentity := client.IdentityURL
	origCompute := client.ComputeURL

	client.Authenticate(context.Background(), "user", "pass", "tenant")

	if client.IdentityURL != origIdentity {
		t.Errorf("explicit IdentityURL was overridden: %q", client.IdentityURL)
	}
	if client.ComputeURL != origCompute {
		t.Errorf("explicit ComputeURL was overridden: %q", client.ComputeURL)
	}
}

// ============================================================
// Credentials
// ============================================================

func TestListCredentials_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"credentials":[{"access":"ak","secret":"sk","user_id":"uid"}]}`))
	})
	defer server.Close()

	creds, err := client.ListCredentials(context.Background(), "user-123")
	assertNoError(t, err)

	if !strings.Contains(capturedPath, "/users/user-123/credentials/OS-EC2") {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(creds) != 1 {
		t.Fatalf("got %d credentials", len(creds))
	}
	if creds[0].Access != "ak" {
		t.Errorf("Access = %q", creds[0].Access)
	}
}

func TestCreateCredential_Success(t *testing.T) {
	var body map[string]string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"credential":{"access":"new-ak","secret":"new-sk","user_id":"uid"}}`))
	})
	defer server.Close()

	cred, err := client.CreateCredential(context.Background(), "user-123", "tenant-456")
	assertNoError(t, err)

	if body["tenant_id"] != "tenant-456" {
		t.Errorf("tenant_id = %q", body["tenant_id"])
	}
	if cred.Access != "new-ak" {
		t.Errorf("Access = %q", cred.Access)
	}
}

func TestDeleteCredential_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteCredential(context.Background(), "user-123", "cred-456")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if !strings.Contains(capturedPath, "/users/user-123/credentials/OS-EC2/cred-456") {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Sub-Users
// ============================================================

func TestListSubUsers_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"users":[{"id":"sub-1","name":"subuser1"}]}`))
	})
	defer server.Close()

	users, err := client.ListSubUsers(context.Background())
	assertNoError(t, err)

	if len(users) != 1 || users[0].ID != "sub-1" {
		t.Errorf("unexpected users: %+v", users)
	}
}

func TestCreateSubUser_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"user":{"id":"new-sub","name":"newuser"}}`))
	})
	defer server.Close()

	user, err := client.CreateSubUser(context.Background(), "pass123", []string{"role-1"})
	assertNoError(t, err)

	if _, ok := body["user"]; !ok {
		t.Error("body should contain 'user' key")
	}
	if user.ID != "new-sub" {
		t.Errorf("ID = %q", user.ID)
	}
}

func TestDeleteSubUser_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteSubUser(context.Background(), "sub-123")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
}

// ============================================================
// Roles
// ============================================================

func TestListRoles_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"roles":[{"id":"role-1","name":"admin"}]}`))
	})
	defer server.Close()

	roles, err := client.ListRoles(context.Background())
	assertNoError(t, err)

	if len(roles) != 1 || roles[0].Name != "admin" {
		t.Errorf("unexpected roles: %+v", roles)
	}
}

func TestCreateRole_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"role":{"id":"new-role","name":"viewer","permissions":["read"]}}`))
	})
	defer server.Close()

	role, err := client.CreateRole(context.Background(), "viewer", []string{"read"})
	assertNoError(t, err)

	if role.Name != "viewer" {
		t.Errorf("Name = %q", role.Name)
	}
}

func TestDeleteRole_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteRole(context.Background(), "role-123")
	assertNoError(t, err)

	if !strings.Contains(capturedPath, "/sub-users/roles/role-123") {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Permissions
// ============================================================

func TestListPermissions_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"permissions":["compute:read","compute:write"]}`))
	})
	defer server.Close()

	perms, err := client.ListPermissions(context.Background())
	assertNoError(t, err)

	if len(perms) != 2 {
		t.Fatalf("got %d permissions", len(perms))
	}
}
