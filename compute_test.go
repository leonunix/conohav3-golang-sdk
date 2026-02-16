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

func TestListServersDetail_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"servers":[]}`))
	})
	defer server.Close()

	opts := &ListServersOptions{Limit: 5, Status: "ACTIVE", Marker: "srv-prev"}
	_, err := client.ListServersDetail(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=5") {
		t.Errorf("URI should contain limit=5: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "status=ACTIVE") {
		t.Errorf("URI should contain status=ACTIVE: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "marker=srv-prev") {
		t.Errorf("URI should contain marker: %q", capturedURI)
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

func TestGetNetworkTraffic_Success(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"interface":{"schema":["timestamp","rx","tx"],"data":[]}}`))
	})
	defer server.Close()

	opts := NetworkMonitoringOptions{PortID: "port-123"}
	opts.Mode = "average"
	_, err := client.GetNetworkTraffic(context.Background(), "srv-123", opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "port_id=port-123") {
		t.Errorf("URI should contain port_id: %q", capturedURI)
	}
}

func TestGetDiskIO_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"disk":{"schema":["timestamp","read","write"],"data":[]}}`))
	})
	defer server.Close()

	opts := &DiskMonitoringOptions{Device: "vda"}
	data, err := client.GetDiskIO(context.Background(), "srv-123", opts)
	assertNoError(t, err)

	if len(data.Schema) != 3 {
		t.Errorf("Schema = %v", data.Schema)
	}
}

// ============================================================
// Additional Server methods
// ============================================================

func TestConfirmResize_Success(t *testing.T) {
	var body map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.ConfirmResize(context.Background(), "srv-123")
	assertNoError(t, err)

	if _, ok := body["confirmResize"]; !ok {
		t.Errorf("body should contain 'confirmResize': %v", body)
	}
}

func TestRevertResize_Success(t *testing.T) {
	var body map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.RevertResize(context.Background(), "srv-123")
	assertNoError(t, err)

	if _, ok := body["revertResize"]; !ok {
		t.Errorf("body should contain 'revertResize': %v", body)
	}
}

func TestSetVideoDevice_Success(t *testing.T) {
	var body map[string]string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.SetVideoDevice(context.Background(), "srv-123", "vga")
	assertNoError(t, err)

	if body["hwVideoModel"] != "vga" {
		t.Errorf("hwVideoModel = %q", body["hwVideoModel"])
	}
}

func TestSetNetworkAdapter_Success(t *testing.T) {
	var body map[string]string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.SetNetworkAdapter(context.Background(), "srv-123", "virtio")
	assertNoError(t, err)

	if body["hwVifModel"] != "virtio" {
		t.Errorf("hwVifModel = %q", body["hwVifModel"])
	}
}

func TestSetStorageController_Success(t *testing.T) {
	var body map[string]string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.SetStorageController(context.Background(), "srv-123", "virtio")
	assertNoError(t, err)

	if body["hwDiskBus"] != "virtio" {
		t.Errorf("hwDiskBus = %q", body["hwDiskBus"])
	}
}

func TestUnmountISO_Success(t *testing.T) {
	var body map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.UnmountISO(context.Background(), "srv-123")
	assertNoError(t, err)

	if _, ok := body["unrescue"]; !ok {
		t.Errorf("body should contain 'unrescue': %v", body)
	}
}

func TestGetServerAddresses_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"addresses":{"public":[{"addr":"1.2.3.4","version":4}]}}`))
	})
	defer server.Close()

	addrs, err := client.GetServerAddresses(context.Background(), "srv-123")
	assertNoError(t, err)

	if len(addrs["public"]) != 1 || addrs["public"][0].Addr != "1.2.3.4" {
		t.Errorf("unexpected addresses: %+v", addrs)
	}
}

func TestGetServerAddressesByNetwork_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"mynet":[{"addr":"10.0.0.1","version":4}]}`))
	})
	defer server.Close()

	addrs, err := client.GetServerAddressesByNetwork(context.Background(), "srv-123", "mynet")
	assertNoError(t, err)

	if len(addrs) != 1 || addrs[0].Addr != "10.0.0.1" {
		t.Errorf("unexpected addresses: %+v", addrs)
	}
}

func TestGetServerSecurityGroups_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"security_groups":[{"id":"sg-1","name":"default"}]}`))
	})
	defer server.Close()

	sgs, err := client.GetServerSecurityGroups(context.Background(), "srv-123")
	assertNoError(t, err)

	if len(sgs) != 1 || sgs[0].Name != "default" {
		t.Errorf("unexpected sgs: %+v", sgs)
	}
}

func TestGetConsoleURL_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"remote_console":{"protocol":"vnc","type":"novnc","url":"https://console.example.com/vnc"}}`))
	})
	defer server.Close()

	console, err := client.GetConsoleURL(context.Background(), "srv-123", RemoteConsoleRequest{Protocol: "vnc", Type: "novnc"})
	assertNoError(t, err)

	if console.URL != "https://console.example.com/vnc" {
		t.Errorf("URL = %q", console.URL)
	}
}

func TestGetVNCConsoleURL_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"remote_console":{"protocol":"vnc","type":"novnc","url":"https://console.example.com/vnc"}}`))
	})
	defer server.Close()

	url, err := client.GetVNCConsoleURL(context.Background(), "srv-123")
	assertNoError(t, err)

	if url != "https://console.example.com/vnc" {
		t.Errorf("URL = %q", url)
	}
}

func TestGetServerMetadata_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"metadata":{"env":"production","app":"web"}}`))
	})
	defer server.Close()

	meta, err := client.GetServerMetadata(context.Background(), "srv-123")
	assertNoError(t, err)

	if meta["env"] != "production" {
		t.Errorf("unexpected metadata: %+v", meta)
	}
}

func TestUpdateServerMetadata_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"metadata":{"env":"staging"}}`))
	})
	defer server.Close()

	meta, err := client.UpdateServerMetadata(context.Background(), "srv-123", map[string]string{"env": "staging"})
	assertNoError(t, err)

	if meta["env"] != "staging" {
		t.Errorf("unexpected metadata: %+v", meta)
	}
}

func TestImportKeypair_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"keypair":{"name":"imported","public_key":"ssh-rsa AAAA...","fingerprint":"aa:bb:cc"}}`))
	})
	defer server.Close()

	kp, err := client.ImportKeypair(context.Background(), "imported", "ssh-rsa AAAA...")
	assertNoError(t, err)

	if kp.Name != "imported" {
		t.Errorf("Name = %q", kp.Name)
	}
}

func TestGetKeypair_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"keypair":{"name":"mykey","fingerprint":"aa:bb:cc"}}`))
	})
	defer server.Close()

	kp, err := client.GetKeypair(context.Background(), "mykey")
	assertNoError(t, err)

	if kp.Fingerprint != "aa:bb:cc" {
		t.Errorf("Fingerprint = %q", kp.Fingerprint)
	}
}

// ============================================================
// Server Interfaces & Volume Attachments
// ============================================================

func TestListServerInterfaces_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"interfaceAttachments":[{"port_id":"port-1","net_id":"net-1"}]}`))
	})
	defer server.Close()

	ifaces, err := client.ListServerInterfaces(context.Background(), "srv-123")
	assertNoError(t, err)

	if len(ifaces) != 1 || ifaces[0].PortID != "port-1" {
		t.Errorf("unexpected interfaces: %+v", ifaces)
	}
}

func TestGetServerInterface_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"interfaceAttachment":{"port_id":"port-1","net_id":"net-1"}}`))
	})
	defer server.Close()

	iface, err := client.GetServerInterface(context.Background(), "srv-123", "port-1")
	assertNoError(t, err)

	if iface.PortID != "port-1" {
		t.Errorf("PortID = %q", iface.PortID)
	}
}

func TestAttachPort_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"interfaceAttachment":{"port_id":"port-new","net_id":"net-1"}}`))
	})
	defer server.Close()

	iface, err := client.AttachPort(context.Background(), "srv-123", "port-new")
	assertNoError(t, err)

	if iface.PortID != "port-new" {
		t.Errorf("PortID = %q", iface.PortID)
	}
}

func TestDetachPort_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.DetachPort(context.Background(), "srv-123", "port-1")
	assertNoError(t, err)

	if capturedPath != "/servers/srv-123/os-interface/port-1" {
		t.Errorf("Path = %q", capturedPath)
	}
}

func TestListServerVolumes_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"volumeAttachments":[{"id":"att-1","volumeId":"vol-1","serverId":"srv-123"}]}`))
	})
	defer server.Close()

	vols, err := client.ListServerVolumes(context.Background(), "srv-123")
	assertNoError(t, err)

	if len(vols) != 1 || vols[0].VolumeID != "vol-1" {
		t.Errorf("unexpected attachments: %+v", vols)
	}
}

func TestGetServerVolume_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"volumeAttachment":{"id":"att-1","volumeId":"vol-1","device":"/dev/vdb"}}`))
	})
	defer server.Close()

	att, err := client.GetServerVolume(context.Background(), "srv-123", "vol-1")
	assertNoError(t, err)

	if att.Device != "/dev/vdb" {
		t.Errorf("Device = %q", att.Device)
	}
}

func TestAttachVolume_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"volumeAttachment":{"id":"att-new","volumeId":"vol-1","serverId":"srv-123"}}`))
	})
	defer server.Close()

	att, err := client.AttachVolume(context.Background(), "srv-123", "vol-1")
	assertNoError(t, err)

	if att.VolumeID != "vol-1" {
		t.Errorf("VolumeID = %q", att.VolumeID)
	}
}

func TestDetachVolume_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.DetachVolume(context.Background(), "srv-123", "vol-1")
	assertNoError(t, err)

	if capturedPath != "/servers/srv-123/os-volume_attachments/vol-1" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// GetCPUUsage with options
// ============================================================

func TestGetCPUUsage_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"cpu":{"schema":["timestamp","cpu"],"data":[[1700000000,50.5]]}}`))
	})
	defer server.Close()

	opts := &MonitoringOptions{
		StartDateRaw: "2024-01-01",
		EndDateRaw:   "2024-01-02",
		Mode:         "average",
	}
	cpu, err := client.GetCPUUsage(context.Background(), "srv-123", opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "/servers/srv-123/rrd/cpu") {
		t.Errorf("URI = %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "mode=average") {
		t.Errorf("URI should contain mode=average: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "start_date_raw=2024-01-01") {
		t.Errorf("URI should contain start_date_raw: %q", capturedURI)
	}
	if len(cpu.Schema) != 2 {
		t.Errorf("Schema length = %d", len(cpu.Schema))
	}
}
