package entity

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createCondition(order int, interval int64) *Condition {
	cond := NewCondition(
		core.ConditionID(order),
		fmt.Sprintf("test_phase_%d_timer", order),
		core.KindTime,
	)

	part := NewConditionPart(
		core.ConditionPartID(order),
		fmt.Sprintf("test_phase_%d_timer_part", order),
	)
	part.ReferenceValueInt = interval
	cond.AddPart(part)

	return cond
}

func TestPhase_NewPhase(t *testing.T) {
	// Arrange
	phaseType := "test_phase"
	order := 1
	cond := createCondition(order, 1)

	// Act
	phase := NewPhase(phaseType, order, cond)

	// Assert
	assert.NotNil(t, phase)
	assert.Equal(t, phaseType, phase.Type)
	assert.Equal(t, order, phase.Order)
	assert.Equal(t, core.StateReady, phase.CurrentState())
	assert.Equal(t, ConditionTypeSingle, phase.ConditionType)
	assert.Len(t, phase.ConditionIDs, 1)
	assert.Len(t, phase.conditions, 1)
}

func TestPhase_IsClear(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, createCondition(1, 1))
	ctx := context.Background()

	// Assert initial state
	assert.False(t, phase.IsClear(), "IsClear should be false initially")

	// Act: activate phase and satisfy condition
	_ = phase.Activate(ctx)
	phase.OnConditionSatisfied(core.ConditionID(1))

	// Assert: IsClear should be true after condition is satisfied
	assert.True(t, phase.IsClear(), "IsClear should be true after condition is satisfied")
}

func TestPhase_StateTransitions(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, createCondition(1, 1))
	ctx := context.Background()

	// Act & Assert
	assert.Equal(t, core.StateReady, phase.CurrentState())

	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateActive, phase.CurrentState())

	err = phase.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateNext, phase.CurrentState())

	err = phase.Finish(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateFinish, phase.CurrentState())
}

func TestPhase_ConditionSatisfaction(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, createCondition(1, 1))
	ctx := context.Background()
	mockObserver := &mockPhaseObserver{
		stateChanges: make([]string, 0),
	}
	phase.AddObserver(mockObserver)

	// Act
	_ = phase.Activate(ctx)
	phase.OnConditionSatisfied(core.ConditionID(1))

	// Assert
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.Equal(t, core.StateNext, phase.CurrentState())
	assert.Contains(t, mockObserver.stateChanges, core.StateNext)
}

func TestPhases_ProcessAndActivateByNextOrder(t *testing.T) {
	// Arrange
	phases := Phases{
		NewPhase("phase1", 1, createCondition(1, 1)),
		NewPhase("phase2", 2, createCondition(2, 2)),
		NewPhase("phase3", 3, createCondition(3, 3)),
	}
	ctx := context.Background()

	// Act & Assert
	// 最初のフェーズを開始
	nextPhase, err := phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "phase1", nextPhase.Type)
	assert.Equal(t, core.StateActive, nextPhase.CurrentState())

	// 次のフェーズに移行
	nextPhase.OnConditionSatisfied(core.ConditionID(1))
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機

	nextPhase, err = phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "phase2", nextPhase.Type)
	assert.Equal(t, core.StateActive, nextPhase.CurrentState())
}

func TestPhases_ResetAll(t *testing.T) {
	// Arrange
	phases := Phases{
		NewPhase("phase1", 1, createCondition(1, 1)),
		NewPhase("phase2", 2, createCondition(2, 2)),
	}
	ctx := context.Background()

	// 最初のフェーズを開始
	_, _ = phases.ProcessAndActivateByNextOrder(ctx)

	// Act
	err := phases.ResetAll(ctx)

	// Assert
	assert.NoError(t, err)
	for _, phase := range phases {
		assert.Equal(t, core.StateReady, phase.CurrentState())
	}
}

type mockPhaseObserver struct {
	stateChanges []string
}

func (m *mockPhaseObserver) OnStateChanged(state string) {
	m.stateChanges = append(m.stateChanges, state)
}
