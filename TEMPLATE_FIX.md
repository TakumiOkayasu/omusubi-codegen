# テンプレート埋め込み問題の修正方法

## 問題

CIでビルドされたバイナリでは、テンプレートファイルが外部ファイルとして存在しないため、以下のエラーが発生します：

```
failed to parse template internal/template/templates/class_header.tmpl:
open internal/template/templates/class_header.tmpl: no such file or directory
```

## 原因

`internal/generator/generator.go` の `loadTemplate` 関数が `template.ParseFiles()` を使用してファイルシステムから直接テンプレートを読み込もうとしています。

```go
func (g *Generator) loadTemplate(name string) (*template.Template, error) {
	tmplPath := filepath.Join(g.templateDir, name)
	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(tmplPath)
	// ...
}
```

## 解決策

Go 1.16+ の `embed` パッケージを使用してテンプレートをバイナリに埋め込みます。

### ステップ1: テンプレートファイルの配置

テンプレートファイルを `internal/generator/templates/` ディレクトリに配置します：

```bash
mkdir -p internal/generator/templates
cp internal/template/templates/*.tmpl internal/generator/templates/
```

### ステップ2: `internal/generator/templates.go` を作成

```go
package generator

import (
	"embed"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS
```

### ステップ3: `internal/generator/generator.go` を修正

**修正前:**
```go
func (g *Generator) loadTemplate(name string) (*template.Template, error) {
	tmplPath := filepath.Join(g.templateDir, name)

	funcMap := template.FuncMap{
		"formatParameters":      formatParameters,
		"formatMethodSignature": formatMethodSignature,
		"toLower":              strings.ToLower,
		"toUpper":              strings.ToUpper,
		"toSnakeCase":          toSnakeCase,
	}

	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	return tmpl, nil
}
```

**修正後:**
```go
func (g *Generator) loadTemplate(name string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"formatParameters":      formatParameters,
		"formatMethodSignature": formatMethodSignature,
		"toLower":              strings.ToLower,
		"toUpper":              strings.ToUpper,
		"toSnakeCase":          toSnakeCase,
	}

	// 埋め込みファイルシステムからテンプレートを読み込む
	tmplPath := "templates/" + name
	tmplContent, err := templatesFS.ReadFile(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return tmpl, nil
}
```

### ステップ4: ビルドして確認

```bash
make build
./omusubi-codegen generate --repo /path/to/pre-omusubi
```

## 代替案: 開発時のフォールバック

開発時はファイルシステムから、本番ビルド時は埋め込みから読み込むハイブリッド方式も可能です：

```go
func (g *Generator) loadTemplate(name string) (*template.Template, error) {
	funcMap := template.FuncMap{
		// ...
	}

	var tmplContent []byte
	var err error

	// まず埋め込みファイルシステムから試す
	tmplPath := "templates/" + name
	tmplContent, err = templatesFS.ReadFile(tmplPath)

	// 埋め込みから読めない場合、ファイルシステムから読む（開発時）
	if err != nil {
		fsPath := filepath.Join(g.templateDir, name)
		tmplContent, err = os.ReadFile(fsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", name, err)
		}
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return tmpl, nil
}
```

## 確認事項

1. `internal/generator/templates/*.tmpl` が存在すること
2. `go:embed` ディレクティブのパスが正しいこと
3. Go 1.16以上を使用していること
4. ビルド後、バイナリにテンプレートが埋め込まれていること

## テスト

```bash
# ビルド
make build

# 生成されたバイナリを別のディレクトリに移動してテスト
cp ./omusubi-codegen /tmp/
cd /tmp
./omusubi-codegen generate --repo /path/to/pre-omusubi
```

これで、テンプレートファイルが外部に存在しなくてもバイナリが正常に動作するはずです。
