# GitHub CLI (gh) の Credential 管理 実装詳細分析

## 1. アーキテクチャ概要

GitHub CLI の credential 管理は、**多層フォールバック戦略**を採用しています。トークンの取得時に以下の優先順位で探索されます。

1. **環境変数** (`GH_TOKEN`, `GITHUB_TOKEN`) — 最優先。CI/CD やスクリプト向け
2. **OS キーリング**（暗号化ストレージ）— デフォルトの安全な保存先
3. **平文設定ファイル** (`~/.config/gh/hosts.yml`) — フォールバック / レガシー

この設計により、環境に応じた柔軟な認証が実現されています。

---

## 2. ディレクトリ・パッケージ構成

```
cli/cli/
├── internal/
│   ├── authflow/
│   │   ├── flow.go          # OAuth Device Flow の実装
│   │   ├── flow_test.go
│   │   └── success.go       # 認証成功時の処理
│   ├── keyring/
│   │   └── keyring.go       # OS キーリングのラッパー（タイムアウト付き）
│   └── config/
│       └── config.go         # 設定ファイル管理 + トークン解決ロジック
├── pkg/cmd/auth/
│   ├── login/
│   │   └── login.go          # `gh auth login` コマンド
│   └── shared/
│       └── login_flow.go     # 共通認証フロー
```

---

## 3. キーリング実装の詳細

### 3.1 外部ライブラリ: zalando/go-keyring

gh CLI は `github.com/zalando/go-keyring` を使用しています。このライブラリの特徴:

- **CGO 不要**: C バインディングに依存せず、静的リンクバイナリを生成可能
- **クロスプラットフォーム対応**:

| OS | バックエンド | 実装方式 |
|----|-------------|---------|
| macOS | Keychain | `/usr/bin/security` コマンド経由 |
| Linux/BSD | Secret Service (GNOME Keyring) | D-Bus インターフェース経由 |
| Windows | Credential Manager | `cmdkey` / PowerShell 経由 |

- **API**: `Set(service, user, secret)`, `Get(service, user)`, `Delete(service, user)` の3操作

### 3.2 gh CLI のラッパー実装 (`internal/keyring/keyring.go`)

gh CLI は go-keyring をそのまま使わず、**タイムアウト機構を追加した薄いラッパー**を実装しています。

```go
package keyring

import (
    "errors"
    "time"
    "github.com/zalando/go-keyring"
)

var ErrNotFound = errors.New("secret not found in keyring")

type TimeoutError struct {
    message string
}

// Set: goroutine + channel で 3秒タイムアウト付き
func Set(service, user, secret string) error {
    ch := make(chan error, 1)
    go func() {
        defer close(ch)
        ch <- keyring.Set(service, user, secret)
    }()
    select {
    case err := <-ch:
        return err
    case <-time.After(3 * time.Second):
        return &TimeoutError{"timeout while trying to set secret in keyring"}
    }
}

// Get: 同様にタイムアウト付き
func Get(service, user string) (string, error) {
    ch := make(chan struct {
        val string
        err error
    }, 1)
    go func() {
        defer close(ch)
        val, err := keyring.Get(service, user)
        ch <- struct {
            val string
            err error
        }{val, err}
    }()
    select {
    case res := <-ch:
        if errors.Is(res.err, keyring.ErrNotFound) {
            return "", ErrNotFound
        }
        return res.val, res.err
    case <-time.After(3 * time.Second):
        return "", &TimeoutError{"timeout while trying to get secret from keyring"}
    }
}

// Delete: 同様
func Delete(service, user string) error { /* 同パターン */ }
```

**設計のポイント**:

- goroutine + channel + `time.After` による非同期タイムアウト（3秒）
- OS のキーチェーンデーモンが応答しない場合のハング防止
- `ErrNotFound` をラップして、呼び出し元でフォールバック判断に使用

### 3.3 トークンの保存キー

```
サービス名: "gh:{hostname}"
ユーザー名: ログインユーザー名

例: keyring.Set("gh:github.com", "octocat", "gho_xxxxxxxxxxxx")
```

---

## 4. 設定ファイル (`config.go`) の設計

### 4.1 ファイル構造 (YAML)

```yaml
version: 1
git_protocol: https
editor: ""
prompt: enabled
pager: ""
browser: ""
hosts:
  github.com:
    git_protocol: https
    users:
      octocat:
        git_protocol: https
```

### 4.2 トークン解決の優先順位（ActiveToken メソッド）

```
ActiveToken(hostname) の探索順序:

1. 環境変数 GH_TOKEN / GITHUB_TOKEN
   → 設定されていれば即座に返す（CI/CD 向け）

2. 平文設定ファイル (~/.config/gh/hosts.yml)
   → レガシー互換。以前のデフォルト保存先

3. キーリング（暗号化ストレージ）
   → keyring.Get("gh:{hostname}", "{username}")
```

### 4.3 ホスト・ユーザー管理の階層構造

```
[key]                                    → グローバル設定
[hosts][hostname][key]                   → ホスト別設定
[hosts][hostname][users][username][key]  → ユーザー別設定
```

GitHub Enterprise Server など複数ホストの同時管理に対応。

---

## 5. 認証フロー (`authflow/flow.go`)

### 5.1 OAuth Device Flow

```
1. CLI → GitHub API: Device Code リクエスト
2. GitHub API → CLI: device_code + user_code + verification_uri
3. CLI → ユーザー: user_code を表示（クリップボードにもコピー）
4. CLI → ブラウザ: verification_uri を自動オープン
5. ユーザー → ブラウザ: user_code を入力して承認
6. CLI → GitHub API: ポーリングで access_token を取得
7. CLI → keyring / config: トークンを保存
```

### 5.2 OAuth スコープ

- デフォルト: `repo`, `read:org`, `gist`
- 追加スコープはオプションで指定可能

### 5.3 埋め込みクライアント情報

```go
oauthClientID     = "178c6fc778ccc68e1d6a"
oauthClientSecret = "34ddeff2b558a23d38fba8a6de74f086ede1cc0b"
```

OAuth Client ID/Secret はバイナリに埋め込み。GitHub のパブリック OAuth アプリなので、Secret の公開は設計上許容されています。

---

## 6. `--insecure-storage` フラグ

```go
// login.go での処理
SecureStorage: !opts.InsecureStorage
```

- **デフォルト**: キーリング使用（SecureStorage = true）
- **`--insecure-storage`**: 平文ファイルにフォールバック
- **キーリング利用不可時**: 自動的に平文保存 + 警告メッセージ表示

```
⚠ Authentication credentials saved in plain text
```

---

## 7. 1Password CLI (`op`) との比較

| 項目 | gh CLI | 1Password CLI (`op`) |
|------|--------|---------------------|
| セッション管理 | OS キーリングに永続保存 | 30分でセッション期限切れ |
| 再認証 | 不要（キーリングが利用可能な限り） | `op signin` が定期的に必要 |
| バックエンド | OS ネイティブ (Keychain等) | 1Password クラウド |
| オフライン | 完全動作 | 制限あり |
| 複数サービス | GitHub のみ | 汎用（任意のシークレット） |
| CLI 統合 | 自前で完結 | `op run` でenv注入 |

---

## 8. 自作ツールへの示唆

gh CLI の設計から学べるポイント:

1. **タイムアウト付きキーリングアクセス**: OS キーリングは応答しない場合があるため必須
2. **多層フォールバック**: 環境変数 → キーリング → ファイル の優先順位
3. **`--insecure-storage` 相当のオプション**: ヘッドレス環境への対応
4. **ホスト×ユーザーのキー設計**: `{service}:{hostname}` + `{username}` で一意に特定
5. **設定ファイルの YAML 管理**: シークレット以外のメタ情報は平文 YAML で管理
6. **マイグレーション機構**: スキーマバージョン管理で将来の構造変更に対応
