package state

import (
	"context"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/value"
	"state_sample/internal/usecase/strategy"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPhaseControllerWithMultipleConditions は複数の条件を持つPhaseControllerの動作をテストします
func TestPhaseControllerWithMultipleConditions(t *testing.T) {
	// テスト用のストラテジーファクトリを作成
	factory := strategy.NewStrategyFactory()
	ctx := context.Background()

	// Phase1: AND条件（2つの条件が両方満たされる必要がある）
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

	// Phase2: OR条件（2つの条件のいずれかが満たされればよい）
	part2_1 := entity.NewConditionPart(3, "Counter_Part_2")
	part2_1.ReferenceValueInt = 1
	part2_1.ComparisonOperator = value.ComparisonOperatorGTE
	cond2_1 := entity.NewCondition(3, "Counter_Condition_2", value.KindCounter)
	cond2_1.AddPart(part2_1)

	part2_2 := entity.NewConditionPart(4, "Time_Part_2")
	part2_2.ReferenceValueInt = 3
	cond2_2 := entity.NewCondition(4, "Time_Condition_2", value.KindTime)
	cond2_2.AddPart(part2_2)

	if err := cond2_1.InitializePartStrategies(factory); err != nil {
		t.Fatalf("Failed to initialize strategies for cond2_1: %v", err)
	}
	if err := cond2_2.InitializePartStrategies(factory); err != nil {
		t.Fatalf("Failed to initialize strategies for cond2_2: %v", err)
	}

	phase2 := entity.NewPhase("PHASE2", 2, []*entity.Condition{cond2_1, cond2_2}, value.ConditionTypeOr, value.GameRule_Animation)
	part2_1.AddConditionPartObserver(cond2_1)
	part2_2.AddConditionPartObserver(cond2_2)
	cond2_1.AddConditionObserver(phase2)
	cond2_2.AddConditionObserver(phase2)

	// Phase3: Single条件（1つの条件のみ）
	part3 := entity.NewConditionPart(5, "Counter_Part_3")
	part3.ReferenceValueInt = 1
	part3.ComparisonOperator = value.ComparisonOperatorGTE
	cond3 := entity.NewCondition(5, "Counter_Condition_3", value.KindCounter)
	cond3.AddPart(part3)

	if err := cond3.InitializePartStrategies(factory); err != nil {
		t.Fatalf("Failed to initialize strategies for cond3: %v", err)
	}

	phase3 := entity.NewPhase("PHASE3", 3, []*entity.Condition{cond3}, value.ConditionTypeSingle, value.GameRule_Animation)
	part3.AddConditionPartObserver(cond3)
	cond3.AddConditionObserver(phase3)

	// PhaseControllerを作成
	phases := entity.Phases{phase1, phase2, phase3}
	controller := NewPhaseController(phases)

	// Phase1（AND条件）のテスト
	t.Run("Phase1_AND_Condition", func(t *testing.T) {
		// リセットして初期状態に
		err := controller.Reset(ctx)
		assert.NoError(t, err)

		// Phase1をアクティブにする
		err = controller.Start(ctx)
		assert.NoError(t, err)
		currentPhase := controller.GetCurrentPhase()
		assert.Equal(t, "PHASE1", currentPhase.Name)
		assert.Equal(t, value.StateActive, currentPhase.CurrentState())

		// 条件の数を確認
		conditions := currentPhase.GetConditions()
		assert.Len(t, conditions, 2, "Phase1 should have 2 conditions")

		// 条件のIDを確認
		var conditionIDs []value.ConditionID
		for id := range conditions {
			conditionIDs = append(conditionIDs, id)
		}
		assert.Contains(t, conditionIDs, value.ConditionID(1))
		assert.Contains(t, conditionIDs, value.ConditionID(2))

		// 1つ目の条件のみ満たす
		err = part1_1.Process(ctx, 1) // カウンターを1増やす
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // 状態更新を待つ

		// まだPhaseは完了していない（AND条件なので両方必要）
		currentPhase = controller.GetCurrentPhase()
		assert.Equal(t, value.StateActive, currentPhase.CurrentState())

		// 2つ目の条件も満たす（時間条件は自動的に満たされるはず）
		time.Sleep(1 * time.Second) // 時間条件を満たすために待機

		// Phaseが完了したことを確認（Next状態に遷移）
		currentPhase = controller.GetCurrentPhase()
		assert.Equal(t, value.StateNext, currentPhase.CurrentState())
	})

	// Phase2（OR条件）のテスト
	t.Run("Phase2_OR_Condition", func(t *testing.T) {
		// リセットして初期状態に
		err := controller.Reset(ctx)
		assert.NoError(t, err)

		// Phase1をスキップしてPhase2をアクティブにする
		err = phase1.Activate(ctx)
		assert.NoError(t, err)
		err = phase1.Next(ctx)
		assert.NoError(t, err)
		err = phase1.Finish(ctx)
		assert.NoError(t, err)

		err = controller.Start(ctx)
		assert.NoError(t, err)
		currentPhase := controller.GetCurrentPhase()
		assert.Equal(t, "PHASE2", currentPhase.Name)
		assert.Equal(t, value.StateActive, currentPhase.CurrentState())

		// 条件の数を確認
		conditions := currentPhase.GetConditions()
		assert.Len(t, conditions, 2, "Phase2 should have 2 conditions")

		// 条件のIDを確認
		var conditionIDs []value.ConditionID
		for id := range conditions {
			conditionIDs = append(conditionIDs, id)
		}
		assert.Contains(t, conditionIDs, value.ConditionID(3))
		assert.Contains(t, conditionIDs, value.ConditionID(4))

		// 1つ目の条件のみ満たす
		err = part2_1.Process(ctx, 1) // カウンターを1増やす
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // 状態更新を待つ

		// ORなので1つの条件だけでPhaseは完了する
		currentPhase = controller.GetCurrentPhase()
		assert.Equal(t, value.StateNext, currentPhase.CurrentState())
	})

	// Phase3（Single条件）のテスト
	t.Run("Phase3_Single_Condition", func(t *testing.T) {
		// リセットして初期状態に
		err := controller.Reset(ctx)
		assert.NoError(t, err)

		// Phase1とPhase2をスキップしてPhase3をアクティブにする
		err = phase1.Activate(ctx)
		assert.NoError(t, err)
		err = phase1.Next(ctx)
		assert.NoError(t, err)
		err = phase1.Finish(ctx)
		assert.NoError(t, err)

		err = phase2.Activate(ctx)
		assert.NoError(t, err)
		err = phase2.Next(ctx)
		assert.NoError(t, err)
		err = phase2.Finish(ctx)
		assert.NoError(t, err)

		err = controller.Start(ctx)
		assert.NoError(t, err)
		currentPhase := controller.GetCurrentPhase()
		assert.Equal(t, "PHASE3", currentPhase.Name)
		assert.Equal(t, value.StateActive, currentPhase.CurrentState())

		// 条件の数を確認
		conditions := currentPhase.GetConditions()
		assert.Len(t, conditions, 1, "Phase3 should have 1 condition")

		// 条件のIDを確認
		var conditionIDs []value.ConditionID
		for id := range conditions {
			conditionIDs = append(conditionIDs, id)
		}
		assert.Contains(t, conditionIDs, value.ConditionID(5))

		// 条件を満たす
		err = part3.Process(ctx, 1) // カウンターを1増やす
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // 状態更新を待つ

		// Phaseが完了したことを確認
		currentPhase = controller.GetCurrentPhase()
		assert.Equal(t, value.StateNext, currentPhase.CurrentState())
	})
}
