package conoha

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// ============================================================
// Domain Operations
// ============================================================

func TestListDomains_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"domains":[{"uuid":"dom-1","name":"example.com.","ttl":3600}],"total_count":1}`))
	})
	defer server.Close()

	domains, err := client.ListDomains(context.Background(), nil)
	assertNoError(t, err)

	if capturedPath != "/domains" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(domains) != 1 || domains[0].Name != "example.com." {
		t.Errorf("unexpected domains: %+v", domains)
	}
}

func TestListDomains_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"domains":[],"total_count":0}`))
	})
	defer server.Close()

	opts := &ListDomainsOptions{Limit: 10, SortType: "asc", SortKey: "name"}
	_, err := client.ListDomains(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=10") {
		t.Errorf("URI should contain limit=10: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "sort_type=asc") {
		t.Errorf("URI should contain sort_type=asc: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "sort_key=name") {
		t.Errorf("URI should contain sort_key=name: %q", capturedURI)
	}
}

func TestGetDomain_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		// GetDomain unmarshals directly into Domain (no wrapper)
		w.Write([]byte(`{"uuid":"dom-123","name":"example.com.","ttl":3600,"email":"admin@example.com"}`))
	})
	defer server.Close()

	domain, err := client.GetDomain(context.Background(), "dom-123")
	assertNoError(t, err)

	if capturedPath != "/domains/dom-123" {
		t.Errorf("Path = %q", capturedPath)
	}
	if domain.Name != "example.com." {
		t.Errorf("Name = %q", domain.Name)
	}
	if domain.TTL != 3600 {
		t.Errorf("TTL = %d", domain.TTL)
	}
}

func TestCreateDomain_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(201)
		w.Write([]byte(`{"uuid":"dom-new","name":"new.example.com.","ttl":7200,"email":"admin@example.com"}`))
	})
	defer server.Close()

	opts := CreateDomainRequest{Name: "new.example.com.", TTL: 7200, Email: "admin@example.com"}
	domain, err := client.CreateDomain(context.Background(), opts)
	assertNoError(t, err)

	if capturedMethod != http.MethodPost {
		t.Errorf("Method = %q", capturedMethod)
	}
	if domain.UUID != "dom-new" {
		t.Errorf("UUID = %q", domain.UUID)
	}
}

func TestUpdateDomain_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(200)
		w.Write([]byte(`{"uuid":"dom-123","name":"example.com.","ttl":1800,"email":"new@example.com"}`))
	})
	defer server.Close()

	opts := UpdateDomainRequest{TTL: 1800, Email: "new@example.com"}
	domain, err := client.UpdateDomain(context.Background(), "dom-123", opts)
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q", capturedMethod)
	}
	if domain.TTL != 1800 {
		t.Errorf("TTL = %d", domain.TTL)
	}
}

func TestDeleteDomain_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteDomain(context.Background(), "dom-123")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/domains/dom-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// DNS Record Operations
// ============================================================

func TestListDNSRecords_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"records":[{"uuid":"rec-1","name":"www.example.com.","type":"A","data":"1.2.3.4","ttl":3600}],"total_count":1}`))
	})
	defer server.Close()

	records, err := client.ListDNSRecords(context.Background(), "dom-123", nil)
	assertNoError(t, err)

	if capturedPath != "/domains/dom-123/records" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(records) != 1 || records[0].Type != "A" {
		t.Errorf("unexpected records: %+v", records)
	}
}

func TestListDNSRecords_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"records":[],"total_count":0}`))
	})
	defer server.Close()

	opts := &ListDNSRecordsOptions{Limit: 20, Offset: 5}
	_, err := client.ListDNSRecords(context.Background(), "dom-123", opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=20") {
		t.Errorf("URI should contain limit=20: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "offset=5") {
		t.Errorf("URI should contain offset=5: %q", capturedURI)
	}
}

func TestGetDNSRecord_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"uuid":"rec-123","name":"www.example.com.","type":"A","data":"1.2.3.4","ttl":3600}`))
	})
	defer server.Close()

	record, err := client.GetDNSRecord(context.Background(), "dom-123", "rec-123")
	assertNoError(t, err)

	if record.Data != "1.2.3.4" {
		t.Errorf("Data = %q", record.Data)
	}
}

func TestCreateDNSRecord_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(201)
		w.Write([]byte(`{"uuid":"rec-new","name":"api.example.com.","type":"A","data":"5.6.7.8","ttl":3600}`))
	})
	defer server.Close()

	opts := CreateDNSRecordRequest{Name: "api.example.com.", Type: "A", Data: "5.6.7.8"}
	record, err := client.CreateDNSRecord(context.Background(), "dom-123", opts)
	assertNoError(t, err)

	if capturedMethod != http.MethodPost {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/domains/dom-123/records" {
		t.Errorf("Path = %q", capturedPath)
	}
	if record.UUID != "rec-new" {
		t.Errorf("UUID = %q", record.UUID)
	}
}

func TestUpdateDNSRecord_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"uuid":"rec-123","name":"www.example.com.","type":"A","data":"9.8.7.6","ttl":1800}`))
	})
	defer server.Close()

	opts := UpdateDNSRecordRequest{Data: "9.8.7.6"}
	record, err := client.UpdateDNSRecord(context.Background(), "dom-123", "rec-123", opts)
	assertNoError(t, err)

	if record.Data != "9.8.7.6" {
		t.Errorf("Data = %q", record.Data)
	}
}

func TestDeleteDNSRecord_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteDNSRecord(context.Background(), "dom-123", "rec-123")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/domains/dom-123/records/rec-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}
