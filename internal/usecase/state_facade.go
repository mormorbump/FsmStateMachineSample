package usecase

import (
	"context"
	"sort"
	"state_sample/internal/domain/entity"
	"time"
)

// StateFacade はフェーズ管理システムのインターフェースを提供します
type StateFacade interface {
	Start(ctx context.Context) error
	Reset(ctx context.Context) error
	GetCurrentPhase() *entity.Phase
	GetController() *PhaseController
}

// stateFacadeImpl はStateFacadeの実装です
type stateFacadeImpl struct {
	controller *PhaseController
}

// NewStateFacade は新しいStateFacadeインスタンスを作成します
func NewStateFacade() StateFacade {
	phases := entity.Phases{
		entity.NewPhase("BUILD_PHASE", 3*time.Second, 1),
		entity.NewPhase("COMBAT_PHASE", 1*time.Second, 2),
		entity.NewPhase("RESOLUTION_PHASE", 5*time.Second, 3),
	}
	// Order順にソート
	sort.Slice(phases, func(i, j int) bool {
		return phases[i].Order < phases[j].Order
	})

	controller := NewPhaseController(phases)

	return &stateFacadeImpl{
		controller: controller,
	}
}

// Start はフェーズシステムを開始します
func (sf *stateFacadeImpl) Start(ctx context.Context) error {
	return sf.controller.Start(ctx)
}

// Reset はフェーズシステムをリセットします
func (sf *stateFacadeImpl) Reset(ctx context.Context) error {
	return sf.controller.Reset(ctx)
}

// GetCurrentPhase は現在アクティブなフェーズを返します
func (sf *stateFacadeImpl) GetCurrentPhase() *entity.Phase {
	return sf.controller.GetCurrentPhase()
}

// GetController はPhaseControllerを返します
func (sf *stateFacadeImpl) GetController() *PhaseController {
	return sf.controller
}
