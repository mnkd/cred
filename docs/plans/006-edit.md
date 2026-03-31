# 006: cred edit コマンド

## 目的

既存のシークレットの値を `$EDITOR` で編集する。複数行の値や JSON 形式のシークレットの更新に有用。

## やること

1. `cmd/edit.go` を新規作成
2. キーリングから現在の値を取得し、一時ファイルに書き出す
3. `$EDITOR`（未設定なら `vi`）で一時ファイルを開く
4. 編集後の内容をキーリングに保存
5. 一時ファイルを確実に削除する

## 使用例

```bash
cred edit OPENAI_API_KEY
EDITOR=code cred edit CONFIG_JSON
```

## 変更箇所

- `cmd/edit.go`: 新規

## 完了条件

- `cred edit KEY` でエディタが起動し、保存後にキーリングが更新される
- 一時ファイルが処理後に削除されている
