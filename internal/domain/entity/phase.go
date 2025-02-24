package entity

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
	logger "state_sample/internal/lib"
	"sync"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// Phase はゲームの各フェーズを表す構造体です
type Phase struct {
	Type     string
	Interval int64 // milliseconds
	Order    int

	isActive bool
	fsm      *fsm.FSM

	*core.StateSubjectImpl // Subject実装
	mu                     sync.RWMutex
	log                    *zap.Logger

	// 条件システム
	conditionType       ConditionType
	conditionIDs        []int64
	satisfiedConditions map[ConditionID]bool
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
func NewPhase(phaseType string, interval int64, order int) *Phase {
	log := logger.DefaultLogger()
	p := &Phase{
		Type:                phaseType,
		Interval:            interval,
		isActive:            false,
		Order:               order,
		StateSubjectImpl:    core.NewStateSubjectImpl(),
		log:                 log,
		conditionType:       ConditionTypeUnspecified,
		conditionIDs:        make([]int64, 0),
		satisfiedConditions: make(map[ConditionID]bool),
	}

	callbacks := fsm.Callbacks{
		"enter_" + core.StateActive: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.isActive = true
		},
		"enter_" + core.StateNext: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.isActive = false
		},
		"enter_" + core.StateFinish: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.isActive = false
		},
		"after_" + core.EventReset: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.isActive = false
			p.satisfiedConditions = make(map[ConditionID]bool)
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Phase transition", zap.String("from", e.Src), zap.String("to", e.Dst))
			p.log.Debug("Phase state changed", zap.String("state", p.CurrentState()))
			if e.Dst != core.StateFinish {
				p.NotifyStateChanged(p.CurrentState())
			}
		},
	}

	p.fsm = fsm.NewFSM(
		core.StateReady,
		fsm.Events{
			{Name: core.EventActivate, Src: []string{core.StateReady, core.StateNext}, Dst: core.StateActive},
			{Name: core.EventNext, Src: []string{core.StateActive}, Dst: core.StateNext},
			{Name: core.EventFinish, Src: []string{core.StateNext}, Dst: core.StateFinish},
			{Name: core.EventReset, Src: []string{core.StateReady, core.StateNext, core.StateFinish}, Dst: core.StateReady},
		},
		callbacks,
	)

	return p
}

// OnConditionSatisfied は条件が満たされた時に呼び出されます
func (p *Phase) OnConditionSatisfied(conditionID ConditionID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.satisfiedConditions[conditionID] = true
	if p.checkConditionsSatisfied() {
		_ = p.Next(context.Background())
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
	p.satisfiedConditions = make(map[ConditionID]bool)
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

// ProcessOrder は次のフェーズに移行します
func (p Phases) ProcessOrder(ctx context.Context) (*Phase, error) {
	log := logger.DefaultLogger()
	current := p.Current()
	if current == nil {
		if len(p) > 0 {
			log.Debug("Phases", zap.String("action", "Starting first phase"))
			return p[0], p[0].Activate(ctx)
		}
		log.Error("Phases", zap.Error(fmt.Errorf("no phases available")))
		return nil, fmt.Errorf("no phases available")
	}

	// 現在のフェーズを完了
	if err := current.Finish(ctx); err != nil {
		log.Error(current.Type, zap.Error(err))
		return nil, err
	}

	// 次のフェーズを探して活性化
	for _, phase := range p {
		if current.Order+1 == phase.Order {
			nextPhase := phase
			log.Debug("Phase action", zap.String("type", nextPhase.Type), zap.String("action", "Activating next phase"))
			return nextPhase, nextPhase.Activate(ctx)
		}
	}

	return nil, nil
}
