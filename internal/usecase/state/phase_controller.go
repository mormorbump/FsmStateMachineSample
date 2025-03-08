package state

import (
	"context"
	"fmt"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PhaseController はフェーズの制御を担当するコントローラーです
type PhaseController struct {
	phases       entity.Phases
	currentPhase *entity.Phase
	observers    []service.ControllerObserver
	mu           sync.RWMutex
	log          *zap.Logger
}

// NewPhaseController は新しいPhaseControllerを作成します
func NewPhaseController(phases entity.Phases) *PhaseController {
	log := logger.DefaultLogger()
	if len(phases) <= 0 {
		log.Error("PhaseController", zap.String("error", "No phases found"))
	}
	pc := &PhaseController{
		phases:    phases,
		observers: make([]service.ControllerObserver, 0),
		log:       log,
	}

	log.Debug("PhaseController initialized", zap.Int("phases count", len(phases)), zap.String("instance", fmt.Sprintf("%p", pc)))
	pc.SetCurrentPhase(phases[0])
	for _, phase := range phases {
		phase.AddObserver(pc)
		log.Debug("Added observer to phase", zap.String("phase", phase.Name), zap.String("observer", fmt.Sprintf("%p", pc)))
		for _, cond := range phase.GetConditions() {
			cond.AddConditionObserver(pc)
			for _, p := range cond.GetParts() {
				p.AddConditionPartObserver(pc)
			}
		}
	}
	return pc
}

// OnPhaseChanged は状態変更通知を受け取るメソッドです
func (pc *PhaseController) OnPhaseChanged(phaseEntity interface{}) {
	// 型チェック
	phase, ok := phaseEntity.(*entity.Phase)
	if !ok {
		pc.log.Error("Invalid phase type in OnPhaseChanged")
		return
	}

	pc.log.Debug("PhaseController.OnPhaseChanged", zap.String("state", phase.CurrentState()),
		zap.String("expected", value.StateNext),
		zap.Bool("equals", phase.CurrentState() == value.StateNext))
	pc.NotifyEntityChanged(phase)

	if phase.CurrentState() == value.StateNext {
		time.Sleep(1 * time.Second)
		pc.log.Debug("start next phase!!!!!!!!!!")
		_ = pc.Start(context.Background())
	}
}

// OnConditionChanged は条件変更通知を受け取るメソッドです
func (pc *PhaseController) OnConditionChanged(condition interface{}) {
	pc.log.Debug("PhaseController.OnConditionChanged", zap.Any("condition", condition))

	// 型チェック
	_, ok := condition.(*entity.Condition)
	if !ok {
		pc.log.Error("Invalid condition type in OnConditionChanged")
		return
	}

	pc.NotifyEntityChanged(condition)
}

// OnConditionPartChanged は条件パーツ変更通知を受け取るメソッドです
func (pc *PhaseController) OnConditionPartChanged(part interface{}) {
	pc.log.Debug("PhaseController.OnConditionPartChanged", zap.Any("part", part))

	// 型チェック
	_, ok := part.(*entity.ConditionPart)
	if !ok {
		pc.log.Error("Invalid part type in OnConditionPartChanged")
		return
	}

	pc.NotifyEntityChanged(part)
}

// GetCurrentPhase は現在のフェーズを取得します
func (pc *PhaseController) GetCurrentPhase() *entity.Phase {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.currentPhase
}

// SetCurrentPhase は現在のフェーズを設定します
func (pc *PhaseController) SetCurrentPhase(phase *entity.Phase) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	oldPhaseName := ""
	if pc.currentPhase != nil {
		oldPhaseName = pc.currentPhase.Name
	}

	pc.currentPhase = phase
	pc.log.Debug("PhaseController", zap.String("old phase", oldPhaseName), zap.String("new phase", phase.Name))
}

// GetPhases は全フェーズを取得します
func (pc *PhaseController) GetPhases() []*entity.Phase {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.phases
}

// Start はフェーズシーケンスを開始します
func (pc *PhaseController) Start(ctx context.Context) error {
	pc.log.Debug("PhaseController.Start", zap.String("action", "Starting phase sequence"))

	// 現在のフェーズを取得
	currentPhase := pc.GetCurrentPhase()
	if currentPhase != nil {
		pc.log.Debug("Current phase before ProcessAndActivateByNextOrder",
			zap.String("name", currentPhase.Name),
			zap.Int("order", currentPhase.Order),
			zap.String("state", currentPhase.CurrentState()))
	}

	// 次のフェーズを取得して活性化
	nextPhase, err := pc.phases.ProcessAndActivateByNextOrder(ctx)

	if nextPhase == nil {
		pc.log.Debug("PhaseController.Start", zap.String("action", "No phases found. notify finish"))
		pc.NotifyEntityChanged(nil)
		//pc.SetCurrentPhase(pc.phases[0])
		return err
	}
	pc.SetCurrentPhase(nextPhase)
	return err
}

// Reset は全てのフェーズをリセットします
func (pc *PhaseController) Reset(ctx context.Context) error {
	if len(pc.phases) <= 0 {
		err := fmt.Errorf("no phases found")
		pc.log.Error("PhaseController.Reset", zap.Error(err))
		return err
	}

	pc.log.Debug("PhaseController.Reset", zap.String("action", "Resetting all phases"))

	// 全フェーズをリセット
	for _, phase := range pc.phases {
		if err := phase.Reset(ctx); err != nil {
			return err
		}
	}

	pc.SetCurrentPhase(pc.phases[0])
	pc.log.Debug("PhaseController.Reset", zap.String("phase name", pc.phases[0].Name))

	return nil
}

func (pc *PhaseController) AddControllerObserver(observer service.ControllerObserver) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.observers = append(pc.observers, observer)
}

func (pc *PhaseController) RemoveControllerObserver(observer service.ControllerObserver) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for i, obs := range pc.observers {
		if obs == observer {
			pc.observers = append(pc.observers[:i], pc.observers[i+1:]...)
			break
		}
	}
}

func (pc *PhaseController) NotifyEntityChanged(entity interface{}) {
	pc.mu.RLock()
	observers := make([]service.ControllerObserver, len(pc.observers))
	copy(observers, pc.observers)
	pc.mu.RUnlock()

	for _, observer := range observers {
		observer.OnEntityChanged(entity)
	}
}
