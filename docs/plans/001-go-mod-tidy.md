# 001: go mod tidy の実行

## 目的

`go.sum` が未生成のためビルドが通らない。`go mod tidy` で依存関係を解決し、ビルド可能な状態にする。

## やること

1. `go mod tidy` を実行して `go.sum` を生成
2. `go build -o cred .` でビルドが通ることを確認
3. `./cred --help` で起動確認

## 完了条件

- `go.sum` が生成されている
- `go build` がエラーなく完了する
- `./cred --help` でヘルプが表示される
