# 004: テストの追加

## 目的

主要パッケージの動作を自動テストで保証する。リファクタリングや機能追加時のデグレを防ぐ。

## やること

1. `internal/env`: .env パーサーのテスト
   - 基本的な `KEY=VALUE`
   - クォート付き `KEY="value"`
   - `export KEY=VALUE`
   - コメント行、空行のスキップ
   - 不正な行の処理
2. `internal/config`: 設定ファイルのテスト
   - Load / Save のラウンドトリップ
   - namespace 操作（EnsureNamespace, SetKeyMeta, DeleteKeyMeta）
   - ActiveNamespace のフォールバック動作
3. `internal/keyring`: キーリングラッパーのテスト
   - `zalando/go-keyring` の `MockInit()` を使った単体テスト
   - Set / Get / Delete の基本操作
   - タイムアウト動作

## 変更箇所

- `internal/env/env_test.go`: 新規
- `internal/config/config_test.go`: 新規
- `internal/keyring/keyring_test.go`: 新規

## 完了条件

- `go test ./...` が全件パスする
