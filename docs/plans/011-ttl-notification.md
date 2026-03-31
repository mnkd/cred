# 011: TTL / ローテーション通知

## 目的

長期間更新されていないシークレットをユーザーに通知し、キーのローテーションを促す。

## やること

1. `cred list` 実行時に、`updated_at` から一定期間（デフォルト 90 日）経過したキーに警告を表示
2. 閾値を設定ファイルで変更可能にする

## 使用例

```bash
cred list
# KEY              ENV_VAR          DESCRIPTION        UPDATED
# OPENAI_API_KEY   OPENAI_API_KEY   OpenAI production  2025-01-01T... ⚠ 90日以上未更新
```

## 変更箇所

- `cmd/list.go`: 警告表示ロジック追加
- `internal/config/config.go`: 閾値設定の追加（任意）

## 完了条件

- 一定期間更新されていないキーに警告が表示される
