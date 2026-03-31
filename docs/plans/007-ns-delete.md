# 007: cred ns delete コマンド

## 目的

不要になった namespace を削除する。キーリング上のエントリと設定ファイルのメタデータを一括で削除する。

## やること

1. `cmd/namespace.go` に `nsDeleteCmd` を追加
2. 対象 namespace の全キーをキーリングから削除
3. 設定ファイルから namespace のメタデータを削除
4. デフォルト namespace の削除は禁止する
5. 確認プロンプトを表示する（`--force` でスキップ可能）

## 使用例

```bash
cred ns delete work
cred ns delete staging --force
```

## 変更箇所

- `cmd/namespace.go`: `nsDeleteCmd` 追加

## 完了条件

- `cred ns delete NAME` で namespace とそのキーが全て削除される
- デフォルト namespace の削除時にエラーが返る
- `--force` なしの場合、確認プロンプトが表示される
