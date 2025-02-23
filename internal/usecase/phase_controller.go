package usecase

import (
	"context"
	"go.uber.org/zap"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/entity"
	logger "state_sample/internal/lib"
	"sync"
	"time"
)

type PhaseController struct {
	phases                 entity.Phases
	currentPhase           *entity.Phase
	*core.StateSubjectImpl // Subject実装
	mu                     sync.RWMutex
	log                    *zap.Logger
}

func NewPhaseController(phases entity.Phases) *PhaseController {
	log := logger.DefaultLogger()
	pc := &PhaseController{
		phases:           phases,
		StateSubjectImpl: core.NewStateSubjectImpl(),
		log:              log,
	}

	log.Debug("PhaseController initialized", zap.Int("phases count", len(phases)))

	// 最初のフェーズを設定
	if len(phases) <= 0 {
		log.Error("PhaseController", zap.String("error", "No phases found"))
	}
	pc.currentPhase = phases[0]
	for _, phase := range phases {
		phase.AddObserver(pc)
	}
	return pc
}

func (pc *PhaseController) OnStateChanged(state string) {
	pc.NotifyStateChanged(state)
	time.Sleep(2 * time.Second)
	pc.log.Debug("PhaseController", zap.String("state", state))
	if state == core.StateNext {
		pc.log.Debug("start next phase!!!!!!!!!!")
		_ = pc.Start(context.Background())
	}
}

func (pc *PhaseController) GetCurrentPhase() *entity.Phase {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.currentPhase
}

func (pc *PhaseController) SetCurrentPhase(phase *entity.Phase) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	oldPhase := ""
	if pc.currentPhase != nil {
		oldPhase = pc.currentPhase.Type
		pc.currentPhase.RemoveObserver(pc)
	}

	pc.currentPhase = phase
	if phase != nil {
		phase.AddObserver(pc)
		pc.log.Debug("PhaseController", zap.String("old phase", oldPhase), zap.String("new phase", phase.Type))
	}
}

func (pc *PhaseController) GetPhases() entity.Phases {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.phases
}

func (pc *PhaseController) Start(ctx context.Context) error {
	pc.log.Debug("PhaseController", zap.String("action", "Starting phase sequence"))
	phase, err := pc.phases.ProcessOrder(ctx)
	// 存在しなければfinishで終了
	if phase == nil {
		pc.NotifyStateChanged(core.StateFinish)
		return err
	}
	pc.currentPhase = phase
	return err
}

// Reset は全てのフェーズをリセットします
func (pc *PhaseController) Reset(ctx context.Context) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.log.Debug("PhaseController", zap.String("action", "Resetting all phases"))

	// 全フェーズをリセット
	for _, phase := range pc.phases {
		pc.log.Debug("PhaseController", zap.String("phase type", phase.Type))
		if err := phase.Reset(ctx); err != nil {
			pc.log.Error("PhaseController", zap.String("phase type", phase.Type), zap.Error(err))
			return err
		}
	}

	// 最初のフェーズに戻る
	if len(pc.phases) > 0 {
		pc.SetCurrentPhase(pc.phases[0])
		pc.log.Debug("PhaseController", zap.String("phase type", pc.phases[0].Type))
	}

	return nil
}
