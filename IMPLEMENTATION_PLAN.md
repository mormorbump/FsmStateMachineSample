# カウンター条件実装計画

## 1. カウンター戦略の実装

### 1.1 ConditionCounterStrategy構造体
```go
type ConditionCounterStrategy struct {
    currentValue int64
    observers   []core.ConditionPart
}
```

### 1.2 インターフェース実装
```go
// PartStrategyインターフェースを実装
type PartStrategy interface {
    Initialize(part ConditionPart) error
    Evaluate(ctx context.Context, part ConditionPart, params map[string]interface{}) error
    Cleanup() error
}

// ConditionCounterStrategyの実装
func (s *ConditionCounterStrategy) Initialize(part core.ConditionPart) error {
    s.currentValue = 0
    return nil
}

func (s *ConditionCounterStrategy) Evaluate(ctx context.Context, part core.ConditionPart, params map[string]interface{}) error {
    // パラメータから増分値を取得
    increment, ok := params["increment"].(int64)
    if !ok {
        return fmt.Errorf("invalid increment value")
    }

    // カウンター値を更新
    s.currentValue += increment

    // ComparisonOperatorを使用して条件を評価
    satisfied := false
    switch part.ComparisonOperator {
    case ComparisonOperatorEQ:
        satisfied = s.currentValue == part.GetReferenceValueInt()
    case ComparisonOperatorGTE:
        satisfied = s.currentValue >= part.GetReferenceValueInt()
    case ComparisonOperatorLTE:
        satisfied = s.currentValue <= part.GetReferenceValueInt()
    // 他の比較演算子も同様に実装
    }

    if satisfied {
        return part.Complete(ctx)
    }
    return nil
}

func (s *ConditionCounterStrategy) Cleanup() error {
    s.currentValue = 0
    return nil
}
```

## 2. 状態管理の実装

### 2.1 状態遷移の定義
- Unsatisfied → Processing: ボタン押下時
- Processing → Satisfied: ComparisonOperatorによる条件満足時
- Processing → Unsatisfied: リセット時

### 2.2 カウンター状態の管理
- カウンター値はStrategy内部で管理
- 条件評価はComparisonOperatorを使用
- パラメータはEvaluateメソッドで渡す

## 3. APIエンドポイントの実装

### 3.1 新規エンドポイント
```go
POST /api/condition/{condition_id}/part/{part_id}/evaluate
Request Body:
{
    "increment": 2  // 増分値をパラメータとして渡す
}
```

### 3.2 レスポンス形式
```json
{
    "current_value": 10,
    "target_value": 20,
    "comparison_operator": "gte",
    "is_satisfied": false
}
```

## 4. フロントエンド実装

### 4.1 UIコンポーネント
- カウンター表示
- 評価ボタン（増分値を設定可能）
- 現在値/目標値の表示
- 比較演算子の表示

### 4.2 イベントハンドリング
```javascript
async function handleEvaluate(conditionId, partId, increment) {
    const response = await fetch(`/api/condition/${conditionId}/part/${partId}/evaluate`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ increment })
    });
    const data = await response.json();
    updateCounterUI(data);
}
```

## 5. テスト計画

### 5.1 ユニットテスト
- ConditionCounterStrategy
  - 各ComparisonOperatorのテスト
  - パラメータ処理のテスト
  - エラーケースのテスト

### 5.2 統合テスト
- フロントエンド-バックエンド連携
- 状態遷移の検証
- ComparisonOperatorの動作確認

## 6. 実装手順

1. ConditionCounterStrategyの実装
2. ComparisonOperatorを使用した条件評価の実装
3. APIエンドポイントの追加
4. フロントエンドUIの実装
5. テストの実装と実行
6. ドキュメントの更新

## 7. 拡張性考慮事項

### 7.1 新規戦略追加への対応
- PartStrategyインターフェースの維持
- パラメータ渡しによる柔軟な拡張

### 7.2 パラメータ設定の柔軟性
- Evaluateメソッドのパラメータによる動的な振る舞いの制御
- 新しいComparisonOperatorの追加容易性

## 8. エラーハンドリング

### 8.1 考慮すべきエラーケース
- 不正なパラメータ
- 未サポートのComparisonOperator
- 並行アクセス

### 8.2 エラーレスポンス
```json
{
    "error": "invalid_parameters",
    "message": "Invalid increment value provided"
}