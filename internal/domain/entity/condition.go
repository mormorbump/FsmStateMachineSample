package entity

import (
	"context"
	"errors"
	"fmt"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/entity/condition_details"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// Condition は状態遷移の条件を表す構造体です
type Condition struct {
	ID          ConditionID
	Label       string
	Kind        ConditionKind
	Parts       []ConditionPart
	Description string

	fsm                    *fsm.FSM
	*core.StateSubjectImpl // Subject実装
	*ConditionSubjectImpl  // Condition Subject実装
	mu                     sync.RWMutex
	log                    *zap.Logger

	satisfiedParts map[ConditionPartID]bool
}

// ConditionKind は条件の種類を表す型です
type ConditionKind int

// ConditionID は条件のIDを表す型です
type ConditionID int

const (
	ConditionKindUnspecified ConditionKind = iota
	ConditionKindTime                      // 時間に基づく条件
	ConditionKindScore                     // スコアに基づく条件
)

// NewCondition は新しいConditionインスタンスを作成します
func NewCondition(id ConditionID, label string, kind ConditionKind) *Condition {
	log := logger.DefaultLogger()
	c := &Condition{
		ID:                   id,
		Label:                label,
		Kind:                 kind,
		Parts:                make([]ConditionPart, 0),
		StateSubjectImpl:     core.NewStateSubjectImpl(),
		ConditionSubjectImpl: NewConditionSubjectImpl(),
		satisfiedParts:       make(map[ConditionPartID]bool),
		log:                  log,
	}

	callbacks := fsm.Callbacks{
		"enter_" + value.StateUnsatisfied: func(ctx context.Context, e *fsm.Event) {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.log.Debug("Condition unsatisfied",
				zap.Int("id", int(c.ID)),
				zap.String("label", c.Label))
		},
		"enter_" + value.StateProcessing: func(ctx context.Context, e *fsm.Event) {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.log.Debug("Condition processing started",
				zap.Int("id", int(c.ID)),
				zap.String("label", c.Label))
		},
		"enter_" + value.StateSatisfied: func(ctx context.Context, e *fsm.Event) {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.log.Debug("Condition satisfied",
				zap.Int("id", int(c.ID)),
				zap.String("label", c.Label))
			c.NotifyStateChanged(value.StateSatisfied)
			c.NotifyConditionSatisfied(c.ID)
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
		},
		callbacks,
	)

	return c
}

// OnPartSatisfied は条件パーツが満たされた時に呼び出されます
func (c *Condition) OnPartSatisfied(partID ConditionPartID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.satisfiedParts[partID] = true
	if c.checkAllPartsSatisfied() {
		_ = c.Complete(context.Background())
	}
}

// checkAllPartsSatisfied はすべての条件パーツが満たされているかチェックします
func (c *Condition) checkAllPartsSatisfied() bool {
	return len(c.satisfiedParts) == len(c.Parts)
}

// Validate は条件の妥当性を検証します
func (c *Condition) Validate() error {
	if c.Kind == ConditionKindUnspecified {
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

// String はConditionKindを文字列に変換します
func (k ConditionKind) String() string {
	switch k {
	case ConditionKindTime:
		return "Time"
	case ConditionKindScore:
		return "Score"
	default:
		return "Unspecified"
	}
}

// CurrentState は現在の状態を返します
func (c *Condition) CurrentState() string {
	return c.fsm.Current()
}

// Activate は条件を有効化します
func (c *Condition) Activate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// すべてのパーツを有効化
	for i := range c.Parts {
		if err := c.Parts[i].Activate(ctx); err != nil {
			return fmt.Errorf("failed to activate part %d: %w", i, err)
		}
	}

	return c.fsm.Event(ctx, value.EventActivate)
}

// StartProcess は条件の処理を開始します
func (c *Condition) StartProcess(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// すべてのパーツの処理を開始
	for i := range c.Parts {
		if err := c.Parts[i].StartProcess(ctx); err != nil {
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
	defer c.mu.Unlock()

	// 満足状態をリセット
	c.satisfiedParts = make(map[ConditionPartID]bool)

	return c.fsm.Event(ctx, value.EventRevert)
}

// AddPart は条件パーツを追加します
func (c *Condition) AddPart(part ConditionPart) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// パーツにObserverを追加
	part.AddConditionPartObserver(c)
	c.Parts = append(c.Parts, part)
}

// InitializePartStrategies は条件パーツの戦略を初期化します
func (c *Condition) InitializePartStrategies(factory condition_details.ConditionStrategyFactory) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.Parts {
		strategy, err := factory.CreateStrategy(c.Kind)
		if err != nil {
			return fmt.Errorf("failed to create strategy for part %d: %w", i, err)
		}

		if err := c.Parts[i].SetStrategy(strategy); err != nil {
			return fmt.Errorf("failed to set strategy for part %d: %w", i, err)
		}
	}

	return nil
}
