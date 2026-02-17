//go:build integration

package conoha

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	testClient   *Client
	testUserID   string
	testTenantID string
	testCtx      context.Context
)

func TestMain(m *testing.M) {
	userID := os.Getenv("CONOHA_USER_ID")
	password := os.Getenv("CONOHA_PASSWORD")
	tenantID := os.Getenv("CONOHA_TENANT_ID")
	region := os.Getenv("CONOHA_REGION")

	if userID == "" || password == "" || tenantID == "" {
		fmt.Println("SKIP: CONOHA_USER_ID, CONOHA_PASSWORD, CONOHA_TENANT_ID not set")
		os.Exit(0)
	}

	testUserID = userID
	testTenantID = tenantID
	testCtx = context.Background()

	var opts []ClientOption
	if region != "" {
		opts = append(opts, WithRegion(region))
	}
	if v := os.Getenv("CONOHA_IDENTITY_URL"); v != "" {
		opts = append(opts, WithIdentityURL(v))
	}
	if v := os.Getenv("CONOHA_COMPUTE_URL"); v != "" {
		opts = append(opts, WithComputeURL(v))
	}
	if v := os.Getenv("CONOHA_BLOCK_STORAGE_URL"); v != "" {
		opts = append(opts, WithBlockStorageURL(v))
	}
	if v := os.Getenv("CONOHA_IMAGE_URL"); v != "" {
		opts = append(opts, WithImageServiceURL(v))
	}
	if v := os.Getenv("CONOHA_NETWORK_URL"); v != "" {
		opts = append(opts, WithNetworkingURL(v))
	}
	if v := os.Getenv("CONOHA_LBAAS_URL"); v != "" {
		opts = append(opts, WithLBaaSURL(v))
	}
	if v := os.Getenv("CONOHA_OBJECT_STORAGE_URL"); v != "" {
		opts = append(opts, WithObjectStorageURL(v))
	}
	if v := os.Getenv("CONOHA_DNS_URL"); v != "" {
		opts = append(opts, WithDNSServiceURL(v))
	}
	testClient = NewClient(opts...)

	token, err := testClient.Authenticate(testCtx, userID, password, tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: authentication failed: %v\n", err)
		os.Exit(1)
	}

	// Log service catalog for debugging endpoint resolution.
	fmt.Println("=== Service Catalog ===")
	for _, svc := range token.Catalog {
		fmt.Printf("  type=%q name=%q\n", svc.Type, svc.Name)
		for _, ep := range svc.Endpoints {
			fmt.Printf("    interface=%s region=%s url=%s\n", ep.Interface, ep.Region, ep.URL)
		}
	}
	fmt.Println("======================")

	os.Exit(m.Run())
}

func randomSuffix() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// waitForServerStatus polls GetServer until the status matches or timeout.
func waitForServerStatus(t *testing.T, serverID, target string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s, err := testClient.GetServer(testCtx, serverID)
		if err != nil {
			// If target is empty string, we're waiting for deletion (404).
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
				return
			}
			t.Logf("GetServer(%s): %v (retrying)", serverID, err)
			time.Sleep(5 * time.Second)
			continue
		}
		t.Logf("Server %s status: %s (waiting for %s)", serverID, s.Status, target)
		if s.Status == target {
			return
		}
		if s.Status == "ERROR" {
			t.Fatalf("server %s entered ERROR state", serverID)
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatalf("timeout waiting for server %s to reach status %s", serverID, target)
}

// waitForVolumeStatus polls GetVolume until the status matches or timeout.
func waitForVolumeStatus(t *testing.T, volumeID, target string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		v, err := testClient.GetVolume(testCtx, volumeID)
		if err != nil {
			t.Logf("GetVolume(%s): %v (retrying)", volumeID, err)
			time.Sleep(5 * time.Second)
			continue
		}
		t.Logf("Volume %s status: %s (waiting for %s)", volumeID, v.Status, target)
		if v.Status == target {
			return
		}
		if v.Status == "error" {
			t.Fatalf("volume %s entered error state", volumeID)
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatalf("timeout waiting for volume %s to reach status %s", volumeID, target)
}

// waitForLBStatus polls GetLoadBalancer until provisioning_status matches or timeout.
func waitForLBStatus(t *testing.T, lbID, target string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		lb, err := testClient.GetLoadBalancer(testCtx, lbID)
		if err != nil {
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
				return
			}
			t.Logf("GetLoadBalancer(%s): %v (retrying)", lbID, err)
			time.Sleep(5 * time.Second)
			continue
		}
		t.Logf("LB %s provisioning_status: %s (waiting for %s)", lbID, lb.ProvisioningStatus, target)
		if lb.ProvisioningStatus == target {
			return
		}
		if lb.ProvisioningStatus == "ERROR" {
			t.Fatalf("load balancer %s entered ERROR state", lbID)
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatalf("timeout waiting for LB %s to reach status %s", lbID, target)
}

// ============================================================
// Authentication
// ============================================================

func TestIntegration_Authentication(t *testing.T) {
	t.Run("ClientHasToken", func(t *testing.T) {
		if testClient.Token == "" {
			t.Fatal("client.Token is empty after authentication")
		}
	})

	t.Run("ClientHasTenantID", func(t *testing.T) {
		if testClient.TenantID == "" {
			t.Fatal("client.TenantID is empty after authentication")
		}
		if testClient.TenantID != testTenantID {
			t.Errorf("TenantID = %q, want %q", testClient.TenantID, testTenantID)
		}
	})

	t.Run("EndpointsResolved", func(t *testing.T) {
		t.Logf("IdentityURL:     %s", testClient.IdentityURL)
		t.Logf("ComputeURL:      %s", testClient.ComputeURL)
		t.Logf("NetworkingURL:   %s", testClient.NetworkingURL)
		t.Logf("BlockStorageURL: %s", testClient.BlockStorageURL)
		t.Logf("ImageServiceURL: %s", testClient.ImageServiceURL)
		t.Logf("ObjectStorageURL:%s", testClient.ObjectStorageURL)
		t.Logf("DNSServiceURL:   %s", testClient.DNSServiceURL)
		t.Logf("LBaaSURL:        %s", testClient.LBaaSURL)
		if testClient.ComputeURL == "" {
			t.Error("ComputeURL is empty")
		}
		if testClient.NetworkingURL == "" {
			t.Error("NetworkingURL is empty")
		}
		if testClient.BlockStorageURL == "" {
			t.Error("BlockStorageURL is empty")
		}
		if testClient.ImageServiceURL == "" {
			t.Error("ImageServiceURL is empty")
		}
		if testClient.ObjectStorageURL == "" {
			t.Error("ObjectStorageURL is empty")
		}
		if testClient.DNSServiceURL == "" {
			t.Error("DNSServiceURL is empty")
		}
		if testClient.LBaaSURL == "" {
			t.Error("LBaaSURL is empty")
		}
	})
}

// ============================================================
// Identity - Credential CRUD
// ============================================================

func TestIntegration_Identity_Credential_CRUD(t *testing.T) {
	var cred *Credential

	t.Run("Create", func(t *testing.T) {
		var err error
		cred, err = testClient.CreateCredential(testCtx, testUserID, testTenantID)
		if err != nil {
			t.Fatalf("CreateCredential: %v", err)
		}
		if cred.Access == "" {
			t.Error("Access should not be empty")
		}
		if cred.Secret == "" {
			t.Error("Secret should not be empty")
		}
		t.Logf("Created credential: access=%s", cred.Access)
	})

	defer func() {
		if cred != nil {
			if err := testClient.DeleteCredential(testCtx, testUserID, cred.Access); err != nil {
				t.Logf("WARNING: cleanup DeleteCredential failed: %v", err)
			}
		}
	}()

	if cred == nil {
		t.Fatal("Create failed")
	}

	t.Run("Get", func(t *testing.T) {
		got, err := testClient.GetCredential(testCtx, testUserID, cred.Access)
		if err != nil {
			t.Fatalf("GetCredential: %v", err)
		}
		if got.Access != cred.Access {
			t.Errorf("Access = %q, want %q", got.Access, cred.Access)
		}
	})

	t.Run("ListContains", func(t *testing.T) {
		creds, err := testClient.ListCredentials(testCtx, testUserID)
		if err != nil {
			t.Fatalf("ListCredentials: %v", err)
		}
		found := false
		for _, c := range creds {
			if c.Access == cred.Access {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("created credential %s not found in list", cred.Access)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := testClient.DeleteCredential(testCtx, testUserID, cred.Access)
		if err != nil {
			t.Fatalf("DeleteCredential: %v", err)
		}
		cred = nil
	})
}

// ============================================================
// Identity - Role CRUD (with permissions)
// ============================================================

func TestIntegration_Identity_Role_CRUD(t *testing.T) {
	// First get available permissions.
	perms, err := testClient.ListPermissions(testCtx)
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(perms) == 0 {
		t.Skip("no permissions available")
	}
	t.Logf("Available permissions: %d", len(perms))

	roleName := "sdk-inttest-role-" + randomSuffix()
	// Use first permission for initial creation.
	initPerms := []string{perms[0].Name}

	var role *RoleDetail
	t.Run("Create", func(t *testing.T) {
		role, err = testClient.CreateRole(testCtx, roleName, initPerms)
		if err != nil {
			t.Fatalf("CreateRole: %v", err)
		}
		if role.ID == "" {
			t.Fatal("role ID is empty")
		}
		if role.Name != roleName {
			t.Errorf("Name = %q, want %q", role.Name, roleName)
		}
		t.Logf("Created role: %s (ID: %s)", role.Name, role.ID)
	})

	defer func() {
		if role != nil {
			if err := testClient.DeleteRole(testCtx, role.ID); err != nil {
				t.Logf("WARNING: cleanup DeleteRole failed: %v", err)
			}
		}
	}()

	if role == nil {
		t.Fatal("Create failed")
	}

	t.Run("Get", func(t *testing.T) {
		got, err := testClient.GetRole(testCtx, role.ID)
		if err != nil {
			t.Fatalf("GetRole: %v", err)
		}
		if got.Name != roleName {
			t.Errorf("Name = %q, want %q", got.Name, roleName)
		}
	})

	updatedName := roleName + "-upd"
	t.Run("Update", func(t *testing.T) {
		got, err := testClient.UpdateRole(testCtx, role.ID, updatedName)
		if err != nil {
			t.Fatalf("UpdateRole: %v", err)
		}
		if got.Name != updatedName {
			t.Errorf("Name = %q, want %q", got.Name, updatedName)
		}
	})

	t.Run("ListContains", func(t *testing.T) {
		roles, err := testClient.ListRoles(testCtx)
		if err != nil {
			t.Fatalf("ListRoles: %v", err)
		}
		found := false
		for _, r := range roles {
			if r.ID == role.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("role %s not found in list", role.ID)
		}
	})

	// Test permission assignment if we have more than one permission.
	if len(perms) >= 2 {
		t.Run("AssignPermissions", func(t *testing.T) {
			got, err := testClient.AssignPermissionsToRole(testCtx, role.ID, []string{perms[1].Name})
			if err != nil {
				t.Fatalf("AssignPermissionsToRole: %v", err)
			}
			t.Logf("Role now has %d permissions", len(got.Permissions))
		})

		t.Run("UnassignPermissions", func(t *testing.T) {
			got, err := testClient.UnassignPermissionsFromRole(testCtx, role.ID, []string{perms[1].Name})
			if err != nil {
				t.Fatalf("UnassignPermissionsFromRole: %v", err)
			}
			t.Logf("Role now has %d permissions", len(got.Permissions))
		})
	}

	t.Run("Delete", func(t *testing.T) {
		err := testClient.DeleteRole(testCtx, role.ID)
		if err != nil {
			t.Fatalf("DeleteRole: %v", err)
		}
		role = nil
	})
}

// ============================================================
// Identity - SubUser CRUD
// ============================================================

func TestIntegration_Identity_SubUser_CRUD(t *testing.T) {
	// Create a role first for assignment tests.
	perms, err := testClient.ListPermissions(testCtx)
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(perms) == 0 {
		t.Skip("no permissions available")
	}

	roleName := "sdk-inttest-surole-" + randomSuffix()
	role, err := testClient.CreateRole(testCtx, roleName, []string{perms[0].Name})
	if err != nil {
		t.Fatalf("CreateRole for sub-user test: %v", err)
	}
	defer func() {
		if role != nil {
			testClient.DeleteRole(testCtx, role.ID)
		}
	}()

	var subUser *SubUser
	subUserPwd := "TestPwd-" + randomSuffix() + "!A1"

	t.Run("Create", func(t *testing.T) {
		subUser, err = testClient.CreateSubUser(testCtx, subUserPwd, []string{role.ID})
		if err != nil {
			t.Fatalf("CreateSubUser: %v", err)
		}
		if subUser.ID == "" {
			t.Fatal("sub-user ID is empty")
		}
		t.Logf("Created sub-user: %s (ID: %s)", subUser.Name, subUser.ID)
	})

	defer func() {
		if subUser != nil {
			if err := testClient.DeleteSubUser(testCtx, subUser.ID); err != nil {
				t.Logf("WARNING: cleanup DeleteSubUser failed: %v", err)
			}
		}
	}()

	if subUser == nil {
		t.Fatal("Create failed")
	}

	t.Run("Get", func(t *testing.T) {
		got, err := testClient.GetSubUser(testCtx, subUser.ID)
		if err != nil {
			t.Fatalf("GetSubUser: %v", err)
		}
		if got.ID != subUser.ID {
			t.Errorf("ID = %q, want %q", got.ID, subUser.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		newPwd := "UpdatedPwd-" + randomSuffix() + "!B2"
		got, err := testClient.UpdateSubUser(testCtx, subUser.ID, newPwd)
		if err != nil {
			t.Fatalf("UpdateSubUser: %v", err)
		}
		if got.ID != subUser.ID {
			t.Errorf("ID = %q, want %q", got.ID, subUser.ID)
		}
	})

	t.Run("ListContains", func(t *testing.T) {
		users, err := testClient.ListSubUsers(testCtx)
		if err != nil {
			t.Fatalf("ListSubUsers: %v", err)
		}
		found := false
		for _, u := range users {
			if u.ID == subUser.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("sub-user %s not found in list", subUser.ID)
		}
	})

	// Create a second role so we can test assign/unassign without removing the last role.
	role2Name := "sdk-inttest-surole2-" + randomSuffix()
	role2, err := testClient.CreateRole(testCtx, role2Name, []string{perms[0].Name})
	if err != nil {
		t.Fatalf("CreateRole (second): %v", err)
	}
	defer func() {
		if role2 != nil {
			testClient.DeleteRole(testCtx, role2.ID)
		}
	}()

	t.Run("AssignRoles", func(t *testing.T) {
		got, err := testClient.AssignRolesToSubUser(testCtx, subUser.ID, []string{role2.ID})
		if err != nil {
			t.Fatalf("AssignRolesToSubUser: %v", err)
		}
		t.Logf("Sub-user now has %d roles", len(got.Roles))
	})

	t.Run("UnassignRoles", func(t *testing.T) {
		got, err := testClient.UnassignRolesFromSubUser(testCtx, subUser.ID, []string{role2.ID})
		if err != nil {
			t.Fatalf("UnassignRolesFromSubUser: %v", err)
		}
		t.Logf("Sub-user now has %d roles", len(got.Roles))
	})

	t.Run("Delete", func(t *testing.T) {
		err := testClient.DeleteSubUser(testCtx, subUser.ID)
		if err != nil {
			t.Fatalf("DeleteSubUser: %v", err)
		}
		subUser = nil
	})
}

// ============================================================
// Compute - Flavor (read-only, provider-defined)
// ============================================================

func TestIntegration_Compute_Flavor(t *testing.T) {
	t.Run("ListFlavors", func(t *testing.T) {
		flavors, err := testClient.ListFlavors(testCtx)
		if err != nil {
			t.Fatalf("ListFlavors: %v", err)
		}
		if len(flavors) == 0 {
			t.Error("expected at least one flavor")
		}
		t.Logf("Found %d flavors", len(flavors))
	})

	t.Run("ListFlavorsDetail", func(t *testing.T) {
		flavors, err := testClient.ListFlavorsDetail(testCtx)
		if err != nil {
			t.Fatalf("ListFlavorsDetail: %v", err)
		}
		if len(flavors) == 0 {
			t.Error("expected at least one flavor")
		}
		for _, f := range flavors {
			if f.ID == "" || f.Name == "" || f.VCPUs <= 0 || f.RAM <= 0 {
				t.Errorf("invalid flavor: ID=%q Name=%q VCPUs=%d RAM=%d", f.ID, f.Name, f.VCPUs, f.RAM)
			}
		}
		t.Logf("Found %d flavors with detail", len(flavors))
	})

	t.Run("GetFlavor", func(t *testing.T) {
		flavors, err := testClient.ListFlavors(testCtx)
		if err != nil {
			t.Fatalf("ListFlavors: %v", err)
		}
		if len(flavors) == 0 {
			t.Skip("no flavors available")
		}
		detail, err := testClient.GetFlavor(testCtx, flavors[0].ID)
		if err != nil {
			t.Fatalf("GetFlavor(%s): %v", flavors[0].ID, err)
		}
		if detail.ID != flavors[0].ID {
			t.Errorf("GetFlavor ID = %q, want %q", detail.ID, flavors[0].ID)
		}
	})
}

// ============================================================
// Keypair CRUD
// ============================================================

func TestIntegration_Keypair_CRUD(t *testing.T) {
	name := "sdk-inttest-kp-" + randomSuffix()

	var created *Keypair
	t.Run("Create", func(t *testing.T) {
		var err error
		created, err = testClient.CreateKeypair(testCtx, name)
		if err != nil {
			t.Fatalf("CreateKeypair: %v", err)
		}
		if created.Name != name {
			t.Errorf("Name = %q, want %q", created.Name, name)
		}
		if created.PrivateKey == "" {
			t.Error("PrivateKey should be returned on creation")
		}
		if created.PublicKey == "" {
			t.Error("PublicKey should not be empty")
		}
		if created.Fingerprint == "" {
			t.Error("Fingerprint should not be empty")
		}
		t.Logf("Created keypair: %s (fingerprint: %s)", created.Name, created.Fingerprint)
	})

	defer func() {
		if created != nil {
			if err := testClient.DeleteKeypair(testCtx, name); err != nil {
				t.Logf("WARNING: cleanup DeleteKeypair(%s) failed: %v", name, err)
			}
		}
	}()

	if created == nil {
		t.Fatal("Create failed")
	}

	t.Run("Get", func(t *testing.T) {
		kp, err := testClient.GetKeypair(testCtx, name)
		if err != nil {
			t.Fatalf("GetKeypair: %v", err)
		}
		if kp.Name != name {
			t.Errorf("Name = %q, want %q", kp.Name, name)
		}
		if kp.Fingerprint != created.Fingerprint {
			t.Errorf("Fingerprint = %q, want %q", kp.Fingerprint, created.Fingerprint)
		}
	})

	t.Run("ListContains", func(t *testing.T) {
		keypairs, err := testClient.ListKeypairs(testCtx, nil)
		if err != nil {
			t.Fatalf("ListKeypairs: %v", err)
		}
		found := false
		for _, kp := range keypairs {
			if kp.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("created keypair %q not found in list", name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := testClient.DeleteKeypair(testCtx, name)
		if err != nil {
			t.Fatalf("DeleteKeypair: %v", err)
		}
		created = nil
	})

	t.Run("VerifyDeleted", func(t *testing.T) {
		_, err := testClient.GetKeypair(testCtx, name)
		if err == nil {
			t.Error("expected error getting deleted keypair")
		}
	})
}

// ============================================================
// Security Group CRUD (with rules)
// ============================================================

func TestIntegration_SecurityGroup_CRUD(t *testing.T) {
	sgName := "sdk-inttest-sg-" + randomSuffix()
	sgDesc := "Integration test security group"

	var sg *SecurityGroup
	t.Run("Create", func(t *testing.T) {
		var err error
		sg, err = testClient.CreateSecurityGroup(testCtx, sgName, sgDesc)
		if err != nil {
			t.Fatalf("CreateSecurityGroup: %v", err)
		}
		if sg.ID == "" {
			t.Fatal("security group ID is empty")
		}
		if sg.Name != sgName {
			t.Errorf("Name = %q, want %q", sg.Name, sgName)
		}
		t.Logf("Created security group: %s (ID: %s)", sg.Name, sg.ID)
	})

	defer func() {
		if sg != nil {
			if err := testClient.DeleteSecurityGroup(testCtx, sg.ID); err != nil {
				t.Logf("WARNING: cleanup DeleteSecurityGroup(%s) failed: %v", sg.ID, err)
			}
		}
	}()

	if sg == nil {
		t.Fatal("Create failed")
	}

	t.Run("Get", func(t *testing.T) {
		got, err := testClient.GetSecurityGroup(testCtx, sg.ID)
		if err != nil {
			t.Fatalf("GetSecurityGroup: %v", err)
		}
		if got.Name != sgName {
			t.Errorf("Name = %q, want %q", got.Name, sgName)
		}
	})

	t.Run("Update", func(t *testing.T) {
		updatedName := sgName + "-upd"
		got, err := testClient.UpdateSecurityGroup(testCtx, sg.ID, updatedName, sgDesc)
		if err != nil {
			t.Fatalf("UpdateSecurityGroup: %v", err)
		}
		if got.Name != updatedName {
			t.Errorf("Name = %q, want %q", got.Name, updatedName)
		}
	})

	t.Run("ListContains", func(t *testing.T) {
		sgs, err := testClient.ListSecurityGroups(testCtx, nil)
		if err != nil {
			t.Fatalf("ListSecurityGroups: %v", err)
		}
		found := false
		for _, s := range sgs {
			if s.ID == sg.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("created security group %s not found in list", sg.ID)
		}
	})

	var ruleID string
	t.Run("CreateRule", func(t *testing.T) {
		proto := "tcp"
		portMin := 8080
		portMax := 8080
		rule, err := testClient.CreateSecurityGroupRule(testCtx, CreateSecurityGroupRuleRequest{
			SecurityGroupID: sg.ID,
			Direction:       "ingress",
			EtherType:       "IPv4",
			Protocol:        &proto,
			PortRangeMin:    &portMin,
			PortRangeMax:    &portMax,
		})
		if err != nil {
			t.Fatalf("CreateSecurityGroupRule: %v", err)
		}
		if rule.ID == "" {
			t.Fatal("rule ID is empty")
		}
		ruleID = rule.ID
		t.Logf("Created rule: %s", ruleID)
	})

	t.Run("GetRule", func(t *testing.T) {
		if ruleID == "" {
			t.Skip("no rule created")
		}
		rule, err := testClient.GetSecurityGroupRule(testCtx, ruleID)
		if err != nil {
			t.Fatalf("GetSecurityGroupRule: %v", err)
		}
		if rule.SecurityGroupID != sg.ID {
			t.Errorf("SecurityGroupID = %q, want %q", rule.SecurityGroupID, sg.ID)
		}
	})

	t.Run("ListRules", func(t *testing.T) {
		rules, err := testClient.ListSecurityGroupRules(testCtx, &ListSecurityGroupRulesOptions{
			SecurityGroupID: sg.ID,
		})
		if err != nil {
			t.Fatalf("ListSecurityGroupRules: %v", err)
		}
		t.Logf("Found %d rules for security group", len(rules))
	})

	t.Run("DeleteRule", func(t *testing.T) {
		if ruleID == "" {
			t.Skip("no rule created")
		}
		err := testClient.DeleteSecurityGroupRule(testCtx, ruleID)
		if err != nil {
			t.Fatalf("DeleteSecurityGroupRule: %v", err)
		}
	})

	t.Run("DeleteSecurityGroup", func(t *testing.T) {
		err := testClient.DeleteSecurityGroup(testCtx, sg.ID)
		if err != nil {
			t.Fatalf("DeleteSecurityGroup: %v", err)
		}
		sg = nil
	})
}

// ============================================================
// Network CRUD (network + subnet + port)
// ============================================================

func TestIntegration_Network_CRUD(t *testing.T) {
	// Create network.
	var net *Network
	t.Run("CreateNetwork", func(t *testing.T) {
		var err error
		net, err = testClient.CreateNetwork(testCtx)
		if err != nil {
			t.Fatalf("CreateNetwork: %v", err)
		}
		if net.ID == "" {
			t.Fatal("network ID is empty")
		}
		t.Logf("Created network: %s", net.ID)
	})

	defer func() {
		if net != nil {
			if err := testClient.DeleteNetwork(testCtx, net.ID); err != nil {
				t.Logf("WARNING: cleanup DeleteNetwork(%s) failed: %v", net.ID, err)
			}
		}
	}()

	if net == nil {
		t.Fatal("CreateNetwork failed")
	}

	t.Run("GetNetwork", func(t *testing.T) {
		got, err := testClient.GetNetwork(testCtx, net.ID)
		if err != nil {
			t.Fatalf("GetNetwork: %v", err)
		}
		if got.ID != net.ID {
			t.Errorf("ID = %q, want %q", got.ID, net.ID)
		}
	})

	t.Run("ListNetworksContains", func(t *testing.T) {
		networks, err := testClient.ListNetworks(testCtx, nil)
		if err != nil {
			t.Fatalf("ListNetworks: %v", err)
		}
		found := false
		for _, n := range networks {
			if n.ID == net.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("network %s not found in list", net.ID)
		}
	})

	// Create subnet on the network.
	var subnet *Subnet
	t.Run("CreateSubnet", func(t *testing.T) {
		var err error
		subnet, err = testClient.CreateSubnet(testCtx, net.ID, "10.0.0.0/24")
		if err != nil {
			t.Fatalf("CreateSubnet: %v", err)
		}
		if subnet.ID == "" {
			t.Fatal("subnet ID is empty")
		}
		if subnet.CIDR != "10.0.0.0/24" {
			t.Errorf("CIDR = %q, want %q", subnet.CIDR, "10.0.0.0/24")
		}
		t.Logf("Created subnet: %s (CIDR: %s)", subnet.ID, subnet.CIDR)
	})

	defer func() {
		if subnet != nil {
			if err := testClient.DeleteSubnet(testCtx, subnet.ID); err != nil {
				t.Logf("WARNING: cleanup DeleteSubnet(%s) failed: %v", subnet.ID, err)
			}
		}
	}()

	if subnet == nil {
		t.Fatal("CreateSubnet failed")
	}

	t.Run("GetSubnet", func(t *testing.T) {
		got, err := testClient.GetSubnet(testCtx, subnet.ID)
		if err != nil {
			t.Fatalf("GetSubnet: %v", err)
		}
		if got.NetworkID != net.ID {
			t.Errorf("NetworkID = %q, want %q", got.NetworkID, net.ID)
		}
	})

	t.Run("ListSubnetsContains", func(t *testing.T) {
		subnets, err := testClient.ListSubnets(testCtx, nil)
		if err != nil {
			t.Fatalf("ListSubnets: %v", err)
		}
		found := false
		for _, s := range subnets {
			if s.ID == subnet.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subnet %s not found in list", subnet.ID)
		}
	})

	// Create port on the network.
	var port *Port
	t.Run("CreatePort", func(t *testing.T) {
		var err error
		port, err = testClient.CreatePort(testCtx, CreatePortRequest{
			NetworkID: net.ID,
		})
		if err != nil {
			t.Fatalf("CreatePort: %v", err)
		}
		if port.ID == "" {
			t.Fatal("port ID is empty")
		}
		t.Logf("Created port: %s", port.ID)
	})

	defer func() {
		if port != nil {
			if err := testClient.DeletePort(testCtx, port.ID); err != nil {
				t.Logf("WARNING: cleanup DeletePort(%s) failed: %v", port.ID, err)
			}
		}
	}()

	if port == nil {
		t.Fatal("CreatePort failed")
	}

	t.Run("GetPort", func(t *testing.T) {
		got, err := testClient.GetPort(testCtx, port.ID)
		if err != nil {
			t.Fatalf("GetPort: %v", err)
		}
		if got.ID != port.ID {
			t.Errorf("ID = %q, want %q", got.ID, port.ID)
		}
	})

	t.Run("UpdatePort", func(t *testing.T) {
		// Re-set the port's current security groups as a no-op update.
		current, err := testClient.GetPort(testCtx, port.ID)
		if err != nil {
			t.Fatalf("GetPort (before update): %v", err)
		}
		got, err := testClient.UpdatePort(testCtx, port.ID, UpdatePortRequest{
			SecurityGroups: current.SecurityGroups,
		})
		if err != nil {
			t.Fatalf("UpdatePort: %v", err)
		}
		if got.ID != port.ID {
			t.Errorf("ID = %q, want %q", got.ID, port.ID)
		}
	})

	t.Run("ListPortsContains", func(t *testing.T) {
		ports, err := testClient.ListPorts(testCtx, nil)
		if err != nil {
			t.Fatalf("ListPorts: %v", err)
		}
		found := false
		for _, p := range ports {
			if p.ID == port.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("port %s not found in list", port.ID)
		}
	})

	t.Run("DeletePort", func(t *testing.T) {
		err := testClient.DeletePort(testCtx, port.ID)
		if err != nil {
			t.Fatalf("DeletePort: %v", err)
		}
		port = nil
	})

	t.Run("DeleteSubnet", func(t *testing.T) {
		err := testClient.DeleteSubnet(testCtx, subnet.ID)
		if err != nil {
			t.Fatalf("DeleteSubnet: %v", err)
		}
		subnet = nil
	})

	t.Run("DeleteNetwork", func(t *testing.T) {
		err := testClient.DeleteNetwork(testCtx, net.ID)
		if err != nil {
			t.Fatalf("DeleteNetwork: %v", err)
		}
		net = nil
	})
}

// ============================================================
// QoS Policy (read-only, provider-defined)
// ============================================================

func TestIntegration_QoSPolicy(t *testing.T) {
	t.Run("ListQoSPolicies", func(t *testing.T) {
		policies, err := testClient.ListQoSPolicies(testCtx, nil)
		if err != nil {
			t.Fatalf("ListQoSPolicies: %v", err)
		}
		t.Logf("Found %d QoS policies", len(policies))

		if len(policies) > 0 {
			got, err := testClient.GetQoSPolicy(testCtx, policies[0].ID)
			if err != nil {
				t.Fatalf("GetQoSPolicy: %v", err)
			}
			if got.ID != policies[0].ID {
				t.Errorf("ID = %q, want %q", got.ID, policies[0].ID)
			}
		}
	})
}

// ============================================================
// Volume CRUD
// ============================================================

func TestIntegration_Volume_CRUD(t *testing.T) {
	// Find a suitable volume type.
	volTypes, err := testClient.ListVolumeTypes(testCtx)
	if err != nil {
		t.Fatalf("ListVolumeTypes: %v", err)
	}
	if len(volTypes) == 0 {
		t.Fatal("no volume types available")
	}
	// Prefer a non-boot volume type for simple test.
	var volTypeName string
	for _, vt := range volTypes {
		if !strings.Contains(vt.Name, "boot") {
			volTypeName = vt.Name
			break
		}
	}
	if volTypeName == "" {
		volTypeName = volTypes[0].Name
	}
	t.Logf("Using volume type: %s", volTypeName)

	volName := "sdk-inttest-vol-" + randomSuffix()
	var vol *Volume

	t.Run("Create", func(t *testing.T) {
		vol, err = testClient.CreateVolume(testCtx, CreateVolumeRequest{
			Size:       200,
			Name:       volName,
			VolumeType: volTypeName,
		})
		if err != nil {
			t.Fatalf("CreateVolume: %v", err)
		}
		if vol.ID == "" {
			t.Fatal("volume ID is empty")
		}
		t.Logf("Created volume: %s (ID: %s)", vol.Name, vol.ID)
	})

	defer func() {
		if vol != nil {
			testClient.DeleteVolume(testCtx, vol.ID, true)
		}
	}()

	if vol == nil {
		t.Fatal("Create failed")
	}

	// Wait for volume to be available.
	waitForVolumeStatus(t, vol.ID, "available", 3*time.Minute)

	t.Run("Get", func(t *testing.T) {
		got, err := testClient.GetVolume(testCtx, vol.ID)
		if err != nil {
			t.Fatalf("GetVolume: %v", err)
		}
		if got.ID != vol.ID {
			t.Errorf("ID = %q, want %q", got.ID, vol.ID)
		}
		if got.Name != volName {
			t.Errorf("Name = %q, want %q", got.Name, volName)
		}
	})

	updatedName := volName + "-upd"
	t.Run("Update", func(t *testing.T) {
		got, err := testClient.UpdateVolume(testCtx, vol.ID, updatedName, nil)
		if err != nil {
			t.Fatalf("UpdateVolume: %v", err)
		}
		if got.Name != updatedName {
			t.Errorf("Name = %q, want %q", got.Name, updatedName)
		}
	})

	t.Run("ListContains", func(t *testing.T) {
		volumes, err := testClient.ListVolumesDetail(testCtx, nil)
		if err != nil {
			t.Fatalf("ListVolumesDetail: %v", err)
		}
		found := false
		for _, v := range volumes {
			if v.ID == vol.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("volume %s not found in list", vol.ID)
		}
	})

	t.Run("GetVolumeType", func(t *testing.T) {
		for _, vt := range volTypes {
			got, err := testClient.GetVolumeType(testCtx, vt.ID)
			if err != nil {
				t.Fatalf("GetVolumeType(%s): %v", vt.ID, err)
			}
			if got.ID != vt.ID {
				t.Errorf("ID = %q, want %q", got.ID, vt.ID)
			}
			break // Just test one.
		}
	})

	t.Run("ListBackups", func(t *testing.T) {
		backups, err := testClient.ListBackups(testCtx, nil)
		if err != nil {
			t.Fatalf("ListBackups: %v", err)
		}
		t.Logf("Found %d backups", len(backups))
	})

	t.Run("Delete", func(t *testing.T) {
		err := testClient.DeleteVolume(testCtx, vol.ID, false)
		if err != nil {
			t.Fatalf("DeleteVolume: %v", err)
		}
		vol = nil
	})
}

// ============================================================
// Image Operations
// ============================================================

func TestIntegration_Image_Operations(t *testing.T) {
	t.Run("ListImages", func(t *testing.T) {
		images, err := testClient.ListImages(testCtx, &ListImagesOptions{
			Visibility: "public",
			Limit:      5,
		})
		if err != nil {
			t.Fatalf("ListImages: %v", err)
		}
		if len(images) == 0 {
			t.Error("expected at least one public image")
		}
		t.Logf("Found %d public images (limited to 5)", len(images))
	})

	t.Run("GetImage", func(t *testing.T) {
		images, err := testClient.ListImages(testCtx, &ListImagesOptions{
			Visibility: "public",
			Limit:      1,
		})
		if err != nil {
			t.Fatalf("ListImages: %v", err)
		}
		if len(images) == 0 {
			t.Skip("no images available")
		}
		img, err := testClient.GetImage(testCtx, images[0].ID)
		if err != nil {
			t.Fatalf("GetImage: %v", err)
		}
		if img.ID != images[0].ID {
			t.Errorf("ID = %q, want %q", img.ID, images[0].ID)
		}
	})

	t.Run("GetImageQuota", func(t *testing.T) {
		quota, err := testClient.GetImageQuota(testCtx)
		if err != nil {
			t.Fatalf("GetImageQuota: %v", err)
		}
		t.Logf("Image quota: %s", quota.ImageSize)
	})

	t.Run("GetImageUsage", func(t *testing.T) {
		usage, err := testClient.GetImageUsage(testCtx)
		if err != nil {
			t.Fatalf("GetImageUsage: %v", err)
		}
		t.Logf("Image usage: %d bytes", usage.Size)
	})
}

// ============================================================
// Server Lifecycle (full CRUD + actions)
// ============================================================

func TestIntegration_Server_Lifecycle(t *testing.T) {
	// 1. Find the smallest flavor.
	flavors, err := testClient.ListFlavorsDetail(testCtx)
	if err != nil {
		t.Fatalf("ListFlavorsDetail: %v", err)
	}
	if len(flavors) == 0 {
		t.Fatal("no flavors available")
	}
	// Use the g2l-t-c2m1 flavor (2 vCPUs, 1GB RAM) which is suitable for VPS.
	var smallest *FlavorDetail
	for i, f := range flavors {
		if f.Name == "g2l-t-c2m1" {
			smallest = &flavors[i]
			break
		}
	}
	if smallest == nil {
		t.Fatal("flavor g2l-t-c2m1 not found")
	}
	t.Logf("Using flavor: %s (%d vCPUs, %d MB RAM)", smallest.Name, smallest.VCPUs, smallest.RAM)

	// 2. Find a Linux boot image.
	images, err := testClient.ListImages(testCtx, &ListImagesOptions{
		Visibility: "public",
		OSType:     "linux",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	if len(images) == 0 {
		t.Fatal("no linux images available")
	}
	imageRef := images[0].ID
	t.Logf("Using image: %s (ID: %s)", images[0].Name, imageRef)

	// 3. Find a boot volume type.
	volTypes, err := testClient.ListVolumeTypes(testCtx)
	if err != nil {
		t.Fatalf("ListVolumeTypes: %v", err)
	}
	var bootVolType string
	for _, vt := range volTypes {
		if strings.Contains(vt.Name, "boot") {
			bootVolType = vt.Name
			break
		}
	}
	if bootVolType == "" {
		t.Fatal("no boot volume type found")
	}
	t.Logf("Using volume type: %s", bootVolType)

	// 4. Create boot volume.
	volName := "sdk-inttest-boot-" + randomSuffix()
	vol, err := testClient.CreateVolume(testCtx, CreateVolumeRequest{
		Size:       100,
		Name:       volName,
		VolumeType: bootVolType,
		ImageRef:   imageRef,
	})
	if err != nil {
		t.Fatalf("CreateVolume: %v", err)
	}
	t.Logf("Created boot volume: %s (ID: %s)", vol.Name, vol.ID)

	defer func() {
		if vol != nil {
			// Wait a bit before trying to delete.
			time.Sleep(5 * time.Second)
			if err := testClient.DeleteVolume(testCtx, vol.ID, true); err != nil {
				t.Logf("WARNING: cleanup DeleteVolume(%s) failed: %v", vol.ID, err)
			}
		}
	}()

	waitForVolumeStatus(t, vol.ID, "available", 5*time.Minute)

	// 5. Create keypair for the server.
	kpName := "sdk-inttest-svkp-" + randomSuffix()
	kp, err := testClient.CreateKeypair(testCtx, kpName)
	if err != nil {
		t.Fatalf("CreateKeypair: %v", err)
	}
	t.Logf("Created keypair: %s", kp.Name)
	defer testClient.DeleteKeypair(testCtx, kpName)

	// 6. Create server.
	var serverID string
	adminPass := "IntTest-" + randomSuffix() + "!Aa1"

	t.Run("CreateServer", func(t *testing.T) {
		resp, err := testClient.CreateServer(testCtx, CreateServerRequest{
			FlavorRef: smallest.ID,
			AdminPass: adminPass,
			BlockDeviceMapping: []BlockDeviceMap{
				{UUID: vol.ID},
			},
			Metadata: map[string]string{
				"instance_name_tag": "sdk-inttest-server",
			},
			KeyName: kpName,
		})
		if err != nil {
			t.Fatalf("CreateServer: %v", err)
		}
		if resp.ID == "" {
			t.Fatal("server ID is empty")
		}
		serverID = resp.ID
		t.Logf("Created server: %s", serverID)
	})

	if serverID == "" {
		t.Fatal("CreateServer failed")
	}

	defer func() {
		if serverID != "" {
			testClient.DeleteServer(testCtx, serverID)
			// Wait for deletion so volume can be cleaned up.
			deadline := time.Now().Add(3 * time.Minute)
			for time.Now().Before(deadline) {
				_, err := testClient.GetServer(testCtx, serverID)
				if err != nil {
					break
				}
				time.Sleep(5 * time.Second)
			}
		}
	}()

	// Wait for server to be ACTIVE.
	waitForServerStatus(t, serverID, "ACTIVE", 5*time.Minute)

	t.Run("GetServer", func(t *testing.T) {
		s, err := testClient.GetServer(testCtx, serverID)
		if err != nil {
			t.Fatalf("GetServer: %v", err)
		}
		if s.Status != "ACTIVE" {
			t.Errorf("Status = %q, want %q", s.Status, "ACTIVE")
		}
		t.Logf("Server %s is ACTIVE", serverID)
	})

	t.Run("ListServersContains", func(t *testing.T) {
		servers, err := testClient.ListServersDetail(testCtx, nil)
		if err != nil {
			t.Fatalf("ListServersDetail: %v", err)
		}
		found := false
		for _, s := range servers {
			if s.ID == serverID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("server %s not found in list", serverID)
		}
	})

	t.Run("GetServerAddresses", func(t *testing.T) {
		addrs, err := testClient.GetServerAddresses(testCtx, serverID)
		if err != nil {
			t.Fatalf("GetServerAddresses: %v", err)
		}
		t.Logf("Server has %d network(s)", len(addrs))
	})

	t.Run("GetServerSecurityGroups", func(t *testing.T) {
		sgs, err := testClient.GetServerSecurityGroups(testCtx, serverID)
		if err != nil {
			t.Fatalf("GetServerSecurityGroups: %v", err)
		}
		t.Logf("Server has %d security group(s)", len(sgs))
	})

	t.Run("GetServerMetadata", func(t *testing.T) {
		meta, err := testClient.GetServerMetadata(testCtx, serverID)
		if err != nil {
			t.Fatalf("GetServerMetadata: %v", err)
		}
		if meta["instance_name_tag"] != "sdk-inttest-server" {
			t.Errorf("instance_name_tag = %q, want %q", meta["instance_name_tag"], "sdk-inttest-server")
		}
	})

	t.Run("UpdateServerMetadata", func(t *testing.T) {
		meta, err := testClient.UpdateServerMetadata(testCtx, serverID, map[string]string{
			"instance_name_tag": "sdk-inttest-updated",
		})
		if err != nil {
			t.Fatalf("UpdateServerMetadata: %v", err)
		}
		if meta["instance_name_tag"] != "sdk-inttest-updated" {
			t.Errorf("instance_name_tag = %q, want %q", meta["instance_name_tag"], "sdk-inttest-updated")
		}
	})

	t.Run("GetVNCConsoleURL", func(t *testing.T) {
		url, err := testClient.GetVNCConsoleURL(testCtx, serverID)
		if err != nil {
			t.Fatalf("GetVNCConsoleURL: %v", err)
		}
		if url == "" {
			t.Error("VNC console URL is empty")
		}
		t.Logf("VNC URL: %s", url)
	})

	t.Run("ListServerInterfaces", func(t *testing.T) {
		ifaces, err := testClient.ListServerInterfaces(testCtx, serverID)
		if err != nil {
			t.Fatalf("ListServerInterfaces: %v", err)
		}
		t.Logf("Server has %d interface(s)", len(ifaces))
	})

	t.Run("ListServerVolumes", func(t *testing.T) {
		vols, err := testClient.ListServerVolumes(testCtx, serverID)
		if err != nil {
			t.Fatalf("ListServerVolumes: %v", err)
		}
		if len(vols) == 0 {
			t.Error("expected at least one volume attached")
		}
		t.Logf("Server has %d volume(s) attached", len(vols))
	})

	// Server actions: Stop -> wait SHUTOFF -> Start -> wait ACTIVE.
	t.Run("StopServer", func(t *testing.T) {
		err := testClient.StopServer(testCtx, serverID)
		if err != nil {
			t.Fatalf("StopServer: %v", err)
		}
		waitForServerStatus(t, serverID, "SHUTOFF", 3*time.Minute)
	})

	t.Run("StartServer", func(t *testing.T) {
		err := testClient.StartServer(testCtx, serverID)
		if err != nil {
			t.Fatalf("StartServer: %v", err)
		}
		waitForServerStatus(t, serverID, "ACTIVE", 3*time.Minute)
	})

	t.Run("RebootServer", func(t *testing.T) {
		err := testClient.RebootServer(testCtx, serverID)
		if err != nil {
			t.Fatalf("RebootServer: %v", err)
		}
		waitForServerStatus(t, serverID, "ACTIVE", 3*time.Minute)
	})

	// Delete server.
	t.Run("DeleteServer", func(t *testing.T) {
		err := testClient.DeleteServer(testCtx, serverID)
		if err != nil {
			t.Fatalf("DeleteServer: %v", err)
		}
		// Wait for server to be gone so volume can be freed.
		deadline := time.Now().Add(3 * time.Minute)
		for time.Now().Before(deadline) {
			_, err := testClient.GetServer(testCtx, serverID)
			if err != nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
		serverID = ""
		t.Log("Server deleted")
	})

	// Delete volume (volume should become available after server deletion).
	t.Run("DeleteVolume", func(t *testing.T) {
		// Wait for volume to detach and become available.
		waitForVolumeStatus(t, vol.ID, "available", 3*time.Minute)
		err := testClient.DeleteVolume(testCtx, vol.ID, false)
		if err != nil {
			t.Fatalf("DeleteVolume: %v", err)
		}
		vol = nil
		t.Log("Volume deleted")
	})
}

// ============================================================
// Object Storage CRUD
// ============================================================

func TestIntegration_ObjectStorage_CRUD(t *testing.T) {
	containerName := "sdk-inttest-ctr-" + randomSuffix()
	objectName := "test-object.txt"
	objectData := []byte("Hello, ConoHa integration test!")

	t.Run("GetAccountInfo", func(t *testing.T) {
		info, err := testClient.GetAccountInfo(testCtx)
		if err != nil {
			t.Fatalf("GetAccountInfo: %v", err)
		}
		t.Logf("Account: %d containers, %d objects, %d bytes",
			info.ContainerCount, info.ObjectCount, info.BytesUsed)
	})

	t.Run("CreateContainer", func(t *testing.T) {
		err := testClient.CreateContainer(testCtx, containerName)
		if err != nil {
			t.Fatalf("CreateContainer: %v", err)
		}
		t.Logf("Created container: %s", containerName)
	})

	defer func() {
		_ = testClient.DeleteObject(testCtx, containerName, objectName)
		_ = testClient.DeleteObject(testCtx, containerName, "test-object-copy.txt")
		if err := testClient.DeleteContainer(testCtx, containerName); err != nil {
			t.Logf("WARNING: cleanup DeleteContainer(%s) failed: %v", containerName, err)
		}
	}()

	t.Run("ListContainersContains", func(t *testing.T) {
		containers, err := testClient.ListContainers(testCtx)
		if err != nil {
			t.Fatalf("ListContainers: %v", err)
		}
		found := false
		for _, c := range containers {
			if c.Name == containerName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("container %q not found in list", containerName)
		}
	})

	t.Run("UploadObject", func(t *testing.T) {
		err := testClient.UploadObject(testCtx, containerName, objectName, bytes.NewReader(objectData))
		if err != nil {
			t.Fatalf("UploadObject: %v", err)
		}
	})

	t.Run("ListObjects", func(t *testing.T) {
		objects, err := testClient.ListObjects(testCtx, containerName, nil)
		if err != nil {
			t.Fatalf("ListObjects: %v", err)
		}
		found := false
		for _, o := range objects {
			if o.Name == objectName {
				found = true
				if o.Bytes != int64(len(objectData)) {
					t.Errorf("object size = %d, want %d", o.Bytes, len(objectData))
				}
				break
			}
		}
		if !found {
			t.Errorf("object %q not found in list", objectName)
		}
	})

	t.Run("DownloadObject", func(t *testing.T) {
		reader, err := testClient.DownloadObject(testCtx, containerName, objectName)
		if err != nil {
			t.Fatalf("DownloadObject: %v", err)
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("reading downloaded object: %v", err)
		}
		if !bytes.Equal(data, objectData) {
			t.Errorf("downloaded data = %q, want %q", string(data), string(objectData))
		}
	})

	t.Run("CopyObject", func(t *testing.T) {
		copiedName := "test-object-copy.txt"
		err := testClient.CopyObject(testCtx, containerName, objectName, containerName, copiedName)
		if err != nil {
			t.Fatalf("CopyObject: %v", err)
		}
		defer testClient.DeleteObject(testCtx, containerName, copiedName)

		reader, err := testClient.DownloadObject(testCtx, containerName, copiedName)
		if err != nil {
			t.Fatalf("DownloadObject(copy): %v", err)
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("reading copied object: %v", err)
		}
		if !bytes.Equal(data, objectData) {
			t.Errorf("copied data = %q, want %q", string(data), string(objectData))
		}
	})

	t.Run("ScheduleObjectDeletion", func(t *testing.T) {
		// Schedule deletion far in the future (we'll delete it manually before then).
		deleteAt := time.Now().Add(24 * time.Hour).Unix()
		err := testClient.ScheduleObjectDeletion(testCtx, containerName, objectName, deleteAt)
		if err != nil {
			t.Fatalf("ScheduleObjectDeletion: %v", err)
		}
	})

	t.Run("DeleteObject", func(t *testing.T) {
		err := testClient.DeleteObject(testCtx, containerName, objectName)
		if err != nil {
			t.Fatalf("DeleteObject: %v", err)
		}
	})

	t.Run("DeleteContainer", func(t *testing.T) {
		err := testClient.DeleteContainer(testCtx, containerName)
		if err != nil {
			t.Fatalf("DeleteContainer: %v", err)
		}
	})
}

// ============================================================
// Load Balancer CRUD (LB + Listener + Pool + Member + HealthMonitor)
// ============================================================

func TestIntegration_LoadBalancer_CRUD(t *testing.T) {
	suffix := randomSuffix()

	// Declare all resource pointers upfront so the defer cleanup can reference them.
	var lb *LoadBalancer
	var listener *Listener
	var pool *Pool
	var hm *HealthMonitor

	// 1. Create Load Balancer.
	t.Run("CreateLB", func(t *testing.T) {
		var err error
		lb, err = testClient.CreateLoadBalancer(testCtx, "sdk-inttest-lb-"+suffix)
		if err != nil {
			t.Fatalf("CreateLoadBalancer: %v", err)
		}
		if lb.ID == "" {
			t.Fatal("LB ID is empty")
		}
		t.Logf("Created LB: %s (ID: %s)", lb.Name, lb.ID)
	})

	defer func() {
		if lb == nil {
			return
		}
		// Delete children first (health monitors, members, pools, listeners)
		// before deleting the LB itself.
		if hm != nil {
			waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)
			if err := testClient.DeleteHealthMonitor(testCtx, hm.ID); err != nil {
				t.Logf("WARNING: cleanup DeleteHealthMonitor(%s) failed: %v", hm.ID, err)
			}
		}
		if pool != nil {
			waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)
			if err := testClient.DeletePool(testCtx, pool.ID); err != nil {
				t.Logf("WARNING: cleanup DeletePool(%s) failed: %v", pool.ID, err)
			}
		}
		if listener != nil {
			waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)
			if err := testClient.DeleteListener(testCtx, listener.ID); err != nil {
				t.Logf("WARNING: cleanup DeleteListener(%s) failed: %v", listener.ID, err)
			}
		}
		waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)
		if err := testClient.DeleteLoadBalancer(testCtx, lb.ID); err != nil {
			t.Logf("WARNING: cleanup DeleteLoadBalancer(%s) failed: %v", lb.ID, err)
		}
	}()

	if lb == nil {
		t.Fatal("CreateLB failed")
	}

	waitForLBStatus(t, lb.ID, "ACTIVE", 5*time.Minute)

	t.Run("GetLB", func(t *testing.T) {
		got, err := testClient.GetLoadBalancer(testCtx, lb.ID)
		if err != nil {
			t.Fatalf("GetLoadBalancer: %v", err)
		}
		if got.ID != lb.ID {
			t.Errorf("ID = %q, want %q", got.ID, lb.ID)
		}
	})

	t.Run("UpdateLB", func(t *testing.T) {
		newName := "sdk-inttest-lb-" + suffix + "-upd"
		got, err := testClient.UpdateLoadBalancer(testCtx, lb.ID, newName)
		if err != nil {
			t.Fatalf("UpdateLoadBalancer: %v", err)
		}
		if got.Name != newName {
			t.Errorf("Name = %q, want %q", got.Name, newName)
		}
	})

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("ListLBContains", func(t *testing.T) {
		lbs, err := testClient.ListLoadBalancers(testCtx)
		if err != nil {
			t.Fatalf("ListLoadBalancers: %v", err)
		}
		found := false
		for _, l := range lbs {
			if l.ID == lb.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("LB %s not found in list", lb.ID)
		}
	})

	// 2. Create Listener.
	t.Run("CreateListener", func(t *testing.T) {
		var err error
		listener, err = testClient.CreateListener(testCtx, "sdk-inttest-ls-"+suffix, "TCP", 80, lb.ID)
		if err != nil {
			t.Fatalf("CreateListener: %v", err)
		}
		if listener.ID == "" {
			t.Fatal("listener ID is empty")
		}
		t.Logf("Created listener: %s (ID: %s)", listener.Name, listener.ID)
	})

	if listener == nil {
		t.Fatal("CreateListener failed")
	}

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("GetListener", func(t *testing.T) {
		got, err := testClient.GetListener(testCtx, listener.ID)
		if err != nil {
			t.Fatalf("GetListener: %v", err)
		}
		if got.Protocol != "TCP" {
			t.Errorf("Protocol = %q, want %q", got.Protocol, "TCP")
		}
	})

	t.Run("UpdateListener", func(t *testing.T) {
		newName := "sdk-inttest-ls-" + suffix + "-upd"
		got, err := testClient.UpdateListener(testCtx, listener.ID, newName)
		if err != nil {
			t.Fatalf("UpdateListener: %v", err)
		}
		if got.Name != newName {
			t.Errorf("Name = %q, want %q", got.Name, newName)
		}
	})

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("ListListenersContains", func(t *testing.T) {
		listeners, err := testClient.ListListeners(testCtx)
		if err != nil {
			t.Fatalf("ListListeners: %v", err)
		}
		found := false
		for _, l := range listeners {
			if l.ID == listener.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("listener %s not found in list", listener.ID)
		}
	})

	// 3. Create Pool.
	t.Run("CreatePool", func(t *testing.T) {
		var err error
		pool, err = testClient.CreatePool(testCtx, "sdk-inttest-pool-"+suffix, "TCP", "ROUND_ROBIN", listener.ID)
		if err != nil {
			t.Fatalf("CreatePool: %v", err)
		}
		if pool.ID == "" {
			t.Fatal("pool ID is empty")
		}
		t.Logf("Created pool: %s (ID: %s)", pool.Name, pool.ID)
	})

	if pool == nil {
		t.Fatal("CreatePool failed")
	}

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("GetPool", func(t *testing.T) {
		got, err := testClient.GetPool(testCtx, pool.ID)
		if err != nil {
			t.Fatalf("GetPool: %v", err)
		}
		if got.LBAlgorithm != "ROUND_ROBIN" {
			t.Errorf("LBAlgorithm = %q, want %q", got.LBAlgorithm, "ROUND_ROBIN")
		}
	})

	t.Run("UpdatePool", func(t *testing.T) {
		newName := "sdk-inttest-pool-" + suffix + "-upd"
		got, err := testClient.UpdatePool(testCtx, pool.ID, newName, "")
		if err != nil {
			t.Fatalf("UpdatePool: %v", err)
		}
		if got.Name != newName {
			t.Errorf("Name = %q, want %q", got.Name, newName)
		}
	})

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("ListPoolsContains", func(t *testing.T) {
		pools, err := testClient.ListPools(testCtx)
		if err != nil {
			t.Fatalf("ListPools: %v", err)
		}
		found := false
		for _, p := range pools {
			if p.ID == pool.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("pool %s not found in list", pool.ID)
		}
	})

	// 4. Member operations - skip AddMember/Get/Update/Delete because
	// ConoHa requires the member IP to be a project-owned public IP.
	// We verify ListMembers works (returns empty list).
	t.Run("ListMembers", func(t *testing.T) {
		members, err := testClient.ListMembers(testCtx, pool.ID)
		if err != nil {
			t.Fatalf("ListMembers: %v", err)
		}
		t.Logf("Pool has %d members", len(members))
	})

	// 5. Create Health Monitor.
	t.Run("CreateHealthMonitor", func(t *testing.T) {
		var err error
		hm, err = testClient.CreateHealthMonitor(testCtx, CreateHealthMonitorRequest{
			Name:       "sdk-inttest-hm-" + suffix,
			PoolID:     pool.ID,
			Delay:      10,
			MaxRetries: 3,
			Timeout:    5,
			Type:       "TCP",
		})
		if err != nil {
			t.Fatalf("CreateHealthMonitor: %v", err)
		}
		if hm.ID == "" {
			t.Fatal("health monitor ID is empty")
		}
		t.Logf("Created health monitor: %s (ID: %s)", hm.Name, hm.ID)
	})

	if hm == nil {
		t.Fatal("CreateHealthMonitor failed")
	}

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("GetHealthMonitor", func(t *testing.T) {
		got, err := testClient.GetHealthMonitor(testCtx, hm.ID)
		if err != nil {
			t.Fatalf("GetHealthMonitor: %v", err)
		}
		if got.Type != "TCP" {
			t.Errorf("Type = %q, want %q", got.Type, "TCP")
		}
	})

	t.Run("UpdateHealthMonitor", func(t *testing.T) {
		newName := "sdk-inttest-hm-" + suffix + "-upd"
		got, err := testClient.UpdateHealthMonitor(testCtx, hm.ID, newName)
		if err != nil {
			t.Fatalf("UpdateHealthMonitor: %v", err)
		}
		if got.Name != newName {
			t.Errorf("Name = %q, want %q", got.Name, newName)
		}
	})

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("ListHealthMonitorsContains", func(t *testing.T) {
		hms, err := testClient.ListHealthMonitors(testCtx)
		if err != nil {
			t.Fatalf("ListHealthMonitors: %v", err)
		}
		found := false
		for _, h := range hms {
			if h.ID == hm.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("health monitor %s not found in list", hm.ID)
		}
	})

	// Teardown in reverse order: HM -> Member -> Pool -> Listener -> LB.
	t.Run("DeleteHealthMonitor", func(t *testing.T) {
		err := testClient.DeleteHealthMonitor(testCtx, hm.ID)
		if err != nil {
			t.Fatalf("DeleteHealthMonitor: %v", err)
		}
		hm = nil
	})

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("DeletePool", func(t *testing.T) {
		err := testClient.DeletePool(testCtx, pool.ID)
		if err != nil {
			t.Fatalf("DeletePool: %v", err)
		}
		pool = nil
	})

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("DeleteListener", func(t *testing.T) {
		err := testClient.DeleteListener(testCtx, listener.ID)
		if err != nil {
			t.Fatalf("DeleteListener: %v", err)
		}
		listener = nil
	})

	waitForLBStatus(t, lb.ID, "ACTIVE", 3*time.Minute)

	t.Run("DeleteLB", func(t *testing.T) {
		err := testClient.DeleteLoadBalancer(testCtx, lb.ID)
		if err != nil {
			t.Fatalf("DeleteLoadBalancer: %v", err)
		}
		lb = nil
	})
}
