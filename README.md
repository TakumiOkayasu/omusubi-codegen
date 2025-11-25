# Omusubi Platform Code Generator

Omusubi組み込みフレームワーク用の自動コード生成ツール。
C++インターフェース定義から実装スケルトン、テストコード、ドキュメントを自動生成します。

## 特徴

- **C++17対応**: Omusubiフレームワークに準拠したコード生成
- **Tree-sitterパーサー**: 高速で正確なC++構文解析
- **マルチリポジトリ対応**: コアライブラリとプラットフォーム実装の自動検出
- **PlatformIOプロジェクト生成**: 完全なプロジェクト構造を自動生成
- **自動テスト生成**: Google Testベースのユニットテスト自動生成 (実装予定)
- **ドキュメント生成**: Doxygen形式のドキュメントコメント生成 (実装予定)
- **テンプレートカスタマイズ**: Go text/templateによる柔軟なテンプレート

## 要件

- Go 1.23以上
- C++17対応コンパイラ (GCC 11以上 / Clang 14以上)
- Google Test (テスト生成を使用する場合)

## インストール

### バイナリダウンロード (推奨)

[GitHub Releases](https://github.com/TakumiOkayasu/omusubi-platform-codegen/releases)から、お使いのOSに合わせたバイナリをダウンロードしてください。

### go install

```bash
go install github.com/TakumiOkayasu/omusubi-platform-codegen/cmd/codegen@latest
# バイナリは $GOPATH/bin/codegen としてインストールされます
# シンボリックリンクを作成する場合:
# ln -s $GOPATH/bin/codegen $GOPATH/bin/omusubi-codegen
```

### ソースからビルド

```bash
git clone https://github.com/TakumiOkayasu/omusubi-platform-codegen
cd omusubi-platform-codegen
make build
# バイナリは ./omusubi-codegen として生成されます
```

詳細は[DISTRIBUTION.md](DISTRIBUTION.md)を参照してください。

### 対応プラットフォーム

- Linux (x86_64)
- macOS (Intel)
- macOS (Apple Silicon)

> **Note**: このプロジェクトはtree-sitterを使用しているため、CGOが必須です。
> Linux ARM64など他のプラットフォームが必要な場合は、該当環境で`make build`を実行してください。

## 使い方

### クイックスタート: PlatformIOプロジェクト生成

ワークスペースディレクトリから実行すると、omusubiコアライブラリとプラットフォーム実装を自動検出し、完全なPlatformIOプロジェクトを生成します:

```bash
cd /path/to/workspace  # omusubi/ と omusubi-m5stack/ があるディレクトリ
./omusubi-codegen generate --project --project-name my-m5stack-project
```

**アルファ版 (pre-omusubi) の場合:**
```bash
# pre-omusubi/ と pre-omusubi-m5stack/ を自動検出
./omusubi-codegen generate --legacy-name --project --project-name my-m5stack-project
```

> **Note**: `--legacy-name` フラグは、正式リリース後に削除される予定です。

これにより以下が生成されます:
- `my-m5stack-project/platformio.ini` (相対パスでライブラリを参照)
- `my-m5stack-project/src/main.cpp` (基本的なArduinoセットアップ)
- `my-m5stack-project/include/omusubi/platform/<device_name>/*.hpp` (ヘッダーファイル)
- `my-m5stack-project/src/*.cpp` (実装ファイル)
- `my-m5stack-project/.gitignore`

### 1. リポジトリ内の抽象クラスを一覧表示

```bash
./omusubi-codegen parse --repo /path/to/omusubi/include
```

詳細表示:
```bash
./omusubi-codegen parse --repo /path/to/omusubi/include --verbose
```

### 2. 実装コードのみを生成 (プロジェクト構造なし)

基本的な使い方(対話式・複数選択):
```bash
./omusubi-codegen generate --repo /path/to/omusubi/include
```

ワークスペースの自動検出を使用:
```bash
./omusubi-codegen generate  # カレントディレクトリから omusubi/ を自動検出
```

このコマンドを実行すると:
1. リポジトリ内の抽象クラスが検索されます
2. **「全て選択しますか？」の選択肢が表示されます**
   - `Yes - Select all classes`: 全てのクラスを選択
   - `No - Choose individually`: 個別選択モードへ
3. 個別選択モードでは**矢印キーとスペースで実装したいクラスを複数選択できます**
   - `↑`/`↓`: カーソル移動
   - `Space`: 選択/解除（選択したものに`[x]`マークが表示されます）
   - `Enter`: 確定
4. 派生クラス名のプレフィックスの入力を求められます（デフォルト: "My"）
5. 選択した各クラスに対して、`<プレフィックス><BaseClassName>`の形式で.hppと.cppが生成されます

#### 表示例
```
? Found 3 abstract classes. Select all?
> No - Choose individually
  Yes - Select all classes
```
個別選択を選んだ場合:
```
? Select abstract classes to implement:
  [x] omusubi::IDevice (from testdata/sample/idevice.hpp)
  [x] omusubi::ISensor (from testdata/sample/isensor.hpp)
> [ ] omusubi::IActuator (from testdata/sample/iactuator.hpp)
```
- `[x]`: 選択済み（緑色で表示）
- `[ ]`: 未選択
- `>`: 現在のカーソル位置

例: プレフィックスに "Custom" を指定し、IDevice と ISensor を選択した場合
- `custom_idevice.hpp` / `custom_idevice.cpp`
- `custom_isensor.hpp` / `custom_isensor.cpp`

**注意**: ファイル名は自動的にスネークケース (`snake_case`) に変換されます。

### 3. コマンドライン引数で指定

```bash
./omusubi-codegen generate \
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
- `my_device.hpp` - 派生クラスのヘッダーファイル
- `my_device.cpp` - 実装スケルトン (関数本体は空、TODOコメント付き)

**注意**:
- ファイル名は自動的にスネークケース (`snake_case`) に変換されます
- インクルードガードは `#pragma once` を使用します（モダンC++標準）
- 元のヘッダーファイルが `.h` の場合、生成されるファイルも `.h` と `.c` になります
- 元のヘッダーファイルが `.hpp` の場合、生成されるファイルは `.hpp` と `.cpp` になります

### 生成例

#### my_device.hpp
```cpp
#pragma once

#include "i_device.hpp"

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
```

#### my_device.cpp
```cpp
#include "my_device.hpp"

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
