package entity

import (
	"context"
	"state_sample/internal/domain/service"
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

// MockPartStrategy は PartStrategy インターフェースのモック実装です
type MockPartStrategy struct {
	InitializeCalled bool
	StartCalled      bool
	EvaluateCalled   bool
	CleanupCalled    bool
	CurrentValue     interface{}
	observers        []service.StrategyObserver
}

// Initialize はモックの初期化関数を呼び出します
func (m *MockPartStrategy) Initialize(part interface{}) error {
	m.InitializeCalled = true
	return nil
}

// GetCurrentValue は現在の値を返します
func (m *MockPartStrategy) GetCurrentValue() interface{} {
	return m.CurrentValue
}

// Start はモックの開始関数を呼び出します
func (m *MockPartStrategy) Start(ctx context.Context, part interface{}) error {
	m.StartCalled = true
	return nil
}

// Evaluate はモックの評価関数を呼び出します
func (m *MockPartStrategy) Evaluate(ctx context.Context, part interface{}, params interface{}) error {
	m.EvaluateCalled = true
	// 評価後に条件が満たされたと通知
	m.NotifyUpdate(value.EventComplete)
	return nil
}

// Cleanup はモックのクリーンアップ関数を呼び出します
func (m *MockPartStrategy) Cleanup() error {
	m.CleanupCalled = true
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

// MockConditionPartObserver は ConditionPartObserver インターフェースのモック実装です
type MockConditionPartObserver struct {
	ChangedParts []interface{}
}

// OnConditionPartChanged はパーツの変更を記録します
func (m *MockConditionPartObserver) OnConditionPartChanged(part interface{}) {
	m.ChangedParts = append(m.ChangedParts, part)
}

func TestNewConditionPart(t *testing.T) {
	// テスト用のIDとラベル
	id := value.ConditionPartID(1)
	label := "Test Part"

	// ConditionPartの作成
	part := NewConditionPart(id, label)

	// 初期状態の検証
	assert.Equal(t, id, part.ID)
	assert.Equal(t, label, part.Label)
	assert.False(t, part.IsClear)
	assert.Nil(t, part.StartTime)
	assert.Nil(t, part.FinishTime)
	assert.Equal(t, value.StateReady, part.CurrentState())
}

func TestConditionPartStateTransitions(t *testing.T) {
	// テスト用のConditionPart
	part := NewConditionPart(1, "Test Part")
	ctx := context.Background()

	// Activate: Ready -> Unsatisfied
	err := part.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateUnsatisfied, part.CurrentState())
	assert.NotNil(t, part.StartTime)

	// Process: Unsatisfied -> Processing
	err = part.Process(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, value.StateProcessing, part.CurrentState())

	// Complete: Processing -> Satisfied
	err = part.Complete(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateSatisfied, part.CurrentState())
	assert.True(t, part.IsClear)
	assert.NotNil(t, part.FinishTime)

	// Reset: Satisfied -> Ready
	err = part.Reset(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateReady, part.CurrentState())
	assert.False(t, part.IsClear)
	assert.Nil(t, part.StartTime)
	assert.Nil(t, part.FinishTime)
}

func TestConditionPartWithStrategy(t *testing.T) {
	// テスト用のConditionPart
	part := NewConditionPart(1, "Test Part")
	ctx := context.Background()

	// モック戦略の作成
	mockStrategy := &MockPartStrategy{
		CurrentValue: int64(10),
	}

	// 戦略の設定
	err := part.SetStrategy(mockStrategy)
	assert.NoError(t, err)
	assert.True(t, mockStrategy.InitializeCalled)

	// Activate
	err = part.Activate(ctx)
	assert.NoError(t, err)
	assert.True(t, mockStrategy.StartCalled)

	// Process
	err = part.Process(ctx, 5)
	assert.NoError(t, err)
	assert.True(t, mockStrategy.EvaluateCalled)

	// GetCurrentValue
	assert.Equal(t, int64(10), part.GetCurrentValue())

	// Reset
	err = part.Reset(ctx)
	assert.NoError(t, err)
	assert.True(t, mockStrategy.CleanupCalled)
}

func TestConditionPartObserver(t *testing.T) {
	// テスト用のConditionPart
	part := NewConditionPart(1, "Test Part")

	// モックオブザーバーの作成
	mockObserver := &MockConditionPartObserver{}

	// オブザーバーの追加
	part.AddConditionPartObserver(mockObserver)

	// 状態変更の通知
	part.NotifyPartChanged()
	assert.Len(t, mockObserver.ChangedParts, 1)
	assert.Equal(t, part, mockObserver.ChangedParts[0])

	// オブザーバーの削除
	part.RemoveConditionPartObserver(mockObserver)

	// 状態変更の通知（オブザーバーが削除されているので通知されない）
	mockObserver.ChangedParts = nil
	part.NotifyPartChanged()
	assert.Len(t, mockObserver.ChangedParts, 0)
}

func TestConditionPartStrategyObserver(t *testing.T) {
	// テスト用のConditionPart
	part := NewConditionPart(1, "Test Part")
	ctx := context.Background()

	// Activate
	err := part.Activate(ctx)
	assert.NoError(t, err)

	// モック戦略の作成
	mockStrategy := &MockPartStrategy{}

	// 戦略の設定
	err = part.SetStrategy(mockStrategy)
	assert.NoError(t, err)

	// OnUpdated: EventComplete
	part.OnUpdated(value.EventComplete)
	assert.Equal(t, value.StateSatisfied, part.CurrentState())

	// Reset
	err = part.Reset(ctx)
	assert.NoError(t, err)

	// Activate
	err = part.Activate(ctx)
	assert.NoError(t, err)

	// OnUpdated: EventTimeout
	part.OnUpdated(value.EventTimeout)
	assert.Equal(t, value.StateSatisfied, part.CurrentState())
}

func TestConditionPartValidation(t *testing.T) {
	// テスト用のConditionPart
	part := NewConditionPart(1, "Test Part")

	// 比較演算子が未指定の場合
	err := part.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comparison operator must be specified")

	// 比較演算子を設定
	part.ComparisonOperator = value.ComparisonOperatorEQ
	err = part.Validate()
	assert.NoError(t, err)

	// Between演算子でMinValue >= MaxValueの場合
	part.ComparisonOperator = value.ComparisonOperatorBetween
	part.MinValue = 10
	part.MaxValue = 5
	err = part.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "min_value must be less than max_value")

	// Between演算子で正しい値の場合
	part.MinValue = 5
	part.MaxValue = 10
	err = part.Validate()
	assert.NoError(t, err)
}

func TestConditionPartIsSatisfied(t *testing.T) {
	// テスト用のConditionPart
	part := NewConditionPart(1, "Test Part")
	ctx := context.Background()

	// 初期状態
	assert.False(t, part.IsSatisfied())

	// Activate -> Process -> Complete
	_ = part.Activate(ctx)
	_ = part.Process(ctx, 1)
	_ = part.Complete(ctx)

	// 満たされた状態
	assert.True(t, part.IsSatisfied())
}
