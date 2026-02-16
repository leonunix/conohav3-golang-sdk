package conoha

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// ============================================================
// Image CRUD
// ============================================================

func TestListImages_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"images":[{"id":"img-1","name":"Ubuntu 22.04","status":"active"}]}`))
	})
	defer server.Close()

	images, err := client.ListImages(context.Background(), nil)
	assertNoError(t, err)

	if capturedPath != "/images" {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(images) != 1 || images[0].Name != "Ubuntu 22.04" {
		t.Errorf("unexpected images: %+v", images)
	}
}

func TestListImages_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"images":[]}`))
	})
	defer server.Close()

	opts := &ListImagesOptions{Limit: 5, Visibility: "public", OSType: "linux"}
	_, err := client.ListImages(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=5") {
		t.Errorf("URI should contain limit=5: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "visibility=public") {
		t.Errorf("URI should contain visibility=public: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "os_type=linux") {
		t.Errorf("URI should contain os_type=linux: %q", capturedURI)
	}
}

func TestGetImage_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		// GetImage unmarshals directly into Image (no wrapper)
		w.Write([]byte(`{"id":"img-123","name":"CentOS 9","status":"active","size":1073741824}`))
	})
	defer server.Close()

	img, err := client.GetImage(context.Background(), "img-123")
	assertNoError(t, err)

	if img.ID != "img-123" || img.Name != "CentOS 9" {
		t.Errorf("unexpected image: %+v", img)
	}
}

func TestDeleteImage_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteImage(context.Background(), "img-123")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/images/img-123" {
		t.Errorf("Path = %q", capturedPath)
	}
}

// ============================================================
// ISO Image Upload
// ============================================================

func TestUploadISOImage_Success(t *testing.T) {
	var capturedContentType string
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	data := strings.NewReader("fake iso data")
	err := client.UploadISOImage(context.Background(), "img-123", data)
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q", capturedMethod)
	}
	if capturedPath != "/images/img-123/file" {
		t.Errorf("Path = %q", capturedPath)
	}
	if capturedContentType != "application/octet-stream" {
		t.Errorf("Content-Type = %q", capturedContentType)
	}
}

func TestUploadISOImage_Error(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(413)
		w.Write([]byte("Request Entity Too Large"))
	})
	defer server.Close()

	data := strings.NewReader("fake iso data")
	err := client.UploadISOImage(context.Background(), "img-123", data)

	assertAPIError(t, err, 413)
}

// ============================================================
// Image Quota
// ============================================================

func TestGetImageQuota_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"quota":{"image_size":"500GB"}}`))
	})
	defer server.Close()

	quota, err := client.GetImageQuota(context.Background())
	assertNoError(t, err)

	if quota.ImageSize != "500GB" {
		t.Errorf("ImageSize = %q", quota.ImageSize)
	}
}

func TestSetImageQuota_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(200)
		w.Write([]byte(`{"quota":{"image_size":"550GB"}}`))
	})
	defer server.Close()

	quota, err := client.SetImageQuota(context.Background(), "550GB")
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q", capturedMethod)
	}
	if quota.ImageSize != "550GB" {
		t.Errorf("ImageSize = %q", quota.ImageSize)
	}
}

func TestCreateISOImage_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte(`{"id":"new-iso","name":"test.iso","disk_format":"iso","status":"queued"}`))
	})
	defer server.Close()

	img, err := client.CreateISOImage(context.Background(), "test.iso")
	assertNoError(t, err)

	if img.ID != "new-iso" || img.DiskFormat != "iso" {
		t.Errorf("unexpected image: %+v", img)
	}
}

// ============================================================
// Image Usage
// ============================================================

func TestGetImageUsage_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"images":{"size":5368709120}}`))
	})
	defer server.Close()

	usage, err := client.GetImageUsage(context.Background())
	assertNoError(t, err)

	if capturedPath != "/images/total" {
		t.Errorf("Path = %q", capturedPath)
	}
	if usage.Size != 5368709120 {
		t.Errorf("Size = %d", usage.Size)
	}
}
