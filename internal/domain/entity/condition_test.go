package entity

import (
	"context"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockStateObserver は PhaseObserver インターフェースのモック実装です
type MockStateObserver struct {
	States []string
}

// OnPhaseChanged は状態変更を記録します
func (m *MockStateObserver) OnPhaseChanged(state string) {
	m.States = append(m.States, state)
}

// MockConditionObserver は ConditionObserver インターフェースのモック実装です
type MockConditionObserver struct {
	Conditions []interface{}
}

// OnConditionChanged は条件変更を記録します
func (m *MockConditionObserver) OnConditionChanged(condition interface{}) {
	m.Conditions = append(m.Conditions, condition)
}

// MockStrategyFactory は StrategyFactory インターフェースのモック実装です
type MockStrategyFactory struct {
	CreatedStrategies []service.PartStrategy
}

// CreateStrategy は指定された種類の戦略を作成します
func (f *MockStrategyFactory) CreateStrategy(kind value.ConditionKind) (service.PartStrategy, error) {
	strategy := &MockPartStrategy{}
	f.CreatedStrategies = append(f.CreatedStrategies, strategy)
	return strategy, nil
}

func TestNewCondition(t *testing.T) {
	// テスト用のIDとラベル
	id := value.ConditionID(1)
	label := "Test Condition"
	kind := value.KindCounter

	// Conditionの作成
	condition := NewCondition(id, label, kind)

	// 初期状態の検証
	assert.Equal(t, id, condition.ID)
	assert.Equal(t, label, condition.Label)
	assert.Equal(t, kind, condition.Kind)
	assert.False(t, condition.IsClear)
	assert.Nil(t, condition.StartTime)
	assert.Nil(t, condition.FinishTime)
	assert.Equal(t, value.StateReady, condition.CurrentState())
	assert.Empty(t, condition.Parts)
}

func TestConditionStateTransitions(t *testing.T) {
	// テスト用のCondition
	condition := NewCondition(1, "Test Condition", value.KindCounter)
	ctx := context.Background()

	// Activate: Ready -> Unsatisfied
	err := condition.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateUnsatisfied, condition.CurrentState())
	assert.NotNil(t, condition.StartTime)

	// Complete: Unsatisfied -> Satisfied
	err = condition.Complete(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateSatisfied, condition.CurrentState())
	assert.True(t, condition.IsClear)
	assert.NotNil(t, condition.FinishTime)

	// Reset: Satisfied -> Ready
	err = condition.Reset(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateReady, condition.CurrentState())
	assert.False(t, condition.IsClear)
	assert.Nil(t, condition.StartTime)
	assert.Nil(t, condition.FinishTime)

	// Activate -> Complete -> Revert
	err = condition.Activate(ctx)
	assert.NoError(t, err)
	err = condition.Complete(ctx)
	assert.NoError(t, err)
	err = condition.Revert(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateUnsatisfied, condition.CurrentState())
}

func TestConditionWithParts(t *testing.T) {
	// テスト用のCondition
	condition := NewCondition(1, "Test Condition", value.KindCounter)
	ctx := context.Background()

	// 条件パーツの作成と追加
	part1 := NewConditionPart(1, "Part 1")
	part2 := NewConditionPart(2, "Part 2")
	condition.AddPart(part1)
	condition.AddPart(part2)

	// パーツの取得
	parts := condition.GetParts()
	assert.Len(t, parts, 2)
	assert.Contains(t, parts, part1)
	assert.Contains(t, parts, part2)

	// 戦略の初期化
	factory := &MockStrategyFactory{}
	err := condition.InitializePartStrategies(factory)
	assert.NoError(t, err)
	assert.Len(t, factory.CreatedStrategies, 2)

	// Activate
	err = condition.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateUnsatisfied, condition.CurrentState())
	assert.Equal(t, value.StateUnsatisfied, part1.CurrentState())
	assert.Equal(t, value.StateUnsatisfied, part2.CurrentState())

	// 一つのパーツが満たされた場合
	part1.OnUpdated(value.EventComplete)
	assert.Equal(t, value.StateSatisfied, part1.CurrentState())
	assert.Equal(t, value.StateUnsatisfied, condition.CurrentState()) // まだ全てのパーツが満たされていない

	// 全てのパーツが満たされた場合
	part2.OnUpdated(value.EventComplete)
	assert.Equal(t, value.StateSatisfied, part2.CurrentState())
	assert.Equal(t, value.StateSatisfied, condition.CurrentState()) // 全てのパーツが満たされた
}

func TestConditionObserver(t *testing.T) {
	// テスト用のCondition
	condition := NewCondition(1, "Test Condition", value.KindCounter)

	// モックオブザーバーの作成
	mockStateObserver := &MockStateObserver{}
	mockConditionObserver := &MockConditionObserver{}

	// オブザーバーの追加
	condition.AddObserver(mockStateObserver)
	condition.AddConditionObserver(mockConditionObserver)

	// 条件変更の通知
	condition.NotifyConditionChanged()
	assert.Len(t, mockConditionObserver.Conditions, 1)
	assert.Equal(t, condition, mockConditionObserver.Conditions[0])

	// オブザーバーの削除
	condition.RemoveObserver(mockStateObserver)
	condition.RemoveConditionObserver(mockConditionObserver)

	// 状態変更の通知（オブザーバーが削除されているので通知されない）
	mockStateObserver.States = nil
	mockConditionObserver.Conditions = nil
	condition.NotifyConditionChanged()
	assert.Len(t, mockStateObserver.States, 0)
	assert.Len(t, mockConditionObserver.Conditions, 0)
}

func TestConditionPartChanged(t *testing.T) {
	// テスト用のCondition
	condition := NewCondition(1, "Test Condition", value.KindCounter)
	ctx := context.Background()

	// 条件パーツの作成と追加
	part := NewConditionPart(1, "Part 1")
	condition.AddPart(part)

	// Activate
	err := condition.Activate(ctx)
	assert.NoError(t, err)

	// パーツの状態変更
	condition.OnConditionPartChanged(part)
	assert.Equal(t, value.StateUnsatisfied, condition.CurrentState()) // まだ満たされていない

	// パーツが満たされた状態に設定
	part.IsClear = true
	// 満たされたパーツとして記録
	condition.satisfiedParts[part.ID] = true
	condition.OnConditionPartChanged(part)
	assert.Equal(t, value.StateSatisfied, condition.CurrentState()) // 全てのパーツが満たされた
}

func TestConditionValidation(t *testing.T) {
	// テスト用のCondition
	condition := NewCondition(1, "Test Condition", value.KindCounter)

	// パーツがない場合
	err := condition.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition must have at least one part")

	// パーツを追加
	part := NewConditionPart(1, "Part 1")
	condition.AddPart(part)

	// パーツが無効な場合
	err = condition.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid condition part")

	// パーツを有効にする
	part.ComparisonOperator = value.ComparisonOperatorEQ
	err = condition.Validate()
	assert.NoError(t, err)
}
