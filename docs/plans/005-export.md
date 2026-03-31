# 005: cred export コマンド

## 目的

キーリングに保存されたシークレットを `.env` 形式で stdout に出力する。`cred import` の逆操作。Docker Compose や direnv など `.env` を読むツールとの連携に使う。

## やること

1. `cmd/export.go` を新規作成
2. 現在の namespace の全キーをキーリングから取得し、`KEY=VALUE` 形式で出力
3. `--keys` フラグで出力するキーを絞り込めるようにする

## 使用例

```bash
cred export > .env.local
cred export -n production
cred export --keys OPENAI_API_KEY,DB_PASSWORD
```

## 変更箇所

- `cmd/export.go`: 新規

## 完了条件

- `cred export` で `.env` 形式の出力が得られる
- `cred import` → `cred export` のラウンドトリップが成立する
- `--keys` で出力を絞り込める
