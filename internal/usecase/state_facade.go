package usecase

import (
	"context"
	"go.uber.org/zap"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/state"
	logger "state_sample/internal/lib"
)

// StateFacade フェーズ管理システムのインターフェースを提供
type StateFacade interface {
	Start(ctx context.Context) error
	Reset(ctx context.Context) error
	GetCurrentPhase() *state.Phase
	GetController() *PhaseController
}

type stateFacadeImpl struct {
	controller *PhaseController
}

func NewStateFacade() StateFacade {
	log := logger.DefaultLogger()
	factory := core.NewDefaultConditionStrategyFactory()
	// Phase1 (1秒)
	part1 := state.NewConditionPart(1, "Phase1_Part")
	part1.ReferenceValueInt = 1
	cond1 := state.NewCondition(1, "Phase1_Condition", core.KindTime)
	cond1.AddPart(part1)
	if err := cond1.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	phase1 := state.NewPhase("PHASE1", 1, []*state.Condition{cond1}, state.ConditionTypeSingle, state.GameRule_Animation)
	part1.AddConditionPartObserver(cond1)
	cond1.AddConditionObserver(phase1)
	log.Debug("StateFacade initialized", zap.Any("phase1", phase1))
	// Phase2 (2秒)
	part2 := state.NewConditionPart(2, "Phase2_Part")
	part2.ReferenceValueInt = 2
	cond2 := state.NewCondition(2, "Phase2_Condition", core.KindTime)
	cond2.AddPart(part2)
	if err := cond2.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	phase2 := state.NewPhase("PHASE2", 2, []*state.Condition{cond2}, state.ConditionTypeSingle, state.GameRule_Animation)
	part2.AddConditionPartObserver(cond2)
	cond2.AddConditionObserver(phase2)
	log.Debug("StateFacade initialized", zap.Any("phase2", phase2))

	// Phase3 (3秒)
	part3 := state.NewConditionPart(3, "Phase3_Part")
	part3.ReferenceValueInt = 3
	cond3 := state.NewCondition(3, "Phase3_Condition", core.KindTime)
	cond3.AddPart(part3)
	if err := cond3.InitializePartStrategies(factory); err != nil {
		panic(err)
	}
	phase3 := state.NewPhase("PHASE3", 3, []*state.Condition{cond3}, state.ConditionTypeSingle, state.GameRule_Animation)
	part3.AddConditionPartObserver(cond3)
	cond3.AddConditionObserver(phase3)
	log.Debug("StateFacade initialized", zap.Any("phase3", phase3))

	phases := state.Phases{phase1, phase2, phase3}
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

func (sf *stateFacadeImpl) GetCurrentPhase() *state.Phase {
	return sf.controller.GetCurrentPhase()
}

func (sf *stateFacadeImpl) GetController() *PhaseController {
	return sf.controller
}
