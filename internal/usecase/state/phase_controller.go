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
	observers    struct {
		state         []service.StateObserver
		condition     []service.ConditionObserver
		conditionPart []service.ConditionPartObserver
	}
	mu  sync.RWMutex
	log *zap.Logger
}

// NewPhaseController は新しいPhaseControllerを作成します
func NewPhaseController(phases entity.Phases) *PhaseController {
	log := logger.DefaultLogger()
	if len(phases) <= 0 {
		log.Error("PhaseController", zap.String("error", "No phases found"))
	}
	pc := &PhaseController{
		phases: phases,
		observers: struct {
			state         []service.StateObserver
			condition     []service.ConditionObserver
			conditionPart []service.ConditionPartObserver
		}{
			state:         make([]service.StateObserver, 0),
			condition:     make([]service.ConditionObserver, 0),
			conditionPart: make([]service.ConditionPartObserver, 0),
		},
		log: log,
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

// OnStateChanged は状態変更通知を受け取るメソッドです
func (pc *PhaseController) OnStateChanged(stateName string) {
	pc.log.Debug("PhaseController.OnStateChanged", zap.String("state", stateName),
		zap.String("expected", value.StateNext),
		zap.Bool("equals", stateName == value.StateNext))
	pc.NotifyStateChanged(stateName)

	if stateName == value.StateNext {
		time.Sleep(1 * time.Second)
		pc.log.Debug("start next phase!!!!!!!!!!")
		_ = pc.Start(context.Background())
	}
}

// OnConditionChanged は条件変更通知を受け取るメソッドです
func (pc *PhaseController) OnConditionChanged(condition interface{}) {
	cond, ok := condition.(*entity.Condition)
	if !ok {
		pc.log.Error("Invalid condition type in OnConditionChanged")
		return
	}
	pc.log.Debug("PhaseController.OnConditionChanged", zap.Int64("conditionId", int64(cond.ID)))
	pc.NotifyConditionChanged(condition)
}

// OnConditionPartChanged は条件パーツ変更通知を受け取るメソッドです
func (pc *PhaseController) OnConditionPartChanged(part interface{}) {
	condPart, ok := part.(*entity.ConditionPart)
	if !ok {
		pc.log.Error("Invalid part type in OnConditionPartChanged")
		return
	}
	pc.log.Debug("PhaseController.OnConditionPartChanged", zap.Int64("partId", int64(condPart.ID)))
	pc.NotifyConditionPartChanged(part)
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

	// 存在しなければ初期化してからfinishで終了
	if nextPhase == nil {
		pc.log.Debug("PhaseController.Start", zap.String("action", "No phases found. notify finish"))
		pc.NotifyStateChanged(value.StateFinish)
		pc.SetCurrentPhase(pc.phases[0])
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

// AddStateObserver 状態オブザーバーを追加します
func (pc *PhaseController) AddStateObserver(observer service.StateObserver) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.observers.state = append(pc.observers.state, observer)
}

// RemoveStateObserver 状態オブザーバーを削除します
func (pc *PhaseController) RemoveStateObserver(observer service.StateObserver) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for i, obs := range pc.observers.state {
		if obs == observer {
			pc.observers.state = append(pc.observers.state[:i], pc.observers.state[i+1:]...)
			break
		}
	}
}

// NotifyStateChanged 状態変更を通知します
func (pc *PhaseController) NotifyStateChanged(state string) {
	pc.mu.RLock()
	observers := make([]service.StateObserver, len(pc.observers.state))
	copy(observers, pc.observers.state)
	pc.mu.RUnlock()

	for _, observer := range observers {
		observer.OnStateChanged(state)
	}
}

// AddConditionObserver 条件オブザーバーを追加します
func (pc *PhaseController) AddConditionObserver(observer service.ConditionObserver) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.observers.condition = append(pc.observers.condition, observer)
}

// RemoveConditionObserver 条件オブザーバーを削除します
func (pc *PhaseController) RemoveConditionObserver(observer service.ConditionObserver) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for i, obs := range pc.observers.condition {
		if obs == observer {
			pc.observers.condition = append(pc.observers.condition[:i], pc.observers.condition[i+1:]...)
			break
		}
	}
}

// NotifyConditionChanged 条件変更を通知します
func (pc *PhaseController) NotifyConditionChanged(condition interface{}) {
	pc.mu.RLock()
	observers := make([]service.ConditionObserver, len(pc.observers.condition))
	copy(observers, pc.observers.condition)
	pc.mu.RUnlock()

	for _, observer := range observers {
		observer.OnConditionChanged(condition)
	}
}

// AddConditionPartObserver 条件パーツオブザーバーを追加します
func (pc *PhaseController) AddConditionPartObserver(observer service.ConditionPartObserver) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.observers.conditionPart = append(pc.observers.conditionPart, observer)
}

// RemoveConditionPartObserver 条件パーツオブザーバーを削除します
func (pc *PhaseController) RemoveConditionPartObserver(observer service.ConditionPartObserver) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for i, obs := range pc.observers.conditionPart {
		if obs == observer {
			pc.observers.conditionPart = append(pc.observers.conditionPart[:i], pc.observers.conditionPart[i+1:]...)
			break
		}
	}
}

// NotifyConditionPartChanged 条件パーツ変更を通知します
func (pc *PhaseController) NotifyConditionPartChanged(part interface{}) {
	pc.mu.RLock()
	observers := make([]service.ConditionPartObserver, len(pc.observers.conditionPart))
	copy(observers, pc.observers.conditionPart)
	pc.mu.RUnlock()

	for _, observer := range observers {
		observer.OnConditionPartChanged(part)
	}
}
