package entity

import (
	"context"
	"fmt"
	"sort"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"
	"time"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// Phase はゲームの各フェーズを表す構造体です
type Phase struct {
	ID                  value.PhaseID
	Order               int
	isActive            bool
	IsClear             bool
	Name                string
	Description         string
	Rule                value.GameRule
	ConditionType       value.ConditionType
	ConditionIDs        []value.ConditionID
	SatisfiedConditions map[value.ConditionID]bool
	Conditions          map[value.ConditionID]*Condition
	StartTime           *time.Time
	FinishTime          *time.Time
	fsm                 *fsm.FSM
	observers           []service.PhaseObserver
	mu                  sync.RWMutex
	log                 *zap.Logger

	// 新しいフィールド - 階層構造のための追加
	ParentID                       value.PhaseID // 親フェーズのID（ルートフェーズの場合は0）
	Parent                         *Phase        // 親フェーズへの参照
	Children                       []*Phase      // 子フェーズのスライス
	AutoProgressOnChildrenComplete bool          // 子フェーズ完了時に自動的に進捗するかどうか
}

// NewPhase は新しいPhaseインスタンスを作成します
func NewPhase(id value.PhaseID, name string, order int, conditions []*Condition, conditionType value.ConditionType, rule value.GameRule, parentID value.PhaseID, autoProgressOnChildrenComplete bool) *Phase {
	log := logger.DefaultLogger()

	p := &Phase{
		ID:                  id,
		Name:                name,
		isActive:            false,
		Order:               order,
		Rule:                rule,
		ConditionType:       conditionType,
		SatisfiedConditions: make(map[value.ConditionID]bool),
		Conditions:          make(map[value.ConditionID]*Condition),
		ConditionIDs:        make([]value.ConditionID, 0),
		observers:           make([]service.PhaseObserver, 0),
		IsClear:             false,
		StartTime:           nil,
		FinishTime:          nil,
		log:                 log,

		// 階層構造のフィールドを初期化
		ParentID:                       parentID,
		Parent:                         nil, // 後で設定
		Children:                       make([]*Phase, 0),
		AutoProgressOnChildrenComplete: autoProgressOnChildrenComplete,
	}

	for _, cond := range conditions {
		p.ConditionIDs = append(p.ConditionIDs, cond.ID)
		p.Conditions[cond.ID] = cond
	}

	callbacks := fsm.Callbacks{
		"enter_" + value.StateActive: func(ctx context.Context, e *fsm.Event) {
			p.isActive = true
			now := time.Now()
			p.StartTime = &now
			for _, c := range p.Conditions {
				p.log.Debug("Phase enter_active: Activating condition", zap.Any("condition", c))
				if err := c.Activate(ctx); err != nil {
					p.log.Error("Failed to activate condition",
						zap.Error(err),
						zap.Int64("condition_id", int64(c.ID)))
				}
			}
		},
		"enter_" + value.StateNext: func(ctx context.Context, e *fsm.Event) {},
		"enter_" + value.StateFinish: func(ctx context.Context, e *fsm.Event) {
			p.isActive = false
			now := time.Now()
			p.FinishTime = &now
		},
		"enter_" + value.StateReady: func(ctx context.Context, e *fsm.Event) {
			p.isActive = false
			p.IsClear = false
			p.StartTime = nil
			p.FinishTime = nil
			p.SatisfiedConditions = make(map[value.ConditionID]bool)
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Phase transition", zap.String("Name", p.Name), zap.String("from", e.Src), zap.String("to", e.Dst))
			p.log.Debug("Phase state changed", zap.String("Name", p.Name), zap.String("state", p.CurrentState()))

			if e.Dst != value.StateFinish {
				p.log.Debug("Calling NotifyPhaseChanged",
					zap.String("phase", p.Name),
					zap.String("state", p.CurrentState()),
					zap.Bool("isNext", p.CurrentState() == value.StateNext))
				p.NotifyPhaseChanged()
			}
		},
	}

	p.fsm = fsm.NewFSM(
		value.StateReady,
		fsm.Events{
			{Name: value.EventActivate, Src: []string{value.StateReady}, Dst: value.StateActive},
			{Name: value.EventNext, Src: []string{value.StateActive}, Dst: value.StateNext},
			{Name: value.EventFinish, Src: []string{value.StateNext}, Dst: value.StateFinish},
			{Name: value.EventReset, Src: []string{value.StateActive, value.StateNext, value.StateFinish}, Dst: value.StateReady},
		},
		callbacks,
	)

	return p
}

// OnConditionChanged は条件が変更された時に呼び出されます
func (p *Phase) OnConditionChanged(condition interface{}) {
	cond, ok := condition.(*Condition)
	if !ok {
		p.log.Error("Invalid condition type in OnConditionChanged")
		return
	}

	if cond.CurrentState() != value.StateSatisfied {
		return
	}

	p.mu.Lock()
	p.SatisfiedConditions[cond.ID] = true
	satisfied := p.checkConditionsSatisfied()
	if satisfied {
		p.IsClear = true
	}
	currentState := p.CurrentState()
	p.mu.Unlock()

	p.log.Debug("Phase.OnConditionChanged",
		zap.String("name", p.Name),
		zap.Bool("satisfied", satisfied),
		zap.Int64("condition_id", int64(cond.ID)),
		zap.String("current_state", currentState))

	// 条件が満たされ、かつフェーズがactive状態の場合のみNextを呼び出す
	if satisfied && currentState == value.StateActive {
		p.log.Debug("Phase.OnConditionChanged: Moving to next state",
			zap.String("phase", p.Name),
			zap.String("from_state", currentState))
		err := p.Next(context.Background())
		if err != nil {
			p.log.Error("Failed to move to next state", zap.Error(err))
		}
	} else if satisfied && currentState != value.StateActive {
		p.log.Debug("Phase.OnConditionChanged: Not moving to next state because phase is not active",
			zap.String("phase", p.Name),
			zap.String("current_state", currentState))
	}
}

// checkConditionsSatisfied は条件が満たされているかチェックします
func (p *Phase) checkConditionsSatisfied() bool {
	if len(p.ConditionIDs) == 0 {
		return false
	}

	switch p.ConditionType {
	case value.ConditionTypeOr:
		return len(p.SatisfiedConditions) > 0
	case value.ConditionTypeAnd:
		return len(p.SatisfiedConditions) == len(p.ConditionIDs)
	default:
		return false
	}
}

// CurrentState は現在の状態を返します
func (p *Phase) CurrentState() string {
	return p.fsm.Current()
}

// SetState はテスト用に状態を直接設定します
func (p *Phase) SetState(state string) {
	p.fsm.SetState(state)
}

// GetStateInfo は状態情報を返します
func (p *Phase) GetStateInfo() *value.GameStateInfo {
	return value.GetGameStateInfo(p.CurrentState())
}

// GetConditions は条件のマップを返します
func (p *Phase) GetConditions() map[value.ConditionID]*Condition {
	return p.Conditions
}

// Activate はフェーズをアクティブにします
func (p *Phase) Activate(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventActivate)
}

// Next は次の状態に進みます
func (p *Phase) Next(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventNext)
}

// Finish はフェーズを終了します
func (p *Phase) Finish(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventFinish)
}

// Reset はフェーズをリセットします
func (p *Phase) Reset(ctx context.Context) error {
	if p.CurrentState() == value.StateReady {
		return nil
	}

	// ログ出力用の時間情報を準備
	var startTimeStr, finishTimeStr string
	if p.StartTime != nil {
		startTimeStr = p.StartTime.Format(time.RFC3339)
	} else {
		startTimeStr = "not set"
	}
	if p.FinishTime != nil {
		finishTimeStr = p.FinishTime.Format(time.RFC3339)
	} else {
		finishTimeStr = "not set"
	}

	p.log.Debug("Phase.Reset: Resetting time information",
		zap.String("start_time", startTimeStr),
		zap.String("finish_time", finishTimeStr),
		zap.String("phase_name", p.Name),
		zap.Int("phase_order", p.Order))

	// 条件とパーツをリセット
	for _, cond := range p.Conditions {
		if err := cond.Reset(ctx); err != nil {
			return fmt.Errorf("failed to reset condition: %w", err)
		}
	}

	return p.fsm.Event(ctx, value.EventReset)
}

// AddObserver オブザーバーを追加します
func (p *Phase) AddObserver(observer service.PhaseObserver) {
	if observer == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers = append(p.observers, observer)
}

// RemoveObserver オブザーバーを削除します
func (p *Phase) RemoveObserver(observer service.PhaseObserver) {
	if observer == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, obs := range p.observers {
		if obs == observer {
			p.observers = append(p.observers[:i], p.observers[i+1:]...)
			return
		}
	}
}

// NotifyStateChanged 状態変更を通知します
func (p *Phase) NotifyPhaseChanged() {
	p.log.Debug("Phase.NotifyPhaseChanged",
		zap.String("phase", p.Name),
		zap.Int("observers", len(p.observers)))

	p.mu.RLock()
	observers := make([]service.PhaseObserver, len(p.observers))
	copy(observers, p.observers)
	p.mu.RUnlock()

	for i, observer := range observers {
		p.log.Debug("Notifying observer",
			zap.Int("index", i),
			zap.String("observer", fmt.Sprintf("%p", observer)))
		observer.OnPhaseChanged(p)
	}
}

// Phases はフェーズのコレクションを表す型です
type Phases []*Phase

// Current は現在アクティブなフェーズを返します
func (p Phases) Current() *Phase {
	log := logger.DefaultLogger()
	log.Debug("Phases.Current")
	for _, phase := range p {
		if phase.isActive {
			log.Debug("Phases.Current", zap.String("name", phase.Name))
			return phase
		}
	}
	return nil
}

// SortByOrder はフェーズをOrderでソートします
func (p Phases) SortByOrder() Phases {
	sorted := make(Phases, len(p))
	copy(sorted, p)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Order < sorted[j].Order
	})
	return sorted
}

// GetByOrder は指定されたOrderを持つフェーズを返します
func (p Phases) GetByOrder(order int) *Phase {
	for _, phase := range p {
		if phase.Order == order {
			return phase
		}
	}
	return nil
}

// GetNextByOrder は現在のOrderの次のOrderを持つフェーズを返します
func (p Phases) GetNextByOrder(currentOrder int) *Phase {
	return p.GetByOrder(currentOrder + 1)
}

// ResetAll は全てのフェーズをリセットします
func (p Phases) ResetAll(ctx context.Context) error {
	for _, phase := range p {
		if err := phase.Reset(ctx); err != nil {
			return err
		}
	}
	return nil
}

// AddChild 子フェーズを追加します
func (p *Phase) AddChild(child *Phase) {
	p.mu.Lock()
	defer p.mu.Unlock()

	child.Parent = p
	p.Children = append(p.Children, child)
}

// GetChildren 子フェーズのスライスを返します（Orderでソート済み）
func (p *Phase) GetChildren() Phases {
	p.mu.RLock()
	defer p.mu.RUnlock()

	children := make(Phases, len(p.Children))
	copy(children, p.Children)
	return children.SortByOrder()
}

// HasChildren 子フェーズを持つかどうかを返します
func (p *Phase) HasChildren() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.Children) > 0
}

// IsActive フェーズがアクティブかどうかを返します
func (p *Phase) IsActive() bool {
	return p.isActive
}

// PhaseMap はParentIDごとにグループ化されたPhasesのマップです
type PhaseMap map[value.PhaseID]Phases

// CurrentPhaseMap は階層ごとの現在アクティブなフェーズを管理するマップです
type CurrentPhaseMap map[value.PhaseID]*Phase

// GroupPhasesByParentID はフェーズをParentIDごとにグループ化します
func GroupPhasesByParentID(phases Phases) PhaseMap {
	phaseMap := make(PhaseMap)
	for _, phase := range phases {
		if _, ok := phaseMap[phase.ParentID]; !ok {
			phaseMap[phase.ParentID] = make(Phases, 0)
		}
		phaseMap[phase.ParentID] = append(phaseMap[phase.ParentID], phase)
	}

	// 各グループをOrderでソート
	for parentID, phases := range phaseMap {
		phaseMap[parentID] = phases.SortByOrder()
	}

	return phaseMap
}

// InitializePhaseHierarchy はフェーズの親子関係を初期化します
func InitializePhaseHierarchy(phases Phases) {
	// IDごとにフェーズをマップ化
	phaseByID := make(map[value.PhaseID]*Phase)
	for _, phase := range phases {
		phaseByID[phase.ID] = phase
	}

	// 親子関係を設定
	for _, phase := range phases {
		if phase.ParentID != 0 {
			if parent, ok := phaseByID[phase.ParentID]; ok {
				parent.AddChild(phase)
				phase.Parent = parent
			}
		}
	}
}

// ProcessAndActivateByNextOrder は次のフェーズに移行します
func (p Phases) ProcessAndActivateByNextOrder(ctx context.Context) (*Phase, error) {
	log := logger.DefaultLogger()
	current := p.Current()

	// 現在アクティブなフェーズがない場合は最初のフェーズを開始
	if current == nil {
		if len(p) <= 0 {
			log.Error("Phases.ProcessAndActivateByNextOrder", zap.Error(fmt.Errorf("no phases available")))
			return nil, fmt.Errorf("no phases available")
		}

		// Orderでソート
		sortedPhases := p.SortByOrder()
		firstPhase := sortedPhases[0]

		log.Debug("Phases.ProcessAndActivateByNextOrder",
			zap.String("action", "Starting first phase"),
			zap.String("name", firstPhase.Name),
		)

		return firstPhase, firstPhase.Activate(ctx)
	}

	log.Debug("ProcessAndActivateByNextOrder: Current phase",
		zap.String("name", current.Name),
		zap.Int("order", current.Order),
		zap.String("state", current.CurrentState()))

	// 現在のフェーズが"next"状態の場合、次のフェーズに進む
	if current.CurrentState() == value.StateNext {
		// 現在のフェーズを終了
		if err := current.Finish(ctx); err != nil {
			log.Error("Failed to finish current phase",
				zap.String("name", current.Name),
				zap.Error(err))
			// エラーが発生しても次のフェーズに進む試みをする
		}

		// 次のフェーズを探す
		nextPhase := p.GetNextByOrder(current.Order)

		if nextPhase != nil {
			log.Debug("Phases.ProcessAndActivateByNextOrder",
				zap.String("name", nextPhase.Name),
				zap.String("action", "Activating next phase"))

			return nextPhase, nextPhase.Activate(ctx)
		}

		log.Debug("No next phase found", zap.Int("current_order", current.Order))
		return nil, nil
	} else {
		log.Debug("Current phase is not in 'next' state, cannot proceed to next phase",
			zap.String("current_state", current.CurrentState()))
	}
	return current, nil
}
