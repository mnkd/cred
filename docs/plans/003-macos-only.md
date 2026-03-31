# 003: macOS 限定のビルド制約

## 目的

cred はローカル macOS 環境での利用に限定する。Windows / Linux サポートのためのコード分岐を不要にし、実装をシンプルに保つ。

## やること

1. `cmd/run.go` に `//go:build darwin` ビルド制約を追加
2. README.md のサポート対象を macOS に限定する記述に更新

## 変更箇所

- `cmd/run.go`: ビルド制約の追加
- `README.md`: サポート対象の記述更新

## 完了条件

- `GOOS=darwin` でビルドが通る
- README にサポート対象が macOS であることが明記されている
