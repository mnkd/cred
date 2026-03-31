# 009: シェル補完

## 目的

サブコマンドやフラグの入力を Tab 補完できるようにし、操作性を向上させる。

## やること

1. cobra の `GenZshCompletion` / `GenBashCompletion` / `GenFishCompletion` を有効化
2. `cred completion zsh` のようなサブコマンドを追加

## 使用例

```bash
# zsh
eval "$(cred completion zsh)"

# bash
eval "$(cred completion bash)"
```

## 変更箇所

- `cmd/root.go` または `cmd/completion.go`: 補完コマンド追加

## 完了条件

- `cred completion zsh` で補完スクリプトが出力される
- サブコマンド、フラグ、namespace 名が補完される
