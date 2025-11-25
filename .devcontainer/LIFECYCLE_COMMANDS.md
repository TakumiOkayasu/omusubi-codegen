# DevContainer Lifecycle Commands

このdevcontainerで使用されているライフサイクルコマンドの説明です。

## 📋 実装されているコマンド

### 1. `initializeCommand` (ホストマシンで実行)

```json
"initializeCommand": "test -f ${localWorkspaceFolder}/.env || cp ${localWorkspaceFolder}/.env.example ${localWorkspaceFolder}/.env"
```

**実行タイミング:** コンテナビルド前 (ホストマシン上)

**目的:**

- `.env`ファイルが存在しない場合、`.env.example`からコピー
- 環境変数の初期設定を保証

**実行頻度:** 毎回

---

### 2. `onCreateCommand` (コンテナ内で実行 - 初回のみ)

```json
"onCreateCommand": {
  "go-deps": "go mod download || true",
  "go-verify": "go mod verify || true"
}
```

**実行タイミング:** コンテナが最初に作成されたとき (1回のみ)

**目的:**

- Go依存関係を事前ダウンロード
- 依存関係の整合性を検証
- `|| true`でエラーがあっても継続

**実行頻度:** コンテナ初回作成時のみ

**利点:**

- 並列実行可能 (複数のキーを指定)
- 失敗してもコンテナ作成が止まらない
- 初回セットアップに最適

---

### 3. `updateContentCommand` (コンテナ内で実行 - 更新時)

```json
"updateContentCommand": "go mod download || true"
```

**実行タイミング:**

- `go.mod`が更新されたとき
- コンテナが再ビルドされたとき
- devcontainer設定が変更されたとき

**目的:**

- 最新の依存関係を取得
- `go.mod`の変更に自動対応

**実行頻度:** コンテンツ更新時

---

## 🔄 ライフサイクルコマンドの実行順序

```text
1. initializeCommand      (ホスト)
   ↓
2. Docker Build
   ↓
3. Container Start
   ↓
4. onCreateCommand        (コンテナ内 - 初回のみ)
   ↓
5. updateContentCommand   (コンテナ内 - 更新時)
   ↓
6. postCreateCommand      (コンテナ内 - 毎回) ← 使用していません
   ↓
7. postStartCommand       (コンテナ内 - 起動時) ← 使用していません
   ↓
8. postAttachCommand      (コンテナ内 - アタッチ時) ← 使用していません
```

## 📝 各コマンドの使い分け

### `initializeCommand`

- ✅ ホスト環境のセットアップ
- ✅ `.env`ファイルの準備
- ✅ git設定の確認

### `onCreateCommand` ⭐ 推奨

- ✅ 初回のみ実行される重い処理
- ✅ 依存関係のダウンロード
- ✅ データベース初期化
- ✅ 並列実行可能

### `updateContentCommand` ⭐ 推奨

- ✅ `go.mod`などの依存関係ファイル更新時
- ✅ 軽量な更新処理

### `postCreateCommand` (非推奨)

- ❌ 毎回実行されるため遅い
- ❌ エラー時にコンテナ作成が止まる
- ➡️ `onCreateCommand`を使用推奨

### `postStartCommand`

- ✅ サービス起動
- ✅ デーモンプロセス起動
- ⚠️ 毎回実行されるので軽量処理のみ

### `postAttachCommand`

- ✅ ウェルカムメッセージ
- ✅ 環境確認コマンド
- ⚠️ 頻繁に実行されるので非常に軽量な処理のみ

## 🎯 このプロジェクトの戦略

### なぜ`onCreateCommand`を使うのか？

1. **初回のみ実行**
   - `go mod download`は初回だけでOK
   - 2回目以降はキャッシュが効く

2. **並列実行**

   ```json
   "onCreateCommand": {
     "go-deps": "go mod download || true",
     "go-verify": "go mod verify || true"
   }
   ```

   - 複数のコマンドを並列実行可能
   - 高速化

3. **失敗に寛容**
   - `|| true`でエラーを無視
   - コンテナ作成が止まらない

### なぜ`updateContentCommand`も使うのか？

1. **go.mod更新検知**
   - `go.mod`が変更されたときのみ実行
   - 無駄な再ダウンロードを回避

2. **最新依存関係**
   - 常に最新の依存関係を保証

## ⚡ パフォーマンス比較

| コマンド | 実行頻度 | 初回起動 | 2回目起動 |
|---------|---------|---------|----------|
| `postCreateCommand` | 毎回 | 遅い | 遅い ❌ |
| `onCreateCommand` | 初回のみ | 遅い | 速い ✅ |
| `updateContentCommand` | 更新時のみ | 遅い | 速い ✅ |

## 🛠️ トラブルシューティング

### コマンドが実行されない

```bash
# コンテナを完全に削除して再作成
docker compose -f .devcontainer/compose.yaml down -v
# VSCodeで "Rebuild Container"
```

### コマンドの実行ログを確認

VSCode: `View` → `Output` → `Dev Containers`でログを確認

### 手動で実行

```bash
# コンテナ内で
go mod download
go mod verify
```

## 📚 参考資料

- [devcontainer.json reference](https://containers.dev/implementors/json_reference/)
- [Lifecycle scripts](https://containers.dev/implementors/json_reference/#lifecycle-scripts)
