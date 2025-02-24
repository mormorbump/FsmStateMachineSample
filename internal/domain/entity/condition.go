package entity

import (
	"context"
	"errors"
	"fmt"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// Condition は状態遷移の条件を表す構造体です
type Condition struct {
	ID                         core.ConditionID
	Label                      string
	Kind                       core.ConditionKind
	Parts                      []*ConditionPart
	Description                string
	isClear                    bool
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
		Parts:                make([]*ConditionPart, 0),
		StateSubjectImpl:     core.NewStateSubjectImpl(),
		ConditionSubjectImpl: core.NewConditionSubjectImpl(),
		satisfiedParts:       make(map[core.ConditionPartID]bool),
		isClear:              false,
		log:                  log,
	}

	callbacks := fsm.Callbacks{
		"enter_" + value.StateUnsatisfied: func(ctx context.Context, e *fsm.Event) {},
		"after_" + value.StateSatisfied: func(ctx context.Context, e *fsm.Event) {
			c.mu.Lock()
			c.isClear = true
			c.mu.Unlock()
			c.NotifyConditionSatisfied(c.ID)
		},
		"enter_" + value.StateReady: func(ctx context.Context, e *fsm.Event) {
			c.mu.Lock()
			c.isClear = false
			c.satisfiedParts = make(map[core.ConditionPartID]bool)
			c.mu.Unlock()
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

// OnPartSatisfied は条件パーツが満たされた時に呼び出されます
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

// checkAllPartsSatisfied はすべての条件パーツが満たされているかチェックします
func (c *Condition) checkAllPartsSatisfied() bool {
	return len(c.satisfiedParts) == len(c.Parts)
}

// Validate は条件の妥当性を検証します
func (c *Condition) Validate() error {
	if c.Kind == core.KindUnspecified {
		return errors.New("condition kind must be specified")
	}

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

// Activate は条件を有効化します
func (c *Condition) Activate(ctx context.Context) error {
	// すべてのパーツを有効化
	for i, part := range c.Parts {
		if err := part.Activate(ctx); err != nil {
			return fmt.Errorf("failed to activate part %d: %w", i, err)
		}
	}

	return c.fsm.Event(ctx, value.EventActivate)
}

// Complete は条件を達成状態に遷移させます
func (c *Condition) Complete(ctx context.Context) error {
	return c.fsm.Event(ctx, value.EventComplete)
}

// Revert は条件を未達成状態に戻します
func (c *Condition) Revert(ctx context.Context) error {
	c.mu.Lock()
	c.satisfiedParts = make(map[core.ConditionPartID]bool)
	c.mu.Unlock()

	return c.fsm.Event(ctx, value.EventRevert)
}

// Reset は条件をリセットします
func (c *Condition) Reset(ctx context.Context) error {
	c.mu.Lock()
	parts := make([]*ConditionPart, len(c.Parts))
	copy(parts, c.Parts)
	c.mu.Unlock()

	// すべてのパーツをリセット
	for i := range parts {
		if err := parts[i].Reset(ctx); err != nil {
			return fmt.Errorf("failed to reset part %d: %w", i, err)
		}
	}

	return c.fsm.Event(ctx, value.EventReset)
}

// AddPart は条件パーツを追加します
func (c *Condition) AddPart(part *ConditionPart) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// パーツにObserverを追加
	part.AddConditionPartObserver(c)
	c.Parts = append(c.Parts, part)
}

// IsClear は条件がクリアされているかどうかを返します
func (c *Condition) IsClear() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isClear
}

// InitializePartStrategies は条件パーツの戦略を初期化します
func (c *Condition) InitializePartStrategies(factory *core.DefaultConditionStrategyFactory) error {
	c.mu.Lock()
	parts := make([]*ConditionPart, len(c.Parts))
	copy(parts, c.Parts)
	c.mu.Unlock()

	for i := range parts {
		strategy, err := factory.CreateStrategy(c.Kind)
		if err != nil {
			return fmt.Errorf("failed to create strategy for part %d: %w", i, err)
		}

		if err := parts[i].SetStrategy(strategy); err != nil {
			return fmt.Errorf("failed to set strategy for part %d: %w", i, err)
		}
	}

	return nil
}
