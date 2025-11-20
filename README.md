# Omusubi Platform Code Generator

Omusubi組み込みフレームワーク用の自動コード生成ツール。
C++インターフェース定義から実装スケルトン、テストコード、ドキュメントを自動生成します。

## 特徴

- **C++14対応**: Omusubiフレームワークに準拠したコード生成
- **Tree-sitterパーサー**: 高速で正確なC++構文解析
- **自動テスト生成**: Google Testベースのユニットテスト自動生成
- **ドキュメント生成**: Doxygen形式のドキュメントコメント生成
- **テンプレートカスタマイズ**: Go html/templateによる柔軟なテンプレート

## 要件

- Go 1.23以上
- C++14対応コンパイラ (GCC 11以上 / Clang 14以上)
- Google Test (テスト生成を使用する場合)

## インストール

### バイナリダウンロード (推奨)

[GitHub Releases](https://github.com/TakumiOkayasu/omusubi-platform-codegen/releases)から、お使いのOSに合わせたバイナリをダウンロードしてください。

### go install

```bash
go install github.com/TakumiOkayasu/omusubi-platform-codegen/cmd/codegen@latest
```

### ソースからビルド

```bash
git clone https://github.com/TakumiOkayasu/omusubi-platform-codegen
cd omusubi-platform-codegen
make build
```

詳細は[DISTRIBUTION.md](DISTRIBUTION.md)を参照してください。

### 対応プラットフォーム

- Linux (x86_64)
- macOS (Intel)
- macOS (Apple Silicon)

> **Note**: このプロジェクトはtree-sitterを使用しているため、CGOが必須です。
> Linux ARM64など他のプラットフォームが必要な場合は、該当環境で`make build`を実行してください。

## 使い方

### 1. リポジトリ内の抽象クラスを一覧表示

```bash
./codegen parse --repo /path/to/pre-omusubi
```

詳細表示:
```bash
./codegen parse --repo /path/to/pre-omusubi --verbose
```

### 2. 実装コードを生成

基本的な使い方(対話式):
```bash
./codegen generate --repo /path/to/pre-omusubi
```

このコマンドを実行すると:
1. リポジトリ内の抽象クラス一覧が表示されます
2. 基底クラス名の入力を求められます
3. 派生クラス名の入力を求められます
4. カレントディレクトリに.hppと.cppが生成されます

### 3. コマンドライン引数で指定

```bash
./codegen generate \
  --repo /path/to/pre-omusubi \
  --base AbstractClassName \
  --class MyImplementation \
  --output ./output
```

### オプション

#### generateコマンド
- `-r, --repo`: pre-omusubiリポジトリのパス (必須)
- `-b, --base`: 基底クラス名 (省略時は対話式で入力)
- `-c, --class`: 派生クラス名 (省略時は対話式で入力)
- `-o, --output`: 出力ディレクトリ (デフォルト: カレントディレクトリ)
- `-t, --templates`: テンプレートディレクトリ (デフォルト: internal/template/templates)

#### parseコマンド
- `-r, --repo`: pre-omusubiリポジトリのパス (必須)
- `-v, --verbose`: 詳細なメソッド情報を表示

## 開発環境

### devcontainer

VS Code Dev Containers対応。以下の環境が自動構築されます:

- Go 1.23
- Clang 18 (latest) + libc++
- Google Test
- tree-sitter CLI
- 各種開発ツール (gopls, golangci-lint等)

**複数プロジェクト運用時の注意:**
`.env`ファイルで`PROJECT_NAME`を設定することで、コンテナ名・ボリューム名・ネットワーク名の衝突を防げます。
詳細は[DEVELOPMENT.md](DEVELOPMENT.md)を参照してください。

```bash
# .envファイルを作成して設定
cp .env.example .env
# PROJECT_NAME を編集
```

### ローカル開発

```bash
# 依存関係のインストール
go mod download

# ビルド
go build -o codegen ./cmd/codegen

# テスト実行
go test ./...
```

## プロジェクト構造

```
omusubi-platform-codegen/
├── .devcontainer/          # Dev Container設定
│   ├── devcontainer.json
│   └── Dockerfile
├── cmd/
│   └── codegen/            # メインコマンド
├── internal/
│   ├── parser/             # tree-sitterベースのC++パーサー
│   ├── generator/          # コード生成エンジン
│   ├── template/           # テンプレート管理
│   │   └── templates/      # 埋め込みテンプレートファイル
│   └── model/              # 内部データモデル
├── testdata/               # テスト用データ
└── go.mod
```

## 生成されるファイル

入力: pre-omusubi内の抽象クラス (例: `IDevice`)

生成されるファイル:
- `mydevice.hpp` - 派生クラスのヘッダーファイル
- `mydevice.cpp` - 実装スケルトン (関数本体は空、TODOコメント付き)

### 生成例

#### mydevice.hpp
```cpp
#ifndef MYDEVICE_HPP_
#define MYDEVICE_HPP_

#include "idevice.hpp"

class MyDevice : public IDevice {
public:
    MyDevice() = default;
    virtual ~MyDevice() = default;

    MyDevice(const MyDevice&) = delete;
    MyDevice& operator=(const MyDevice&) = delete;

    // Override pure virtual methods from base class
    void initialize() override;
    int read(uint8_t* buffer, size_t size) override;
};

#endif // MYDEVICE_HPP_
```

#### mydevice.cpp
```cpp
#include "mydevice.hpp"

void MyDevice::initialize() {
    // TODO: Implement initialize
}

int MyDevice::read(uint8_t* buffer, size_t size) {
    // TODO: Implement read
}
```

## ライセンス

MIT License

## 関連プロジェクト

- [pre-omusubi](https://github.com/TakumiOkayasu/pre-omusubi) - Omusubi組み込みフレームワーク本体
