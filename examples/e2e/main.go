// End-to-end example: create a VPS, manage lifecycle, and enable daily backup.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	conoha "github.com/leonunix/conohav3-golang-sdk"
)

// waitForServerStatus polls until the server reaches the target status.
func waitForServerStatus(ctx context.Context, client *conoha.Client, serverID, target string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		server, err := client.GetServer(ctx, serverID)
		if err != nil {
			return err
		}
		fmt.Printf("  status: %s\n", server.Status)
		if server.Status == target {
			return nil
		}
		if server.Status == "ERROR" {
			return fmt.Errorf("server entered ERROR state")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for status %q (last: %q)", target, server.Status)
		}
		time.Sleep(5 * time.Second)
	}
}

// waitForVolumeStatus polls until the volume reaches the target status.
func waitForVolumeStatus(ctx context.Context, client *conoha.Client, volumeID, target string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		vol, err := client.GetVolume(ctx, volumeID)
		if err != nil {
			return err
		}
		fmt.Printf("  status: %s\n", vol.Status)
		if vol.Status == target {
			return nil
		}
		if vol.Status == "error" {
			return fmt.Errorf("volume entered error state")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for status %q (last: %q)", target, vol.Status)
		}
		time.Sleep(5 * time.Second)
	}
}

func main() {
	ctx := context.Background()

	// ── Initialize client ──────────────────────────────────
	fmt.Println("=== Initializing ConoHa client ===")
	userID := os.Getenv("CONOHA_USER_ID")
	password := os.Getenv("CONOHA_PASSWORD")
	tenantID := os.Getenv("CONOHA_TENANT_ID")

	if userID == "" || password == "" || tenantID == "" {
		log.Fatal("Please set CONOHA_USER_ID, CONOHA_PASSWORD, and CONOHA_TENANT_ID")
	}

	client := conoha.NewClient()
	token, err := client.Authenticate(ctx, userID, password, tenantID)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Printf("Authenticated! Token expires at: %s\n", token.ExpiresAt)

	// ── Find 1GB flavor ────────────────────────────────────
	fmt.Println("\n=== Finding 1GB flavor ===")
	flavors, err := client.ListFlavorsDetail(ctx)
	if err != nil {
		log.Fatalf("List flavors failed: %v", err)
	}
	var flavorID string
	for _, f := range flavors {
		if f.Name == "g2l-t-c2m1" {
			flavorID = f.ID
			fmt.Printf("Flavor: %s (id=%s, ram=%dMB)\n", f.Name, f.ID, f.RAM)
			break
		}
	}
	if flavorID == "" {
		log.Fatal("1GB flavor (g2l-t-c2m1) not found")
	}

	// ── Find Ubuntu 24.04 image ────────────────────────────
	fmt.Println("\n=== Finding Ubuntu 24.04 image ===")
	images, err := client.ListImages(ctx, &conoha.ListImagesOptions{Visibility: "public"})
	if err != nil {
		log.Fatalf("List images failed: %v", err)
	}
	var imageID string
	for _, img := range images {
		if img.Name == "vmi-ubuntu-24.04-amd64" {
			imageID = img.ID
			fmt.Printf("Image: %s (id=%s)\n", img.Name, img.ID)
			break
		}
	}
	if imageID == "" {
		log.Fatal("Image 'vmi-ubuntu-24.04-amd64' not found")
	}

	// ── Create 100GB boot volume ───────────────────────────
	fmt.Println("\n=== Creating 100GB boot volume ===")
	vol, err := client.CreateVolume(ctx, conoha.CreateVolumeRequest{
		Size:     100,
		Name:     "sdk-test-daily-backup-boot",
		ImageRef: imageID,
	})
	if err != nil {
		log.Fatalf("Create volume failed: %v", err)
	}
	volumeID := vol.ID
	fmt.Printf("Volume created: %s\n", volumeID)

	fmt.Println("Waiting for volume to become available...")
	if err := waitForVolumeStatus(ctx, client, volumeID, "available", 3*time.Minute); err != nil {
		log.Fatalf("Volume wait failed: %v", err)
	}
	fmt.Println("Volume is available.")

	// ── Create server ──────────────────────────────────────
	fmt.Println("\n=== Creating server (1GB plan) ===")
	server, err := client.CreateServer(ctx, conoha.CreateServerRequest{
		FlavorRef: flavorID,
		AdminPass: "SdkTest#2026",
		BlockDeviceMapping: []conoha.BlockDeviceMap{
			{UUID: volumeID},
		},
		Metadata: map[string]string{
			"instance_name_tag": "sdk-test-daily-backup",
		},
	})
	if err != nil {
		log.Fatalf("Create server failed: %v", err)
	}
	serverID := server.ID
	fmt.Printf("Server created: %s\n", serverID)

	fmt.Println("Waiting for server to become ACTIVE...")
	if err := waitForServerStatus(ctx, client, serverID, "ACTIVE", 5*time.Minute); err != nil {
		log.Fatalf("Server wait failed: %v", err)
	}
	fmt.Println("Server is ACTIVE.")

	// ── Stop server ────────────────────────────────────────
	fmt.Println("\n=== Stopping server ===")
	if err := client.StopServer(ctx, serverID); err != nil {
		log.Fatalf("Stop server failed: %v", err)
	}
	fmt.Println("Stop requested. Waiting for SHUTOFF...")
	if err := waitForServerStatus(ctx, client, serverID, "SHUTOFF", 3*time.Minute); err != nil {
		log.Fatalf("Server stop wait failed: %v", err)
	}
	fmt.Println("Server is SHUTOFF.")

	// ── Start server ───────────────────────────────────────
	fmt.Println("\n=== Starting server ===")
	if err := client.StartServer(ctx, serverID); err != nil {
		log.Fatalf("Start server failed: %v", err)
	}
	fmt.Println("Start requested. Waiting for ACTIVE...")
	if err := waitForServerStatus(ctx, client, serverID, "ACTIVE", 3*time.Minute); err != nil {
		log.Fatalf("Server start wait failed: %v", err)
	}
	fmt.Println("Server is ACTIVE again.")

	// ── Enable daily backup ────────────────────────────────
	fmt.Println("\n=== Enabling daily backup (retention=14) ===")
	backup, err := client.EnableAutoBackup(ctx, serverID, &conoha.EnableAutoBackupOptions{
		Schedule:  "daily",
		Retention: 14,
	})
	if err != nil {
		log.Fatalf("Enable daily backup failed: %v", err)
	}
	fmt.Printf("Daily backup enabled: id=%s, status=%s\n", backup.ID, backup.Status)

	fmt.Println("\n=== All steps completed successfully! ===")
	fmt.Printf("Server ID: %s\n", serverID)

	// ── Cleanup prompt ─────────────────────────────────────
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("WARNING: The server and volume are still running and incurring charges.")
	fmt.Printf("  Server ID: %s\n", serverID)
	fmt.Printf("  Volume ID: %s\n", volumeID)
	fmt.Print("\nPress Enter to delete them, or type 'keep' to keep: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	reply := strings.TrimSpace(scanner.Text())

	if strings.EqualFold(reply, "keep") {
		fmt.Println("Keeping resources. Remember to delete them manually later!")
		return
	}

	fmt.Println("Disabling auto-backup...")
	if err := client.DisableAutoBackup(ctx, serverID); err != nil {
		fmt.Printf("Warning: disable backup failed: %v\n", err)
	}

	fmt.Println("Deleting server...")
	if err := client.DeleteServer(ctx, serverID); err != nil {
		fmt.Printf("Warning: delete server failed: %v\n", err)
	} else {
		fmt.Println("Waiting for server to be deleted...")
		deadline := time.Now().Add(2 * time.Minute)
		for time.Now().Before(deadline) {
			s, err := client.GetServer(ctx, serverID)
			if err != nil {
				var apiErr *conoha.APIError
				if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
					fmt.Println("Server deleted.")
					break
				}
				fmt.Printf("  error checking: %v\n", err)
				break
			}
			fmt.Printf("  status: %s\n", s.Status)
			time.Sleep(5 * time.Second)
		}
	}

	fmt.Println("Deleting volume...")
	if err := client.DeleteVolume(ctx, volumeID, false); err != nil {
		fmt.Printf("Could not delete volume (may already be gone): %v\n", err)
	} else {
		fmt.Println("Volume deleted.")
	}

	fmt.Println("\nCleanup complete.")
}
