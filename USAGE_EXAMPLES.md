# 使用例

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

出力例:
```
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

詳細表示:
```bash
./omusubi-codegen parse --repo ~/projects/omusubi/include --verbose
```

出力例(詳細):
```
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

### 5. 対話式で実装コードを生成

```bash
./omusubi-codegen generate --repo ~/projects/omusubi/include
```

実行例:
```
Searching for abstract classes in repository...

Available abstract classes:
  - IDevice (from /home/user/projects/omusubi/include/device/idevice.hpp)
  - IGpio (from /home/user/projects/omusubi/include/hal/igpio.hpp)
  - ITimer (from /home/user/projects/omusubi/include/hal/itimer.hpp)

Enter base class name: IDevice
Searching for abstract class 'IDevice'...
Found abstract class: IDevice
Namespace: omusubi
Pure virtual methods: 4

Enter derived class name: MyDevice

Generating files for MyDevice...
Generated: ./mydevice.hpp
Generated: ./mydevice.cpp

✓ Code generation completed successfully!
Generated files in: .
```

### 6. コマンドライン引数で直接指定

```bash
./omusubi-codegen generate \
  --repo ~/projects/omusubi/include \
  --base IDevice \
  --class MyDevice \
  --output ./my_implementation
```

出力:
```
Searching for abstract class 'IDevice'...
Found abstract class: IDevice
Namespace: omusubi
Pure virtual methods: 4

Generating files for MyDevice...
Generated: ./my_implementation/mydevice.hpp
Generated: ./my_implementation/mydevice.cpp

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

### 複数のクラスを一括生成(シェルスクリプト)

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

## トラブルシューティング

### エラー: abstract class 'XXX' not found

```bash
# 正確なクラス名を確認
./omusubi-codegen parse --repo ~/projects/omusubi/include --verbose
```

クラス名は大文字小文字を区別します。また、名前空間は含めません。

### エラー: failed to parse repository

tree-sitter-cppが正しくインストールされているか確認:
```bash
# devcontainer内で
tree-sitter --version
```

### 生成されたファイルのエンコーディングがおかしい

テンプレートファイルがUTF-8であることを確認してください:
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
