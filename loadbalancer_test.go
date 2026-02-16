package conoha

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// ============================================================
// Load Balancers
// ============================================================

func TestListLoadBalancers_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"loadbalancers":[{"id":"lb-1","name":"mylb","provisioning_status":"ACTIVE"}]}`))
	})
	defer server.Close()

	lbs, err := client.ListLoadBalancers(context.Background())
	assertNoError(t, err)

	if capturedPath != "/lbaas/loadbalancers" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(lbs) != 1 || lbs[0].ID != "lb-1" {
		t.Errorf("unexpected lbs: %+v", lbs)
	}
}

func TestGetLoadBalancer_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"loadbalancer":{"id":"lb-123","name":"mylb","vip_address":"10.0.0.1"}}`))
	})
	defer server.Close()

	lb, err := client.GetLoadBalancer(context.Background(), "lb-123")
	assertNoError(t, err)

	if lb.VIPAddress != "10.0.0.1" {
		t.Errorf("VIPAddress = %q", lb.VIPAddress)
	}
}

func TestCreateLoadBalancer_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"loadbalancer":{"id":"lb-new","name":"newlb"}}`))
	})
	defer server.Close()

	lb, err := client.CreateLoadBalancer(context.Background(), "newlb")
	assertNoError(t, err)

	if _, ok := body["loadbalancer"]; !ok {
		t.Error("body should contain 'loadbalancer'")
	}
	if lb.ID != "lb-new" {
		t.Errorf("ID = %q", lb.ID)
	}
}

func TestUpdateLoadBalancer_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(200)
		w.Write([]byte(`{"loadbalancer":{"id":"lb-123","name":"updated"}}`))
	})
	defer server.Close()

	lb, err := client.UpdateLoadBalancer(context.Background(), "lb-123", "updated")
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q", capturedMethod)
	}
	if lb.Name != "updated" {
		t.Errorf("Name = %q", lb.Name)
	}
}

func TestDeleteLoadBalancer_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteLoadBalancer(context.Background(), "lb-123")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/lbaas/loadbalancers/lb-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Listeners
// ============================================================

func TestListListeners_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"listeners":[{"id":"ls-1","name":"http","protocol":"HTTP","protocol_port":80}]}`))
	})
	defer server.Close()

	listeners, err := client.ListListeners(context.Background())
	assertNoError(t, err)

	if len(listeners) != 1 || listeners[0].Protocol != "HTTP" {
		t.Errorf("unexpected listeners: %+v", listeners)
	}
}

func TestGetListener_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"listener":{"id":"ls-123","name":"http","protocol_port":80}}`))
	})
	defer server.Close()

	ls, err := client.GetListener(context.Background(), "ls-123")
	assertNoError(t, err)

	if ls.ProtocolPort != 80 {
		t.Errorf("ProtocolPort = %d", ls.ProtocolPort)
	}
}

func TestCreateListener_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"listener":{"id":"ls-new","name":"https","protocol":"HTTPS","protocol_port":443}}`))
	})
	defer server.Close()

	ls, err := client.CreateListener(context.Background(), "https", "HTTPS", 443, "lb-123")
	assertNoError(t, err)

	if _, ok := body["listener"]; !ok {
		t.Error("body should contain 'listener'")
	}
	if ls.ProtocolPort != 443 {
		t.Errorf("ProtocolPort = %d", ls.ProtocolPort)
	}
}

func TestUpdateListener_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"listener":{"id":"ls-123","name":"renamed"}}`))
	})
	defer server.Close()

	ls, err := client.UpdateListener(context.Background(), "ls-123", "renamed")
	assertNoError(t, err)

	if ls.Name != "renamed" {
		t.Errorf("Name = %q", ls.Name)
	}
}

func TestDeleteListener_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteListener(context.Background(), "ls-123")
	assertNoError(t, err)

	if capturedPath != "/lbaas/listeners/ls-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Pools
// ============================================================

func TestListPools_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"pools":[{"id":"pool-1","name":"mypool","protocol":"HTTP","lb_algorithm":"ROUND_ROBIN"}]}`))
	})
	defer server.Close()

	pools, err := client.ListPools(context.Background())
	assertNoError(t, err)

	if len(pools) != 1 || pools[0].LBAlgorithm != "ROUND_ROBIN" {
		t.Errorf("unexpected pools: %+v", pools)
	}
}

func TestGetPool_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"pool":{"id":"pool-123","name":"mypool","lb_algorithm":"LEAST_CONNECTIONS"}}`))
	})
	defer server.Close()

	pool, err := client.GetPool(context.Background(), "pool-123")
	assertNoError(t, err)

	if pool.LBAlgorithm != "LEAST_CONNECTIONS" {
		t.Errorf("LBAlgorithm = %q", pool.LBAlgorithm)
	}
}

func TestCreatePool_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"pool":{"id":"pool-new","name":"newpool","protocol":"HTTP"}}`))
	})
	defer server.Close()

	pool, err := client.CreatePool(context.Background(), "newpool", "HTTP", "ROUND_ROBIN", "ls-123")
	assertNoError(t, err)

	if _, ok := body["pool"]; !ok {
		t.Error("body should contain 'pool'")
	}
	if pool.ID != "pool-new" {
		t.Errorf("ID = %q", pool.ID)
	}
}

func TestUpdatePool_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"pool":{"id":"pool-123","name":"updated","lb_algorithm":"LEAST_CONNECTIONS"}}`))
	})
	defer server.Close()

	pool, err := client.UpdatePool(context.Background(), "pool-123", "updated", "LEAST_CONNECTIONS")
	assertNoError(t, err)

	if pool.Name != "updated" {
		t.Errorf("Name = %q", pool.Name)
	}
}

func TestDeletePool_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeletePool(context.Background(), "pool-123")
	assertNoError(t, err)

	if capturedPath != "/lbaas/pools/pool-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Members
// ============================================================

func TestListMembers_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"members":[{"id":"mem-1","address":"10.0.0.5","protocol_port":8080}]}`))
	})
	defer server.Close()

	members, err := client.ListMembers(context.Background(), "pool-123")
	assertNoError(t, err)

	if !strings.Contains(capturedPath, "/lbaas/pools/pool-123/members") {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(members) != 1 || members[0].Address != "10.0.0.5" {
		t.Errorf("unexpected members: %+v", members)
	}
}

func TestGetMember_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"member":{"id":"mem-123","address":"10.0.0.5","protocol_port":8080,"weight":1}}`))
	})
	defer server.Close()

	member, err := client.GetMember(context.Background(), "pool-123", "mem-123")
	assertNoError(t, err)

	if member.Weight != 1 {
		t.Errorf("Weight = %d", member.Weight)
	}
}

func TestAddMember_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"member":{"id":"mem-new","name":"backend1","address":"10.0.0.10","protocol_port":8080}}`))
	})
	defer server.Close()

	member, err := client.AddMember(context.Background(), "pool-123", "backend1", "10.0.0.10", 8080)
	assertNoError(t, err)

	if _, ok := body["member"]; !ok {
		t.Error("body should contain 'member'")
	}
	if member.Address != "10.0.0.10" {
		t.Errorf("Address = %q", member.Address)
	}
}

func TestUpdateMember_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"member":{"id":"mem-123","admin_state_up":false}}`))
	})
	defer server.Close()

	member, err := client.UpdateMember(context.Background(), "pool-123", "mem-123", false)
	assertNoError(t, err)

	if member.AdminStateUp != false {
		t.Errorf("AdminStateUp = %v", member.AdminStateUp)
	}
}

func TestDeleteMember_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteMember(context.Background(), "pool-123", "mem-123")
	assertNoError(t, err)

	if capturedPath != "/lbaas/pools/pool-123/members/mem-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Health Monitors
// ============================================================

func TestListHealthMonitors_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"healthmonitors":[{"id":"hm-1","type":"HTTP","delay":5}]}`))
	})
	defer server.Close()

	monitors, err := client.ListHealthMonitors(context.Background())
	assertNoError(t, err)

	if len(monitors) != 1 || monitors[0].Type != "HTTP" {
		t.Errorf("unexpected monitors: %+v", monitors)
	}
}

func TestGetHealthMonitor_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"healthmonitor":{"id":"hm-123","type":"HTTP","delay":5,"timeout":3,"max_retries":3}}`))
	})
	defer server.Close()

	hm, err := client.GetHealthMonitor(context.Background(), "hm-123")
	assertNoError(t, err)

	if hm.Delay != 5 || hm.MaxRetries != 3 {
		t.Errorf("unexpected monitor: %+v", hm)
	}
}

func TestCreateHealthMonitor_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"healthmonitor":{"id":"hm-new","type":"HTTP","delay":10}}`))
	})
	defer server.Close()

	opts := CreateHealthMonitorRequest{
		Name:       "http-check",
		PoolID:     "pool-123",
		Delay:      10,
		MaxRetries: 3,
		Timeout:    5,
		Type:       "HTTP",
	}
	hm, err := client.CreateHealthMonitor(context.Background(), opts)
	assertNoError(t, err)

	if _, ok := body["healthmonitor"]; !ok {
		t.Error("body should contain 'healthmonitor'")
	}
	if hm.Delay != 10 {
		t.Errorf("Delay = %d", hm.Delay)
	}
}

func TestUpdateHealthMonitor_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"healthmonitor":{"id":"hm-123","name":"renamed"}}`))
	})
	defer server.Close()

	hm, err := client.UpdateHealthMonitor(context.Background(), "hm-123", "renamed")
	assertNoError(t, err)

	if hm.Name != "renamed" {
		t.Errorf("Name = %q", hm.Name)
	}
}

func TestDeleteHealthMonitor_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteHealthMonitor(context.Background(), "hm-123")
	assertNoError(t, err)

	if capturedPath != "/lbaas/healthmonitors/hm-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}
