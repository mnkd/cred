# 002: cred set のセキュアプロンプト

## 目的

`cred set KEY` で値を対話入力する際、現在は `bufio.Scanner` を使っているため入力が画面に表示される。シークレットの入力時にエコーバックを抑制する。

## やること

1. `golang.org/x/term` を依存に追加
2. `cmd/set.go` の対話入力部分を `term.ReadPassword` に置き換え
3. 入力後に改行を出力（`ReadPassword` は改行を出さないため）

## 変更箇所

- `cmd/set.go`: プロンプト入力部分
- `go.mod`: `golang.org/x/term` 追加

## 完了条件

- `cred set KEY` 実行時に入力値が画面に表示されない
- stdin パイプ（`echo "xxx" | cred set KEY -`）は引き続き動作する
