package state

import (
	"context"
	"errors"
	"fmt"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"
	"time"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// Condition は状態遷移の条件を表す構造体です
type Condition struct {
	ID                         core.ConditionID
	Label                      string
	Kind                       core.ConditionKind
	Parts                      map[core.ConditionPartID]*ConditionPart
	Name                       string
	Description                string
	IsClear                    bool
	StartTime                  *time.Time
	FinishTime                 *time.Time
	fsm                        *fsm.FSM
	*core.StateSubjectImpl     // Subject実装
	*core.ConditionSubjectImpl // Condition Subject実装
	mu                         sync.RWMutex
	log                        *zap.Logger

	satisfiedParts map[core.ConditionPartID]bool
}

// NewCondition は新しいConditionインスタンスを作成します
func NewCondition(id core.ConditionID, label string, kind core.ConditionKind) *Condition {
	log := logger.DefaultLogger()
	c := &Condition{
		ID:                   id,
		Label:                label,
		Kind:                 kind,
		Parts:                make(map[core.ConditionPartID]*ConditionPart),
		StateSubjectImpl:     core.NewStateSubjectImpl(),
		ConditionSubjectImpl: core.NewConditionSubjectImpl(),
		satisfiedParts:       make(map[core.ConditionPartID]bool),
		IsClear:              false,
		StartTime:            nil,
		FinishTime:           nil,
		log:                  log,
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
			c.NotifyConditionSatisfied(c.ID)
		},
		"enter_" + value.StateReady: func(ctx context.Context, e *fsm.Event) {
			c.IsClear = false
			c.StartTime = nil
			c.FinishTime = nil
			c.satisfiedParts = make(map[core.ConditionPartID]bool)

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

func (c *Condition) OnPartSatisfied(partID core.ConditionPartID) {
	c.mu.Lock()
	c.satisfiedParts[partID] = true
	satisfied := c.checkAllPartsSatisfied()
	c.mu.Unlock()

	c.log.Debug("Condition: OnPartSatisfied",
		zap.Int64("condition_id", int64(c.ID)),
		zap.Int64("part_id", int64(partID)),
		zap.Bool("satisfied", satisfied),
	)
	if satisfied {
		_ = c.Complete(context.Background())
	}
}

func (c *Condition) checkAllPartsSatisfied() bool {
	return len(c.satisfiedParts) == len(c.Parts)
}

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

func (c *Condition) CurrentState() string {
	return c.fsm.Current()
}

func (c *Condition) Activate(ctx context.Context) error {
	return c.fsm.Event(ctx, value.EventActivate)
}

func (c *Condition) Complete(ctx context.Context) error {
	return c.fsm.Event(ctx, value.EventComplete)
}

func (c *Condition) Revert(ctx context.Context) error {
	c.mu.Lock()
	c.satisfiedParts = make(map[core.ConditionPartID]bool)
	c.mu.Unlock()

	return c.fsm.Event(ctx, value.EventRevert)
}

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

func (c *Condition) AddPart(part *ConditionPart) {
	c.mu.Lock()
	defer c.mu.Unlock()

	part.AddConditionPartObserver(c)
	c.Parts[part.ID] = part
}

func (c *Condition) InitializePartStrategies(factory *core.DefaultConditionStrategyFactory) error {
	strategy, err := factory.CreateStrategy(c.Kind)
	if err != nil {
		return fmt.Errorf("failed to create strategy %w", err)
	}

	for i, part := range c.Parts {
		if err = part.SetStrategy(strategy); err != nil {
			return fmt.Errorf("failed to set strategy for part %d: %w", i, err)
		}
	}

	return nil
}
