package usecase

import (
	"context"
	"state_sample/internal/domain/entity"
	"time"
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
	phases := entity.Phases{
		entity.NewPhase("PHASE1", 1*time.Second, 1),
		entity.NewPhase("PHASE2", 2*time.Second, 2),
		entity.NewPhase("PHASE3", 3*time.Second, 3),
	}

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
