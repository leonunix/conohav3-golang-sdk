package conoha

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// ============================================================
// Server CRUD
// ============================================================

func TestListServers_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"servers":[{"id":"srv-1","name":"server1"}]}`))
	})
	defer server.Close()

	servers, err := client.ListServers(context.Background(), nil)
	assertNoError(t, err)

	if capturedPath != "/servers" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(servers) != 1 || servers[0].ID != "srv-1" {
		t.Errorf("unexpected servers: %+v", servers)
	}
}

func TestListServers_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"servers":[]}`))
	})
	defer server.Close()

	opts := &ListServersOptions{Limit: 10, Status: "ACTIVE", Name: "test"}
	_, err := client.ListServers(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=10") {
		t.Errorf("URI should contain limit=10: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "status=ACTIVE") {
		t.Errorf("URI should contain status=ACTIVE: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "name=test") {
		t.Errorf("URI should contain name=test: %q", capturedURI)
	}
}

func TestListServersDetail_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"servers":[{"id":"srv-1","name":"server1","status":"ACTIVE"}]}`))
	})
	defer server.Close()

	servers, err := client.ListServersDetail(context.Background(), nil)
	assertNoError(t, err)

	if capturedPath != "/servers/detail" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(servers) != 1 || servers[0].Status != "ACTIVE" {
		t.Errorf("unexpected servers: %+v", servers)
	}
}

func TestGetServer_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"server":{"id":"srv-123","name":"myserver","status":"ACTIVE"}}`))
	})
	defer server.Close()

	srv, err := client.GetServer(context.Background(), "srv-123")
	assertNoError(t, err)

	if capturedPath != "/servers/srv-123" {
		t.Errorf("Path = %q", capturedPath)
	}
	if srv.ID != "srv-123" || srv.Name != "myserver" {
		t.Errorf("unexpected server: %+v", srv)
	}
}

func TestGetServer_NotFound(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"itemNotFound":{"message":"Server not found","code":404}}`))
	})
	defer server.Close()

	_, err := client.GetServer(context.Background(), "missing")
	assertAPIError(t, err, 404)
}

func TestCreateServer_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
		w.Write([]byte(`{"server":{"id":"new-srv","adminPass":"secret123"}}`))
	})
	defer server.Close()

	opts := CreateServerRequest{
		FlavorRef: "flavor-1",
		AdminPass: "mypass",
		BlockDeviceMapping: []BlockDeviceMap{
			{UUID: "image-uuid"},
		},
	}
	result, err := client.CreateServer(context.Background(), opts)
	assertNoError(t, err)

	// Verify body wrapping
	if _, ok := body["server"]; !ok {
		t.Error("body should contain 'server' key")
	}
	if result.ID != "new-srv" {
		t.Errorf("ID = %q", result.ID)
	}
	if result.AdminPass != "secret123" {
		t.Errorf("AdminPass = %q", result.AdminPass)
	}
}

func TestDeleteServer_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteServer(context.Background(), "srv-123")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/servers/srv-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Server Actions
// ============================================================

func TestStartServer_Success(t *testing.T) {
	var body map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.StartServer(context.Background(), "srv-123")
	assertNoError(t, err)

	if _, ok := body["os-start"]; !ok {
		t.Errorf("body should contain 'os-start': %v", body)
	}
}

func TestStopServer_Success(t *testing.T) {
	var body map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.StopServer(context.Background(), "srv-123")
	assertNoError(t, err)

	if _, ok := body["os-stop"]; !ok {
		t.Errorf("body should contain 'os-stop': %v", body)
	}
}

func TestRebootServer_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.RebootServer(context.Background(), "srv-123")
	assertNoError(t, err)

	if _, ok := body["reboot"]; !ok {
		t.Errorf("body should contain 'reboot': %v", body)
	}
}

func TestForceStopServer_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.ForceStopServer(context.Background(), "srv-123")
	assertNoError(t, err)

	raw, ok := body["os-stop"]
	if !ok {
		t.Fatalf("body should contain 'os-stop': %v", body)
	}
	var inner map[string]bool
	json.Unmarshal(raw, &inner)
	if !inner["force_shutdown"] {
		t.Errorf("force_shutdown should be true: %v", inner)
	}
}

func TestRebuildServer_Success(t *testing.T) {
	var capturedPath string
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	opts := RebuildServerRequest{ImageRef: "img-123", AdminPass: "newpass"}
	err := client.RebuildServer(context.Background(), "srv-123", opts)
	assertNoError(t, err)

	if capturedPath != "/servers/srv-123/action" {
		t.Errorf("Path = %q", capturedPath)
	}
	if _, ok := body["rebuild"]; !ok {
		t.Errorf("body should contain 'rebuild'")
	}
}

func TestResizeServer_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.ResizeServer(context.Background(), "srv-123", "new-flavor")
	assertNoError(t, err)

	if _, ok := body["resize"]; !ok {
		t.Errorf("body should contain 'resize'")
	}
}

func TestMountISO_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"adminPass":"rescue-pass"}`))
	})
	defer server.Close()

	pass, err := client.MountISO(context.Background(), "srv-123", "iso-img-ref")
	assertNoError(t, err)

	if pass != "rescue-pass" {
		t.Errorf("adminPass = %q", pass)
	}
}

func TestServerAction_Error(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		w.Write([]byte(`{"conflictingRequest":{"message":"Server is locked","code":409}}`))
	})
	defer server.Close()

	err := client.StartServer(context.Background(), "srv-123")
	assertAPIError(t, err, 409)
}

// ============================================================
// Flavors
// ============================================================

func TestListFlavors_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"flavors":[{"id":"f-1","name":"g-2gb"}]}`))
	})
	defer server.Close()

	flavors, err := client.ListFlavors(context.Background())
	assertNoError(t, err)

	if capturedPath != "/flavors" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(flavors) != 1 || flavors[0].ID != "f-1" {
		t.Errorf("unexpected flavors: %+v", flavors)
	}
}

func TestListFlavorsDetail_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"flavors":[{"id":"f-1","name":"g-2gb","vcpus":2,"ram":2048,"disk":100}]}`))
	})
	defer server.Close()

	flavors, err := client.ListFlavorsDetail(context.Background())
	assertNoError(t, err)

	if capturedPath != "/flavors/detail" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(flavors) != 1 || flavors[0].VCPUs != 2 {
		t.Errorf("unexpected flavors: %+v", flavors)
	}
}

func TestGetFlavor_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"flavor":{"id":"f-1","name":"g-2gb","vcpus":2,"ram":2048}}`))
	})
	defer server.Close()

	flavor, err := client.GetFlavor(context.Background(), "f-1")
	assertNoError(t, err)

	if flavor.Name != "g-2gb" {
		t.Errorf("Name = %q", flavor.Name)
	}
}

// ============================================================
// Keypairs
// ============================================================

func TestListKeypairs_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"keypairs":[{"keypair":{"name":"mykey","fingerprint":"aa:bb:cc"}}]}`))
	})
	defer server.Close()

	keypairs, err := client.ListKeypairs(context.Background(), nil)
	assertNoError(t, err)

	if capturedPath != "/os-keypairs" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(keypairs) != 1 || keypairs[0].Name != "mykey" {
		t.Errorf("unexpected keypairs: %+v", keypairs)
	}
}

func TestListKeypairs_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"keypairs":[]}`))
	})
	defer server.Close()

	opts := &ListKeypairsOptions{Limit: 5, Marker: "marker-id"}
	_, err := client.ListKeypairs(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=5") {
		t.Errorf("URI should contain limit=5: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "marker=marker-id") {
		t.Errorf("URI should contain marker: %q", capturedURI)
	}
}

func TestCreateKeypair_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"keypair":{"name":"newkey","public_key":"ssh-rsa ...","private_key":"-----BEGIN RSA PRIVATE KEY-----\n..."}}`))
	})
	defer server.Close()

	kp, err := client.CreateKeypair(context.Background(), "newkey")
	assertNoError(t, err)

	if _, ok := body["keypair"]; !ok {
		t.Error("body should contain 'keypair'")
	}
	if kp.Name != "newkey" {
		t.Errorf("Name = %q", kp.Name)
	}
	if kp.PrivateKey == "" {
		t.Error("PrivateKey should be set for new keypair")
	}
}

func TestDeleteKeypair_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteKeypair(context.Background(), "mykey")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/os-keypairs/mykey" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Monitoring
// ============================================================

func TestGetCPUUsage_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"cpu":{"schema":["timestamp","cpu"],"data":[[1234567890,50.5]]}}`))
	})
	defer server.Close()

	data, err := client.GetCPUUsage(context.Background(), "srv-123", nil)
	assertNoError(t, err)

	if capturedPath != "/servers/srv-123/rrd/cpu" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(data.Schema) != 2 {
		t.Errorf("Schema = %v", data.Schema)
	}
}

func TestGetNetworkTraffic_MissingPortID(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	defer server.Close()

	_, err := client.GetNetworkTraffic(context.Background(), "srv-123", NetworkMonitoringOptions{})
	assertError(t, err)

	if !strings.Contains(err.Error(), "PortID is required") {
		t.Errorf("error should mention PortID: %v", err)
	}
}
