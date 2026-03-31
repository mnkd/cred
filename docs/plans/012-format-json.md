# 012: --format json フラグ

## 目的

`cred get` や `cred list` の出力を JSON 形式で取得可能にする。スクリプトや他のツールとの連携を容易にする。

## やること

1. `cmd/root.go` にグローバルフラグ `--format` を追加（`text` / `json`）
2. `cmd/get.go`: JSON 出力対応
3. `cmd/list.go`: JSON 出力対応
4. `cmd/namespace.go` (`ns list`): JSON 出力対応

## 使用例

```bash
cred get OPENAI_API_KEY --format json
# {"key":"OPENAI_API_KEY","value":"sk-xxx"}

cred list --format json
# [{"key":"OPENAI_API_KEY","env_var":"OPENAI_API_KEY","description":"..."}]

cred ns list --format json
# [{"name":"default","default":true,"key_count":3}]
```

## 変更箇所

- `cmd/root.go`: `--format` グローバルフラグ追加
- `cmd/get.go`: JSON 出力分岐
- `cmd/list.go`: JSON 出力分岐
- `cmd/namespace.go`: JSON 出力分岐

## 完了条件

- `--format json` で有効な JSON が出力される
- デフォルト（`--format text`）の動作は変わらない
