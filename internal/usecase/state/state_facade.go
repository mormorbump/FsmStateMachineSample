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

// StateFacade フェーズ管理システムのインターフェースを提供します
type StateFacade interface {
	Start(ctx context.Context) error
	Reset(ctx context.Context) error
	GetCurrentPhase() *entity.Phase
	GetController() *PhaseController
	GetConditionPart(conditionID, partID int64) (*entity.ConditionPart, error)
}

// stateFacadeImpl はStateFacadeの実装です
type stateFacadeImpl struct {
	controller *PhaseController
}

// NewStateFacade は新しいStateFacadeを作成します
func NewStateFacade() StateFacade {
	log := logger.DefaultLogger()
	factory := strategy.NewStrategyFactory()

	// Phase1 (カウンター条件)
	part1 := entity.NewConditionPart(1, "Counter_Part")
	part1.ReferenceValueInt = 5                            // 目標値: 1
	part1.ComparisonOperator = value.ComparisonOperatorGTE // 以上
	cond1 := entity.NewCondition(1, "Counter_Condition", value.KindCounter)
	cond1.AddPart(part1)
	if err := cond1.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	phase1 := entity.NewPhase("PHASE1", 1, []*entity.Condition{cond1}, value.ConditionTypeSingle, value.GameRule_Animation)
	part1.AddConditionPartObserver(cond1)
	cond1.AddConditionObserver(phase1)
	log.Debug("StateFacade initialized", zap.Any("phase1", phase1))

	// Phase2 (2秒)
	part2 := entity.NewConditionPart(2, "Phase2_Part")
	part2.ReferenceValueInt = 2
	cond2 := entity.NewCondition(2, "Phase2_Condition", value.KindTime)
	cond2.AddPart(part2)
	if err := cond2.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	phase2 := entity.NewPhase("PHASE2", 2, []*entity.Condition{cond2}, value.ConditionTypeSingle, value.GameRule_Animation)
	part2.AddConditionPartObserver(cond2)
	cond2.AddConditionObserver(phase2)
	log.Debug("StateFacade initialized", zap.Any("phase2", phase2))

	// Phase3 (3秒)
	part3 := entity.NewConditionPart(3, "Phase3_Part")
	part3.ReferenceValueInt = 3
	cond3 := entity.NewCondition(3, "Phase3_Condition", value.KindTime)
	cond3.AddPart(part3)
	if err := cond3.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	phase3 := entity.NewPhase("PHASE3", 3, []*entity.Condition{cond3}, value.ConditionTypeSingle, value.GameRule_Animation)
	part3.AddConditionPartObserver(cond3)
	cond3.AddConditionObserver(phase3)
	log.Debug("StateFacade initialized", zap.Any("phase3", phase3))

	phases := []*entity.Phase{phase1, phase2, phase3}
	controller := NewPhaseController(phases)

	return &stateFacadeImpl{
		controller: controller,
	}
}

// Start はフェーズシーケンスを開始します
func (sf *stateFacadeImpl) Start(ctx context.Context) error {
	return sf.controller.Start(ctx)
}

// Reset は全てのフェーズをリセットします
func (sf *stateFacadeImpl) Reset(ctx context.Context) error {
	return sf.controller.Reset(ctx)
}

// GetCurrentPhase は現在のフェーズを取得します
func (sf *stateFacadeImpl) GetCurrentPhase() *entity.Phase {
	return sf.controller.GetCurrentPhase()
}

// GetController はPhaseControllerを取得します
func (sf *stateFacadeImpl) GetController() *PhaseController {
	return sf.controller
}

// GetConditionPart は指定されたIDの条件パーツを取得します
func (sf *stateFacadeImpl) GetConditionPart(conditionID, partID int64) (*entity.ConditionPart, error) {
	phase := sf.GetCurrentPhase()
	if phase == nil {
		return nil, fmt.Errorf("no active phase")
	}

	for _, condition := range phase.GetConditions() {
		if condition.ID == value.ConditionID(conditionID) {
			for _, part := range condition.GetParts() {
				if part.ID == value.ConditionPartID(partID) {
					return part, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("condition part not found")
}
