# cred - Credential Manager for Local Development

`.env` ファイルに散らばった API Key を OS キーリングで一元管理する CLI ツール。
GitHub CLI (`gh`) の credential 管理アーキテクチャに着想を得ています。

## 特徴

- **セッション管理不要**: macOS Keychain を使用するため、`op signin` のような再認証が不要
- **macOS 専用**: ローカル開発環境（macOS）に特化
- **namespace で整理**: プロジェクトや環境ごとに credential をグルーピング
- **env 注入**: `cred run -- python app.py` で環境変数として自動注入
- **.env インポート**: 既存の `.env` ファイルから一括移行

## インストール

```bash
go install github.com/mnkd/cred@latest
```

または、リポジトリをクローンしてビルド:

```bash
git clone https://github.com/mnkd/cred.git
cd cred
go build -o cred .
```

## 使い方

### `cred set` — シークレットの保存

```bash
# 入力プロンプトで値を入力（画面に表示される）
cred set OPENAI_API_KEY

# 値を直接指定して保存
cred set OPENAI_API_KEY sk-xxxxxxxxxxxx

# 説明を付けて保存
cred set OPENAI_API_KEY sk-xxxxxxxxxxxx -d "OpenAI production"

# stdin から読み込み（パイプ連携）
echo "sk-xxxx" | cred set OPENAI_API_KEY -
cat token.txt | cred set GITHUB_TOKEN -

# 特定の namespace に保存
cred set ANTHROPIC_API_KEY -n work

# 既存キーの値を更新（説明はそのまま保持される）
cred set OPENAI_API_KEY sk-new-key
```

### `cred get` — シークレットの取得

```bash
# 値を stdout に出力（末尾改行なし）
cred get OPENAI_API_KEY

# コマンド置換でシェル変数に代入
export OPENAI_API_KEY=$(cred get OPENAI_API_KEY)

# パイプでクリップボードにコピー（macOS）
cred get OPENAI_API_KEY | pbcopy

# 特定の namespace から取得
cred get DB_PASSWORD -n production

# スクリプト内で利用
curl -H "Authorization: Bearer $(cred get GITHUB_TOKEN)" https://api.github.com/user
```

### `cred delete` — シークレットの削除

```bash
# キーリングと設定ファイルの両方から削除
cred delete OPENAI_API_KEY

# エイリアス（rm / remove）でも同様
cred rm OPENAI_API_KEY
cred remove OPENAI_API_KEY

# 特定の namespace から削除
cred delete AWS_SECRET_KEY -n work
```

### `cred list` — 登録済みキーの一覧

```bash
# 現在の namespace のキー一覧
cred list

# エイリアス
cred ls

# 特定の namespace のキー一覧
cred list -n production

# 全 namespace のキーを横断表示
cred list --all
cred ls -a
```

### `cred import` — .env ファイルからインポート

```bash
# .env ファイルの内容をキーリングに一括登録
cred import .env

# プレビュー（実際には保存しない）
cred import .env --dry-run

# 既存キーを上書き
cred import .env --overwrite

# 特定の namespace にインポート
cred import .env.production -n production

# プレビューしてから上書きインポート
cred import .env --dry-run
cred import .env --overwrite
```

### `cred export` — .env 形式でエクスポート

```bash
# 現在の namespace の全キーを .env 形式で stdout に出力
cred export

# ファイルに保存
cred export > .env.local

# 特定の namespace からエクスポート
cred export -n production

# 特定のキーだけ出力
cred export --keys OPENAI_API_KEY,DB_PASSWORD
```

### `cred edit` — エディタでシークレットを編集

```bash
# $EDITOR（デフォルト: vi）で値を編集
cred edit OPENAI_API_KEY

# エディタを指定
EDITOR=code cred edit CONFIG_JSON

# 特定の namespace のキーを編集
cred edit DB_PASSWORD -n production
```

### `cred run` — 環境変数を注入してコマンド実行

```bash
# 現在の namespace の全 credential を環境変数として注入
cred run -- python app.py
cred run -- node server.js

# 特定の namespace から注入
cred run -n production -- ./deploy.sh

# 注入された環境変数を確認
cred run -- env | grep OPENAI

# Docker に渡す
cred run -- docker run --env OPENAI_API_KEY app:latest

# 複数コマンドをシェル経由で実行
cred run -- sh -c 'echo "key=$OPENAI_API_KEY" && python app.py'
```

### `cred namespace` — namespace 管理

```bash
# namespace 一覧（現在のデフォルトに * が付く）
cred ns list

# 新しい namespace を作成
cred ns create work
cred ns create production -d "本番環境用 API キー"

# デフォルト namespace を確認
cred ns default

# デフォルト namespace を切り替え
cred ns default work

# namespace を削除（確認プロンプトあり）
cred ns delete work

# 確認をスキップして削除
cred ns delete staging --force

# namespace を切り替えながら操作
cred set API_KEY -n work
cred get API_KEY -n work
cred list -n work
cred run -n work -- python app.py
```

### シェル連携の応用

```bash
# .bashrc / .zshrc で自動 export
export OPENAI_API_KEY=$(cred get OPENAI_API_KEY)
export ANTHROPIC_API_KEY=$(cred get ANTHROPIC_API_KEY)

# パイプでクリップボードにコピー
cred get OPENAI_API_KEY | pbcopy    # macOS
cred get OPENAI_API_KEY | xclip     # Linux

# 別のツールにシークレットを渡す
gh secret set OPENAI_API_KEY --body "$(cred get OPENAI_API_KEY)"

# .env を cred に移行してから削除
cred import .env --dry-run   # まずプレビュー
cred import .env              # インポート実行
rm .env                       # 元ファイルを削除
```

### `cred doctor` — 環境診断

```bash
# キーリングの可用性、設定ファイル、整合性をチェック
cred doctor
# ✓ Keyring: accessible
# ✓ Config: ~/.config/cred/config.yml
# ✓ default/OPENAI_API_KEY
# ⚠ default/OLD_KEY: in config but not in keyring
```

### `--format json` — JSON 出力

```bash
# cred get の JSON 出力
cred get OPENAI_API_KEY --format json
# {"key":"OPENAI_API_KEY","value":"sk-xxx"}

# cred list の JSON 出力
cred list --format json
# [{"key":"OPENAI_API_KEY","env_var":"OPENAI_API_KEY","description":"..."}]

# cred ns list の JSON 出力
cred ns list --format json
# [{"name":"default","default":true,"key_count":3}]
```

## 設定ファイル

`~/.config/cred/config.yml` にメタデータが保存されます。
**シークレットの値は設定ファイルには保存されません** — OS キーリングにのみ保存されます。

```yaml
version: 1
default_namespace: default
namespaces:
  default:
    keys:
      OPENAI_API_KEY:
        description: "OpenAI API Key (production)"
        env_var: OPENAI_API_KEY
        created_at: "2026-03-31T00:00:00Z"
        updated_at: "2026-03-31T00:00:00Z"
      ANTHROPIC_API_KEY:
        env_var: ANTHROPIC_API_KEY
  work:
    description: "業務用"
    keys:
      AWS_ACCESS_KEY_ID:
        env_var: AWS_ACCESS_KEY_ID
```

## アーキテクチャ

```
cred
├── main.go                       # エントリポイント
├── cmd/
│   ├── root.go                   # ルートコマンド + グローバルフラグ
│   ├── set.go                    # cred set
│   ├── get.go                    # cred get
│   ├── delete.go                 # cred delete
│   ├── list.go                   # cred list
│   ├── run.go                    # cred run -- COMMAND
│   ├── import.go                 # cred import .env
│   ├── export.go                 # cred export
│   ├── edit.go                   # cred edit
│   ├── doctor.go                 # cred doctor
│   └── namespace.go              # cred ns {list,create,default,delete}
└── internal/
    ├── keyring/
    │   └── keyring.go            # OS キーリングラッパー（3秒タイムアウト）
    ├── config/
    │   └── config.go             # YAML 設定ファイル管理
    └── env/
        └── env.go                # .env ファイルパーサー
```

### キーリングでの保存形式

```
service: "cred:{namespace}"    (例: "cred:default", "cred:work")
user:    "{key_name}"          (例: "OPENAI_API_KEY")
secret:  "{actual_value}"      (例: "sk-xxxxxxxxxxxx")
```

## gh CLI との比較

| 項目 | gh CLI | cred |
|------|--------|------|
| バックエンド | zalando/go-keyring | zalando/go-keyring（同じ） |
| キー設計 | `gh:{hostname}` + `{username}` | `cred:{namespace}` + `{key_name}` |
| 用途 | GitHub トークン専用 | 汎用（任意の API Key） |
| env 注入 | なし | `cred run` で対応 |
| .env インポート | なし | `cred import` で対応 |
| .env エクスポート | なし | `cred export` で対応 |
| JSON 出力 | なし | `--format json` で対応 |

## 依存ライブラリ

- [spf13/cobra](https://github.com/spf13/cobra) — CLI フレームワーク
- [zalando/go-keyring](https://github.com/zalando/go-keyring) — クロスプラットフォーム キーリング
- [go-yaml/yaml](https://github.com/go-yaml/yaml) — YAML パーサー
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) — セキュアプロンプト入力

## ライセンス

MIT
