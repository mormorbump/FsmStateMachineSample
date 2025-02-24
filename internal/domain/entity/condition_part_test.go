package entity

import (
	"context"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/value"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConditionPart_NewConditionPart(t *testing.T) {
	// Arrange
	id := core.ConditionPartID(1)
	label := "test_part"

	// Act
	part := NewConditionPart(id, label)

	// Assert
	assert.NotNil(t, part)
	assert.Equal(t, id, part.ID)
	assert.Equal(t, label, part.Label)
	assert.Equal(t, value.StateReady, part.CurrentState())
}

func TestConditionPart_Validate(t *testing.T) {
	tests := []struct {
		name    string
		part    *ConditionPart
		wantErr bool
	}{
		{
			name: "valid part",
			part: &ConditionPart{
				ComparisonOperator: ComparisonOperatorEQ,
			},
			wantErr: false,
		},
		{
			name: "unspecified operator",
			part: &ConditionPart{
				ComparisonOperator: ComparisonOperatorUnspecified,
			},
			wantErr: true,
		},
		{
			name: "invalid between values",
			part: &ConditionPart{
				ComparisonOperator: ComparisonOperatorBetween,
				MinValue:           10,
				MaxValue:           5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.part.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConditionPart_IsClear(t *testing.T) {
	// Arrange
	part := NewConditionPart(1, "test_part")
	ctx := context.Background()

	// Assert initial state
	assert.False(t, part.IsClear, "IsClear should be false initially")

	// Act: transition to satisfied state
	_ = part.Activate(ctx)
	_ = part.StartProcess(ctx)
	_ = part.Timeout(ctx) // タイムアウトで直接Satisfiedに遷移

	// Assert: IsClear should be true after satisfied
	assert.True(t, part.IsClear, "IsClear should be true after satisfied")
}

func TestConditionPart_StateTransitions(t *testing.T) {
	// Arrange
	part := NewConditionPart(1, "test_part")
	ctx := context.Background()

	// Act & Assert: Ready → Unsatisfied
	assert.Equal(t, value.StateReady, part.CurrentState())
	err := part.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateUnsatisfied, part.CurrentState())

	// Unsatisfied → Processing
	err = part.StartProcess(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateProcessing, part.CurrentState())

	// Processing → Unsatisfied (Revert)
	err = part.Revert(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateUnsatisfied, part.CurrentState())

	// Unsatisfied → Satisfied (Timeout)
	err = part.Timeout(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateSatisfied, part.CurrentState())

	// Satisfied → Ready (Reset)
	err = part.Reset(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateReady, part.CurrentState())
}

func TestConditionPart_ObserverNotification(t *testing.T) {
	// Arrange
	part := NewConditionPart(1, "test_part")
	mockObserver := &mockConditionPartObserver{
		satisfiedParts: make(map[core.ConditionPartID]bool),
		stateChanges:   make([]string, 0),
	}
	part.AddObserver(mockObserver)
	ctx := context.Background()

	// Act: 状態遷移シーケンス
	_ = part.Activate(ctx)
	_ = part.StartProcess(ctx)
	_ = part.Timeout(ctx) // タイムアウトで直接Satisfiedに遷移

	// Assert
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.True(t, part.IsClear, "Part should be marked as clear")
	assert.Equal(t, value.StateSatisfied, part.CurrentState(), "Part should be in satisfied state")
}

type mockConditionPartObserver struct {
	satisfiedParts map[core.ConditionPartID]bool
	stateChanges   []string
}

func (m *mockConditionPartObserver) OnPartSatisfied(partID core.ConditionPartID) {
	m.satisfiedParts[partID] = true
}

func (m *mockConditionPartObserver) OnStateChanged(state string) {
	m.stateChanges = append(m.stateChanges, state)
}
