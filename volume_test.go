package conoha

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// ============================================================
// Volume CRUD
// ============================================================

func TestListVolumes_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"volumes":[{"id":"vol-1","name":"volume1","size":50}]}`))
	})
	defer server.Close()

	volumes, err := client.ListVolumes(context.Background(), nil)
	assertNoError(t, err)

	// Path should contain tenantID
	if !strings.Contains(capturedPath, "/test-tenant-id/volumes") {
		t.Errorf("Path should contain tenantID: %q", capturedPath)
	}
	if len(volumes) != 1 || volumes[0].ID != "vol-1" {
		t.Errorf("unexpected volumes: %+v", volumes)
	}
}

func TestListVolumes_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"volumes":[]}`))
	})
	defer server.Close()

	opts := &ListVolumesOptions{Limit: 10, Sort: "created_at:desc"}
	_, err := client.ListVolumes(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=10") {
		t.Errorf("URI should contain limit=10: %q", capturedURI)
	}
}

func TestGetVolume_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"volume":{"id":"vol-123","name":"myvolume","size":100,"status":"available"}}`))
	})
	defer server.Close()

	vol, err := client.GetVolume(context.Background(), "vol-123")
	assertNoError(t, err)

	if !strings.Contains(capturedPath, "/test-tenant-id/volumes/vol-123") {
		t.Errorf("Path = %q", capturedPath)
	}
	if vol.ID != "vol-123" || vol.Size != 100 {
		t.Errorf("unexpected volume: %+v", vol)
	}
}

func TestCreateVolume_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
		w.Write([]byte(`{"volume":{"id":"new-vol","name":"newvolume","size":50}}`))
	})
	defer server.Close()

	opts := CreateVolumeRequest{Size: 50, Name: "newvolume", VolumeType: "c3j1-ds02-boot"}
	vol, err := client.CreateVolume(context.Background(), opts)
	assertNoError(t, err)

	if _, ok := body["volume"]; !ok {
		t.Error("body should contain 'volume' key")
	}
	if vol.ID != "new-vol" {
		t.Errorf("ID = %q", vol.ID)
	}
}

func TestDeleteVolume_Success(t *testing.T) {
	var capturedMethod string
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.DeleteVolume(context.Background(), "vol-123", false)
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if strings.Contains(capturedURI, "force=true") {
		t.Errorf("URI should NOT contain force=true: %q", capturedURI)
	}
}

func TestDeleteVolume_Force(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(202)
	})
	defer server.Close()

	err := client.DeleteVolume(context.Background(), "vol-123", true)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "force=true") {
		t.Errorf("URI should contain force=true: %q", capturedURI)
	}
}

func TestUpdateVolume_Success(t *testing.T) {
	var capturedMethod string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(200)
		w.Write([]byte(`{"volume":{"id":"vol-123","name":"updated"}}`))
	})
	defer server.Close()

	desc := "new description"
	vol, err := client.UpdateVolume(context.Background(), "vol-123", "updated", &desc)
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q", capturedMethod)
	}
	if vol.Name != "updated" {
		t.Errorf("Name = %q", vol.Name)
	}
}

func TestSaveVolumeAsImage_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
		w.Write([]byte(`{"os-volume_upload_image":{"id":"vol-123","image_id":"img-456","image_name":"myimage"}}`))
	})
	defer server.Close()

	result, err := client.SaveVolumeAsImage(context.Background(), "vol-123", "myimage")
	assertNoError(t, err)

	if _, ok := body["os-volume_upload_image"]; !ok {
		t.Error("body should contain 'os-volume_upload_image'")
	}
	if result.ImageID != "img-456" {
		t.Errorf("ImageID = %q", result.ImageID)
	}
}

// ============================================================
// Volume Types
// ============================================================

func TestListVolumeTypes_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"volume_types":[{"id":"vt-1","name":"ssd"}]}`))
	})
	defer server.Close()

	types, err := client.ListVolumeTypes(context.Background())
	assertNoError(t, err)

	if !strings.Contains(capturedPath, "/test-tenant-id/types") {
		t.Errorf("Path = %q", capturedPath)
	}
	if len(types) != 1 || types[0].Name != "ssd" {
		t.Errorf("unexpected types: %+v", types)
	}
}

func TestGetVolumeType_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"volume_type":{"id":"vt-1","name":"ssd","is_public":true}}`))
	})
	defer server.Close()

	vt, err := client.GetVolumeType(context.Background(), "vt-1")
	assertNoError(t, err)

	if vt.Name != "ssd" || !vt.IsPublic {
		t.Errorf("unexpected volume type: %+v", vt)
	}
}

// ============================================================
// Backups
// ============================================================

func TestListBackups_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"backups":[{"id":"bk-1","name":"backup1","status":"available"}]}`))
	})
	defer server.Close()

	backups, err := client.ListBackups(context.Background(), nil)
	assertNoError(t, err)

	if len(backups) != 1 || backups[0].ID != "bk-1" {
		t.Errorf("unexpected backups: %+v", backups)
	}
}

func TestListBackups_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"backups":[]}`))
	})
	defer server.Close()

	opts := &ListBackupsOptions{Limit: 10, Offset: 5, Sort: "created_at:desc"}
	_, err := client.ListBackups(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=10") {
		t.Errorf("URI should contain limit=10: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "offset=5") {
		t.Errorf("URI should contain offset=5: %q", capturedURI)
	}
}

func TestGetBackup_Success(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"backup":{"id":"bk-123","name":"mybackup","size":100}}`))
	})
	defer server.Close()

	backup, err := client.GetBackup(context.Background(), "bk-123")
	assertNoError(t, err)

	if backup.ID != "bk-123" || backup.Size != 100 {
		t.Errorf("unexpected backup: %+v", backup)
	}
}

func TestEnableAutoBackup_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
		w.Write([]byte(`{"backup":{"id":"bk-new","status":"creating"}}`))
	})
	defer server.Close()

	backup, err := client.EnableAutoBackup(context.Background(), "srv-123", nil)
	assertNoError(t, err)

	if _, ok := body["backup"]; !ok {
		t.Error("body should contain 'backup'")
	}
	if backup.ID != "bk-new" {
		t.Errorf("ID = %q", backup.ID)
	}
}

func TestEnableAutoBackup_DefaultNoExtraParams(t *testing.T) {
	var rawBody map[string]map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &rawBody)
		w.WriteHeader(201)
		w.Write([]byte(`{"backup":{"id":"bk-1","status":"creating"}}`))
	})
	defer server.Close()

	_, err := client.EnableAutoBackup(context.Background(), "srv-123", nil)
	assertNoError(t, err)

	backupBody := rawBody["backup"]
	if backupBody["instance_uuid"] != "srv-123" {
		t.Errorf("instance_uuid = %v", backupBody["instance_uuid"])
	}
	if _, ok := backupBody["schedule"]; ok {
		t.Error("nil opts should not include schedule")
	}
	if _, ok := backupBody["retention"]; ok {
		t.Error("nil opts should not include retention")
	}
}

func TestEnableAutoBackup_WithDailySchedule(t *testing.T) {
	var rawBody map[string]map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &rawBody)
		w.WriteHeader(201)
		w.Write([]byte(`{"backup":{"id":"bk-daily","status":"creating"}}`))
	})
	defer server.Close()

	backup, err := client.EnableAutoBackup(context.Background(), "srv-123", &EnableAutoBackupOptions{
		Schedule:  "daily",
		Retention: 14,
	})
	assertNoError(t, err)

	backupBody := rawBody["backup"]
	if backupBody["schedule"] != "daily" {
		t.Errorf("schedule = %v", backupBody["schedule"])
	}
	if int(backupBody["retention"].(float64)) != 14 {
		t.Errorf("retention = %v", backupBody["retention"])
	}
	if backup.ID != "bk-daily" {
		t.Errorf("ID = %q", backup.ID)
	}
}

func TestEnableAutoBackup_ScheduleOnly(t *testing.T) {
	var rawBody map[string]map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &rawBody)
		w.WriteHeader(201)
		w.Write([]byte(`{"backup":{"id":"bk-2","status":"creating"}}`))
	})
	defer server.Close()

	_, err := client.EnableAutoBackup(context.Background(), "srv-123", &EnableAutoBackupOptions{
		Schedule: "daily",
	})
	assertNoError(t, err)

	backupBody := rawBody["backup"]
	if backupBody["schedule"] != "daily" {
		t.Errorf("schedule = %v", backupBody["schedule"])
	}
	if _, ok := backupBody["retention"]; ok {
		t.Error("zero retention should not be included")
	}
}

// ============================================================
// UpdateBackupRetention
// ============================================================

func TestUpdateBackupRetention_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	var rawBody map[string]map[string]interface{}
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		readJSONBody(t, r, &rawBody)
		w.WriteHeader(200)
		w.Write([]byte(`{"backup":{"id":"bk-1","status":"available"}}`))
	})
	defer server.Close()

	backup, err := client.UpdateBackupRetention(context.Background(), "srv-123", 30)
	assertNoError(t, err)

	if capturedMethod != http.MethodPut {
		t.Errorf("Method = %q, want PUT", capturedMethod)
	}
	if !strings.Contains(capturedPath, "/test-tenant-id/backups/srv-123") {
		t.Errorf("Path = %q", capturedPath)
	}
	backupBody := rawBody["backup"]
	if int(backupBody["retention"].(float64)) != 30 {
		t.Errorf("retention = %v", backupBody["retention"])
	}
	if backup.ID != "bk-1" {
		t.Errorf("ID = %q", backup.ID)
	}
}

func TestUpdateBackupRetention_NotFound(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":{"message":"Daily backup not found"}}`))
	})
	defer server.Close()

	_, err := client.UpdateBackupRetention(context.Background(), "srv-123", 14)
	assertAPIError(t, err, 404)
}

func TestRestoreBackup_Success(t *testing.T) {
	var body map[string]json.RawMessage
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		readJSONBody(t, r, &body)
		w.WriteHeader(202)
		w.Write([]byte(`{"restore":{"backup_id":"bk-123","volume_id":"vol-456"}}`))
	})
	defer server.Close()

	result, err := client.RestoreBackup(context.Background(), "bk-123", "vol-456")
	assertNoError(t, err)

	if _, ok := body["restore"]; !ok {
		t.Error("body should contain 'restore'")
	}
	if result.BackupID != "bk-123" || result.VolumeID != "vol-456" {
		t.Errorf("unexpected restore: %+v", result)
	}
}

// ============================================================
// ListVolumesDetail
// ============================================================

func TestListVolumesDetail_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"volumes":[{"id":"vol-1","name":"volume1","size":50,"status":"available"}]}`))
	})
	defer server.Close()

	volumes, err := client.ListVolumesDetail(context.Background(), nil)
	assertNoError(t, err)

	if !strings.Contains(capturedPath, "/test-tenant-id/volumes/detail") {
		t.Errorf("Path should contain /volumes/detail: %q", capturedPath)
	}
	if len(volumes) != 1 || volumes[0].ID != "vol-1" {
		t.Errorf("unexpected volumes: %+v", volumes)
	}
}

func TestListVolumesDetail_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"volumes":[]}`))
	})
	defer server.Close()

	opts := &ListVolumesOptions{Limit: 10, Offset: 5, Sort: "name:asc"}
	_, err := client.ListVolumesDetail(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=10") {
		t.Errorf("URI should contain limit=10: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "offset=5") {
		t.Errorf("URI should contain offset=5: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "sort=name") {
		t.Errorf("URI should contain sort: %q", capturedURI)
	}
}

// ============================================================
// ListBackupsDetail
// ============================================================

func TestListBackupsDetail_Success(t *testing.T) {
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"backups":[{"id":"bk-1","name":"backup1","status":"available","size":100}]}`))
	})
	defer server.Close()

	backups, err := client.ListBackupsDetail(context.Background(), nil)
	assertNoError(t, err)

	if !strings.Contains(capturedPath, "/test-tenant-id/backups/detail") {
		t.Errorf("Path should contain /backups/detail: %q", capturedPath)
	}
	if len(backups) != 1 || backups[0].ID != "bk-1" {
		t.Errorf("unexpected backups: %+v", backups)
	}
}

func TestListBackupsDetail_WithOptions(t *testing.T) {
	var capturedURI string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.URL.RequestURI()
		w.WriteHeader(200)
		w.Write([]byte(`{"backups":[]}`))
	})
	defer server.Close()

	opts := &ListBackupsOptions{Limit: 5, Offset: 10}
	_, err := client.ListBackupsDetail(context.Background(), opts)
	assertNoError(t, err)

	if !strings.Contains(capturedURI, "limit=5") {
		t.Errorf("URI should contain limit=5: %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "offset=10") {
		t.Errorf("URI should contain offset=10: %q", capturedURI)
	}
}

// ============================================================
// DisableAutoBackup
// ============================================================

func TestDisableAutoBackup_Success(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DisableAutoBackup(context.Background(), "srv-123")
	assertNoError(t, err)

	if capturedMethod != http.MethodDelete {
		t.Errorf("Method = %q", capturedMethod)
	}
	if !strings.Contains(capturedPath, "/test-tenant-id/backups/srv-123") {
		t.Errorf("Path = %q", capturedPath)
	}
}
