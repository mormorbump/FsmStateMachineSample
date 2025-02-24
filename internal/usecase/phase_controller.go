package usecase

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/state"
	logger "state_sample/internal/lib"
	"sync"
	"time"
)

type PhaseController struct {
	phases                 state.Phases
	currentPhase           *state.Phase
	*core.StateSubjectImpl // Subject実装
	mu                     sync.RWMutex
	log                    *zap.Logger
}

func NewPhaseController(phases state.Phases) *PhaseController {
	log := logger.DefaultLogger()
	if len(phases) <= 0 {
		log.Error("PhaseController", zap.String("error", "No phases found"))
	}
	pc := &PhaseController{
		phases:           phases,
		StateSubjectImpl: core.NewStateSubjectImpl(),
		log:              log,
	}

	log.Debug("PhaseController initialized", zap.Int("phases count", len(phases)))
	pc.SetCurrentPhase(phases[0])
	for _, phase := range phases {
		phase.AddObserver(pc)
	}
	return pc
}

func (pc *PhaseController) OnStateChanged(state string) {
	pc.log.Debug("PhaseController.OnStateChanged", zap.String("state", state))
	pc.NotifyStateChanged(state)
	time.Sleep(1 * time.Second)
	if state == core.StateNext {
		pc.log.Debug("start next phase!!!!!!!!!!")
		_ = pc.Start(context.Background())
	}
}

func (pc *PhaseController) GetCurrentPhase() *state.Phase {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.currentPhase
}

func (pc *PhaseController) SetCurrentPhase(phase *state.Phase) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	oldPhaseName := ""
	if pc.currentPhase != nil {
		oldPhaseName = pc.currentPhase.Name
	}

	pc.currentPhase = phase
	pc.log.Debug("PhaseController", zap.String("old phase", oldPhaseName), zap.String("new phase", phase.Name))
}

func (pc *PhaseController) GetPhases() state.Phases {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.phases
}

func (pc *PhaseController) Start(ctx context.Context) error {
	pc.log.Debug("PhaseController.Start", zap.String("action", "Starting phase sequence"))
	phase, err := pc.phases.ProcessAndActivateByNextOrder(ctx)
	// 存在しなければ初期化してからfinishで終了
	if phase == nil {
		pc.log.Debug("PhaseController.Start", zap.String("action", "No phases found. notify finish"))
		pc.SetCurrentPhase(pc.phases[0])
		pc.NotifyStateChanged(core.StateFinish)
		return err
	}
	pc.SetCurrentPhase(phase)
	return err
}

// Reset は全てのフェーズをリセットします
// SetCurrentPhaseの中でmutexをかけてるので、この中でもmutexをかけるとデッドロックになる。
func (pc *PhaseController) Reset(ctx context.Context) error {
	if len(pc.phases) <= 0 {
		err := fmt.Errorf("no phases found")
		pc.log.Error("PhaseController.Reset", zap.Error(err))
		return err
	}

	pc.log.Debug("PhaseController.Reset", zap.String("action", "Resetting all phases"))

	// 全フェーズをリセット
	err := pc.phases.ResetAll(ctx)
	if err != nil {
		return err
	}

	pc.SetCurrentPhase(pc.phases[0])
	pc.log.Debug("PhaseController.Reset", zap.String("phase name", pc.phases[0].Name))

	return nil
}
