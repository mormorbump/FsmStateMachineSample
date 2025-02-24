package state

import (
	"context"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/value"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCondition_NewCondition(t *testing.T) {
	// Arrange
	id := core.ConditionID(1)
	label := "test_condition"
	kind := core.KindTime

	// Act
	cond := NewCondition(id, label, kind)

	// Assert
	assert.NotNil(t, cond)
	assert.Equal(t, id, cond.ID)
	assert.Equal(t, label, cond.Label)
	assert.Equal(t, kind, cond.Kind)
	assert.Equal(t, value.StateReady, cond.CurrentState())
}

func TestCondition_AddPart(t *testing.T) {
	// Arrange
	cond := NewCondition(1, "test_condition", core.KindTime)
	part := NewConditionPart(1, "test_part")

	// Act
	cond.AddPart(part)

	// Assert
	assert.Len(t, cond.Parts, 1)
	assert.Equal(t, part.ID, cond.Parts[part.ID].ID)
}

func TestCondition_IsClear(t *testing.T) {
	// Arrange
	cond := NewCondition(1, "test_condition", core.KindTime)
	part := NewConditionPart(1, "test_part")
	cond.AddPart(part)
	ctx := context.Background()

	// Assert initial state
	assert.False(t, cond.IsClear, "IsClear should be false initially")

	// Act: transition to satisfied state
	_ = cond.Activate(ctx)
	cond.OnPartSatisfied(part.ID)

	// Assert: IsClear should be true after satisfied
	assert.True(t, cond.IsClear, "IsClear should be true after satisfied")
}

func TestCondition_StateTransitions(t *testing.T) {
	// Arrange
	cond := NewCondition(1, "test_condition", core.KindTime)
	part := NewConditionPart(1, "test_part")
	cond.AddPart(part)
	ctx := context.Background()

	// Act & Assert
	assert.Equal(t, value.StateReady, cond.CurrentState())

	err := cond.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateUnsatisfied, cond.CurrentState())

	err = cond.Complete(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateSatisfied, cond.CurrentState())
}

func TestCondition_PartSatisfaction(t *testing.T) {
	// Arrange
	cond := NewCondition(1, "test_condition", core.KindTime)
	part := NewConditionPart(1, "test_part")
	cond.AddPart(part)
	ctx := context.Background()
	mockObserver := &mockConditionObserver{
		satisfiedConditions: make([]core.ConditionID, 0),
		stateChanges:        make([]string, 0),
	}
	cond.AddConditionObserver(mockObserver)

	// Act
	_ = cond.Activate(ctx)
	cond.OnPartSatisfied(part.ID)

	// Assert
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.Equal(t, value.StateSatisfied, cond.CurrentState(), "Condition should be in satisfied state")
	assert.True(t, cond.IsClear, "Condition should be marked as clear")
}

type mockConditionObserver struct {
	satisfiedConditions []core.ConditionID
	stateChanges        []string
}

func (m *mockConditionObserver) OnConditionSatisfied(conditionID core.ConditionID) {
	m.satisfiedConditions = append(m.satisfiedConditions, conditionID)
}

func (m *mockConditionObserver) OnStateChanged(state string) {
	m.stateChanges = append(m.stateChanges, state)
}

func TestCondition_TimeManagement(t *testing.T) {
	// Arrange
	cond := NewCondition(1, "test_condition", core.KindTime)
	part := NewConditionPart(1, "test_part")
	cond.AddPart(part)
	ctx := context.Background()

	// 初期状態の確認
	assert.Nil(t, cond.StartTime, "初期状態ではStartTimeはnilのはず")
	assert.Nil(t, cond.FinishTime, "初期状態ではFinishTimeはnilのはず")

	// StateUnsatisfied遷移時のStartTime設定を確認
	err := cond.Activate(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, cond.StartTime, "StateUnsatisfied遷移後はStartTimeが設定されているはず")
	assert.Nil(t, cond.FinishTime, "StateUnsatisfied遷移後もFinishTimeはnilのはず")
	activateTime := *cond.StartTime

	// StateSatisfied遷移時のFinishTime設定を確認
	cond.OnPartSatisfied(part.ID)
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.NotNil(t, cond.FinishTime, "StateSatisfied遷移後はFinishTimeが設定されているはず")
	assert.Equal(t, activateTime, *cond.StartTime, "StateSatisfied遷移後もStartTimeは変更されないはず")

	// Reset時の時間情報初期化を確認
	err = cond.Reset(ctx)
	assert.NoError(t, err)
	assert.Nil(t, cond.StartTime, "Reset後はStartTimeがnilになるはず")
	assert.Nil(t, cond.FinishTime, "Reset後はFinishTimeがnilになるはず")
}
