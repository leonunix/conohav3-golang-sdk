# ConoHa VPS v3 Go SDK

A Go SDK for the [ConoHa VPS v3 API](https://doc.conoha.jp/reference/api-vps3/).

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
client := conoha.NewClient(conoha.WithRegion("other-region"))
```

### Custom HTTP Client

```go
client := conoha.NewClient(conoha.WithHTTPClient(&http.Client{
	Timeout: 30 * time.Second,
}))
```

### Custom Endpoint URLs

```go
client := conoha.NewClient()
client.ComputeURL = "https://compute.custom.conoha.io"
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

// Auto-backup
backup, err := client.EnableAutoBackup(ctx, serverID)
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
// Create domain
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

---

# ConoHa VPS v3 Go SDK (日本語)

[ConoHa VPS v3 API](https://doc.conoha.jp/reference/api-vps3/) 用の Go SDK です。

## インストール

```bash
go get github.com/leonunix/conohav3-golang-sdk
```

## クイックスタート

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

	// 認証
	token, err := client.Authenticate(ctx, "ユーザーID", "パスワード", "テナントID")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("トークン有効期限: %s\n", token.ExpiresAt)

	// サーバー一覧取得
	servers, err := client.ListServersDetail(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range servers {
		fmt.Printf("サーバー: %s (%s)\n", s.Metadata["instance_name_tag"], s.Status)
	}
}
```

## 対応API一覧

| サービス | 説明 | エンドポイント数 |
|---------|------|----------------|
| **Identity** | 認証、クレデンシャル、サブユーザー、ロール、パーミッション | 20 |
| **Compute** | サーバー管理、フレーバー、SSHキーペア、サーバー操作、モニタリング | 34 |
| **Volume** | ブロックストレージ、ボリュームタイプ、バックアップ | 15 |
| **Image** | OSイメージ、ISOアップロード、クォータ | 8 |
| **Network** | ネットワーク、サブネット、ポート、セキュリティグループ、QoS | 25 |
| **Load Balancer** | ロードバランサー、リスナー、プール、メンバー、ヘルスモニター | 25 |
| **Object Storage** | コンテナ、オブジェクト、大容量ファイルアップロード、バージョニング | 18 |
| **DNS** | ドメイン、DNSレコード | 10 |

## 設定

### リージョン

デフォルトリージョンは `c3j1` です。変更する場合：

```go
client := conoha.NewClient(conoha.WithRegion("other-region"))
```

### カスタムHTTPクライアント

```go
client := conoha.NewClient(conoha.WithHTTPClient(&http.Client{
	Timeout: 30 * time.Second,
}))
```

## 主な使い方

### 認証

```go
// ユーザーID + テナントID で認証
token, err := client.Authenticate(ctx, "user-id", "password", "tenant-id")

// ユーザー名 + テナント名 で認証
token, err := client.AuthenticateByName(ctx, "user-name", "password", "tenant-name")
```

### サーバー管理

```go
// サーバー作成
server, err := client.CreateServer(ctx, conoha.CreateServerRequest{
	FlavorRef: "フレーバーUUID",
	AdminPass: "SecurePassword123!",
	BlockDeviceMapping: []conoha.BlockDeviceMap{
		{UUID: "ブートボリュームUUID"},
	},
	Metadata: map[string]string{
		"instance_name_tag": "my-server",
	},
})

// 起動 / 停止 / 再起動
err = client.StartServer(ctx, serverID)
err = client.StopServer(ctx, serverID)
err = client.RebootServer(ctx, serverID)
```

### ボリューム管理

```go
// ブートボリューム作成
vol, err := client.CreateVolume(ctx, conoha.CreateVolumeRequest{
	Size:       100,
	Name:       "boot-vol",
	VolumeType: "c3j1-ds02-boot",
	ImageRef:   "イメージUUID",
})

// 自動バックアップ
backup, err := client.EnableAutoBackup(ctx, serverID)
```

### DNS管理

```go
// ドメイン作成（末尾にピリオドが必要）
domain, err := client.CreateDomain(ctx, conoha.CreateDomainRequest{
	Name:  "example.com.",
	TTL:   3600,
	Email: "admin@example.com",
})

// Aレコード追加
record, err := client.CreateDNSRecord(ctx, domain.UUID, conoha.CreateDNSRecordRequest{
	Name: "www.example.com.",
	Type: "A",
	Data: "203.0.113.1",
})
```

## エラーハンドリング

APIエラーは `*conoha.APIError` として返されます：

```go
servers, err := client.ListServers(ctx)
if err != nil {
	var apiErr *conoha.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("HTTP %d: %s\n", apiErr.StatusCode, apiErr.Body)
	}
}
```

## ライセンス

[MIT](LICENSE)
