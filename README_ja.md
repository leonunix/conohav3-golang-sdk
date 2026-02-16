# ConoHa VPS v3 Go SDK

[ConoHa VPS v3 API](https://doc.conoha.jp/reference/api-vps3/) 用の Go SDK です。

[English](README.md)

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
client := conoha.NewClient(conoha.WithRegion("c3j2"))
```

### エンドポイント解決順序

エンドポイントURLは以下の優先順位で決定されます：

1. **明示的に指定** (`With*URL()` / `WithEndpoints()`) — 上書きされない
2. **自動検出** — `Authenticate()` 後にサービスカタログから取得
3. **リージョンパターン** — `https://{service}.{region}.conoha.io` から生成

```go
// Identity URLのみ指定、残りは認証後に自動検出
client := conoha.NewClient(conoha.WithIdentityURL("https://identity.c3j2.conoha.io"))

// 全URLを手動で指定
client := conoha.NewClient(conoha.WithEndpoints(conoha.Endpoints{
	Identity: "https://identity.c3j1.conoha.io",
	Compute:  "https://compute.c3j1.conoha.io",
	// ...
}))

// 一部を指定、残りを自動検出
client := conoha.NewClient(
	conoha.WithIdentityURL("https://identity.c3j2.conoha.io"),
	conoha.WithComputeURL("https://compute.custom.example.com"),
)
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

// プラン変更（リサイズ）
err = client.ResizeServer(ctx, serverID, "新フレーバーUUID")
err = client.ConfirmResize(ctx, serverID)

// VNCコンソールURL取得
url, err := client.GetVNCConsoleURL(ctx, serverID)
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

// アタッチ / デタッチ
attachment, err := client.AttachVolume(ctx, serverID, volumeID)
err = client.DetachVolume(ctx, serverID, volumeID)

// 自動バックアップ
backup, err := client.EnableAutoBackup(ctx, serverID)
err = client.DisableAutoBackup(ctx, serverID)
```

### ネットワーク・セキュリティグループ

```go
// セキュリティグループ作成
sg, err := client.CreateSecurityGroup(ctx, "web-server", "HTTP/HTTPSアクセス")

// ルール追加（TCP 80番ポートを許可）
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

// 追加IPアドレスの割り当て
port, err := client.AllocateAdditionalIP(ctx, 1, nil)
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

### オブジェクトストレージ

```go
// コンテナ作成とアップロード
err = client.CreateContainer(ctx, "my-bucket")
err = client.UploadObject(ctx, "my-bucket", "hello.txt", strings.NewReader("Hello!"))

// ダウンロード
reader, err := client.DownloadObject(ctx, "my-bucket", "hello.txt")
defer reader.Close()

// コピー
err = client.CopyObject(ctx, "src-bucket", "file.txt", "dst-bucket", "file-copy.txt")
```

### ロードバランサー

```go
// ロードバランサー → リスナー → プール → メンバー の順に作成
lb, err := client.CreateLoadBalancer(ctx, "my-lb")
listener, err := client.CreateListener(ctx, "web", "TCP", 80, lb.ID)
pool, err := client.CreatePool(ctx, "web-pool", "TCP", "ROUND_ROBIN", listener.ID)
member, err := client.AddMember(ctx, pool.ID, "server1", "203.0.113.1", 80)

// ヘルスモニター
hm, err := client.CreateHealthMonitor(ctx, conoha.CreateHealthMonitorRequest{
	Name:       "tcp-check",
	PoolID:     pool.ID,
	Delay:      30,
	MaxRetries: 3,
	Timeout:    10,
	Type:       "TCP",
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
