package fsm

import (
	"context"
	"sort"
)

// StateFacade はフェーズ管理システムのインターフェースを提供します
type StateFacade interface {
	Start(ctx context.Context) error
	Stop() error
	Reset(ctx context.Context) error
	GetCurrentPhase() *Phase
	GetController() *PhaseController
}

// stateFacadeImpl はStateFacadeの実装です
type stateFacadeImpl struct {
	controller *PhaseController
}

// NewStateFacade は新しいStateFacadeインスタンスを作成します
func NewStateFacade(phases Phases) StateFacade {
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

// Stop はフェーズシステムを停止します
func (sf *stateFacadeImpl) Stop() error {
	// 必要に応じて停止処理を実装
	return nil
}

// Reset はフェーズシステムをリセットします
func (sf *stateFacadeImpl) Reset(ctx context.Context) error {
	return sf.controller.Reset(ctx)
}

// GetCurrentPhase は現在アクティブなフェーズを返します
func (sf *stateFacadeImpl) GetCurrentPhase() *Phase {
	return sf.controller.GetCurrentPhase()
}

// GetController はPhaseControllerを返します
func (sf *stateFacadeImpl) GetController() *PhaseController {
	return sf.controller
}
