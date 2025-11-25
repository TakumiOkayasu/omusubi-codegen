# 使用例

## クイックスタート: PlatformIOプロジェクト生成

最も一般的な使用方法です。ワークスペースディレクトリから実行すると、完全なPlatformIOプロジェクトを生成できます。

```bash
# ワークスペース構造
# workspace/
# ├── omusubi/              # コアライブラリ
# └── omusubi-m5stack/      # プラットフォーム実装

cd /path/to/workspace
./omusubi-codegen generate --project --project-name my-m5stack-project
```

生成されるファイル：

- `my-m5stack-project/platformio.ini`
- `my-m5stack-project/src/main.cpp`
- `my-m5stack-project/include/omusubi/platform/<device_name>/*.hpp`
- `my-m5stack-project/src/*.cpp`
- `my-m5stack-project/.gitignore`

## 基本的なワークフロー

### 1. omusubiリポジトリをクローン

```bash
cd ~/projects
git clone https://github.com/TakumiOkayasu/omusubi
```

### 2. devcontainerでcodegen環境をセットアップ

```bash
cd omusubi-codegen
code .  # VSCodeで開く
# "Reopen in Container"を選択
```

### 3. ビルド

```bash
make build
```

### 4. リポジトリ内の抽象クラスを確認

```bash
./omusubi-codegen parse --repo ~/projects/omusubi/include
```

出力例：

```text
Parsing repository: /home/user/projects/omusubi/include

Found 3 file(s) with abstract classes:

File: /home/user/projects/omusubi/include/device/idevice.hpp
Namespace: omusubi

  Class: IDevice
  Pure virtual methods: 4

File: /home/user/projects/omusubi/include/hal/igpio.hpp
Namespace: omusubi::hal

  Class: IGpio
  Pure virtual methods: 6
...
```

詳細表示：

```bash
./omusubi-codegen parse --repo ~/projects/omusubi/include --verbose
```

出力例(詳細)：

```text
File: /home/user/projects/omusubi/include/device/idevice.hpp
Namespace: omusubi

  Class: IDevice
  Pure virtual methods: 4
  Methods:
    - void initialize()
    - int read(uint8_t* buffer, size_t size)
    - int write(const uint8_t* buffer, size_t size)
    - void reset()
```

### 5. 対話式で実装コードを生成（マルチセレクト対応）

```bash
./omusubi-codegen generate --repo ~/projects/omusubi/include
```

実行例：

```text
Searching for abstract classes in repository...

? Found 3 abstract classes. Select all?
> No - Choose individually
  Yes - Select all classes
```

個別選択を選んだ場合：

```text
? Select abstract classes to implement:
  [x] omusubi::IDevice (from device/idevice.hpp)
  [x] omusubi::IGpio (from hal/igpio.hpp)
> [ ] omusubi::ITimer (from hal/itimer.hpp)
```

操作方法：

- `↑`/`↓`: カーソル移動
- `Space`: 選択/解除
- `Enter`: 確定

```text
? Enter class name prefix (default: My): Custom

Generating files...
Generated: ./custom_device.hpp
Generated: ./custom_device.cpp
Generated: ./custom_gpio.hpp
Generated: ./custom_gpio.cpp

✓ Code generation completed successfully!
Generated files in: .
```

**注意**: ファイル名は自動的にスネークケース (`snake_case`) に変換されます。

### 6. コマンドライン引数で直接指定

```bash
./omusubi-codegen generate \
  --repo ~/projects/omusubi/include \
  --base IDevice \
  --class MyDevice \
  --output ./my_implementation
```

出力：

```text
Searching for abstract class 'IDevice'...
Found abstract class: IDevice
Namespace: omusubi
Pure virtual methods: 4

Generating files for MyDevice...
Generated: ./my_implementation/my_device.hpp
Generated: ./my_implementation/my_device.cpp

✓ Code generation completed successfully!
Generated files in: ./my_implementation
```

## 高度な使用例

### カスタムテンプレートを使用

```bash
# 独自のテンプレートディレクトリを作成
cp -r internal/generator/templates my_templates
# my_templates/*.tmpl を編集

# カスタムテンプレートで生成
./omusubi-codegen generate \
  --repo ~/projects/omusubi/include \
  --base IDevice \
  --class MyDevice \
  --templates ./my_templates
```

### 複数のクラスを一括生成

対話式のマルチセレクト機能を使って、複数のクラスを一度に生成できます:

```bash
./omusubi-codegen generate --repo ~/projects/omusubi/include

# "Yes - Select all classes" を選択するか、
# 個別選択モードでSpaceキーで複数選択してEnter
```

または、シェルスクリプトで自動化する場合:

```bash
#!/bin/bash
REPO_PATH=~/projects/omusubi/include
OUTPUT_DIR=./generated

CLASSES=(
  "IDevice:MyDevice"
  "IGpio:MyGpio"
  "ITimer:MyTimer"
)

for entry in "${CLASSES[@]}"; do
  IFS=':' read -r base derived <<< "$entry"
  echo "Generating $derived from $base..."

  ./omusubi-codegen generate \
    --repo "$REPO_PATH" \
    --base "$base" \
    --class "$derived" \
    --output "$OUTPUT_DIR"
done

echo "All classes generated in $OUTPUT_DIR"
```

### PlatformIOプロジェクトの詳細な設定

```bash
# ボードとライブラリパスを明示的に指定
./omusubi-codegen generate \
  --project \
  --project-name my-custom-project \
  --core-lib ./omusubi \
  --platform-lib ./omusubi-m5stack \
  --board esp32-s3-devkitc-1
```

### アルファ版 (pre-omusubi) の場合

```bash
# pre-omusubi/ と pre-omusubi-m5stack/ を自動検出
./omusubi-codegen generate --legacy-name --project --project-name my-project
```

> **Note**: `--legacy-name` フラグは正式リリース後に削除される予定です。

## トラブルシューティング

### エラー: abstract class 'XXX' not found

```bash
# 正確なクラス名を確認
./omusubi-codegen parse --repo ~/projects/omusubi/include --verbose
```

クラス名は大文字小文字を区別します。また、名前空間は含めません。

### エラー: failed to parse repository

tree-sitter-cppが正しくインストールされているか確認：

```bash
# devcontainer内で
tree-sitter --version
```

### 生成されたファイルのエンコーディングがおかしい

テンプレートファイルがUTF-8であることを確認してください：

```bash
file internal/generator/templates/*.tmpl
```

## ヘルプ

```bash
# 全般的なヘルプ
./omusubi-codegen --help

# generateコマンドのヘルプ
./omusubi-codegen generate --help

# parseコマンドのヘルプ
./omusubi-codegen parse --help

# バージョン情報
./omusubi-codegen --version
```
