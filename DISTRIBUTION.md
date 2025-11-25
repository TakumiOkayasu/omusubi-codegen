# バイナリ配布方法

## 配布方法の比較

| 方法 | 難易度 | 自動化 | おすすめ度 | 用途 | CGO対応 |
|------|--------|--------|-----------|------|---------|
| GitHub Actions (マトリックス) | 低 | ◎ | ★★★★★ | 本番リリース | ◎ |
| Docker buildx | 中 | ◎ | ★★★★☆ | ローカルマルチビルド | ◎ |
| ローカルビルド | 低 | △ | ★★★☆☆ | テスト配布 | ○ (現在のプラットフォームのみ) |
| go install | 最低 | - | ★★☆☆☆ | 開発者向け | ○ (要CGO環境) |

## 1. GitHub Actions (推奨)

### セットアップ

すでに設定済みです:
- `.github/workflows/release.yml`: マトリックスビルドワークフロー

### リリース手順

```bash
# 1. コミットして変更を確定
git add .
git commit -m "Release v1.0.0"

# 2. タグを作成
git tag -a v1.0.0 -m "Release version 1.0.0"

# 3. タグをプッシュ (これでGitHub Actionsが自動実行される)
git push origin v1.0.0
```

これだけで、以下が自動的に行われます:
- Linux (amd64) をUbuntuランナーでビルド
- macOS (amd64) をIntel Macランナーでビルド
- macOS (arm64) をApple Silicon Macランナーでビルド
- tar.gzアーカイブを作成
- チェックサムファイルを生成
- GitHub Releasesにアップロード
- リリースノートを自動生成

### CGOによる制限

tree-sitterはCGOを必要とするため:
- 各プラットフォームはネイティブランナーでビルド
- Linux ARM64は含まれません(GitHub Actionsにネイティブランナーがないため)
- 必要な場合は、ARM64環境で手動ビルドまたはDocker buildxを使用

### ユーザーのインストール方法

リリース後、ユーザーは以下の方法でインストールできます:

```bash
# Homebrew (macOS/Linux) - 将来対応予定
# brew install TakumiOkayasu/tap/codegen

# 手動ダウンロード
# GitHub ReleasesからOS別のバイナリをダウンロード
# https://github.com/TakumiOkayasu/omusubi-codegen/releases
```

## 2. ローカルビルド

### 重要な注意点

**このプロジェクトはtree-sitterを使用しているため、CGOが必須です。**

これにより:
- 通常のGoクロスコンパイル(`GOOS=linux GOARCH=amd64 go build`)は動作しません
- ホストプラットフォームのみビルド可能です
- マルチプラットフォームビルドには、GitHub ActionsまたはDocker buildxが必要です

### 現在のプラットフォーム向けビルド

```bash
# 現在のOS/アーキテクチャ向けにビルド
make release

# または直接スクリプト実行
./scripts/build-all.sh
```

生成されるファイル例(macOS ARM64の場合):
```
dist/
├── codegen-v1.0.0-darwin-arm64.tar.gz
└── checksums.txt
```

### Docker buildxを使ったマルチプラットフォームビルド

```bash
# すべてのプラットフォーム向けにビルド
make release-docker

# または直接スクリプト実行
./scripts/docker-build-all.sh
```

生成されるファイル:
```
dist/
├── codegen-v1.0.0-linux-amd64.tar.gz
├── codegen-v1.0.0-linux-arm64.tar.gz
├── codegen-v1.0.0-darwin-amd64.tar.gz
├── codegen-v1.0.0-darwin-arm64.tar.gz
└── checksums.txt
```

### 配布

1. **GitHub Releasesに手動アップロード**
   - GitHubのReleasesページで手動アップロード

2. **ファイルサーバーで配布**
   ```bash
   # dist/をサーバーにアップロード
   scp dist/* user@server:/path/to/downloads/
   ```

3. **Google Drive/Dropbox**
   - dist/内のファイルをクラウドストレージにアップロード

## 3. go install (開発者向け)

Goがインストールされている環境で直接インストール可能です。

### 前提条件

リポジトリがGitHubに公開されている必要があります。

### インストール

```bash
go install github.com/TakumiOkayasu/omusubi-codegen/cmd/codegen@latest

# または特定バージョン
go install github.com/TakumiOkayasu/omusubi-codegen/cmd/codegen@v1.0.0
```

### メリット/デメリット

**メリット:**
- 配布作業不要
- 常に最新版を取得可能

**デメリット:**
- Goのインストールが必要
- 一般ユーザーには不向き

## 4. GoReleaserのローカルテスト

リリース前にローカルでテストできます。

### GoReleaserのインストール

```bash
# macOS
brew install goreleaser

# Linux
# https://goreleaser.com/install/
```

### スナップショットビルド

```bash
# バージョン情報なしでビルド
make snapshot

# または直接実行
goreleaser release --snapshot --clean
```

`dist/`ディレクトリに成果物が生成されます。

## 5. Docker経由の配布

### Dockerfileの例

```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod download
RUN go build -ldflags="-s -w" -o codegen ./cmd/codegen

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/codegen /usr/local/bin/
ENTRYPOINT ["codegen"]
```

### ビルドと実行

```bash
# ビルド
docker build -t codegen:latest .

# 実行
docker run --rm -v $(pwd):/workspace codegen generate --repo /workspace/omusubi/include
```

## 推奨フロー

### 開発中
```bash
make build
./omusubi-codegen --help
```

### テスト配布
```bash
make release
# dist/内のファイルを配布
```

### 本番リリース
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
# GitHub Actionsが自動実行
```

## バージョン管理のベストプラクティス

### セマンティックバージョニング

- `v1.0.0`: メジャーリリース (破壊的変更)
- `v1.1.0`: マイナーリリース (機能追加)
- `v1.1.1`: パッチリリース (バグ修正)

### タグの例

```bash
# 最初のリリース
git tag -a v1.0.0 -m "Initial release"

# 機能追加
git tag -a v1.1.0 -m "Add support for custom templates"

# バグ修正
git tag -a v1.1.1 -m "Fix parser crash on empty files"
```

## チェックサム検証

ユーザーがダウンロードしたファイルの整合性を検証できます。

```bash
# macOS/Linux
shasum -a 256 -c checksums.txt

# または個別に確認
shasum -a 256 codegen-v1.0.0-darwin-arm64.tar.gz
```

## トラブルシューティング

### GoReleaserが動かない

```bash
# 設定ファイルの検証
goreleaser check

# ドライラン
goreleaser release --snapshot --skip=publish --clean
```

### クロスコンパイルエラー

> **注意**: このプロジェクトはtree-sitterを使用しているため、CGOが必須です。
> `CGO_ENABLED=0` ではビルドできません。
> クロスコンパイルが必要な場合は、GitHub ActionsまたはDocker buildxを使用してください。

```bash
# ローカルでのビルドは現在のプラットフォームのみ対応
make release
```

### GitHub Actionsが失敗

1. リポジトリのSettings > Actions > General
2. "Workflow permissions"を"Read and write permissions"に設定
3. タグを再プッシュ

## 参考リンク

- [GoReleaser Documentation](https://goreleaser.com/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Cross Compilation](https://go.dev/doc/install/source#environment)
