# 008: 設定ファイルのマイグレーション

## 目的

設定ファイルのスキーマを将来変更した際に、既存の config.yml を自動的に新しい形式に移行する仕組みを用意する。

## やること

1. `internal/config/migrate.go` を新規作成
2. `config.Load()` 内で `version` フィールドを確認し、現在のバージョンより古い場合にマイグレーション関数を実行する
3. マイグレーション関数をバージョンごとに登録できる構造にする

## 変更箇所

- `internal/config/migrate.go`: 新規
- `internal/config/config.go`: `Load()` にマイグレーション呼び出しを追加

## 完了条件

- 古い `version` の config.yml を読み込んだ際に、自動でマイグレーションが実行される
- マイグレーション後に `version` が更新される
