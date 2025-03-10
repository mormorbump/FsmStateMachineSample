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
	phaseFacade *entity.PhaseFacade
	observers   []service.ControllerObserver
	mu          sync.RWMutex
	log         *zap.Logger
}

// NewPhaseController は新しいPhaseControllerを作成します
func NewPhaseController(phases entity.Phases) *PhaseController {
	log := logger.DefaultLogger()
	if len(phases) <= 0 {
		log.Error("PhaseController", zap.String("error", "No phases found"))
	}

	// PhaseFacadeを作成
	phaseFacade := entity.NewPhaseFacade(phases)

	pc := &PhaseController{
		phaseFacade: phaseFacade,
		observers:   make([]service.ControllerObserver, 0),
		log:         log,
	}

	log.Debug("PhaseController initialized",
		zap.Int("phases count", len(phases)),
		zap.String("instance", fmt.Sprintf("%p", pc)))

	// オブザーバーを設定
	for _, phase := range phases {
		phase.AddObserver(pc)
		log.Debug("Added observer to phase",
			zap.String("phase", phase.Name),
			zap.String("observer", fmt.Sprintf("%p", pc)))

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

		ctx := context.Background()

		// フェーズの親IDを取得
		parentID := phase.ParentID

		// 同じ親を持つフェーズのグループを取得
		siblingPhases := pc.phaseFacade.GetPhasesByParentID(parentID)

		// 現在のフェーズを終了
		if err := phase.Finish(ctx); err != nil {
			pc.log.Error("Failed to finish current phase", zap.Error(err))
			// エラーが発生しても次のフェーズに進む試みをする
		}

		// 次のフェーズを探す
		nextPhase := siblingPhases.GetNextByOrder(phase.Order)

		if nextPhase != nil {
			// 次のフェーズが見つかった場合、それをアクティブ化
			pc.log.Debug("Found next phase",
				zap.String("next_phase", nextPhase.Name),
				zap.Int("next_order", nextPhase.Order))
			_ = pc.ActivatePhaseRecursively(ctx, nextPhase)
		} else if parentID != 0 {
			// 次のフェーズがなく、親がルートでない場合、親の次のフェーズを探す
			pc.log.Debug("No next phase found, checking parent's siblings")

			// 親フェーズが子フェーズ完了時に自動的に進捗する設定の場合
			if phase.Parent != nil && phase.Parent.AutoProgressOnChildrenComplete {
				pc.log.Debug("Moving parent phase to next state (auto progress enabled)",
					zap.String("parent_name", phase.Parent.Name))

				if err := phase.Parent.Next(ctx); err != nil {
					pc.log.Error("Failed to move parent to next state", zap.Error(err))
				}
			}
		} else {
			// 親IDが0（ルートフェーズ）で次のフェーズがない場合、次のルートフェーズを探す
			pc.log.Debug("No next phase found for root phase, looking for next root phase")

			rootPhases := pc.phaseFacade.GetPhasesByParentID(0)
			nextRootPhase := rootPhases.GetNextByOrder(phase.Order)

			if nextRootPhase != nil {
				pc.log.Debug("Found next root phase",
					zap.String("next_root", nextRootPhase.Name),
					zap.Int("next_order", nextRootPhase.Order))
				_ = pc.ActivatePhaseRecursively(ctx, nextRootPhase)
			} else {
				pc.log.Debug("No next root phase found, all phases completed")
				pc.NotifyEntityChanged(nil)
			}
		}
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

// GetPhases は全フェーズを取得します
func (pc *PhaseController) GetPhases() entity.Phases {
	return pc.phaseFacade.GetAllPhases()
}

// ActivatePhaseRecursively はフェーズを再帰的にアクティブ化します
func (pc *PhaseController) ActivatePhaseRecursively(ctx context.Context, phase *entity.Phase) error {
	if phase == nil {
		return fmt.Errorf("phase is nil")
	}

	pc.log.Debug("ActivatePhaseRecursively",
		zap.String("phase", phase.Name),
		zap.Int("order", phase.Order))

	// フェーズをアクティブ化
	if err := phase.Activate(ctx); err != nil {
		return err
	}

	// 現在のフェーズとして設定
	pc.phaseFacade.SetCurrentPhase(phase)

	// 子フェーズがある場合は最初の子フェーズをアクティブ化
	if phase.HasChildren() {
		children := phase.GetChildren()
		if len(children) == 0 {
			pc.log.Error("ActivatePhaseRecursively: HasChildren() is true but GetChildren() returned empty slice",
				zap.String("phase", phase.Name))
			return fmt.Errorf("inconsistent phase state: HasChildren() is true but GetChildren() returned empty slice")
		}

		firstChild := children[0]
		if firstChild == nil {
			pc.log.Error("ActivatePhaseRecursively: first child is nil",
				zap.String("phase", phase.Name))
			return fmt.Errorf("first child is nil")
		}

		pc.log.Debug("ActivatePhaseRecursively: Activating first child",
			zap.String("child", firstChild.Name),
			zap.Int("child_order", firstChild.Order))

		return pc.ActivatePhaseRecursively(ctx, firstChild)
	}

	return nil
}

// Reset は全てのフェーズをリセットします
func (pc *PhaseController) Reset(ctx context.Context) error {
	allPhases := pc.GetPhases()
	if len(allPhases) <= 0 {
		err := fmt.Errorf("no phases found")
		pc.log.Error("PhaseController.Reset", zap.Error(err))
		return err
	}

	pc.log.Debug("PhaseController.Reset", zap.String("action", "Resetting all phases"))

	// 全フェーズをリセット
	for _, phase := range allPhases {
		if err := phase.Reset(ctx); err != nil {
			return err
		}
	}

	// 現在のフェーズマップをクリア
	pc.phaseFacade.ResetCurrentPhaseMap()

	// 最初のルートフェーズを取得
	rootPhases := pc.phaseFacade.GetPhasesByParentID(0)
	if len(rootPhases) > 0 {
		firstRootPhase := rootPhases[0]
		pc.phaseFacade.SetCurrentPhase(firstRootPhase)
		pc.log.Debug("PhaseController.Reset", zap.String("phase name", firstRootPhase.Name))
	}

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
