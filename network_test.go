package conoha

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// ============================================================
// Networks
// ============================================================

func TestListNetworks_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"networks":[{"id":"net-1","name":"public-net","status":"ACTIVE"}]}`))
	})
	defer server.Close()

	networks, err := client.ListNetworks(context.Background(), nil)
	assertNoError(t, err)

	if capturedPath != "/networks" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(networks) != 1 || networks[0].ID != "net-1" {
		t.Errorf("unexpected networks: %+v", networks)
	}
}

func TestGetNetwork_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"network":{"id":"net-123","name":"mynetwork","status":"ACTIVE","subnets":["sub-1"]}}`))
	})
	defer server.Close()

	net, err := client.GetNetwork(context.Background(), "net-123")
	assertNoError(t, err)

	if net.ID != "net-123" || len(net.Subnets) != 1 {
		t.Errorf("unexpected network: %+v", net)
	}
}

func TestCreateNetwork_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(201)
		w.Write([]byte(`{"network":{"id":"new-net","name":"","admin_state_up":true}}`))
	})
	defer server.Close()

	net, err := client.CreateNetwork(context.Background())
	assertNoError(t, err)

	if capturedMethod != http.MethodPost {
		t.Errorf("Method = %q", capturedMethod)
	}
	if net.ID != "new-net" {
		t.Errorf("ID = %q", net.ID)
	}
}

func TestDeleteNetwork_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteNetwork(context.Background(), "net-123")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/networks/net-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Subnets
// ============================================================

func TestListSubnets_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"subnets":[{"id":"sub-1","cidr":"192.168.0.0/24","ip_version":4}]}`))
	})
	defer server.Close()

	subnets, err := client.ListSubnets(context.Background(), nil)
	assertNoError(t, err)

	if len(subnets) != 1 || subnets[0].CIDR != "192.168.0.0/24" {
		t.Errorf("unexpected subnets: %+v", subnets)
	}
}

func TestCreateSubnet_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"subnet":{"id":"sub-new","network_id":"net-1","cidr":"10.0.0.0/24"}}`))
	})
	defer server.Close()

	subnet, err := client.CreateSubnet(context.Background(), "net-1", "10.0.0.0/24")
	assertNoError(t, err)

	if _, ok := body["subnet"]; !ok {
		t.Error("body should contain 'subnet'")
	}
	if subnet.CIDR != "10.0.0.0/24" {
		t.Errorf("CIDR = %q", subnet.CIDR)
	}
}

func TestDeleteSubnet_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteSubnet(context.Background(), "sub-123")
	assertNoError(t, err)

	if capturedPath != "/subnets/sub-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Security Groups
// ============================================================

func TestListSecurityGroups_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"security_groups":[{"id":"sg-1","name":"default","description":"default security group"}]}`))
	})
	defer server.Close()

	sgs, err := client.ListSecurityGroups(context.Background(), nil)
	assertNoError(t, err)

	if len(sgs) != 1 || sgs[0].Name != "default" {
		t.Errorf("unexpected security groups: %+v", sgs)
	}
}

func TestGetSecurityGroup_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"security_group":{"id":"sg-123","name":"websg","security_group_rules":[]}}`))
	})
	defer server.Close()

	sg, err := client.GetSecurityGroup(context.Background(), "sg-123")
	assertNoError(t, err)

	if sg.Name != "websg" {
		t.Errorf("Name = %q", sg.Name)
	}
}

func TestCreateSecurityGroup_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"security_group":{"id":"sg-new","name":"websg","description":"web servers"}}`))
	})
	defer server.Close()

	sg, err := client.CreateSecurityGroup(context.Background(), "websg", "web servers")
	assertNoError(t, err)

	if _, ok := body["security_group"]; !ok {
		t.Error("body should contain 'security_group'")
	}
	if sg.Name != "websg" {
		t.Errorf("Name = %q", sg.Name)
	}
}

func TestDeleteSecurityGroup_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteSecurityGroup(context.Background(), "sg-123")
	assertNoError(t, err)

	if capturedPath != "/security-groups/sg-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

func TestCreateSecurityGroupRule_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"security_group_rule":{"id":"rule-1","direction":"ingress","ethertype":"IPv4"}}`))
	})
	defer server.Close()

	proto := "tcp"
	portMin := 80
	portMax := 80
	opts := CreateSecurityGroupRuleRequest{
		SecurityGroupID: "sg-123",
		Direction:       "ingress",
		EtherType:       "IPv4",
		Protocol:        &proto,
		PortRangeMin:    &portMin,
		PortRangeMax:    &portMax,
	}
	rule, err := client.CreateSecurityGroupRule(context.Background(), opts)
	assertNoError(t, err)

	if _, ok := body["security_group_rule"]; !ok {
		t.Error("body should contain 'security_group_rule'")
	}
	if rule.Direction != "ingress" {
		t.Errorf("Direction = %q", rule.Direction)
	}
}

func TestDeleteSecurityGroupRule_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteSecurityGroupRule(context.Background(), "rule-123")
	assertNoError(t, err)

	if capturedPath != "/security-group-rules/rule-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// Ports
// ============================================================

func TestListPorts_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ports":[{"id":"port-1","network_id":"net-1","status":"ACTIVE"}]}`))
	})
	defer server.Close()

	ports, err := client.ListPorts(context.Background(), nil)
	assertNoError(t, err)

	if len(ports) != 1 || ports[0].ID != "port-1" {
		t.Errorf("unexpected ports: %+v", ports)
	}
}

func TestListPorts_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"ports":[]}`))
	})
	defer server.Close()

	opts := &ListPortsOptions{NetworkID: "net-1", DeviceID: "dev-1"}
	_, err := client.ListPorts(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "network_id=net-1") {
		t.Errorf("URI should contain network_id: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "device_id=dev-1") {
		t.Errorf("URI should contain device_id: %q", capturedURI)
	}
}

func TestCreatePort_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(201)
		w.Write([]byte(`{"port":{"id":"port-new","network_id":"net-1"}}`))
	})
	defer server.Close()

	opts := CreatePortRequest{NetworkID: "net-1"}
	port, err := client.CreatePort(context.Background(), opts)
	assertNoError(t, err)

	if _, ok := body["port"]; !ok {
		t.Error("body should contain 'port'")
	}
	if port.ID != "port-new" {
		t.Errorf("ID = %q", port.ID)
	}
}

func TestDeletePort_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeletePort(context.Background(), "port-123")
	assertNoError(t, err)

	if capturedPath != "/ports/port-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// QoS Policies
// ============================================================

func TestListQoSPolicies_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"policies":[{"id":"qos-1","name":"default-qos"}]}`))
	})
	defer server.Close()

	policies, err := client.ListQoSPolicies(context.Background(), nil)
	assertNoError(t, err)

	if len(policies) != 1 || policies[0].Name != "default-qos" {
		t.Errorf("unexpected policies: %+v", policies)
	}
}

func TestGetQoSPolicy_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"policy":{"id":"qos-1","name":"premium","rules":[{"max_kbps":100000}]}}`))
	})
	defer server.Close()

	policy, err := client.GetQoSPolicy(context.Background(), "qos-1")
	assertNoError(t, err)

	if policy.Name != "premium" {
		t.Errorf("Name = %q", policy.Name)
	}
	if len(policy.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(policy.Rules))
	}
}
