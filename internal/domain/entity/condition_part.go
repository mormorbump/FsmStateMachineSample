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

// ConditionPart は条件の部分的な評価を表す構造体です
type ConditionPart struct {
	ID                   value.ConditionPartID
	Label                string
	ComparisonOperator   value.ComparisonOperator
	IsClear              bool
	TargetEntityType     string
	TargetEntityID       int64
	ReferenceValueInt    int64
	ReferenceValueFloat  float64
	ReferenceValueString string
	MinValue             int64
	MaxValue             int64
	Priority             int32
	StartTime            *time.Time
	FinishTime           *time.Time
	fsm                  *fsm.FSM
	mu                   sync.RWMutex
	log                  *zap.Logger

	strategy      service.PartStrategy
	partObservers []service.ConditionPartObserver
}

// NewConditionPart は新しいConditionPartインスタンスを作成します
func NewConditionPart(id value.ConditionPartID, label string) *ConditionPart {
	log := logger.DefaultLogger()
	p := &ConditionPart{
		ID:            id,
		Label:         label,
		partObservers: make([]service.ConditionPartObserver, 0),
		IsClear:       false,
		StartTime:     nil,
		FinishTime:    nil,
		log:           log,
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
				if err := p.strategy.Start(ctx, p); err != nil {
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
			{Name: value.EventProcess, Src: []string{value.StateUnsatisfied, value.StateProcessing}, Dst: value.StateProcessing},
			{Name: value.EventComplete, Src: []string{value.StateProcessing, value.StateUnsatisfied}, Dst: value.StateSatisfied},
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

func (p *ConditionPart) GetComparisonOperator() value.ComparisonOperator {
	return p.ComparisonOperator
}

func (p *ConditionPart) GetMaxValue() int64 {
	return p.MaxValue
}

func (p *ConditionPart) GetMinValue() int64 {
	return p.MinValue
}

func (p *ConditionPart) IsSatisfied() bool {
	return p.fsm.Is(value.StateSatisfied)
}

func (p *ConditionPart) GetCurrentValue() interface{} {
	if p.strategy == nil {
		return int64(0) // strategyがnilの場合はデフォルト値を返す
	}
	return p.strategy.GetCurrentValue()
}

func (p *ConditionPart) OnUpdated(event string) {
	p.log.Debug("ConditionPart.OnUpdated called",
		zap.String("event", event),
		zap.String("current_state", p.fsm.Current()),
		zap.Int64("id", int64(p.ID)))

	switch {
	case event == value.EventTimeout:
		p.log.Debug("ConditionPart.OnUpdated: Calling Timeout")
		p.Timeout(context.Background())
	case event == value.EventComplete:
		p.log.Debug("ConditionPart.OnUpdated: Calling Complete")
		p.Complete(context.Background())
	case event == value.EventProcess:
		p.log.Debug("ConditionPart.OnUpdated: Calling Process for EventProcess")
	}
	p.NotifyPartChanged(p)
}

// Validate は条件パーツの妥当性を検証します
func (p *ConditionPart) Validate() error {
	if p.ComparisonOperator == value.ComparisonOperatorUnspecified {
		return errors.New("comparison operator must be specified")
	}

	// 比較演算子がBetweenの場合、MinValueとMaxValueが必要
	if p.ComparisonOperator == value.ComparisonOperatorBetween {
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

// isNotTransitionError はエラーがfsm.NoTransitionErrorかどうかを判定します
func isNotTransitionError(err error) bool {
	var noTransitionError fsm.NoTransitionError
	return errors.As(err, &noTransitionError)
}

func (p *ConditionPart) Process(ctx context.Context, increment int64) error {
	currentState := p.fsm.Current()
	p.log.Debug("Part Process",
		zap.Int64("id", int64(p.ID)),
		zap.String("current_state", currentState),
		zap.Int64("increment", increment))

	// すでに状態が満たされていたらスキップ
	if p.fsm.Current() == value.StateSatisfied {
		p.log.Debug("Part Process: State changed to satisfied during evaluation, skipping process event")
		return nil
	}

	// 状態遷移を先にしてからStrategyを実行(じゃないと、OnUpdatedでの通知で状態変更がUIに反映されない)
	err := p.fsm.Event(ctx, value.EventProcess)
	if err != nil && !isNotTransitionError(err) {
		// NoTransitionError以外のエラーの場合のみエラーとして扱う
		p.log.Error("Failed to transition state", zap.Error(err))
		return err
	} else {
		p.log.Debug("State transition successful", zap.String("new_state", p.fsm.Current()))
	}

	// 複数人から呼ばれる部分なのでmutex
	if p.strategy != nil {
		if err := p.strategy.Evaluate(ctx, p, increment); err != nil {
			p.log.Error("failed to evaluate strategy", zap.Error(err))
			// エラーが発生した場合は状態遷移を行わない
			return err
		}
	}

	// Evaluate後、状態が満たされている可能性があるので再度確認
	if p.fsm.Current() == value.StateSatisfied {
		p.log.Debug("Part Process: State changed to satisfied during evaluation, skipping process event")
		return nil
	}

	p.log.Debug("Part Process: Calling FSM Event",
		zap.String("event", value.EventProcess),
		zap.String("current_state", p.fsm.Current()))

	return nil
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

	// 戦略をリセット
	if p.strategy != nil {
		// 現在の戦略をクリーンアップ
		if err := p.strategy.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup strategy: %w", err)
		}

		// 戦略を再初期化
		if err := p.strategy.Initialize(p); err != nil {
			return fmt.Errorf("failed to reinitialize strategy: %w", err)
		}

		p.log.Debug("ConditionPart.Reset: Strategy reinitialized",
			zap.Int64("id", int64(p.ID)),
			zap.String("label", p.Label))
	}

	return p.fsm.Event(ctx, value.EventReset)
}

func (p *ConditionPart) SetStrategy(strategy service.PartStrategy) error {
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

// AddConditionPartObserver オブザーバーを追加
func (p *ConditionPart) AddConditionPartObserver(observer service.ConditionPartObserver) {
	if observer == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.partObservers = append(p.partObservers, observer)
}

// RemoveConditionPartObserver オブザーバーを削除
func (p *ConditionPart) RemoveConditionPartObserver(observer service.ConditionPartObserver) {
	if observer == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, obs := range p.partObservers {
		if obs == observer {
			p.partObservers = append(p.partObservers[:i], p.partObservers[i+1:]...)
			break
		}
	}
}

// NotifyPartChanged 状態変更を通知
func (p *ConditionPart) NotifyPartChanged(part interface{}) {
	p.mu.RLock()
	observers := make([]service.ConditionPartObserver, len(p.partObservers))
	copy(observers, p.partObservers)
	p.mu.RUnlock()

	for _, observer := range observers {
		observer.OnConditionPartChanged(p)
	}
}
