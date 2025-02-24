package entity

import (
	"context"
	"errors"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/entity/condition_details"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// ConditionPart は条件の詳細な部分を表す構造体です
type ConditionPart struct {
	ID                 ConditionPartID
	Label              string
	ComparisonOperator ComparisonOperator

	TargetEntityType string
	TargetEntityID   int64

	ReferenceValueInt    int64
	ReferenceValueFloat  float64
	ReferenceValueString string

	MinValue float64
	MaxValue float64

	Priority int32

	fsm                    *fsm.FSM
	*core.StateSubjectImpl // Subject実装
	*ConditionSubjectImpl  // Condition Subject実装
	strategy               condition_details.ConditionStrategy
	strategyMu             sync.RWMutex
	mu                     sync.RWMutex
	log                    *zap.Logger
}

type ComparisonOperator int
type ConditionPartID int

const (
	ComparisonOperatorUnspecified ComparisonOperator = iota
	ComparisonOperatorEQ                             // 等しい
	ComparisonOperatorNEQ                            // 等しくない
	ComparisonOperatorGT                             // より大きい
	ComparisonOperatorGTE                            // 以上
	ComparisonOperatorLT                             // より小さい
	ComparisonOperatorLTE                            // 以下
	ComparisonOperatorBetween                        // 範囲内
	ComparisonOperatorIn                             // 含まれる
	ComparisonOperatorNotIn                          // 含まれない
)

// NewConditionPart は新しいConditionPartインスタンスを作成します
func NewConditionPart(id ConditionPartID, label string) *ConditionPart {
	log := logger.DefaultLogger()
	p := &ConditionPart{
		ID:                   id,
		Label:                label,
		StateSubjectImpl:     core.NewStateSubjectImpl(),
		ConditionSubjectImpl: NewConditionSubjectImpl(),
		log:                  log,
	}

	callbacks := fsm.Callbacks{
		"enter_" + value.StateUnsatisfied: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.log.Debug("ConditionPart unsatisfied",
				zap.Int("id", int(p.ID)),
				zap.String("label", p.Label))
		},
		"enter_" + value.StateProcessing: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.log.Debug("ConditionPart processing started",
				zap.Int("id", int(p.ID)),
				zap.String("label", p.Label))

			// Strategyの評価を開始
			if err := p.EvaluateWithStrategy(ctx); err != nil {
				p.log.Error("Strategy evaluation failed",
					zap.Error(err),
					zap.Int("id", int(p.ID)))
			}
		},
		"enter_" + value.StateSatisfied: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.log.Debug("ConditionPart satisfied",
				zap.Int("id", int(p.ID)),
				zap.String("label", p.Label))
			p.NotifyStateChanged(value.StateSatisfied)
			p.NotifyPartSatisfied(p.ID)
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("ConditionPart state transition",
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
		},
		callbacks,
	)

	return p
}

// SetStrategy は評価戦略を設定します
func (p *ConditionPart) SetStrategy(strategy condition_details.ConditionStrategy) error {
	p.strategyMu.Lock()
	defer p.strategyMu.Unlock()

	if p.strategy != nil {
		if err := p.strategy.Cleanup(); err != nil {
			return err
		}
	}

	p.strategy = strategy
	return p.strategy.Initialize(p)
}

// EvaluateWithStrategy は設定された戦略で条件を評価します
func (p *ConditionPart) EvaluateWithStrategy(ctx context.Context) error {
	p.strategyMu.RLock()
	defer p.strategyMu.RUnlock()

	if p.strategy == nil {
		return errors.New("strategy not set")
	}

	return p.strategy.Evaluate(ctx, p)
}

// OnTimeTicked はタイマーイベントを処理します
func (p *ConditionPart) OnTimeTicked() {
	p.log.Debug("ConditionPart.OnTimeTicked")
	_ = p.Complete(context.Background())
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

// String はComparisonOperatorを文字列に変換します
func (o ComparisonOperator) String() string {
	switch o {
	case ComparisonOperatorEQ:
		return "="
	case ComparisonOperatorNEQ:
		return "!="
	case ComparisonOperatorGT:
		return ">"
	case ComparisonOperatorGTE:
		return ">="
	case ComparisonOperatorLT:
		return "<"
	case ComparisonOperatorLTE:
		return "<="
	case ComparisonOperatorBetween:
		return "between"
	case ComparisonOperatorIn:
		return "in"
	case ComparisonOperatorNotIn:
		return "not in"
	default:
		return "unspecified"
	}
}

// CurrentState は現在の状態を返します
func (p *ConditionPart) CurrentState() string {
	return p.fsm.Current()
}

// Activate は条件パーツを有効化します
func (p *ConditionPart) Activate(ctx context.Context) error {
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

// AddConditionPartObserver は条件パーツの監視者を追加します
func (p *ConditionPart) AddConditionPartObserver(observer ConditionPartObserver) {
	p.ConditionSubjectImpl.AddConditionPartObserver(observer)
}
