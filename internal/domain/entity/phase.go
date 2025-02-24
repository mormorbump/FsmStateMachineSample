package entity

import (
	"context"
	"fmt"
	"state_sample/internal/domain/condition"
	"state_sample/internal/domain/core"
	logger "state_sample/internal/lib"
	"sync"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// Phase はゲームの各フェーズを表す構造体です
type Phase struct {
	Type  string
	Order int

	isActive               bool
	fsm                    *fsm.FSM
	isClear                bool
	*core.StateSubjectImpl // Subject実装
	mu                     sync.RWMutex
	log                    *zap.Logger

	// 条件システム
	conditionType       ConditionType
	conditionIDs        []int64
	satisfiedConditions map[condition.ConditionID]bool
	conditions          map[condition.ConditionID]*Condition
}

// ConditionType は条件の組み合わせ方を表す型です
type ConditionType int

const (
	ConditionTypeUnspecified ConditionType = iota
	ConditionTypeAnd                       // すべての条件を満たす必要がある
	ConditionTypeOr                        // いずれかの条件を満たせばよい
	ConditionTypeSingle                    // 単一条件
)

// NewPhase は新しいPhaseインスタンスを作成します
func NewPhase(phaseType string, order int, cond *Condition) *Phase {
	log := logger.DefaultLogger()

	p := &Phase{
		Type:                phaseType,
		isActive:            false,
		Order:               order,
		StateSubjectImpl:    core.NewStateSubjectImpl(),
		conditionType:       ConditionTypeSingle,
		conditionIDs:        []int64{int64(order)},
		satisfiedConditions: make(map[condition.ConditionID]bool),
		conditions:          make(map[condition.ConditionID]*Condition),
		isClear:             false,
		log:                 log,
	}

	// 条件の設定
	p.conditions[cond.ID] = cond

	callbacks := fsm.Callbacks{
		"enter_" + core.StateActive: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			p.isActive = true
			p.mu.Unlock()

			// 条件のアクティベート
			for _, cond := range p.conditions {
				if err := cond.Activate(ctx); err != nil {
					p.log.Error("Failed to activate condition",
						zap.Error(err),
						zap.Int64("condition_id", int64(cond.ID)))
				}
			}
		},
		"enter_" + core.StateNext: func(ctx context.Context, e *fsm.Event) {},
		"enter_" + core.StateFinish: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			p.isActive = false
			p.mu.Unlock()
		},
		"enter_" + core.StateReady: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			p.isActive = false
			p.satisfiedConditions = make(map[condition.ConditionID]bool)
			p.mu.Unlock()
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Phase transition", zap.String("Type", p.Type), zap.String("from", e.Src), zap.String("to", e.Dst))
			p.log.Debug("Phase state changed", zap.String("Type", p.Type), zap.String("state", p.CurrentState()))
			if e.Dst != core.StateFinish {
				p.NotifyStateChanged(p.CurrentState())
			}
		},
	}

	p.fsm = fsm.NewFSM(
		core.StateReady,
		fsm.Events{
			{Name: core.EventActivate, Src: []string{core.StateReady}, Dst: core.StateActive},
			{Name: core.EventNext, Src: []string{core.StateActive}, Dst: core.StateNext},
			{Name: core.EventFinish, Src: []string{core.StateNext}, Dst: core.StateFinish},
			{Name: core.EventReset, Src: []string{core.StateActive, core.StateNext, core.StateFinish}, Dst: core.StateReady},
		},
		callbacks,
	)

	return p
}

// OnConditionSatisfied は条件が満たされた時に呼び出されます
func (p *Phase) OnConditionSatisfied(conditionID condition.ConditionID) {
	p.mu.Lock()
	p.satisfiedConditions[conditionID] = true
	satisfied := p.checkConditionsSatisfied()
	if satisfied {
		p.isClear = true
	}
	p.mu.Unlock()

	p.log.Debug("Phase", zap.String("type", p.Type), zap.Bool("satisfied", satisfied), zap.Int64("condition_id", int64(conditionID)))
	if satisfied {
		err := p.Next(context.Background())
		if err != nil {
			p.log.Error("Failed to move to next state", zap.Error(err))
		}
	}
}

// checkConditionsSatisfied は条件が満たされているかチェックします
func (p *Phase) checkConditionsSatisfied() bool {
	if len(p.conditionIDs) == 0 {
		return false
	}

	switch p.conditionType {
	case ConditionTypeOr, ConditionTypeSingle:
		return len(p.satisfiedConditions) > 0
	case ConditionTypeAnd:
		return len(p.satisfiedConditions) == len(p.conditionIDs)
	default:
		return false
	}
}

// SetConditions は条件を設定します
func (p *Phase) SetConditions(conditionType ConditionType, conditionIDs []int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.conditionType = conditionType
	p.conditionIDs = conditionIDs
	p.satisfiedConditions = make(map[condition.ConditionID]bool)
}

func (p *Phase) CurrentState() string {
	return p.fsm.Current()
}

func (p *Phase) GetStateInfo() *core.GameStateInfo {
	return core.GetGameStateInfo(p.CurrentState())
}

func (p *Phase) Activate(ctx context.Context) error {
	return p.fsm.Event(ctx, core.EventActivate)
}

func (p *Phase) Next(ctx context.Context) error {
	return p.fsm.Event(ctx, core.EventNext)
}

func (p *Phase) Finish(ctx context.Context) error {
	return p.fsm.Event(ctx, core.EventFinish)
}

func (p *Phase) Reset(ctx context.Context) error {
	if p.CurrentState() == core.StateReady {
		return nil
	}

	// 条件とパーツをリセット
	for _, cond := range p.conditions {
		if err := cond.Reset(ctx); err != nil {
			return fmt.Errorf("failed to reset condition: %w", err)
		}
	}

	return p.fsm.Event(ctx, core.EventReset)
}

// Phases はフェーズのコレクションを表す型です
type Phases []*Phase

func (p Phases) Current() *Phase {
	for _, phase := range p {
		if phase.isActive {
			return phase
		}
	}
	return nil
}

func (p Phases) ResetAll(ctx context.Context) error {
	for _, phase := range p {
		if err := phase.Reset(ctx); err != nil {
			return err
		}
	}
	return nil
}

// IsClear はフェーズがクリアされているかどうかを返します
func (p *Phase) IsClear() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isClear
}

// ProcessAndActivateByNextOrder 次のフェーズに移行
func (p Phases) ProcessAndActivateByNextOrder(ctx context.Context) (*Phase, error) {
	log := logger.DefaultLogger()
	current := p.Current()

	if current == nil {
		if len(p) <= 0 {
			log.Error("Phases", zap.Error(fmt.Errorf("no phases available")))
			return nil, fmt.Errorf("no phases available")
		}

		log.Debug("Phases", zap.String("action", "Starting first phase"))
		return p[0], p[0].Activate(ctx)
	}

	if err := current.Finish(ctx); err != nil {
		log.Error(current.Type, zap.Error(err))
		return nil, err
	}

	for _, phase := range p {
		if current.Order+1 == phase.Order {
			log.Debug("Phase action", zap.String("type", phase.Type), zap.String("action", "Activating next phase"))
			return phase, phase.Activate(ctx)
		}
	}

	return nil, nil
}
