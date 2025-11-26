# omusubi-codegen 問題点一覧

本ドキュメントは omusubi-codegen プロジェクトの問題点を洗い出した結果をまとめたものです。

**調査日**: 2025-11-26
**テストカバレッジ**: 31.2%

---

## 目次

1. [優先度 HIGH - 即座に対応推奨](#優先度-high---即座に対応推奨)
2. [優先度 MEDIUM - 次回リリースで対応](#優先度-medium---次回リリースで対応)
3. [優先度 LOW - 将来対応](#優先度-low---将来対応)
4. [評価サマリー](#評価サマリー)

---

## 優先度 HIGH - 即座に対応推奨

### 1. 未使用の変数・パラメータ

**ファイル**: `cmd/codegen/main.go`

`baseClass` 変数が宣言され、フラグとして定義されているが、`runGenerate()` 関数内で**一切使用されていない**。

```go
// 行48: 宣言
baseClass       string

// 行81-83: フラグ定義
cmd.Flags().StringVarP(&baseClass, "base", "b", "",
    "Base class name to search for (optional, will prompt if not provided)\n"+
    "Example: --base IDevice")

// 行132: 関数パラメータとして受け取るが未使用
func runGenerate(repoPath, baseClass, className, ...) error {
    // baseClass は関数内で参照されていない
}
```

**対応案**:z
- 削除する、または
- `--base` フラグで指定されたクラスを自動選択する機能を実装する

---

### 2. generator.go の重複コード

**ファイル**: `internal/generator/generator.go`

以下の5つの関数が同じパターン（テンプレート読込 → Execute → WriteFile）を繰り返している:

| 関数 | 行番号 |
|------|--------|
| `generateHeader()` | 69-95 |
| `generateSource()` | 99-122 |
| `generatePlatformIOConfig()` | 297-313 |
| `generateMainCpp()` | 317-333 |
| `generateGitignore()` | 337-353 |

**対応案**: 汎用の `renderAndWrite()` ヘルパー関数を作成して約100行削減

---

### 3. テストカバレッジ不足

**現状カバレッジ**: 31.2%

| パッケージ | カバレッジ | 状態 |
|-----------|-----------|------|
| `cmd/codegen` | 0.0% | 未テスト |
| `internal/model` | 0.0% | 未テスト |
| `internal/generator` | 48.2% | 部分的 |
| `internal/parser` | 47.3% | 部分的 |

**未テストの重要関数**:

| 関数 | ファイル | 重要度 |
|------|----------|--------|
| `ParseFile()` | parser.go:34 | 高 |
| `ParseDirectory()` | parser.go:346 | 高 |
| `FindAbstractClass()` | parser.go:387 | 高 |
| `DetectWorkspace()` | parser.go:416 | 高 |
| `extractFieldInfo()` | parser.go:275 | 中 |
| `generateGitignore()` | generator.go:337 | 低 |
| `AccessLevel.String()` | model.go:54 | 低 |

**対応案**:
- `cmd/codegen/` に統合テストを追加
- `DetectWorkspace()` と `ParseDirectory()` のテストを優先作成
- 目標カバレッジ: 60%以上

---

### 4. CI/CD にテストステップがない

**ファイル**: `.github/workflows/release.yml`

リリースワークフローにテスト実行ステップが含まれていない。テストが失敗してもリリースが実行される可能性がある。

**対応案**:
1. `release.yml` にテストステップを追加
2. PR用の `test.yml` ワークフローを新規作成

```yaml
# 追加すべきステップ
- name: Run tests
  run: go test ./...

- name: Run linter
  run: golangci-lint run ./...
```

---

### 5. panic の使用

**ファイル**: `internal/parser/parser.go:25`

```go
func New() *Parser {
    parser := sitter.NewParser()
    lang := cpp.GetLanguage()
    if lang == nil {
        panic("failed to get C++ language from tree-sitter")
    }
    // ...
}
```

`panic` は予期しないプログラム終了を引き起こす。エラーハンドリングに変更すべき。

**対応案**:
```go
func New() (*Parser, error) {
    parser := sitter.NewParser()
    lang := cpp.GetLanguage()
    if lang == nil {
        return nil, fmt.Errorf("failed to get C++ language from tree-sitter")
    }
    // ...
    return p, nil
}
```

---

## 優先度 MEDIUM - 次回リリースで対応

### 6. 過度に複雑な関数

**ファイル**: `cmd/codegen/main.go`
**関数**: `runGenerate()` (行132-365)

- **234行**の超長関数
- **11個のパラメータ**
- 5つの異なる責任を持つ:
  1. ワークスペース検出
  2. ディレクトリ解析
  3. ユーザープロンプト処理
  4. プロジェクト生成
  5. ファイル生成ループ

**対応案**: 以下のように分割
- `detectWorkspacePaths()` - ワークスペース検出
- `collectAbstractClasses()` - クラス収集
- `promptUserSelection()` - ユーザー選択
- `generateSelectedClasses()` - ファイル生成

---

### 7. エラーハンドリングの不一貫性

**ファイル**: `internal/generator/generator.go`

一部の関数は `fmt.Errorf` でラップしているが、他はそのまま `return err` している。

```go
// ラップあり（良い）
return fmt.Errorf("failed to write header file: %w", err)

// ラップなし（改善の余地あり）
if err != nil {
    return err
}
```

**対応案**: すべてのエラーに `fmt.Errorf` でコンテキストを追加

---

### 8. golangci.yml がない

プロジェクト固有の lint ルールが定義されていない。

**対応案**: `.golangci.yml` を作成

```yaml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - misspell

linters-settings:
  errcheck:
    check-type-assertions: true
```

---

### 9. GoDoc コメント不足

パブリック関数に GoDoc コメントがない、または不十分。

**例**: `internal/parser/parser.go`

```go
// コメントなし
func (p *Parser) ParseFile(filePath string) (*model.FileInfo, error) {
```

**対応案**: すべてのパブリック関数に GoDoc コメントを追加

```go
// ParseFile parses a single C++ header file and extracts class information.
// It returns a FileInfo containing all classes found in the file.
func (p *Parser) ParseFile(filePath string) (*model.FileInfo, error) {
```

---

### 10. parser.go と generator.go が長すぎる

| ファイル | 行数 | 推奨 |
|----------|------|------|
| parser.go | 482行 | 300行以下 |
| generator.go | 355行 | 300行以下 |
| main.go | 458行 | 200行以下 |

**対応案**:
- `parser.go`: workspace 検出を `workspace.go` に分離
- `generator.go`: project 生成を `project.go` に分離
- `main.go`: コマンドハンドラを別ファイルに分離

---

## 優先度 LOW - 将来対応

### 11. マジックナンバーのハードコーディング

ファイルパーミッション `0755` や `0644` が複数箇所でハードコードされている。

**対応案**: 定数として定義

```go
const (
    DirPermission  = 0755
    FilePermission = 0644
)
```

---

### 12. devcontainer.json のセキュリティ

**ファイル**: `.devcontainer/devcontainer.json`

```json
"postStartCommand": "sudo chmod -R 777 /go 2>/dev/null || true"
```

`chmod 777` は過剰な権限。`755` で十分。

---

### 13. 日本語コメントの混在

`parser.go` 162行に日本語コメントがある。プロジェクト全体の言語を統一すべき。

---

### 14. ベンチマークテストがない

大規模リポジトリでのパフォーマンス測定ができない。

**対応案**: `internal/parser/parser_bench_test.go` を作成

```go
func BenchmarkParseDirectory(b *testing.B) {
    // ...
}
```

---

### 15. testify ライブラリの未使用

現在は標準の `testing` パッケージのみ使用。`testify` を使用するとテストコードが簡潔になる。

**対応案**: `github.com/stretchr/testify` を依存関係に追加

---

## 評価サマリー

| 項目 | 評価 | コメント |
|------|------|---------|
| ディレクトリ構成 | ⭐⭐⭐⭐⭐ | 明確で階層的 |
| コード品質 | ⭐⭐⭐⭐ | 一部改善の余地あり |
| テストカバレッジ | ⭐⭐⭐ | 31.2%（目標: 60%以上） |
| ドキュメント | ⭐⭐⭐⭐⭐ | 非常に充実 |
| 依存関係管理 | ⭐⭐⭐⭐⭐ | 厳密に管理 |
| エラーハンドリング | ⭐⭐⭐⭐ | 概ね良好 |
| ビルド・リリース | ⭐⭐⭐⭐ | テストステップ追加で満点 |
| CI/CD | ⭐⭐⭐ | PRテストワークフロー必要 |

**総合評価**: 6/10 → 改善後目標: 8/10

---

## 推奨アクションプラン

### フェーズ1: 即座に対応（1-2日）
1. `baseClass` 変数の削除または実装
2. `panic` をエラー返却に変更
3. `release.yml` にテストステップ追加

### フェーズ2: 短期対応（1週間）
4. `test.yml` ワークフロー新規作成
5. `DetectWorkspace()` と `ParseDirectory()` のテスト追加
6. `.golangci.yml` 作成

### フェーズ3: 中期対応（2-3週間）
7. `runGenerate()` のリファクタリング
8. `generator.go` の重複コード削減
9. GoDoc コメント追加
10. テストカバレッジ 60% 達成
