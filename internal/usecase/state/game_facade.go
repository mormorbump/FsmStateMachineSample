package state

import (
	"context"
	"fmt"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"state_sample/internal/usecase/strategy"

	"go.uber.org/zap"
)

type GameFacade struct {
	controller *PhaseController
}

// NewStateFacade は新しいStateFacadeを作成します
func NewStateFacade() *GameFacade {
	log := logger.DefaultLogger()
	factory := strategy.NewStrategyFactory()

	// ルートフェーズ1
	RootParentPhaseID := value.PhaseID(0)
	part1 := entity.NewConditionPart(1, "Time_Part")
	part1.ReferenceValueInt = 5
	cond1 := entity.NewCondition(1, "Time_Condition", value.KindTime)
	cond1.AddPart(part1)
	if err := cond1.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	rootPhaseID_1 := value.PhaseID(1)
	rootPhase1 := entity.NewPhase(rootPhaseID_1, "ROOT_PHASE", 1, []*entity.Condition{cond1}, value.ConditionTypeAnd, value.GameRule_Animation, RootParentPhaseID, false)
	part1.AddConditionPartObserver(cond1)
	cond1.AddConditionObserver(rootPhase1)
	log.Debug("GameFacade initialized", zap.Any("rootPhase1", rootPhase1))

	// ルートフェーズ2
	part2 := entity.NewConditionPart(2, "Time_Part")
	part2.ReferenceValueInt = 5
	cond2 := entity.NewCondition(2, "Time_Condition", value.KindTime)
	cond2.AddPart(part2)                                            // cond1ではなくcond2に追加
	if err := cond2.InitializePartStrategies(factory); err != nil { // cond1ではなくcond2を初期化
		panic(err)
	}
	rootPhaseID_2 := value.PhaseID(2)                                                                                                                                       // 一意のID（2）を割り当て
	rootPhase2 := entity.NewPhase(rootPhaseID_2, "ROOT_PHASE_2", 2, []*entity.Condition{cond2}, value.ConditionTypeAnd, value.GameRule_Animation, RootParentPhaseID, false) // 名前も変更し、cond2を使用
	part2.AddConditionPartObserver(cond2)
	cond2.AddConditionObserver(rootPhase2)
	log.Debug("GameFacade initialized", zap.Any("rootPhase2", rootPhase2)) // ログメッセージも修正

	// 子フェーズ1: CHILD_PHASE1（親=ROOT_PHASE）
	childPart1 := entity.NewConditionPart(3, "Child1_Part")
	childPart1.ReferenceValueInt = 2
	childPart1.ComparisonOperator = value.ComparisonOperatorGTE
	childCond1 := entity.NewCondition(3, "Child1_Condition", value.KindCounter)
	childCond1.AddPart(childPart1)
	if err := childCond1.InitializePartStrategies(factory); err != nil {
		panic(err)
	}

	childPhaseID_1 := value.PhaseID(4) // 一意のID（4）を割り当て（2から変更）
	childPhase1 := entity.NewPhase(childPhaseID_1, "CHILD_PHASE1", 1, []*entity.Condition{childCond1}, value.ConditionTypeOr, value.GameRule_Animation, rootPhaseID_1, false)
	childPart1.AddConditionPartObserver(childCond1)
	childCond1.AddConditionObserver(childPhase1)
	log.Debug("GameFacade initialized", zap.Any("childPhase1", childPhase1))

	// 子フェーズ2: CHILD_PHASE2（親=ROOT_PHASE）
	childPart2 := entity.NewConditionPart(4, "Child2_Part")
	childPart2.ReferenceValueInt = 3
	childPart2.ComparisonOperator = value.ComparisonOperatorGTE
	childCond2 := entity.NewCondition(4, "Child2_Condition", value.KindCounter)
	childCond2.AddPart(childPart2)
	if err := childCond2.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	childPhaseID_2 := value.PhaseID(3)
	childPhase2 := entity.NewPhase(childPhaseID_2, "CHILD_PHASE2", 2, []*entity.Condition{childCond2}, value.ConditionTypeOr, value.GameRule_Animation, rootPhaseID_1, true)
	childPart2.AddConditionPartObserver(childCond2)
	childCond2.AddConditionObserver(childPhase2)
	log.Debug("GameFacade initialized", zap.Any("childPhase2", childPhase2))

	//// 孫フェーズ: GRANDCHILD_PHASE（親=CHILD_PHASE2）
	//grandchildPart := entity.NewConditionPart(5, "Grandchild_Part")
	//grandchildPart.ReferenceValueInt = 2
	//grandchildCond := entity.NewCondition(5, "Grandchild_Condition", value.KindTime)
	//grandchildCond.AddPart(grandchildPart)
	//if err := grandchildCond.InitializePartStrategies(factory); err != nil {
	//	panic(err)
	//}
	//grandchildPhase := entity.NewPhase("GRANDCHILD_PHASE", 1, []*entity.Condition{grandchildCond}, value.ConditionTypeOr, value.GameRule_Animation, 3, false)
	//grandchildPhase.ID = 4 // IDを明示的に設定
	//grandchildPart.AddConditionPartObserver(grandchildCond)
	//grandchildCond.AddConditionObserver(grandchildPhase)
	//log.Debug("GameFacade initialized", zap.Any("grandchildPhase", grandchildPhase))

	// 全フェーズをスライスに追加
	phases := []*entity.Phase{rootPhase1, rootPhase2, childPhase1, childPhase2}

	// PhaseControllerを作成
	controller := NewPhaseController(phases)

	return &GameFacade{
		controller: controller,
	}
}

// Start はフェーズシーケンスを開始します
func (sf *GameFacade) Start(ctx context.Context) error {
	// 最初のルートフェーズを取得
	rootPhases := sf.controller.phaseFacade.GetPhasesByParentID(0)
	if len(rootPhases) <= 0 {
		return fmt.Errorf("no root phases available")
	}

	firstRootPhase := rootPhases[0]

	// 最初のルートフェーズを直接アクティブ化
	return sf.controller.ActivatePhaseRecursively(ctx, firstRootPhase)
}

// Reset は全てのフェーズをリセットします
func (sf *GameFacade) Reset(ctx context.Context) error {
	return sf.controller.Reset(ctx)
}

// GetCurrentPhase は指定された親IDに対する現在のフェーズを取得します
func (sf *GameFacade) GetCurrentPhase(parentID value.PhaseID) *entity.Phase {
	return sf.controller.phaseFacade.GetCurrentPhase(parentID)
}

// GetCurrentLeafPhase は現在アクティブな最下層のフェーズを取得します
func (sf *GameFacade) GetCurrentLeafPhase() *entity.Phase {
	return sf.controller.phaseFacade.GetCurrentLeafPhase()
}

// GetController はPhaseControllerを取得します
func (sf *GameFacade) GetController() *PhaseController {
	return sf.controller
}

// GetConditionPart は指定されたIDの条件パーツを取得します
func (sf *GameFacade) GetConditionPart(conditionID, partID int64) (*entity.ConditionPart, error) {
	// まずルートフェーズから条件パーツを探す
	rootPhase := sf.GetCurrentPhase(0)
	if rootPhase != nil {
		for _, condition := range rootPhase.GetConditions() {
			if condition.ID == value.ConditionID(conditionID) {
				for _, part := range condition.GetParts() {
					if part.ID == value.ConditionPartID(partID) {
						return part, nil
					}
				}
			}
		}
	}

	// ルートフェーズで見つからない場合は、すべてのフェーズから探す
	allPhases := sf.controller.GetPhases()
	for _, phase := range allPhases {
		for _, condition := range phase.GetConditions() {
			if condition.ID == value.ConditionID(conditionID) {
				for _, part := range condition.GetParts() {
					if part.ID == value.ConditionPartID(partID) {
						return part, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("condition part not found")
}
