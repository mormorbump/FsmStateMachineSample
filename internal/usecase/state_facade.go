package usecase

import (
	"context"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/entity"
)

// StateFacade フェーズ管理システムのインターフェースを提供
type StateFacade interface {
	Start(ctx context.Context) error
	Reset(ctx context.Context) error
	GetCurrentPhase() *entity.Phase
	GetController() *PhaseController
}

type stateFacadeImpl struct {
	controller *PhaseController
}

func NewStateFacade() StateFacade {
	// Phase1 (1秒)
	part1 := entity.NewConditionPart(1, "Phase1_Part")
	part1.ReferenceValueInt = 1
	cond1 := entity.NewCondition(1, "Phase1_Condition", core.KindTime)
	cond1.AddPart(part1)
	phase1 := entity.NewPhase("PHASE1", 1, cond1)
	part1.AddConditionPartObserver(cond1)
	cond1.AddConditionObserver(phase1)

	// Phase2 (2秒)
	part2 := entity.NewConditionPart(2, "Phase2_Part")
	part2.ReferenceValueInt = 2
	cond2 := entity.NewCondition(2, "Phase2_Condition", core.KindTime)
	cond2.AddPart(part2)
	phase2 := entity.NewPhase("PHASE2", 2, cond2)
	part2.AddConditionPartObserver(cond2)
	cond2.AddConditionObserver(phase2)

	// Phase3 (3秒)
	part3 := entity.NewConditionPart(3, "Phase3_Part")
	part3.ReferenceValueInt = 3
	cond3 := entity.NewCondition(3, "Phase3_Condition", core.KindTime)
	cond3.AddPart(part3)
	phase3 := entity.NewPhase("PHASE3", 3, cond3)
	part3.AddConditionPartObserver(cond3)
	cond3.AddConditionObserver(phase3)

	// Strategyの初期化
	factory := core.NewDefaultConditionStrategyFactory()
	if err := cond1.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	if err := cond2.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	if err := cond3.InitializePartStrategies(factory); err != nil {
		panic(err)
	}

	phases := entity.Phases{phase1, phase2, phase3}
	controller := NewPhaseController(phases)

	return &stateFacadeImpl{
		controller: controller,
	}
}

func (sf *stateFacadeImpl) Start(ctx context.Context) error {
	return sf.controller.Start(ctx)
}

func (sf *stateFacadeImpl) Reset(ctx context.Context) error {
	return sf.controller.Reset(ctx)
}

func (sf *stateFacadeImpl) GetCurrentPhase() *entity.Phase {
	return sf.controller.GetCurrentPhase()
}

func (sf *stateFacadeImpl) GetController() *PhaseController {
	return sf.controller
}
