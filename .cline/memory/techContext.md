# 技術コンテキスト

## 使用技術

### バックエンド

- **言語**: Go 1.20+
- **フレームワーク/ライブラリ**:
  - `github.com/looplab/fsm`: 有限状態機械の実装
  - `github.com/gorilla/websocket`: WebSocket通信
  - `github.com/gorilla/mux`: HTTPルーティング
  - `go.uber.org/zap`: 構造化ロギング

### フロントエンド

- **言語**: HTML, CSS, JavaScript
- **ライブラリ**:
  - 純粋なJavaScriptを使用（フレームワークなし）
  - WebSocket APIを使用した通信

## 開発環境

### 必要なツール

- Go 1.20以上
- Git
- エディタ（VSCode推奨）
- ブラウザ（Chrome, Firefox, Safari）

### セットアップ手順

1. リポジトリのクローン:
   ```bash
   git clone <repository-url>
   cd state_sample
   ```

2. 依存関係のインストール:
   ```bash
   go mod download
   ```

3. アプリケーションの実行:
   ```bash
   go run main.go
   ```

4. ブラウザでアクセス:
   ```
   http://localhost:8080
   ```

## テスト環境

### テストフレームワーク

- Go標準のテストパッケージ (`testing`)
- `github.com/stretchr/testify/assert`: アサーション
- `github.com/stretchr/testify/mock`: モック

### テスト実行

```bash
# すべてのテストを実行
go test ./...

# 特定のパッケージのテストを実行
go test ./internal/usecase/state/...

# 詳細なログ出力でテストを実行
go test -v ./...

# カバレッジレポートの生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 技術的制約

1. **パフォーマンス**:
   - 多数の状態と条件を効率的に処理する必要がある
   - WebSocket接続のリソース管理に注意

2. **並行処理**:
   - 複数のゴルーチンが同時に状態を変更する可能性がある
   - 適切な同期機構（ミューテックス）を使用する必要がある

3. **メモリ管理**:
   - 長時間実行されるタイマーやゴルーチンのリソース解放
   - オブザーバーパターン使用時のメモリリーク防止

4. **エラー処理**:
   - 不正な状態遷移の適切な処理
   - WebSocket接続エラーの処理
   - クライアント側でのエラー表示

## 依存関係

### 直接的な依存関係

```
github.com/looplab/fsm v1.0.2
github.com/gorilla/websocket v1.5.0
github.com/gorilla/mux v1.8.0
go.uber.org/zap v1.24.0
github.com/stretchr/testify v1.8.2
```

### 間接的な依存関係

```
go.uber.org/atomic v1.10.0
go.uber.org/multierr v1.11.0
github.com/davecgh/go-spew v1.1.1
github.com/pmezard/go-difflib v1.0.0
gopkg.in/yaml.v3 v3.0.1
```

## デプロイ環境

現在はローカル開発環境のみを想定しています。将来的には以下の環境へのデプロイを検討する可能性があります：

- Docker コンテナ
- Kubernetes クラスター
- クラウドプラットフォーム（AWS, GCP, Azure）