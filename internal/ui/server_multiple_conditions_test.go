package ui

import (
	"context"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/value"
	"state_sample/internal/usecase/state"
	"state_sample/internal/usecase/strategy"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestServerWithMultipleConditions は複数の条件がUIに正しく表示されることをテストします
func TestServerWithMultipleConditions(t *testing.T) {
	// テスト用のストラテジーファクトリを作成
	factory := strategy.NewStrategyFactory()

	// 複数の条件を持つPhaseを作成
	// Phase1: AND条件（2つの条件）
	part1_1 := entity.NewConditionPart(1, "Counter_Part_1")
	part1_1.ReferenceValueInt = 1
	part1_1.ComparisonOperator = value.ComparisonOperatorGTE
	cond1_1 := entity.NewCondition(1, "Counter_Condition_1", value.KindCounter)
	cond1_1.AddPart(part1_1)

	part1_2 := entity.NewConditionPart(2, "Time_Part_1")
	part1_2.ReferenceValueInt = 1
	cond1_2 := entity.NewCondition(2, "Time_Condition_1", value.KindTime)
	cond1_2.AddPart(part1_2)

	if err := cond1_1.InitializePartStrategies(factory); err != nil {
		t.Fatalf("Failed to initialize strategies for cond1_1: %v", err)
	}
	if err := cond1_2.InitializePartStrategies(factory); err != nil {
		t.Fatalf("Failed to initialize strategies for cond1_2: %v", err)
	}

	phase1 := entity.NewPhase("PHASE1", 1, []*entity.Condition{cond1_1, cond1_2}, value.ConditionTypeAnd, value.GameRule_Animation)
	part1_1.AddConditionPartObserver(cond1_1)
	part1_2.AddConditionPartObserver(cond1_2)
	cond1_1.AddConditionObserver(phase1)
	cond1_2.AddConditionObserver(phase1)

	// StateFacadeとStateServerを作成
	phases := entity.Phases{phase1}
	controller := state.NewPhaseController(phases)
	facade := &mockStateFacade{
		controller:   controller,
		currentPhase: phase1,
	}
	server := NewStateServer(facade)

	// EditResponseメソッドを呼び出して、レスポンスを取得
	stateInfo := server.getGameStateInfo(phase1)
	response := server.EditResponse(value.StateActive, phase1, stateInfo)

	// レスポンスに複数の条件が含まれていることを確認
	assert.Len(t, response.Conditions, 2, "Response should contain 2 conditions")

	// 条件のIDとラベルを確認
	conditionMap := make(map[value.ConditionID]ConditionInfo)
	for _, cond := range response.Conditions {
		conditionMap[cond.ID] = cond
	}

	// 条件1の確認
	cond1, exists := conditionMap[value.ConditionID(1)]
	assert.True(t, exists, "Condition with ID 1 should exist")
	assert.Equal(t, "Counter_Condition_1", cond1.Label)
	assert.Equal(t, value.KindCounter, cond1.Kind)
	assert.Len(t, cond1.Parts, 1, "Condition 1 should have 1 part")

	// 条件2の確認
	cond2, exists := conditionMap[value.ConditionID(2)]
	assert.True(t, exists, "Condition with ID 2 should exist")
	assert.Equal(t, "Time_Condition_1", cond2.Label)
	assert.Equal(t, value.KindTime, cond2.Kind)
	assert.Len(t, cond2.Parts, 1, "Condition 2 should have 1 part")

	// 条件パーツの確認
	assert.Equal(t, "Counter_Part_1", cond1.Parts[0].Label)
	assert.Equal(t, value.ComparisonOperatorGTE, cond1.Parts[0].ComparisonOperator)
	assert.Equal(t, int64(1), cond1.Parts[0].ReferenceValueInt)

	assert.Equal(t, "Time_Part_1", cond2.Parts[0].Label)
	assert.Equal(t, int64(1), cond2.Parts[0].ReferenceValueInt)
}

// mockStateFacade はテスト用のStateFacadeモック
type mockStateFacade struct {
	controller   *state.PhaseController
	currentPhase *entity.Phase
}

func (m *mockStateFacade) Start(ctx context.Context) error {
	return nil
}

func (m *mockStateFacade) Reset(ctx context.Context) error {
	return nil
}

func (m *mockStateFacade) GetCurrentPhase() *entity.Phase {
	return m.currentPhase
}

func (m *mockStateFacade) GetController() *state.PhaseController {
	return m.controller
}

func (m *mockStateFacade) GetConditionPart(conditionID, partID int64) (*entity.ConditionPart, error) {
	for _, cond := range m.currentPhase.GetConditions() {
		if cond.ID == value.ConditionID(conditionID) {
			for _, part := range cond.GetParts() {
				if part.ID == value.ConditionPartID(partID) {
					return part, nil
				}
			}
		}
	}
	return nil, nil
}
