# cred

OS キーリング（macOS Keychain）で API Key 等のシークレットを一元管理する Go CLI ツール。

## Build & Test

```bash
make build       # go mod tidy + go build
make test        # go test ./...
make install     # go install .
```

## Architecture

- `cmd/` — cobra サブコマンド
- `internal/keyring/` — zalando/go-keyring のタイムアウトラッパー
- `internal/config/` — ~/.config/cred/config.yml の読み書き
- `internal/env/` — .env ファイルパーサー

## Coding Rules

- macOS 専用。Windows / Linux 対応は不要
- シークレットの値は設定ファイルに書かない（OS キーリングにのみ保存）
- ユーザー向けメッセージは `os.Stderr` に出力（`os.Stdout` はデータ出力専用）
- エラーは `fmt.Errorf("context: %w", err)` でラップ

## Workflow

- 作業開始時は docs/plans/ にプランを作成してから実装する
- TDD の流れで実装する: テストを先に書く → 失敗を確認 → 実装 → テストが通ることを確認
- コード変更後は README.md の更新が必要か確認し、必要なら更新する
  - 対象: 使い方、アーキテクチャ図、比較表、依存ライブラリ
