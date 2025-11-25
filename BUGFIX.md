# バグ修正: パーサーのnil pointer dereference

## 問題の概要

CIでビルドされたバイナリを実行すると、以下のエラーが発生していました：

```text
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x2 addr=0x20 pc=0x100aba22c]
```

## 原因

`internal/parser/parser.go` の `ParseSource` メソッドで、`p.parser.Parse(nil, source)` を呼び出していましたが、この第1引数の `nil` が問題でした。

### 詳細

1. **主な原因**: `Parse(nil, source)` の第1引数に `nil` を渡していた
   - `Parse` メソッドは内部的に `ParseCtx` を呼び出します
   - `ParseCtx` の第1引数は `context.Context` で、これが `nil` だとパニックが発生します

2. **なぜローカルでは動作していたか**:
   - ローカル環境とCI環境でのtree-sitterライブラリのバージョンや最適化レベルの違い
   - CGOを使用しているため、環境依存の問題が発生しやすい

## 修正内容

### 1. contextパッケージのインポート追加

```go
import (
    "context"
    // ... 他のインポート
)
```

### 2. ParseSourceメソッドの修正

**修正前:**

```go
func (p *Parser) ParseSource(source []byte) (*model.FileInfo, error) {
    tree := p.parser.Parse(nil, source)  // ❌ nilを渡していた
    if tree == nil {
        return nil, fmt.Errorf("failed to parse source: tree is nil")
    }
    defer tree.Close()
    // ...
}
```

**修正後:**

```go
func (p *Parser) ParseSource(source []byte) (*model.FileInfo, error) {
    ctx := context.Background()  // ✅ 有効なcontextを作成
    tree, err := p.parser.ParseCtx(ctx, nil, source)  // ✅ ParseCtxを明示的に使用
    if err != nil {
        return nil, fmt.Errorf("failed to parse source: %w", err)
    }
    if tree == nil {
        return nil, fmt.Errorf("failed to parse source: tree is nil")
    }
    defer tree.Close()
    // ...
}
```

### 主な変更点

1. `context.Background()` で有効なcontextを作成
2. `Parse` の代わりに `ParseCtx` を明示的に使用
3. エラーハンドリングを追加（`tree` と `err` の両方をチェック）

## 影響範囲

- `internal/parser/parser.go` の `ParseSource` メソッドのみ
- すべての parse/generate コマンドの安定性が向上

## テスト方法

```bash
# ビルド
make build

# 修正されたバイナリでテスト
./omusubi-codegen parse --repo /path/to/omusubi/include --verbose
./omusubi-codegen generate --repo /path/to/omusubi/include
```

## 今後の推奨事項

1. **CI/CDパイプラインでの統合テスト追加**
   - 実際のリポジトリに対してparse/generateコマンドを実行するテスト

2. **tree-sitterのバージョン固定**
   - go.mod で tree-sitter のバージョンを明示的に管理

3. **エラーハンドリングの強化**
   - 他の箇所でも同様のパターンがないか確認

## 参考

- tree-sitter Go bindings: <https://github.com/smacker/go-tree-sitter>
- Go context package: <https://pkg.go.dev/context>
