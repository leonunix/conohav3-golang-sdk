# ConoHa VPS v3 Go SDK

A Go SDK for the [ConoHa VPS v3 API](https://doc.conoha.jp/reference/api-vps3/).

[日本語ドキュメント](README_ja.md)

## Installation

```bash
go get github.com/leonunix/conohav3-golang-sdk
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	conoha "github.com/leonunix/conohav3-golang-sdk"
)

func main() {
	ctx := context.Background()
	client := conoha.NewClient()

	// Authenticate
	token, err := client.Authenticate(ctx, "user-id", "password", "tenant-id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Token expires at: %s\n", token.ExpiresAt)

	// List servers
	servers, err := client.ListServersDetail(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range servers {
		fmt.Printf("Server: %s (%s)\n", s.Metadata["instance_name_tag"], s.Status)
	}
}
```

## Supported APIs

| Service | Description | Endpoints |
|---------|-------------|-----------|
| **Identity** | Authentication, credentials, sub-users, roles, permissions | 20 |
| **Compute** | Servers, flavors, SSH keypairs, server actions, monitoring | 34 |
| **Volume** | Block storage, volume types, backups | 15 |
| **Image** | OS images, ISO upload, quotas | 8 |
| **Network** | Networks, subnets, ports, security groups, QoS | 25 |
| **Load Balancer** | Load balancers, listeners, pools, members, health monitors | 25 |
| **Object Storage** | Containers, objects, large file uploads, versioning | 18 |
| **DNS** | Domains, DNS records | 10 |

## Configuration

### Region

The default region is `c3j1`. You can change it with:

```go
client := conoha.NewClient(conoha.WithRegion("c3j2"))
```

### Endpoint Discovery

Endpoints are resolved in this order (highest priority first):

1. **Explicitly set** via `With*URL()` or `WithEndpoints()` — never overridden
2. **Auto-discovered** from Service Catalog after `Authenticate()`
3. **Generated** from Region pattern `https://{service}.{region}.conoha.io`

```go
// Only set Identity URL, auto-discover rest after auth
client := conoha.NewClient(conoha.WithIdentityURL("https://identity.c3j2.conoha.io"))

// Set all URLs manually
client := conoha.NewClient(conoha.WithEndpoints(conoha.Endpoints{
	Identity: "https://identity.c3j1.conoha.io",
	Compute:  "https://compute.c3j1.conoha.io",
	// ...
}))

// Mix: set some explicitly, auto-discover the rest
client := conoha.NewClient(
	conoha.WithIdentityURL("https://identity.c3j2.conoha.io"),
	conoha.WithComputeURL("https://compute.custom.example.com"),
)
```

### Custom HTTP Client

```go
client := conoha.NewClient(conoha.WithHTTPClient(&http.Client{
	Timeout: 30 * time.Second,
}))
```

## Usage Examples

### Authentication

```go
// By user ID + tenant ID
token, err := client.Authenticate(ctx, "user-id", "password", "tenant-id")

// By user name + tenant name
token, err := client.AuthenticateByName(ctx, "user-name", "password", "tenant-name")
```

### Server Management

```go
// Create a server
server, err := client.CreateServer(ctx, conoha.CreateServerRequest{
	FlavorRef: "flavor-uuid",
	AdminPass: "SecurePassword123!",
	BlockDeviceMapping: []conoha.BlockDeviceMap{
		{UUID: "boot-volume-uuid"},
	},
	Metadata: map[string]string{
		"instance_name_tag": "my-server",
	},
	KeyName: "my-ssh-key",
})

// Start / Stop / Reboot
err = client.StartServer(ctx, serverID)
err = client.StopServer(ctx, serverID)
err = client.RebootServer(ctx, serverID)

// Resize (plan change)
err = client.ResizeServer(ctx, serverID, "new-flavor-uuid")
err = client.ConfirmResize(ctx, serverID)

// Get VNC console URL
url, err := client.GetVNCConsoleURL(ctx, serverID)
```

### Volume Management

```go
// Create a boot volume
vol, err := client.CreateVolume(ctx, conoha.CreateVolumeRequest{
	Size:       100,
	Name:       "boot-vol",
	VolumeType: "c3j1-ds02-boot",
	ImageRef:   "image-uuid",
})

// Attach / Detach
attachment, err := client.AttachVolume(ctx, serverID, volumeID)
err = client.DetachVolume(ctx, serverID, volumeID)

// Auto-backup (weekly, default)
backup, err := client.EnableAutoBackup(ctx, serverID, nil)

// Auto-backup (daily with 14-day retention)
backup, err = client.EnableAutoBackup(ctx, serverID, &conoha.EnableAutoBackupOptions{
	Schedule:  "daily",
	Retention: 14,
})

// Update daily backup retention to 30 days
backup, err = client.UpdateBackupRetention(ctx, serverID, 30)

// Disable auto-backup (cancels both weekly and daily)
err = client.DisableAutoBackup(ctx, serverID)
```

### Network & Security Groups

```go
// Create security group
sg, err := client.CreateSecurityGroup(ctx, "web-server", "HTTP/HTTPS access")

// Add a rule (allow TCP port 80)
portMin, portMax := 80, 80
protocol := "tcp"
rule, err := client.CreateSecurityGroupRule(ctx, conoha.CreateSecurityGroupRuleRequest{
	SecurityGroupID: sg.ID,
	Direction:       "ingress",
	EtherType:       "IPv4",
	Protocol:        &protocol,
	PortRangeMin:    &portMin,
	PortRangeMax:    &portMax,
})

// Allocate additional IP
port, err := client.AllocateAdditionalIP(ctx, 1, nil)
```

### DNS

```go
// Create domain (trailing period required)
domain, err := client.CreateDomain(ctx, conoha.CreateDomainRequest{
	Name:  "example.com.",
	TTL:   3600,
	Email: "admin@example.com",
})

// Add A record
record, err := client.CreateDNSRecord(ctx, domain.UUID, conoha.CreateDNSRecordRequest{
	Name: "www.example.com.",
	Type: "A",
	Data: "203.0.113.1",
})
```

### Object Storage

```go
// Create container and upload
err = client.CreateContainer(ctx, "my-bucket")
err = client.UploadObject(ctx, "my-bucket", "hello.txt", strings.NewReader("Hello!"))

// Download
reader, err := client.DownloadObject(ctx, "my-bucket", "hello.txt")
defer reader.Close()

// Copy
err = client.CopyObject(ctx, "src-bucket", "file.txt", "dst-bucket", "file-copy.txt")
```

### Load Balancer

```go
// Create load balancer -> listener -> pool -> member
lb, err := client.CreateLoadBalancer(ctx, "my-lb")
listener, err := client.CreateListener(ctx, "web", "TCP", 80, lb.ID)
pool, err := client.CreatePool(ctx, "web-pool", "TCP", "ROUND_ROBIN", listener.ID)
member, err := client.AddMember(ctx, pool.ID, "server1", "203.0.113.1", 80)

// Health monitor
hm, err := client.CreateHealthMonitor(ctx, conoha.CreateHealthMonitorRequest{
	Name:       "tcp-check",
	PoolID:     pool.ID,
	Delay:      30,
	MaxRetries: 3,
	Timeout:    10,
	Type:       "TCP",
})
```

## Error Handling

API errors are returned as `*conoha.APIError`:

```go
servers, err := client.ListServers(ctx)
if err != nil {
	var apiErr *conoha.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("HTTP %d: %s\n", apiErr.StatusCode, apiErr.Body)
	}
}
```

## License

[MIT](LICENSE)
