package entity

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
	assert.Equal(t, part.ID, cond.Parts[0].ID)
}

func TestCondition_IsClear(t *testing.T) {
	// Arrange
	cond := NewCondition(1, "test_condition", core.KindTime)
	part := NewConditionPart(1, "test_part")
	cond.AddPart(part)
	ctx := context.Background()

	// Assert initial state
	assert.False(t, cond.IsClear(), "isClear should be false initially")

	// Act: transition to satisfied state
	_ = cond.Activate(ctx)
	_ = cond.StartProcess(ctx)
	cond.OnPartSatisfied(part.ID)

	// Assert: isClear should be true after satisfied
	assert.True(t, cond.IsClear(), "isClear should be true after satisfied")
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

	err = cond.StartProcess(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateProcessing, cond.CurrentState())

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
	}
	cond.AddConditionObserver(mockObserver)

	// Act
	_ = cond.Activate(ctx)
	_ = cond.StartProcess(ctx)
	cond.OnPartSatisfied(part.ID)

	// Assert
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.Equal(t, value.StateSatisfied, cond.CurrentState())
	assert.Contains(t, mockObserver.satisfiedConditions, cond.ID)
}

type mockConditionObserver struct {
	satisfiedConditions []core.ConditionID
}

func (m *mockConditionObserver) OnConditionSatisfied(conditionID core.ConditionID) {
	m.satisfiedConditions = append(m.satisfiedConditions, conditionID)
}
