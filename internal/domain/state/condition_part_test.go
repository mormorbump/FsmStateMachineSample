package state

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
				ComparisonOperator: core.ComparisonOperatorEQ,
			},
			wantErr: false,
		},
		{
			name: "unspecified operator",
			part: &ConditionPart{
				ComparisonOperator: core.ComparisonOperatorUnspecified,
			},
			wantErr: true,
		},
		{
			name: "invalid between values",
			part: &ConditionPart{
				ComparisonOperator: core.ComparisonOperatorBetween,
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
	_ = part.Process(ctx)
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
	err = part.Process(ctx)
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
	_ = part.Process(ctx)
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

func TestConditionPart_TimeManagement(t *testing.T) {
	// Arrange
	part := NewConditionPart(1, "test_part")
	ctx := context.Background()

	// 初期状態の確認
	assert.Nil(t, part.StartTime, "初期状態ではStartTimeはnilのはず")
	assert.Nil(t, part.FinishTime, "初期状態ではFinishTimeはnilのはず")

	// StateUnsatisfied遷移時のStartTime設定を確認
	err := part.Activate(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, part.StartTime, "StateUnsatisfied遷移後はStartTimeが設定されているはず")
	assert.Nil(t, part.FinishTime, "StateUnsatisfied遷移後もFinishTimeはnilのはず")
	activateTime := *part.StartTime

	// Processing状態への遷移
	err = part.Process(ctx)
	assert.NoError(t, err)
	assert.Equal(t, activateTime, *part.StartTime, "Processing遷移後もStartTimeは変更されないはず")
	assert.Nil(t, part.FinishTime, "Processing遷移後もFinishTimeはnilのはず")

	// Complete経由でのStateSatisfied遷移時のFinishTime設定を確認
	err = part.Complete(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, part.FinishTime, "Complete後はFinishTimeが設定されているはず")
	assert.Equal(t, activateTime, *part.StartTime, "Complete後もStartTimeは変更されないはず")

	// Reset時の時間情報初期化を確認
	err = part.Reset(ctx)
	assert.NoError(t, err)
	assert.Nil(t, part.StartTime, "Reset後はStartTimeがnilになるはず")
	assert.Nil(t, part.FinishTime, "Reset後はFinishTimeがnilになるはず")

	// Timeout経由でのStateSatisfied遷移時のFinishTime設定を確認
	err = part.Activate(ctx)
	assert.NoError(t, err)
	activateTime = *part.StartTime

	err = part.Timeout(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, part.FinishTime, "Timeout後はFinishTimeが設定されているはず")
	assert.Equal(t, activateTime, *part.StartTime, "Timeout後もStartTimeは変更されないはず")
}
