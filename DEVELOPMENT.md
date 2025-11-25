# Development Guide

## 開発環境のセットアップ

### Dev Containerの使用 (推奨)

このプロジェクトはDocker Compose形式のdevcontainerを使用しています。

1. VS Codeでリポジトリを開く
2. 「Reopen in Container」を選択
3. 必要な開発環境が自動的に構築されます

`.env`ファイルが存在しない場合、初回起動時に自動的に`.env.example`からコピーされます。

#### ユーザー設定について

devcontainerはホストの環境変数から自動的にユーザー情報を取得します:

- `USER`: ホストのユーザー名
- `UID`: ホストのユーザーID
- `GID`: ホストのグループID

**macOSの場合:**
環境変数`UID`と`GID`が自動設定されていない場合、以下のコマンドで確認し、`.env`ファイルに設定してください:

```bash
# 現在のUIDとGIDを確認
id -u  # UID
id -g  # GID

# .envファイルを作成
cp .env.example .env
# .envファイルを編集してUID/GIDを設定
```

**Linuxの場合:**
通常は自動的に環境変数が設定されているため、追加設定は不要です。

#### プロジェクト識別子の設定 (REQUIRED)

**重要:** devcontainerを起動する前に、必ず`.env`ファイルを作成してください。

```bash
# .envファイルを作成 (初回のみ)
cp .env.example .env
```

`.env`ファイルが存在しない場合、devcontainer起動時に自動的に`.env.example`からコピーされます。

複数のプロジェクトで同じdevcontainer設定を使い回す場合は、`.env`ファイルの`PROJECT_NAME`を**プロジェクトごとに異なる値**に設定してください。これにより、コンテナ名、ボリューム名、ネットワーク名の衝突を防げます。

`PROJECT_NAME`を設定すると、以下のリソースがプロジェクト固有の名前になります:

- **コンテナ名**: `${PROJECT_NAME}-devcontainer`
- **ボリューム名**:
  - `${PROJECT_NAME}-go-modules` (Goモジュールキャッシュ)
  - `${PROJECT_NAME}-vscode-extensions` (VSCode拡張キャッシュ)
- **ネットワーク名**: `${PROJECT_NAME}-network`
- **ホスト名**: `${PROJECT_NAME}-dev`

**設定例:**

```bash
# プロジェクトA (.env)
PROJECT_NAME=omusubi-codegen
USER=yamazaki
UID=501
GID=20

# プロジェクトB (.env) - 別プロジェクトで同時起動可能
PROJECT_NAME=omusubi-device-impl
USER=yamazaki
UID=501
GID=20
```

**デフォルト値:** `.env.example`に記載されている`omusubi-codegen`が使用されます。

#### Docker Composeでの起動 (オプション)

より細かい制御が必要な場合は、Docker Composeを直接使用できます:

```bash
# .envファイルを読み込んでコンテナ起動
docker compose -f .devcontainer/compose.yaml up -d

# コンテナに接続
docker exec -it ${PROJECT_NAME}-devcontainer /bin/bash

# コンテナ停止
docker compose -f .devcontainer/compose.yaml down
```

**注意:** Docker Compose v2以降は`docker compose`コマンド(ハイフンなし)を使用します。

### ローカル環境

以下のツールが必要です:

- Go 1.23以上
- Clang 18以上 (推奨) または GCC 11以上
- Make

依存関係のインストール:

```bash
make deps
```

## ビルド

```bash
make build
```

バイナリは `omusubi-codegen` として生成されます。

## テスト

```bash
# 全テスト実行
make test

# カバレッジ付き
make test-coverage
```

## コードフォーマット

```bash
make fmt
```

## Lint

```bash
make lint
```

golangci-lintが必要です:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## アーキテクチャ

### パッケージ構成

- `cmd/codegen`: メインエントリーポイント、CLIコマンド定義
- `internal/parser`: C++ソースコード解析
- `internal/generator`: コード生成ロジック
- `internal/generator/templates`: テンプレートファイル
- `internal/model`: 内部データ構造

### データフロー

```
C++ Header File (.hpp/.h)
    ↓
Parser (tree-sitter)
    ↓
Model (ClassInfo, MethodInfo, etc.)
    ↓
Generator + Templates
    ↓
Generated Files (.hpp/.h + .cpp)
```

## tree-sitterの使い方

tree-sitterは構文解析ライブラリで、C++のASTを生成します。

### 基本的な使用例

```go
import (
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/cpp"
)

parser := sitter.NewParser()
parser.SetLanguage(cpp.GetLanguage())

source := []byte("class Foo { public: void bar(); };")
tree, _ := parser.ParseCtx(nil, nil, source)
defer tree.Close()

// ASTのトラバース
root := tree.RootNode()
// ...
```

### ノードタイプの確認

C++のクラス定義は `class_specifier` ノードとして表現されます。
メソッドは `function_definition` または `field_declaration` として表現されます。

デバッグ時は以下のコマンドでASTを確認できます:

```bash
echo "class Foo {};" | tree-sitter parse --language cpp
```

## テンプレートのカスタマイズ

テンプレートは `internal/generator/templates/` に配置されています。

### テンプレート変数

- `.ClassName`: 派生クラス名
- `.BaseClass`: 基底クラス名
- `.BaseClassExt`: ファイル拡張子 (h/hpp)
- `.Namespace`: 名前空間
- `.Methods`: メソッドのリスト (MethodInfo配列)
- `.HasNamespace`: 名前空間の有無 (Boolean)

### 新しいテンプレートの追加

1. `internal/generator/templates/` に `.tmpl` ファイルを追加
2. `internal/generator/generator.go` に新しいレンダリングメソッドを追加
3. `go:embed` ディレクティブにより自動的にバイナリに埋め込まれます

## デバッグ

### パーサーのデバッグ

```bash
./omusubi-codegen parse --repo testdata --verbose
```

### 生成されたコードの確認

```bash
./omusubi-codegen generate --repo testdata --output /tmp/output
ls -la /tmp/output
```

## コントリビューション

1. Forkしてブランチを作成
2. 変更を実装
3. テストを追加・実行
4. Lintを実行
5. Pull Requestを作成

## トラブルシューティング

### tree-sitterのビルドエラー

Dev Containerを使用することで、tree-sitterのC++バインディングが正しくビルドされます。

### テンプレートが更新されない

テンプレートは `go:embed` で埋め込まれているため、変更後は再ビルドが必要です:

```bash
make clean
make build
```
