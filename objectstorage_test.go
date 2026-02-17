package conoha

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// ============================================================
// objectStoragePath
// ============================================================

func TestObjectStoragePath_Simple(t *testing.T) {
	c := NewClient()
	c.ObjectStorageURL = "https://object-storage.c3j1.conoha.io/v1"
	c.TenantID = "tenant-abc"

	path := c.objectStoragePath("mycontainer")
	expected := "https://object-storage.c3j1.conoha.io/v1/AUTH_tenant-abc/mycontainer"
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}

func TestObjectStoragePath_WithObject(t *testing.T) {
	c := NewClient()
	c.ObjectStorageURL = "https://object-storage.c3j1.conoha.io/v1"
	c.TenantID = "tenant-abc"

	path := c.objectStoragePath("container", "path/to/object.txt")
	// Slashes within object names should be preserved
	if !strings.Contains(path, "AUTH_tenant-abc/container/path/to/object.txt") {
		t.Errorf("slashes should be preserved in object name: %q", path)
	}
}

func TestObjectStoragePath_SpecialChars(t *testing.T) {
	c := NewClient()
	c.ObjectStorageURL = "https://object-storage.c3j1.conoha.io/v1"
	c.TenantID = "tenant-abc"

	path := c.objectStoragePath("container", "file with spaces.txt")
	if !strings.Contains(path, "file%20with%20spaces.txt") {
		t.Errorf("spaces should be encoded: %q", path)
	}
}

func TestObjectStoragePath_NoParts(t *testing.T) {
	c := NewClient()
	c.ObjectStorageURL = "https://object-storage.c3j1.conoha.io/v1"
	c.TenantID = "tenant-abc"

	path := c.objectStoragePath()
	expected := "https://object-storage.c3j1.conoha.io/v1/AUTH_tenant-abc"
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}

// ============================================================
// Account Operations
// ============================================================

func TestGetAccountInfo_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.Header().Set("X-Account-Container-Count", "5")
		w.Header().Set("X-Account-Object-Count", "42")
		w.Header().Set("X-Account-Bytes-Used", "1073741824")
		w.Header().Set("X-Account-Bytes-Used-Actual", "1073741824")
		w.Header().Set("X-Account-Meta-Quota-Bytes", "10737418240")
		w.WriteHeader(204)
	})
	defer server.Close()

	info, err := client.GetAccountInfo(context.Background())
	assertNoError(t, err)

	if capturedMethod != http.MethodHead {
		t.Errorf("Method = %q, want HEAD", capturedMethod)
	}
	if info.ContainerCount != 5 {
		t.Errorf("ContainerCount = %d", info.ContainerCount)
	}
	if info.ObjectCount != 42 {
		t.Errorf("ObjectCount = %d", info.ObjectCount)
	}
	if info.BytesUsed != 1073741824 {
		t.Errorf("BytesUsed = %d", info.BytesUsed)
	}
	if info.QuotaBytes != 10737418240 {
		t.Errorf("QuotaBytes = %d", info.QuotaBytes)
	}
}

func TestGetAccountInfo_Error(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	})
	defer server.Close()

	_, err := client.GetAccountInfo(context.Background())
	assertAPIError(t, err, 401)
}

func TestSetAccountQuota_Success(t *testing.T) {
	var capturedHeader string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Account-Meta-Quota-Giga-Bytes")
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.SetAccountQuota(context.Background(), "200")
	assertNoError(t, err)

	if capturedHeader != "200" {
		t.Errorf("X-Account-Meta-Quota-Giga-Bytes = %q", capturedHeader)
	}
}

// ============================================================
// Container Operations
// ============================================================

func TestListContainers_Success(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`[{"name":"container1","count":10,"bytes":1024}]`))
	})
	defer server.Close()

	containers, err := client.ListContainers(context.Background())
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "format=json") {
		t.Errorf("URI should contain format=json: %q", capturedURI)
	}
	if len(containers) != 1 || containers[0].Name != "container1" {
		t.Errorf("unexpected containers: %+v", containers)
	}
}

func TestCreateContainer_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(201)
	})
	defer server.Close()

	err := client.CreateContainer(context.Background(), "new-container")
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q, want PUT", capturedMethod)
	}
	if !strings.Contains(capturedPath, "new-container") {
		t.Errorf("Path should contain container name: %q", capturedPath)
	}
}

func TestDeleteContainer_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteContainer(context.Background(), "old-container")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
}

func TestGetContainerInfo_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.Header().Set("X-Container-Object-Count", "7")
		w.Header().Set("X-Container-Bytes-Used", "4096")
		w.Header().Set("X-Container-Read", ".r:*")
		w.Header().Set("X-Container-Meta-Color", "blue")
		w.WriteHeader(204)
	})
	defer server.Close()

	info, err := client.GetContainerInfo(context.Background(), "mycontainer")
	assertNoError(t, err)

	if capturedMethod != http.MethodHead {
		t.Errorf("Method = %q, want HEAD", capturedMethod)
	}
	if info.ObjectCount != 7 {
		t.Errorf("ObjectCount = %d", info.ObjectCount)
	}
	if info.BytesUsed != 4096 {
		t.Errorf("BytesUsed = %d", info.BytesUsed)
	}
	if info.ReadACL != ".r:*" {
		t.Errorf("ReadACL = %q", info.ReadACL)
	}
	if info.Metadata["color"] != "blue" {
		t.Errorf("Metadata[color] = %q", info.Metadata["color"])
	}
}

func TestGetContainerInfo_Error(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	})
	defer server.Close()

	_, err := client.GetContainerInfo(context.Background(), "missing-container")
	assertAPIError(t, err, 404)
}

// ============================================================
// Object Operations
// ============================================================

func TestListObjects_Success(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`[{"name":"file1.txt","bytes":1024,"hash":"abc123"}]`))
	})
	defer server.Close()

	objects, err := client.ListObjects(context.Background(), "mycontainer", nil)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "format=json") {
		t.Errorf("URI should contain format=json: %q", capturedURI)
	}
	if len(objects) != 1 || objects[0].Name != "file1.txt" {
		t.Errorf("unexpected objects: %+v", objects)
	}
}

func TestListObjects_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	})
	defer server.Close()

	opts := &ListObjectsOptions{Limit: 10, Prefix: "docs/", Delimiter: "/"}
	_, err := client.ListObjects(context.Background(), "mycontainer", opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=10") {
		t.Errorf("URI should contain limit=10: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "prefix=docs") {
		t.Errorf("URI should contain prefix: %q", capturedURI)
	}
}

func TestUploadObject_Success(t *testing.T) {
	var capturedMethod string
	var capturedBody string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		w.WriteHeader(201)
	})
	defer server.Close()

	data := strings.NewReader("file content here")
	err := client.UploadObject(context.Background(), "container", "test.txt", data)
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q, want PUT", capturedMethod)
	}
	if capturedBody != "file content here" {
		t.Errorf("body = %q", capturedBody)
	}
}

func TestDownloadObject_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("downloaded content"))
	})
	defer server.Close()

	reader, err := client.DownloadObject(context.Background(), "container", "file.txt")
	assertNoError(t, err)
	defer reader.Close()

	data, _ := io.ReadAll(reader)
	if string(data) != "downloaded content" {
		t.Errorf("data = %q", string(data))
	}
}

func TestDownloadObject_Error(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	})
	defer server.Close()

	_, err := client.DownloadObject(context.Background(), "container", "missing.txt")
	assertAPIError(t, err, 404)
}

func TestDeleteObject_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteObject(context.Background(), "container", "file.txt")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
}

func TestGetObjectInfo_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.Header().Set("Content-Length", "1024")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("ETag", "abc123")
		w.Header().Set("Last-Modified", "Mon, 01 Jan 2024 00:00:00 GMT")
		w.Header().Set("X-Delete-At", "1700000000")
		w.Header().Set("X-Object-Meta-Owner", "sdk")
		w.WriteHeader(200)
	})
	defer server.Close()

	info, err := client.GetObjectInfo(context.Background(), "mycontainer", "myfile.txt")
	assertNoError(t, err)

	if capturedMethod != http.MethodHead {
		t.Errorf("Method = %q, want HEAD", capturedMethod)
	}
	if info.ContentLength != 1024 {
		t.Errorf("ContentLength = %d", info.ContentLength)
	}
	if info.ContentType != "text/plain" {
		t.Errorf("ContentType = %q", info.ContentType)
	}
	if info.ETag != "abc123" {
		t.Errorf("ETag = %q", info.ETag)
	}
	if info.DeleteAt != 1700000000 {
		t.Errorf("DeleteAt = %d", info.DeleteAt)
	}
	if info.Metadata["owner"] != "sdk" {
		t.Errorf("Metadata[owner] = %q", info.Metadata["owner"])
	}
}

func TestGetObjectInfo_Error(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	})
	defer server.Close()

	_, err := client.GetObjectInfo(context.Background(), "container", "missing.txt")
	assertAPIError(t, err, 404)
}

func TestCopyObject_Success(t *testing.T) {
	var capturedMethod string
	var capturedDestination string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedDestination = r.Header.Get("Destination")
		w.WriteHeader(201)
	})
	defer server.Close()

	err := client.CopyObject(context.Background(), "src-container", "src-obj", "dst-container", "dst-obj")
	assertNoError(t, err)

	if capturedMethod != "COPY" {
		t.Errorf("Method = %q, want COPY", capturedMethod)
	}
	if capturedDestination != "dst-container/dst-obj" {
		t.Errorf("Destination = %q", capturedDestination)
	}
}

func TestScheduleObjectDeletion_Success(t *testing.T) {
	var capturedHeader string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Delete-At")
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.ScheduleObjectDeletion(context.Background(), "container", "file.txt", 1700000000)
	assertNoError(t, err)

	if capturedHeader != fmt.Sprintf("%d", 1700000000) {
		t.Errorf("X-Delete-At = %q", capturedHeader)
	}
}

func TestScheduleObjectDeletionAfter_Success(t *testing.T) {
	var capturedHeader string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Delete-After")
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.ScheduleObjectDeletionAfter(context.Background(), "container", "file.txt", 3600)
	assertNoError(t, err)

	if capturedHeader != "3600" {
		t.Errorf("X-Delete-After = %q", capturedHeader)
	}
}

// ============================================================
// Container Configuration
// ============================================================

func TestEnableVersioning_Success(t *testing.T) {
	var capturedHeader string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Versions-Location")
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.EnableVersioning(context.Background(), "mycontainer", "versions-container")
	assertNoError(t, err)

	if capturedHeader != "versions-container" {
		t.Errorf("X-Versions-Location = %q", capturedHeader)
	}
}

func TestDisableVersioning_Success(t *testing.T) {
	var headerPresent bool
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		_, headerPresent = r.Header["X-Remove-Versions-Location"]
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DisableVersioning(context.Background(), "mycontainer")
	assertNoError(t, err)

	if !headerPresent {
		t.Error("X-Remove-Versions-Location header should be present")
	}
}

func TestEnableWebPublishing_Success(t *testing.T) {
	var capturedHeader string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Container-Read")
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.EnableWebPublishing(context.Background(), "mycontainer")
	assertNoError(t, err)

	if capturedHeader != ".r:*" {
		t.Errorf("X-Container-Read = %q, want '.r:*'", capturedHeader)
	}
}

func TestSetTempURLKey_Success(t *testing.T) {
	var capturedHeader string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Account-Meta-Temp-URL-Key")
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.SetTempURLKey(context.Background(), "my-secret-key")
	assertNoError(t, err)

	if capturedHeader != "my-secret-key" {
		t.Errorf("X-Account-Meta-Temp-URL-Key = %q", capturedHeader)
	}
}

func TestRemoveTempURLKey_Success(t *testing.T) {
	var headerPresent bool
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		for k := range r.Header {
			if strings.EqualFold(k, "X-Remove-Account-Meta-Temp-URL-Key") {
				headerPresent = true
				break
			}
		}
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.RemoveTempURLKey(context.Background())
	assertNoError(t, err)

	if !headerPresent {
		t.Error("X-Remove-Account-Meta-Temp-URL-Key header should be present")
	}
}

func TestGenerateTempURL_Success(t *testing.T) {
	c := NewClient()
	c.ObjectStorageURL = "https://object-storage.c3j1.conoha.io/v1"
	c.TenantID = "tenant-abc"

	key := "my-secret-key"
	expires := int64(1700000000)

	tempURL, err := c.GenerateTempURL("get", "mycontainer", "path/to file.txt", key, expires)
	assertNoError(t, err)

	u, err := url.Parse(tempURL)
	assertNoError(t, err)

	if u.EscapedPath() != "/v1/AUTH_tenant-abc/mycontainer/path/to%20file.txt" {
		t.Errorf("EscapedPath = %q", u.EscapedPath())
	}
	if u.Query().Get("temp_url_expires") != "1700000000" {
		t.Errorf("temp_url_expires = %q", u.Query().Get("temp_url_expires"))
	}

	payload := "GET\n1700000000\n/v1/AUTH_tenant-abc/mycontainer/path/to%20file.txt"
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if u.Query().Get("temp_url_sig") != expectedSig {
		t.Errorf("temp_url_sig = %q, want %q", u.Query().Get("temp_url_sig"), expectedSig)
	}
}

func TestGenerateTempURL_Validation(t *testing.T) {
	c := NewClient()
	c.ObjectStorageURL = "https://object-storage.c3j1.conoha.io/v1"
	c.TenantID = "tenant-abc"

	_, err := c.GenerateTempURL("", "container", "file.txt", "key", 1700000000)
	assertError(t, err)

	_, err = c.GenerateTempURL("GET", "container", "", "key", 1700000000)
	assertError(t, err)

	_, err = c.GenerateTempURL("GET", "container", "file.txt", "", 1700000000)
	assertError(t, err)

	_, err = c.GenerateTempURL("GET", "container", "file.txt", "key", 0)
	assertError(t, err)
}

// ============================================================
// parseHeaderInt64
// ============================================================

func TestParseHeaderInt64(t *testing.T) {
	h := http.Header{}
	h.Set("X-Count", "42")
	h.Set("X-Empty", "")

	var count int64
	parseHeaderInt64(h, "X-Count", &count)
	if count != 42 {
		t.Errorf("count = %d, want 42", count)
	}

	var empty int64
	parseHeaderInt64(h, "X-Empty", &empty)
	if empty != 0 {
		t.Errorf("empty = %d, want 0", empty)
	}

	var missing int64
	parseHeaderInt64(h, "X-Missing", &missing)
	if missing != 0 {
		t.Errorf("missing = %d, want 0", missing)
	}
}

// ============================================================
// DisableWebPublishing
// ============================================================

func TestDisableWebPublishing_Success(t *testing.T) {
	var capturedHeader string
	var headerPresent bool
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Container-Read")
		_, headerPresent = r.Header["X-Container-Read"]
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DisableWebPublishing(context.Background(), "mycontainer")
	assertNoError(t, err)

	if !headerPresent {
		t.Error("X-Container-Read header should be present")
	}
	if capturedHeader != "" {
		t.Errorf("X-Container-Read = %q, want empty", capturedHeader)
	}
}

// ============================================================
// CreateDLOManifest
// ============================================================

func TestCreateDLOManifest_Success(t *testing.T) {
	var capturedMethod string
	var capturedManifestHeader string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedManifestHeader = r.Header.Get("X-Object-Manifest")
		w.WriteHeader(201)
	})
	defer server.Close()

	err := client.CreateDLOManifest(context.Background(), "container", "manifest.dat", "segments", "prefix_")
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q, want PUT", capturedMethod)
	}
	if capturedManifestHeader != "segments/prefix_" {
		t.Errorf("X-Object-Manifest = %q", capturedManifestHeader)
	}
}

// ============================================================
// CreateSLOManifest
// ============================================================

func TestCreateSLOManifest_Success(t *testing.T) {
	var capturedMethod string
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(201)
	})
	defer server.Close()

	segments := []SLOSegment{
		{Path: "segments/part1", Etag: "abc", SizeBytes: 1024},
		{Path: "segments/part2", Etag: "def", SizeBytes: 2048},
	}
	err := client.CreateSLOManifest(context.Background(), "container", "bigfile.dat", segments)
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q, want PUT", capturedMethod)
	}
	if !strings.Contains(capturedURI, "multipart-manifest=put") {
		t.Errorf("URI should contain multipart-manifest=put: %q", capturedURI)
	}
}
