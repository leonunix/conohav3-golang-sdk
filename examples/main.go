package main

import (
	"context"
	"fmt"
	"log"
	"os"

	conoha "github.com/leonunix/conohav3-golang-sdk"
)

func main() {
	ctx := context.Background()

	// Create client
	client := conoha.NewClient()

	// Get credentials from environment variables
	userID := os.Getenv("CONOHA_USER_ID")
	password := os.Getenv("CONOHA_PASSWORD")
	tenantID := os.Getenv("CONOHA_TENANT_ID")

	if userID == "" || password == "" || tenantID == "" {
		log.Fatal("Please set CONOHA_USER_ID, CONOHA_PASSWORD, and CONOHA_TENANT_ID")
	}

	// --------------------------------------------------------
	// 1. Authenticate
	// --------------------------------------------------------
	token, err := client.Authenticate(ctx, userID, password, tenantID)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Printf("Authenticated! Token expires at: %s\n", token.ExpiresAt)
	fmt.Printf("Tenant ID: %s\n", client.TenantID)

	// --------------------------------------------------------
	// 2. List Flavors (server plans)
	// --------------------------------------------------------
	flavors, err := client.ListFlavorsDetail(ctx)
	if err != nil {
		log.Fatalf("List flavors failed: %v", err)
	}
	fmt.Printf("\nAvailable Flavors (%d):\n", len(flavors))
	for _, f := range flavors {
		fmt.Printf("  - %s: %d vCPUs, %d MB RAM, %d GB Disk\n", f.Name, f.VCPUs, f.RAM, f.Disk)
	}

	// --------------------------------------------------------
	// 3. List Servers
	// --------------------------------------------------------
	servers, err := client.ListServersDetail(ctx, nil)
	if err != nil {
		log.Fatalf("List servers failed: %v", err)
	}
	fmt.Printf("\nServers (%d):\n", len(servers))
	for _, s := range servers {
		nameTag := s.Metadata["instance_name_tag"]
		fmt.Printf("  - %s (ID: %s, Status: %s)\n", nameTag, s.ID, s.Status)
	}

	// --------------------------------------------------------
	// 4. List Images
	// --------------------------------------------------------
	images, err := client.ListImages(ctx, &conoha.ListImagesOptions{
		Visibility: "public",
		OSType:     "linux",
		Limit:      5,
	})
	if err != nil {
		log.Fatalf("List images failed: %v", err)
	}
	fmt.Printf("\nLinux Images (first 5):\n")
	for _, img := range images {
		fmt.Printf("  - %s (ID: %s)\n", img.Name, img.ID)
	}

	// --------------------------------------------------------
	// 5. List Volumes
	// --------------------------------------------------------
	volumes, err := client.ListVolumesDetail(ctx, nil)
	if err != nil {
		log.Fatalf("List volumes failed: %v", err)
	}
	fmt.Printf("\nVolumes (%d):\n", len(volumes))
	for _, v := range volumes {
		fmt.Printf("  - %s: %d GB (%s, %s)\n", v.Name, v.Size, v.VolumeType, v.Status)
	}

	// --------------------------------------------------------
	// 6. List Security Groups
	// --------------------------------------------------------
	sgs, err := client.ListSecurityGroups(ctx, nil)
	if err != nil {
		log.Fatalf("List security groups failed: %v", err)
	}
	fmt.Printf("\nSecurity Groups (%d):\n", len(sgs))
	for _, sg := range sgs {
		fmt.Printf("  - %s (Rules: %d)\n", sg.Name, len(sg.Rules))
	}

	// --------------------------------------------------------
	// 7. List SSH Keypairs
	// --------------------------------------------------------
	keypairs, err := client.ListKeypairs(ctx, nil)
	if err != nil {
		log.Fatalf("List keypairs failed: %v", err)
	}
	fmt.Printf("\nSSH Keypairs (%d):\n", len(keypairs))
	for _, kp := range keypairs {
		fmt.Printf("  - %s (Fingerprint: %s)\n", kp.Name, kp.Fingerprint)
	}

	// --------------------------------------------------------
	// 8. List DNS Domains
	// --------------------------------------------------------
	domains, err := client.ListDomains(ctx, nil)
	if err != nil {
		log.Fatalf("List domains failed: %v", err)
	}
	fmt.Printf("\nDNS Domains (%d):\n", len(domains))
	for _, d := range domains {
		fmt.Printf("  - %s (TTL: %d)\n", d.Name, d.TTL)
	}

	// --------------------------------------------------------
	// Example: Create a server (commented out to prevent charges)
	// --------------------------------------------------------
	/*
		// Step 1: Create a boot volume
		vol, err := client.CreateVolume(ctx, conoha.CreateVolumeRequest{
			Size:       100,
			Name:       "my-boot-volume",
			VolumeType: "c3j1-ds02-boot",
			ImageRef:   "<image-uuid>",
		})
		if err != nil {
			log.Fatalf("Create volume failed: %v", err)
		}
		fmt.Printf("Volume created: %s\n", vol.ID)

		// Step 2: Create the server
		server, err := client.CreateServer(ctx, conoha.CreateServerRequest{
			FlavorRef: "<flavor-uuid>",
			AdminPass: "YourSecurePassword123!",
			BlockDeviceMapping: []conoha.BlockDeviceMap{
				{UUID: vol.ID},
			},
			Metadata: map[string]string{
				"instance_name_tag": "my-server",
			},
			KeyName: "my-ssh-key",
		})
		if err != nil {
			log.Fatalf("Create server failed: %v", err)
		}
		fmt.Printf("Server created: %s\n", server.ID)

		// Step 3: Stop/Start the server
		err = client.StopServer(ctx, server.ID)
		err = client.StartServer(ctx, server.ID)
	*/

	fmt.Println("\nDone!")
}
