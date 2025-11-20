# devcontainer での開発ガイド

## devcontainer内でのビルドと実行

### 基本的な使い方

1. **devcontainerを開く**
   ```
   VSCode: Cmd+Shift+P → "Dev Containers: Reopen in Container"
   ```

2. **ビルド**
   ```bash
   # devcontainer内のターミナルで
   make build
   ```

3. **実行**
   ```bash
   ./omusubi-codegen --help
   ```

## pre-omusubiリポジトリへのアクセス

### 方法1: ホストのディレクトリをマウント（推奨）

`.devcontainer/compose.yaml`のvolumesセクションに追加:

```yaml
volumes:
  # ... 既存のvolumes ...

  # pre-omusubiをマウント
  - type: bind
    source: ${PRE_OMUSUBI_PATH:-../pre-omusubi}
    target: /pre-omusubi
    read_only: true
```

`.env`に追加:

```bash
# pre-omusubiリポジトリのパス
PRE_OMUSUBI_PATH=/Users/yamazaki/prog/omusubi_project/pre-omusubi
```

devcontainer再起動後:

```bash
# devcontainer内で
./omusubi-codegen parse --repo /pre-omusubi
./omusubi-codegen generate --repo /pre-omusubi
```

### 方法2: devcontainer内でクローン

```bash
# devcontainer内で
cd /workspace
git clone https://github.com/TakumiOkayasu/pre-omusubi ../pre-omusubi

# 使用
./omusubi-codegen parse --repo ../pre-omusubi
```

### 方法3: /workspaceの親ディレクトリを使う

デフォルトでは`/workspace`がマウントされています:
- ホスト: `/Users/yamazaki/prog/omusubi_project/platform_builder`
- devcontainer: `/workspace`

pre-omusubiが同じ階層にある場合:
```
/Users/yamazaki/prog/omusubi_project/
├── platform_builder/  (マウント済み)
└── pre-omusubi/       (アクセスしたい)
```

残念ながらデフォルトでは親ディレクトリはマウントされていないため、方法1または2を使用してください。

## 実行例

### 抽象クラスの一覧表示

```bash
# devcontainer内で
./omusubi-codegen parse --repo /pre-omusubi --verbose
```

### 実装コードの生成

```bash
# 対話式
./omusubi-codegen generate --repo /pre-omusubi

# コマンドライン指定
./omusubi-codegen generate \
  --repo /pre-omusubi \
  --base IDevice \
  --class MyDevice \
  --output /workspace/output
```

生成されたファイルは`/workspace/output`に作成され、ホストの`platform_builder/output/`からもアクセスできます。

## ビルドされたバイナリについて

devcontainer(Ubuntu)でビルドされたバイナリは**Linux ARM64用**です:

```bash
file omusubi-codegen
# omusubi-codegen: ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV)...
```

このバイナリは:
- ✅ devcontainer内で実行可能
- ✅ Linux ARM64環境で実行可能
- ❌ macOSでは実行不可（exec format error）

## macOSで実行したい場合

### オプション1: macOSにGoをインストール

```bash
# ホスト(macOS)で
brew install go
cd /Users/yamazaki/prog/omusubi_project/platform_builder
make build

# macOS用バイナリが生成される
./omusubi-codegen --help  # ✅ 動作する
```

### オプション2: Docker経由で実行

```bash
# ホスト(macOS)で
./scripts/run-in-docker.sh parse --repo /path/to/pre-omusubi
```

### オプション3: GitHub Releasesからダウンロード

リリース後、macOS用ビルド済みバイナリをダウンロード:

```bash
# 将来のリリース後
curl -L https://github.com/TakumiOkayasu/omusubi-codegen/releases/download/v1.0.0/omusubi-codegen-1.0.0-darwin-arm64.tar.gz -o omusubi-codegen.tar.gz
tar -xzf omusubi-codegen.tar.gz
cd omusubi-codegen-1.0.0-darwin-arm64/
./omusubi-codegen --help  # ✅ 動作する
```

## 推奨ワークフロー

### 開発時

```bash
# 1. devcontainerを開く
# 2. コードを編集
# 3. devcontainer内でビルド・テスト
make build
make test
./omusubi-codegen parse --repo /pre-omusubi
```

### 実際の使用時

```bash
# macOSにGoをインストールして
brew install go

# macOS用にビルド
make build

# macOSで実行
./omusubi-codegen generate --repo ~/projects/pre-omusubi
```

## トラブルシューティング

### エラー: exec format error

```bash
./omusubi-codegen --help
# zsh: exec format error: ./omusubi-codegen
```

**原因**: Linux用バイナリをmacOSで実行しようとしています

**解決策**:
1. devcontainer内で実行する
2. または、macOSでリビルド: `brew install go && make build`

### エラー: /pre-omusubi: no such file or directory

**原因**: pre-omusubiがマウントされていません

**解決策**:
1. `.devcontainer/compose.yaml`にvolumeを追加（上記「方法1」参照）
2. devcontainerを再起動

### エラー: permission denied

**原因**: ファイルの書き込み権限がありません

**解決策**:
```bash
# devcontainer内で
chmod +x omusubi-codegen
```

## まとめ

| 実行環境 | ビルド方法 | 実行可否 |
|---------|-----------|---------|
| devcontainer | `make build` | ✅ |
| macOS | `brew install go && make build` | ✅ |
| macOS | devcontainerでビルド | ❌ (format error) |
| Linux ARM64 | devcontainerでビルド | ✅ |

**推奨**: 開発はdevcontainer、実使用はmacOSでビルド
