# 010: cred doctor コマンド

## 目的

cred の動作環境を診断する。キーリングの可用性チェックや、設定ファイルとキーリングの不整合を検出する。

## やること

1. `cmd/doctor.go` を新規作成
2. 以下のチェックを実行:
   - OS キーリングへの読み書きが可能か
   - 設定ファイルが存在し、パース可能か
   - 設定ファイルに記録されているキーがキーリングに存在するか（不整合検出）

## 使用例

```bash
cred doctor
# ✓ Keyring: accessible
# ✓ Config: ~/.config/cred/config.yml
# ⚠ OPENAI_API_KEY: in config but not in keyring
```

## 変更箇所

- `cmd/doctor.go`: 新規

## 完了条件

- `cred doctor` でキーリングの可用性が確認できる
- 設定ファイルとキーリングの不整合が検出・表示される
