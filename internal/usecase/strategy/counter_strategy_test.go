package strategy

import (
	"context"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/value"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockStrategyObserver は StrategyObserver インターフェースのモック実装です
type MockStrategyObserver struct {
	Events []string
}

// OnUpdated はイベントを記録します
func (m *MockStrategyObserver) OnUpdated(event string) {
	m.Events = append(m.Events, event)
}

func TestNewCounterStrategy(t *testing.T) {
	// 新しいCounterStrategyを作成
	strategy := NewCounterStrategy()

	// 初期状態の検証
	assert.Equal(t, int64(0), strategy.currentValue)
	assert.Empty(t, strategy.observers)
}

func TestCounterStrategyInitialize(t *testing.T) {
	// 新しいCounterStrategyを作成
	strategy := NewCounterStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")

	// 初期化
	err := strategy.Initialize(part)
	assert.NoError(t, err)

	// オブザーバーが追加されていることを確認
	assert.Equal(t, 1, len(strategy.observers))

	// 無効なパラメータでの初期化
	err = strategy.Initialize("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid part type")
}

func TestCounterStrategyGetCurrentValue(t *testing.T) {
	// 新しいCounterStrategyを作成
	strategy := NewCounterStrategy()
	strategy.currentValue = 42

	// 現在値の取得
	value := strategy.GetCurrentValue()
	assert.Equal(t, int64(42), value)
}

func TestCounterStrategyStart(t *testing.T) {
	// 新しいCounterStrategyを作成
	strategy := NewCounterStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")

	// 開始
	ctx := context.Background()
	err := strategy.Start(ctx, part)
	assert.NoError(t, err)
}

func TestCounterStrategyEvaluate(t *testing.T) {
	// テストケース
	testCases := []struct {
		name              string
		initialValue      int64
		increment         int64
		operator          value.ComparisonOperator
		referenceValue    int64
		minValue          int64
		maxValue          int64
		expectedSatisfied bool
		expectedEvent     string
	}{
		{
			name:              "EQ_Satisfied",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorEQ,
			referenceValue:    5,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "EQ_Unsatisfied",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorEQ,
			referenceValue:    10,
			expectedSatisfied: false,
			expectedEvent:     value.EventProcess,
		},
		{
			name:              "NEQ_Satisfied",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorNEQ,
			referenceValue:    10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "NEQ_Unsatisfied",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorNEQ,
			referenceValue:    5,
			expectedSatisfied: false,
			expectedEvent:     value.EventProcess,
		},
		{
			name:              "GT_Satisfied",
			initialValue:      0,
			increment:         15,
			operator:          value.ComparisonOperatorGT,
			referenceValue:    10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "GT_Unsatisfied",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorGT,
			referenceValue:    10,
			expectedSatisfied: false,
			expectedEvent:     value.EventProcess,
		},
		{
			name:              "GTE_Satisfied_Equal",
			initialValue:      0,
			increment:         10,
			operator:          value.ComparisonOperatorGTE,
			referenceValue:    10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "GTE_Satisfied_Greater",
			initialValue:      0,
			increment:         15,
			operator:          value.ComparisonOperatorGTE,
			referenceValue:    10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "GTE_Unsatisfied",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorGTE,
			referenceValue:    10,
			expectedSatisfied: false,
			expectedEvent:     value.EventProcess,
		},
		{
			name:              "LT_Satisfied",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorLT,
			referenceValue:    10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "LT_Unsatisfied",
			initialValue:      0,
			increment:         15,
			operator:          value.ComparisonOperatorLT,
			referenceValue:    10,
			expectedSatisfied: false,
			expectedEvent:     value.EventProcess,
		},
		{
			name:              "LTE_Satisfied_Equal",
			initialValue:      0,
			increment:         10,
			operator:          value.ComparisonOperatorLTE,
			referenceValue:    10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "LTE_Satisfied_Less",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorLTE,
			referenceValue:    10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "LTE_Unsatisfied",
			initialValue:      0,
			increment:         15,
			operator:          value.ComparisonOperatorLTE,
			referenceValue:    10,
			expectedSatisfied: false,
			expectedEvent:     value.EventProcess,
		},
		{
			name:              "Between_Satisfied_Min",
			initialValue:      0,
			increment:         5,
			operator:          value.ComparisonOperatorBetween,
			minValue:          5,
			maxValue:          10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "Between_Satisfied_Max",
			initialValue:      0,
			increment:         10,
			operator:          value.ComparisonOperatorBetween,
			minValue:          5,
			maxValue:          10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "Between_Satisfied_Middle",
			initialValue:      0,
			increment:         7,
			operator:          value.ComparisonOperatorBetween,
			minValue:          5,
			maxValue:          10,
			expectedSatisfied: true,
			expectedEvent:     value.EventComplete,
		},
		{
			name:              "Between_Unsatisfied_Less",
			initialValue:      0,
			increment:         3,
			operator:          value.ComparisonOperatorBetween,
			minValue:          5,
			maxValue:          10,
			expectedSatisfied: false,
			expectedEvent:     value.EventProcess,
		},
		{
			name:              "Between_Unsatisfied_Greater",
			initialValue:      0,
			increment:         15,
			operator:          value.ComparisonOperatorBetween,
			minValue:          5,
			maxValue:          10,
			expectedSatisfied: false,
			expectedEvent:     value.EventProcess,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 新しいCounterStrategyを作成
			strategy := NewCounterStrategy()
			strategy.currentValue = tc.initialValue

			// モックオブザーバーの作成
			mockObserver := &MockStrategyObserver{}
			strategy.AddObserver(mockObserver)

			// テスト用のConditionPartを作成
			part := entity.NewConditionPart(1, "Test Part")
			part.ComparisonOperator = tc.operator
			part.ReferenceValueInt = tc.referenceValue
			part.MinValue = tc.minValue
			part.MaxValue = tc.maxValue

			// 評価
			ctx := context.Background()
			err := strategy.Evaluate(ctx, part, tc.increment)
			assert.NoError(t, err)

			// 現在値の確認
			assert.Equal(t, tc.initialValue+tc.increment, strategy.currentValue)

			// イベントの確認
			assert.Len(t, mockObserver.Events, 1)
			assert.Equal(t, tc.expectedEvent, mockObserver.Events[0])
		})
	}

	// 無効なパラメータでの評価
	strategy := NewCounterStrategy()
	err := strategy.Evaluate(context.Background(), "invalid", int64(5))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid part type")

	// nilパラメータでの評価
	err = strategy.Evaluate(context.Background(), entity.NewConditionPart(1, "Test Part"), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nil params")

	// 未サポートの比較演算子
	part := entity.NewConditionPart(1, "Test Part")
	part.ComparisonOperator = value.ComparisonOperatorIn // 未サポート
	err = strategy.Evaluate(context.Background(), part, int64(5))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported comparison operator")
}

func TestCounterStrategyCleanup(t *testing.T) {
	// 新しいCounterStrategyを作成
	strategy := NewCounterStrategy()
	strategy.currentValue = 42
	strategy.observers = append(strategy.observers, &MockStrategyObserver{})

	// クリーンアップ
	err := strategy.Cleanup()
	assert.NoError(t, err)

	// 状態の確認
	assert.Equal(t, int64(0), strategy.currentValue)
	assert.NotNil(t, strategy.observers, "observers should not be nil")
	assert.Empty(t, strategy.observers, "observers should be empty")
}

func TestCounterStrategyResetAndRetry(t *testing.T) {
	// 新しいCounterStrategyを作成
	strategy := NewCounterStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")
	part.ComparisonOperator = value.ComparisonOperatorGTE
	part.ReferenceValueInt = 5

	// 初期化
	err := strategy.Initialize(part)
	assert.NoError(t, err)

	// モックオブザーバーが追加されていることを確認
	assert.Len(t, strategy.observers, 1)

	// 評価（まだ条件を満たさない）
	ctx := context.Background()
	mockObserver := &MockStrategyObserver{}
	strategy.observers[0] = mockObserver // モックに置き換え

	err = strategy.Evaluate(ctx, part, int64(3))
	assert.NoError(t, err)
	assert.Equal(t, int64(3), strategy.currentValue)
	assert.Len(t, mockObserver.Events, 1)
	assert.Equal(t, value.EventProcess, mockObserver.Events[0])

	// クリーンアップ（リセット）
	err = strategy.Cleanup()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), strategy.currentValue)
	assert.Empty(t, strategy.observers)

	// 再初期化
	err = strategy.Initialize(part)
	assert.NoError(t, err)
	assert.Len(t, strategy.observers, 1)

	// 再評価（今度は条件を満たす）
	mockObserver = &MockStrategyObserver{}
	strategy.observers[0] = mockObserver // モックに置き換え

	err = strategy.Evaluate(ctx, part, int64(5))
	assert.NoError(t, err)
	assert.Equal(t, int64(5), strategy.currentValue)
	assert.Len(t, mockObserver.Events, 1)
	assert.Equal(t, value.EventComplete, mockObserver.Events[0])
}

func TestCounterStrategyObserver(t *testing.T) {
	// 新しいCounterStrategyを作成
	strategy := NewCounterStrategy()

	// モックオブザーバーの作成
	mockObserver1 := &MockStrategyObserver{}
	mockObserver2 := &MockStrategyObserver{}

	// オブザーバーの追加
	strategy.AddObserver(mockObserver1)
	strategy.AddObserver(mockObserver2)
	assert.Len(t, strategy.observers, 2)

	// 通知
	strategy.NotifyUpdate("test_event")
	assert.Len(t, mockObserver1.Events, 1)
	assert.Equal(t, "test_event", mockObserver1.Events[0])
	assert.Len(t, mockObserver2.Events, 1)
	assert.Equal(t, "test_event", mockObserver2.Events[0])

	// オブザーバーの削除
	strategy.RemoveObserver(mockObserver1)
	assert.Len(t, strategy.observers, 1)

	// 通知（削除したオブザーバーには通知されない）
	mockObserver1.Events = nil
	mockObserver2.Events = nil
	strategy.NotifyUpdate("another_event")
	assert.Len(t, mockObserver1.Events, 0)
	assert.Len(t, mockObserver2.Events, 1)
	assert.Equal(t, "another_event", mockObserver2.Events[0])
}
