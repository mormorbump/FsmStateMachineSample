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

// ConditionPart は条件の部分的な評価を表す構造体です
type ConditionPart struct {
	ID                     core.ConditionPartID
	Label                  string
	ComparisonOperator     core.ComparisonOperator
	IsClear                bool
	TargetEntityType       string
	TargetEntityID         int64
	ReferenceValueInt      int64
	ReferenceValueFloat    float64
	ReferenceValueString   string
	MinValue               int64
	MaxValue               int64
	Priority               int32
	StartTime              *time.Time
	FinishTime             *time.Time
	fsm                    *fsm.FSM
	*core.StateSubjectImpl // Subject実装
	*ConditionPartSubjectImpl
	mu  sync.RWMutex
	log *zap.Logger

	strategy PartStrategy
}

// NewConditionPart は新しいConditionPartインスタンスを作成します
func NewConditionPart(id core.ConditionPartID, label string) *ConditionPart {
	log := logger.DefaultLogger()
	p := &ConditionPart{
		ID:                       id,
		Label:                    label,
		StateSubjectImpl:         core.NewStateSubjectImpl(),
		ConditionPartSubjectImpl: NewConditionPartSubjectImpl(),
		IsClear:                  false,
		StartTime:                nil,
		FinishTime:               nil,
		log:                      log,
	}

	callbacks := fsm.Callbacks{
		"enter_" + value.StateUnsatisfied: func(ctx context.Context, e *fsm.Event) {
			now := time.Now()
			p.StartTime = &now
			log.Debug("ConditionPart enter_unsatisfied",
				zap.Int64("id", int64(p.ID)),
				zap.Time("start_time", now),
			)
			if p.strategy != nil {
				if err := p.strategy.Evaluate(ctx, p, nil); err != nil {
					p.log.Error("failed to evaluate strategy", zap.Error(err))
				}
			}
		},
		"enter_" + value.StateProcessing: func(ctx context.Context, e *fsm.Event) {},
		"enter_" + value.StateSatisfied: func(ctx context.Context, e *fsm.Event) {
			now := time.Now()
			p.FinishTime = &now
			p.IsClear = true
			p.log.Debug("part satisfied",
				zap.Bool("IsClear", p.IsClear),
				zap.Time("finish_time", now),
			)
			if p.strategy != nil {
				if err := p.strategy.Cleanup(); err != nil {
					p.log.Error("failed to cleanup strategy", zap.Error(err))
				}
			}
			p.NotifyPartSatisfied(p.ID)
		},
		"enter_" + value.StateReady: func(ctx context.Context, e *fsm.Event) {
			p.IsClear = false
			p.StartTime = nil
			p.FinishTime = nil
			p.log.Debug("ConditionPart enter_ready: resetting time information",
				zap.Int64("id", int64(p.ID)))
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("ConditionPart Info",
				zap.Int64("id", int64(p.ID)),
				zap.String("label", p.Label))
			p.log.Debug("ConditionPart state transition",
				zap.String("from", e.Src),
				zap.String("to", e.Dst))
		},
	}

	p.fsm = fsm.NewFSM(
		value.StateReady,
		fsm.Events{
			{Name: value.EventActivate, Src: []string{value.StateReady}, Dst: value.StateUnsatisfied},
			{Name: value.EventProcess, Src: []string{value.StateUnsatisfied}, Dst: value.StateProcessing},
			{Name: value.EventComplete, Src: []string{value.StateProcessing}, Dst: value.StateSatisfied},
			{Name: value.EventTimeout, Src: []string{value.StateProcessing, value.StateUnsatisfied}, Dst: value.StateSatisfied},
			{Name: value.EventRevert, Src: []string{value.StateProcessing}, Dst: value.StateUnsatisfied},
			{Name: value.EventReset, Src: []string{value.StateUnsatisfied, value.StateProcessing, value.StateSatisfied}, Dst: value.StateReady},
		},
		callbacks,
	)

	return p
}

func (p *ConditionPart) GetReferenceValueInt() int64 {
	return p.ReferenceValueInt
}

func (p *ConditionPart) GetComparisonOperator() core.ComparisonOperator {
	return p.ComparisonOperator
}

func (p *ConditionPart) GetMaxValue() int64 {
	return p.MaxValue
}

func (p *ConditionPart) GetMinValue() int64 {
	return p.MinValue
}

func (p *ConditionPart) OnTimeTicked() {
	p.Timeout(context.Background())
}

// Validate は条件パーツの妥当性を検証します
func (p *ConditionPart) Validate() error {
	if p.ComparisonOperator == core.ComparisonOperatorUnspecified {
		return errors.New("comparison operator must be specified")
	}

	// 比較演算子がBetweenの場合、MinValueとMaxValueが必要
	if p.ComparisonOperator == core.ComparisonOperatorBetween {
		if p.MinValue >= p.MaxValue {
			return errors.New("min_value must be less than max_value")
		}
	}

	return nil
}

func (p *ConditionPart) CurrentState() string {
	return p.fsm.Current()
}

func (p *ConditionPart) Activate(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventActivate)
}

func (p *ConditionPart) Process(ctx context.Context, increment int64) error {
	// 複数人から呼ばれる部分なのでmutex
	if p.strategy != nil {
		if err := p.strategy.Evaluate(ctx, p, increment); err != nil {
			p.log.Error("failed to evaluate strategy", zap.Error(err))
		}
	}
	return p.fsm.Event(ctx, value.EventProcess)
}

func (p *ConditionPart) Complete(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventComplete)
}

func (p *ConditionPart) Timeout(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventTimeout)
}

func (p *ConditionPart) Revert(ctx context.Context) error {
	return p.fsm.Event(ctx, value.EventRevert)
}

func (p *ConditionPart) Reset(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

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

	p.log.Debug("ConditionPart.Reset: Resetting time information",
		zap.String("start_time", startTimeStr),
		zap.String("finish_time", finishTimeStr),
		zap.Int64("id", int64(p.ID)),
		zap.String("label", p.Label))

	// 時間情報をリセット
	p.StartTime = nil
	p.FinishTime = nil

	if p.strategy != nil {
		if err := p.strategy.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup strategy: %w", err)
		}
	}

	return p.fsm.Event(ctx, value.EventReset)
}

func (p *ConditionPart) SetStrategy(strategy PartStrategy) error {
	if p.strategy != nil {
		if err := p.strategy.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup old strategy: %w", err)
		}
	}

	if err := strategy.Initialize(p); err != nil {
		return fmt.Errorf("failed to setup strategy: %w", err)
	}
	p.strategy = strategy
	return nil
}
