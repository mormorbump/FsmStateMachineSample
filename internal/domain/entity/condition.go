package entity

import (
	"context"
	"errors"
	"fmt"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"
	"time"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// Condition は状態遷移の条件を表す構造体です
type Condition struct {
	ID             value.ConditionID
	Label          string
	Kind           value.ConditionKind
	Parts          map[value.ConditionPartID]*ConditionPart
	Name           string
	Description    string
	IsClear        bool
	StartTime      *time.Time
	FinishTime     *time.Time
	fsm            *fsm.FSM
	stateObservers []service.StateObserver
	condObservers  []service.ConditionObserver
	mu             sync.RWMutex
	log            *zap.Logger
	satisfiedParts map[value.ConditionPartID]bool
}

// NewCondition は新しいConditionインスタンスを作成します
func NewCondition(id value.ConditionID, label string, kind value.ConditionKind) *Condition {
	log := logger.DefaultLogger()
	c := &Condition{
		ID:             id,
		Label:          label,
		Kind:           kind,
		Parts:          make(map[value.ConditionPartID]*ConditionPart),
		stateObservers: make([]service.StateObserver, 0),
		condObservers:  make([]service.ConditionObserver, 0),
		satisfiedParts: make(map[value.ConditionPartID]bool),
		IsClear:        false,
		StartTime:      nil,
		FinishTime:     nil,
		log:            log,
	}

	callbacks := fsm.Callbacks{
		"enter_" + value.StateUnsatisfied: func(ctx context.Context, e *fsm.Event) {
			now := time.Now()
			c.StartTime = &now
			c.log.Debug("Condition enter_unsatisfied: setting start time",
				zap.Time("start_time", now),
				zap.Int64("condition_id", int64(c.ID)))

			for i, part := range c.Parts {
				c.log.Debug("Condition enter_unsatisfied: activating part",
					zap.Any("condition_part", part),
				)
				if err := part.Activate(ctx); err != nil {
					c.log.Error("failed to activate part", zap.Int("part", int(i)), zap.Error(err))
				}
			}
		},
		"enter_" + value.StateSatisfied: func(ctx context.Context, e *fsm.Event) {
			now := time.Now()
			c.FinishTime = &now
			c.log.Debug("Condition enter_satisfied: setting finish time",
				zap.Time("finish_time", now),
				zap.Int64("condition_id", int64(c.ID)))

			c.IsClear = true
			c.NotifyConditionChanged(c)
		},
		"enter_" + value.StateReady: func(ctx context.Context, e *fsm.Event) {
			c.IsClear = false
			c.StartTime = nil
			c.FinishTime = nil
			c.satisfiedParts = make(map[value.ConditionPartID]bool)

			c.log.Debug("Condition enter_ready: resetting time information",
				zap.Int64("condition_id", int64(c.ID)))
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			c.log.Debug("Condition info",
				zap.Int64("id", int64(c.ID)),
				zap.String("label", c.Label))
			c.log.Debug("Condition state transition",
				zap.String("from", e.Src),
				zap.String("to", e.Dst))
			c.NotifyStateChanged(e.Dst)
		},
	}

	c.fsm = fsm.NewFSM(
		value.StateReady,
		fsm.Events{
			{Name: value.EventActivate, Src: []string{value.StateReady}, Dst: value.StateUnsatisfied},
			{Name: value.EventComplete, Src: []string{value.StateUnsatisfied}, Dst: value.StateSatisfied},
			{Name: value.EventRevert, Src: []string{value.StateReady, value.StateSatisfied}, Dst: value.StateUnsatisfied},
			{Name: value.EventReset, Src: []string{value.StateUnsatisfied, value.StateSatisfied}, Dst: value.StateReady},
		},
		callbacks,
	)

	return c
}

// GetParts は条件パーツのスライスを返します
func (c *Condition) GetParts() []*ConditionPart {
	c.mu.RLock()
	defer c.mu.RUnlock()

	parts := make([]*ConditionPart, 0, len(c.Parts))
	for _, part := range c.Parts {
		parts = append(parts, part)
	}
	return parts
}

// OnConditionPartChanged は条件パーツの状態が変更された時に呼び出されます
func (c *Condition) OnConditionPartChanged(part interface{}) {
	condPart, ok := part.(*ConditionPart)
	if !ok {
		c.log.Error("Invalid part type in OnConditionPartChanged")
		return
	}

	c.mu.Lock()
	if condPart.IsSatisfied() {
		c.satisfiedParts[condPart.ID] = true
	}
	c.mu.Unlock()

	c.log.Debug("Condition: OnConditionPartChanged",
		zap.Int64("condition_id", int64(c.ID)),
		zap.Int64("part_id", int64(condPart.ID)),
		zap.Bool("satisfied", c.checkAllPartsSatisfied()),
	)
	if c.checkAllPartsSatisfied() {
		_ = c.Complete(context.Background())
	}
}

// checkAllPartsSatisfied は全ての条件パーツが満たされているかチェックします
func (c *Condition) checkAllPartsSatisfied() bool {
	return len(c.satisfiedParts) == len(c.Parts)
}

// Validate は条件の妥当性を検証します
func (c *Condition) Validate() error {
	if len(c.Parts) == 0 {
		return errors.New("condition must have at least one part")
	}

	// 各パーツの検証
	for _, part := range c.Parts {
		if err := part.Validate(); err != nil {
			return fmt.Errorf("invalid condition part: %w", err)
		}
	}

	return nil
}

// CurrentState は現在の状態を返します
func (c *Condition) CurrentState() string {
	return c.fsm.Current()
}

// Activate は条件をアクティブにします
func (c *Condition) Activate(ctx context.Context) error {
	return c.fsm.Event(ctx, value.EventActivate)
}

// Complete は条件を完了状態にします
func (c *Condition) Complete(ctx context.Context) error {
	return c.fsm.Event(ctx, value.EventComplete)
}

// Revert は条件を未達成状態に戻します
func (c *Condition) Revert(ctx context.Context) error {
	c.mu.Lock()
	c.satisfiedParts = make(map[value.ConditionPartID]bool)
	c.mu.Unlock()

	return c.fsm.Event(ctx, value.EventRevert)
}

// Reset は条件をリセットします
func (c *Condition) Reset(ctx context.Context) error {
	// ログ出力用の時間情報を準備
	var startTimeStr, finishTimeStr string
	if c.StartTime != nil {
		startTimeStr = c.StartTime.Format(time.RFC3339)
	} else {
		startTimeStr = "not set"
	}
	if c.FinishTime != nil {
		finishTimeStr = c.FinishTime.Format(time.RFC3339)
	} else {
		finishTimeStr = "not set"
	}

	c.log.Debug("Condition.Reset: Resetting time information",
		zap.String("start_time", startTimeStr),
		zap.String("finish_time", finishTimeStr),
		zap.Int64("condition_id", int64(c.ID)),
		zap.String("label", c.Label))

	// 時間情報をリセット
	c.StartTime = nil
	c.FinishTime = nil

	// パーツをリセット
	for i, part := range c.Parts {
		if err := part.Reset(ctx); err != nil {
			return fmt.Errorf("failed to reset part %d: %w", i, err)
		}
	}

	return c.fsm.Event(ctx, value.EventReset)
}

// AddPart は条件パーツを追加します
func (c *Condition) AddPart(part *ConditionPart) {
	c.mu.Lock()
	defer c.mu.Unlock()

	part.AddConditionPartObserver(c)
	c.Parts[part.ID] = part
}

// InitializePartStrategies は条件パーツの戦略を初期化します
func (c *Condition) InitializePartStrategies(factory service.StrategyFactory) error {
	// 各パーツに対して個別のStrategyインスタンスを作成するように修正
	for i, part := range c.Parts {
		// 各パーツごとに新しいstrategyインスタンスを作成
		strategy, err := factory.CreateStrategy(c.Kind)
		if err != nil {
			return fmt.Errorf("failed to create strategy %w", err)
		}

		if err = part.SetStrategy(strategy); err != nil {
			return fmt.Errorf("failed to set strategy for part %d: %w", i, err)
		}
	}

	return nil
}

// AddObserver オブザーバーを追加します
func (c *Condition) AddObserver(observer service.StateObserver) {
	if observer == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stateObservers = append(c.stateObservers, observer)
}

// RemoveObserver オブザーバーを削除します
func (c *Condition) RemoveObserver(observer service.StateObserver) {
	if observer == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, obs := range c.stateObservers {
		if obs == observer {
			c.stateObservers = append(c.stateObservers[:i], c.stateObservers[i+1:]...)
			return
		}
	}
}

// NotifyStateChanged 状態変更を通知します
func (c *Condition) NotifyStateChanged(state string) {
	c.mu.RLock()
	observers := make([]service.StateObserver, len(c.stateObservers))
	copy(observers, c.stateObservers)
	c.mu.RUnlock()

	for _, observer := range observers {
		observer.OnStateChanged(state)
	}
}

// AddConditionObserver 条件オブザーバーを追加します
func (c *Condition) AddConditionObserver(observer service.ConditionObserver) {
	if observer == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.condObservers = append(c.condObservers, observer)
}

// RemoveConditionObserver 条件オブザーバーを削除します
func (c *Condition) RemoveConditionObserver(observer service.ConditionObserver) {
	if observer == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, obs := range c.condObservers {
		if obs == observer {
			c.condObservers = append(c.condObservers[:i], c.condObservers[i+1:]...)
			break
		}
	}
}

// NotifyConditionChanged 条件変更を通知します
func (c *Condition) NotifyConditionChanged(condition interface{}) {
	c.mu.RLock()
	observers := make([]service.ConditionObserver, len(c.condObservers))
	copy(observers, c.condObservers)
	c.mu.RUnlock()

	for _, observer := range observers {
		observer.OnConditionChanged(condition)
	}
}
