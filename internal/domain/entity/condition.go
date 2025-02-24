package entity

import (
	"context"
	"errors"
	"fmt"
	"state_sample/internal/domain/condition"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// Condition は状態遷移の条件を表す構造体です
type Condition struct {
	ID                              condition.ConditionID
	Label                           string
	Kind                            condition.Kind
	Parts                           []*ConditionPart
	Description                     string
	isClear                         bool
	fsm                             *fsm.FSM
	*core.StateSubjectImpl          // Subject実装
	*condition.ConditionSubjectImpl // Condition Subject実装
	mu                              sync.RWMutex
	log                             *zap.Logger

	satisfiedParts map[condition.ConditionPartID]bool
}

// NewCondition は新しいConditionインスタンスを作成します
func NewCondition(id condition.ConditionID, label string, kind condition.Kind) *Condition {
	log := logger.DefaultLogger()
	c := &Condition{
		ID:                   id,
		Label:                label,
		Kind:                 kind,
		Parts:                make([]*ConditionPart, 0),
		StateSubjectImpl:     core.NewStateSubjectImpl(),
		ConditionSubjectImpl: condition.NewConditionSubjectImpl(),
		satisfiedParts:       make(map[condition.ConditionPartID]bool),
		isClear:              false,
		log:                  log,
	}

	callbacks := fsm.Callbacks{
		"enter_" + value.StateUnsatisfied: func(ctx context.Context, e *fsm.Event) {
			c.log.Debug("Condition unsatisfied",
				zap.Int64("id", int64(c.ID)),
				zap.String("label", c.Label))
		},
		"enter_" + value.StateProcessing: func(ctx context.Context, e *fsm.Event) {
			c.log.Debug("Condition processing started",
				zap.Int64("id", int64(c.ID)),
				zap.String("label", c.Label))
		},
		"enter_" + value.StateSatisfied: func(ctx context.Context, e *fsm.Event) {
			c.log.Debug("Condition satisfied",
				zap.Int64("id", int64(c.ID)),
				zap.String("label", c.Label))
			c.mu.Lock()
			c.isClear = true
			c.mu.Unlock()
			c.NotifyStateChanged(value.StateSatisfied)
			c.NotifyConditionSatisfied(c.ID)
		},
		"enter_" + value.StateReady: func(ctx context.Context, e *fsm.Event) {
			c.mu.Lock()
			c.satisfiedParts = make(map[condition.ConditionPartID]bool)
			c.mu.Unlock()
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			c.log.Debug("Condition state transition",
				zap.String("from", e.Src),
				zap.String("to", e.Dst))
		},
	}

	c.fsm = fsm.NewFSM(
		value.StateReady,
		fsm.Events{
			{Name: value.EventActivate, Src: []string{value.StateReady}, Dst: value.StateUnsatisfied},
			{Name: value.EventStartProcess, Src: []string{value.StateUnsatisfied}, Dst: value.StateProcessing},
			{Name: value.EventComplete, Src: []string{value.StateProcessing}, Dst: value.StateSatisfied},
			{Name: value.EventRevert, Src: []string{value.StateProcessing}, Dst: value.StateUnsatisfied},
			{Name: value.EventReset, Src: []string{value.StateUnsatisfied, value.StateProcessing, value.StateSatisfied}, Dst: value.StateReady},
		},
		callbacks,
	)

	return c
}

// OnPartSatisfied は条件パーツが満たされた時に呼び出されます
func (c *Condition) OnPartSatisfied(partID condition.ConditionPartID) {
	c.mu.Lock()
	c.satisfiedParts[partID] = true
	satisfied := c.checkAllPartsSatisfied()
	c.mu.Unlock()

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
	if c.Kind == condition.KindUnspecified {
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
	c.mu.Lock()
	parts := make([]*ConditionPart, len(c.Parts))
	copy(parts, c.Parts)
	c.mu.Unlock()

	// すべてのパーツを有効化
	for i := range parts {
		if err := parts[i].Activate(ctx); err != nil {
			return fmt.Errorf("failed to activate part %d: %w", i, err)
		}
	}

	return c.fsm.Event(ctx, value.EventActivate)
}

// StartProcess は条件の処理を開始します
func (c *Condition) StartProcess(ctx context.Context) error {
	c.mu.Lock()
	parts := make([]*ConditionPart, len(c.Parts))
	copy(parts, c.Parts)
	c.mu.Unlock()

	// すべてのパーツの処理を開始
	for i := range parts {
		if err := parts[i].StartProcess(ctx); err != nil {
			return fmt.Errorf("failed to start process for part %d: %w", i, err)
		}
	}

	return c.fsm.Event(ctx, value.EventStartProcess)
}

// Complete は条件を達成状態に遷移させます
func (c *Condition) Complete(ctx context.Context) error {
	return c.fsm.Event(ctx, value.EventComplete)
}

// Revert は条件を未達成状態に戻します
func (c *Condition) Revert(ctx context.Context) error {
	c.mu.Lock()
	c.satisfiedParts = make(map[condition.ConditionPartID]bool)
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
func (c *Condition) InitializePartStrategies(factory condition.Factory) error {
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
