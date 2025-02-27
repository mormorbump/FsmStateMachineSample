# ステートマシン実装のテスト計画

## テスト計画の概要

リファクタリング後のコードの品質を確保するために、以下のテスト計画を実施します。

## 1. ユニットテスト

### 1.1 エンティティのテスト

#### 1.1.1 ConditionPartのテスト (`internal/domain/entity/condition_part_test.go`)

```go
// テストの概要
// 1. 基本的な状態遷移のテスト
// 2. 戦略パターンとの連携テスト
// 3. オブザーバーパターンのテスト
// 4. バリデーションのテスト
// 5. エッジケースのテスト
```

テストケース:
- 新規作成時の初期状態が `StateReady` であることを確認
- `Activate` 後の状態が `StateUnsatisfied` であることを確認
- `Process` 後の状態が `StateProcessing` であることを確認
- `Complete` 後の状態が `StateSatisfied` であることを確認
- `Reset` 後の状態が `StateReady` に戻ることを確認
- 無効な状態遷移（例: `Ready` から `Satisfied` への直接遷移）がエラーを返すことを確認
- モック戦略を使用して、戦略の評価結果に基づいて状態が正しく遷移することを確認
- オブザーバーが正しく通知を受け取ることを確認
- バリデーションが正しく機能することを確認（無効な比較演算子など）

#### 1.1.2 Conditionのテスト (`internal/domain/entity/condition_test.go`)

```go
// テストの概要
// 1. 基本的な状態遷移のテスト
// 2. 条件パーツの管理テスト
// 3. オブザーバーパターンのテスト
// 4. バリデーションのテスト
// 5. エッジケースのテスト
```

テストケース:
- 新規作成時の初期状態が `StateReady` であることを確認
- `Activate` 後の状態が `StateUnsatisfied` であることを確認
- すべての条件パーツが満たされた後、状態が `StateSatisfied` に遷移することを確認
- `Reset` 後の状態が `StateReady` に戻ることを確認
- 条件パーツの追加と取得が正しく機能することを確認
- オブザーバーが正しく通知を受け取ることを確認
- バリデーションが正しく機能することを確認（条件パーツがない場合など）

### 1.2 戦略のテスト

#### 1.2.1 CounterStrategyのテスト (`internal/usecase/strategy/counter_strategy_test.go`)

```go
// テストの概要
// 1. 初期化のテスト
// 2. 評価ロジックのテスト
// 3. オブザーバー通知のテスト
// 4. クリーンアップのテスト
```

テストケース:
- 初期化後の現在値が0であることを確認
- 評価後の現在値が正しく更新されることを確認
- 各比較演算子（EQ, NEQ, GT, GTE, LT, LTE, Between）に対して正しく評価されることを確認
- 条件が満たされた場合に `EventComplete` が通知されることを確認
- 条件が満たされていない場合に `EventProcess` が通知されることを確認
- クリーンアップ後にリソースが正しく解放されることを確認

#### 1.2.2 TimeStrategyのテスト (`internal/usecase/strategy/time_strategy_test.go`)

```go
// テストの概要
// 1. 初期化のテスト
// 2. タイマー開始のテスト
// 3. タイマーイベントのテスト
// 4. クリーンアップのテスト
```

テストケース:
- 初期化後の間隔が正しく設定されることを確認
- `Start` 後にタイマーが開始されることを確認
- タイマーイベントが発生した後に `EventTimeout` が通知されることを確認
- クリーンアップ後にタイマーが停止されることを確認

#### 1.2.3 StrategyFactoryのテスト (`internal/usecase/strategy/strategy_factory_test.go`)

```go
// テストの概要
// 1. 各種戦略の作成テスト
// 2. 未知の戦略種類に対するエラー処理のテスト
```

テストケース:
- `KindTime` に対して `TimeStrategy` が作成されることを確認
- `KindCounter` に対して `CounterStrategy` が作成されることを確認
- 未知の種類に対してエラーが返されることを確認

## 2. 統合テスト

### 2.1 状態遷移の統合テスト

```go
// テストの概要
// 1. 条件パーツと条件の連携テスト
// 2. 戦略と条件パーツの連携テスト
// 3. 複雑なシナリオのテスト
```

テストケース:
- 複数の条件パーツを持つ条件が、すべてのパーツが満たされた場合にのみ満たされることを確認
- 実際の戦略を使用して、条件パーツの状態が正しく遷移することを確認
- 複雑なシナリオ（例: 一部の条件パーツが満たされ、一部がリセットされる）が正しく処理されることを確認

## 3. モックの作成

テストで使用するモックを作成します。

### 3.1 StrategyObserverのモック

```go
// MockStrategyObserver は StrategyObserver インターフェースのモック実装です
type MockStrategyObserver struct {
	OnUpdatedFunc func(event string)
	Events        []string
}

// OnUpdated はイベントを記録します
func (m *MockStrategyObserver) OnUpdated(event string) {
	m.Events = append(m.Events, event)
	if m.OnUpdatedFunc != nil {
		m.OnUpdatedFunc(event)
	}
}
```

### 3.2 ConditionPartObserverのモック

```go
// MockConditionPartObserver は ConditionPartObserver インターフェースのモック実装です
type MockConditionPartObserver struct {
	OnConditionPartChangedFunc func(part interface{})
	Parts                      []interface{}
}

// OnConditionPartChanged はパーツの変更を記録します
func (m *MockConditionPartObserver) OnConditionPartChanged(part interface{}) {
	m.Parts = append(m.Parts, part)
	if m.OnConditionPartChangedFunc != nil {
		m.OnConditionPartChangedFunc(part)
	}
}
```

### 3.3 PartStrategyのモック

```go
// MockPartStrategy は PartStrategy インターフェースのモック実装です
type MockPartStrategy struct {
	InitializeFunc    func(part interface{}) error
	GetCurrentValueFunc func() interface{}
	StartFunc         func(ctx context.Context, part interface{}) error
	EvaluateFunc      func(ctx context.Context, part interface{}, params interface{}) error
	CleanupFunc       func() error
	observers         []service.StrategyObserver
}

// Initialize はモックの初期化関数を呼び出します
func (m *MockPartStrategy) Initialize(part interface{}) error {
	if m.InitializeFunc != nil {
		return m.InitializeFunc(part)
	}
	return nil
}

// GetCurrentValue は現在の値を返します
func (m *MockPartStrategy) GetCurrentValue() interface{} {
	if m.GetCurrentValueFunc != nil {
		return m.GetCurrentValueFunc()
	}
	return nil
}

// Start はモックの開始関数を呼び出します
func (m *MockPartStrategy) Start(ctx context.Context, part interface{}) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx, part)
	}
	return nil
}

// Evaluate はモックの評価関数を呼び出します
func (m *MockPartStrategy) Evaluate(ctx context.Context, part interface{}, params interface{}) error {
	if m.EvaluateFunc != nil {
		return m.EvaluateFunc(ctx, part, params)
	}
	return nil
}

// Cleanup はモックのクリーンアップ関数を呼び出します
func (m *MockPartStrategy) Cleanup() error {
	if m.CleanupFunc != nil {
		return m.CleanupFunc()
	}
	return nil
}

// AddObserver はオブザーバーを追加します
func (m *MockPartStrategy) AddObserver(observer service.StrategyObserver) {
	m.observers = append(m.observers, observer)
}

// RemoveObserver はオブザーバーを削除します
func (m *MockPartStrategy) RemoveObserver(observer service.StrategyObserver) {
	for i, obs := range m.observers {
		if obs == observer {
			m.observers = append(m.observers[:i], m.observers[i+1:]...)
			break
		}
	}
}

// NotifyUpdate はオブザーバーに通知します
func (m *MockPartStrategy) NotifyUpdate(event string) {
	for _, observer := range m.observers {
		observer.OnUpdated(event)
	}
}
```

## 4. テスト実行計画

1. 各コンポーネントのユニットテストを実装
2. 統合テストを実装
3. テストの自動化（CI/CDパイプラインへの組み込み）
4. コードカバレッジの測定と改善

## 5. テスト実装の優先順位

1. 基本的なエンティティのテスト（ConditionPart, Condition）
2. 戦略のテスト（CounterStrategy, TimeStrategy）
3. ファクトリのテスト（StrategyFactory）
4. 統合テスト

## 6. テストの実行方法

```bash
# 全てのテストを実行
go test ./...

# 特定のパッケージのテストを実行
go test ./internal/domain/entity/...
go test ./internal/usecase/strategy/...

# カバレッジレポートの生成
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out