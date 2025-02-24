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

// ConditionPart は条件の部分的な評価を表す構造体です
type ConditionPart struct {
	ID                 condition.ConditionPartID
	Label              string
	PartKind           condition.Kind
	ComparisonOperator ComparisonOperator
	isClear            bool
	TargetEntityType   string
	TargetEntityID     int64

	ReferenceValueInt    int64
	ReferenceValueFloat  float64
	ReferenceValueString string

	MinValue float64
	MaxValue float64

	Priority int32

	fsm                    *fsm.FSM
	*core.StateSubjectImpl // Subject実装
	*condition.ConditionPartSubjectImpl
	mu  sync.RWMutex
	log *zap.Logger

	strategy condition.Strategy
}

// ComparisonOperator は比較演算子を表す型です
type ComparisonOperator int

const (
	ComparisonOperatorUnspecified ComparisonOperator = iota
	ComparisonOperatorEQ
	ComparisonOperatorNEQ
	ComparisonOperatorGT
	ComparisonOperatorGTE
	ComparisonOperatorLT
	ComparisonOperatorLTE
	ComparisonOperatorBetween
	ComparisonOperatorIn
	ComparisonOperatorNotIn
)

// NewConditionPart は新しいConditionPartインスタンスを作成します
func NewConditionPart(id condition.ConditionPartID, label string) *ConditionPart {
	log := logger.DefaultLogger()
	p := &ConditionPart{
		ID:                       id,
		Label:                    label,
		StateSubjectImpl:         core.NewStateSubjectImpl(),
		ConditionPartSubjectImpl: condition.NewConditionPartSubjectImpl(),
		isClear:                  false,
		log:                      log,
	}

	callbacks := fsm.Callbacks{
		"enter_" + value.StateUnsatisfied: func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Part unsatisfied",
				zap.Int64("id", int64(p.ID)),
				zap.String("label", p.Label))
		},
		"enter_" + value.StateProcessing: func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Part processing started",
				zap.Int64("id", int64(p.ID)),
				zap.String("label", p.Label))
		},
		"enter_" + value.StateSatisfied: func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Part satisfied",
				zap.Int64("id", int64(p.ID)),
				zap.String("label", p.Label))
			p.mu.Lock()
			p.isClear = true
			p.mu.Unlock()
			p.NotifyStateChanged(value.StateSatisfied)
			p.NotifyPartSatisfied(p.ID)
		},
		"enter_" + value.StateReady: func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Part ready",
				zap.Int64("id", int64(p.ID)),
				zap.String("label", p.Label))
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Part state transition",
				zap.String("from", e.Src),
				zap.String("to", e.Dst))
		},
	}

	p.fsm = fsm.NewFSM(
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

	return p
}

// GetID はConditionPartIDを返します
func (p *ConditionPart) GetID() condition.ConditionPartID {
	return p.ID
}

// GetReferenceValueInt はReferenceValueIntを返します
func (p *ConditionPart) GetReferenceValueInt() int64 {
	return p.ReferenceValueInt
}

// AddObserver はオブザーバーを追加します
func (p *ConditionPart) AddObserver(observer interface{}) {
	if stateObserver, ok := observer.(core.StateObserver); ok {
		p.StateSubjectImpl.AddObserver(stateObserver)
	}
	if partObserver, ok := observer.(condition.ConditionPartObserver); ok {
		p.ConditionPartSubjectImpl.AddConditionPartObserver(partObserver)
	}
}

// OnTimeTicked はタイマーのティック時に呼び出されます
func (p *ConditionPart) OnTimeTicked() {
	p.Complete(context.Background())
}

// Validate は条件パーツの妥当性を検証します
func (p *ConditionPart) Validate() error {
	if p.ComparisonOperator == ComparisonOperatorUnspecified {
		return errors.New("comparison operator must be specified")
	}

	// 比較演算子がBetweenの場合、MinValueとMaxValueが必要
	if p.ComparisonOperator == ComparisonOperatorBetween {
		if p.MinValue >= p.MaxValue {
			return errors.New("min_value must be less than max_value")
		}
	}

	return nil
}

// CurrentState は現在の状態を返します
func (p *ConditionPart) CurrentState() string {
	return p.fsm.Current()
}

// Activate は条件パーツを有効化します
func (p *ConditionPart) Activate(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.strategy != nil {
		if err := p.strategy.Initialize(p); err != nil {
			return fmt.Errorf("failed to initialize strategy: %w", err)
		}
		if err := p.strategy.Evaluate(ctx, p); err != nil {
			return fmt.Errorf("failed to evaluate strategy: %w", err)
		}
	}

	return p.fsm.Event(ctx, value.EventActivate)
}

// StartProcess は条件パーツの処理を開始します
func (p *ConditionPart) StartProcess(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventStartProcess)
}

// Complete は条件パーツを達成状態に遷移させます
func (p *ConditionPart) Complete(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventComplete)
}

// Revert は条件パーツを未達成状態に戻します
func (p *ConditionPart) Revert(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventRevert)
}

// Reset は条件パーツをリセットします
func (p *ConditionPart) Reset(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.strategy != nil {
		if err := p.strategy.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup strategy: %w", err)
		}
	}

	return p.fsm.Event(ctx, value.EventReset)
}

// IsClear は条件パーツがクリアされているかどうかを返します
func (p *ConditionPart) IsClear() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isClear
}

// SetStrategy は評価戦略を設定します
func (p *ConditionPart) SetStrategy(strategy condition.Strategy) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.strategy != nil {
		if err := p.strategy.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup old strategy: %w", err)
		}
	}

	p.strategy = strategy
	return nil
}
