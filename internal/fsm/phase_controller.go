package fsm

import (
	"context"
	"sync"
)

// PhaseController はフェーズの状態を監視・管理します
type PhaseController struct {
	phases            Phases
	currentPhase      *Phase
	*StateSubjectImpl // Subject実装
	*ObserverImpl     // Observer実装
	mu                sync.RWMutex
}

// NewPhaseController は新しいPhaseControllerインスタンスを作成します
func NewPhaseController(phases Phases) *PhaseController {
	pc := &PhaseController{
		phases:           phases,
		StateSubjectImpl: NewStateSubjectImpl(),
	}

	// ObserverImplの初期化
	pc.ObserverImpl = NewObserverImpl(
		func(state string) {
			pc.NotifyStateChanged(state) // 状態変更を上位に伝播
		},
		func(err error) {
			pc.NotifyError(err) // エラーを上位に伝播
		},
	)

	// 最初のフェーズを設定
	if len(phases) > 0 {
		pc.currentPhase = phases[0]
		pc.currentPhase.AddObserver(pc)
	}

	return pc
}

// GetCurrentPhase は現在のフェーズを返します
func (pc *PhaseController) GetCurrentPhase() *Phase {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.currentPhase
}

// SetCurrentPhase は現在のフェーズを設定します
func (pc *PhaseController) SetCurrentPhase(phase *Phase) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.currentPhase != nil {
		pc.currentPhase.RemoveObserver(pc)
	}

	pc.currentPhase = phase
	if phase != nil {
		phase.AddObserver(pc)
	}
}

// GetPhases は管理しているフェーズのスライスを返します
func (pc *PhaseController) GetPhases() Phases {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.phases
}

func (pc *PhaseController) Start(ctx context.Context) error {
	return pc.phases.MoveNext(ctx)
}

// Reset は全てのフェーズをリセットします
func (pc *PhaseController) Reset(ctx context.Context) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// 全フェーズをリセット
	for _, phase := range pc.phases {
		if err := phase.Reset(ctx); err != nil {
			return err
		}
	}

	// 最初のフェーズに戻る
	if len(pc.phases) > 0 {
		pc.SetCurrentPhase(pc.phases[0])
	}

	return nil
}
