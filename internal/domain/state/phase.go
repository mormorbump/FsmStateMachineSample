package state

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
	logger "state_sample/internal/lib"
	"sync"
	"time"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

type PhaseID int
type GameRule int

const (
	GameRule_Shooting GameRule = iota
	GameRule_PushSwitch
	GameRule_Animation
)

// Phase はゲームの各フェーズを表す構造体です
type Phase struct {
	ID                     PhaseID
	Order                  int
	isActive               bool
	IsClear                bool
	Name                   string
	Description            string
	Rule                   GameRule
	ConditionType          ConditionType
	ConditionIDs           []core.ConditionID
	SatisfiedConditions    map[core.ConditionID]bool
	Conditions             map[core.ConditionID]*Condition
	StartTime              *time.Time
	FinishTime             *time.Time
	fsm                    *fsm.FSM
	*core.StateSubjectImpl // Subject実装
	mu                     sync.RWMutex
	log                    *zap.Logger
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
func NewPhase(name string, order int, conditions []*Condition, conditionType ConditionType, rule GameRule) *Phase {
	log := logger.DefaultLogger()

	p := &Phase{
		Name:                name,
		isActive:            false,
		Order:               order,
		StateSubjectImpl:    core.NewStateSubjectImpl(),
		Rule:                rule,
		ConditionType:       conditionType,
		SatisfiedConditions: make(map[core.ConditionID]bool),
		Conditions:          make(map[core.ConditionID]*Condition),
		ConditionIDs:        make([]core.ConditionID, 0),
		IsClear:             false,
		StartTime:           nil,
		FinishTime:          nil,
		log:                 log,
	}

	for _, cond := range conditions {
		p.ConditionIDs = append(p.ConditionIDs, cond.ID)
		p.Conditions[cond.ID] = cond
	}

	callbacks := fsm.Callbacks{
		"enter_" + core.StateActive: func(ctx context.Context, e *fsm.Event) {
			p.isActive = true
			now := time.Now()
			p.StartTime = &now
			for _, c := range p.Conditions {
				p.log.Debug("Phase enter_active: Activating condition", zap.Any("condition", c))
				if err := c.Activate(ctx); err != nil {
					p.log.Error("Failed to activate condition",
						zap.Error(err),
						zap.Int64("condition_id", int64(c.ID)))
				}
			}
		},
		"enter_" + core.StateNext: func(ctx context.Context, e *fsm.Event) {},
		"enter_" + core.StateFinish: func(ctx context.Context, e *fsm.Event) {
			p.isActive = false
			now := time.Now()
			p.FinishTime = &now
		},
		"enter_" + core.StateReady: func(ctx context.Context, e *fsm.Event) {
			p.isActive = false
			p.StartTime = nil
			p.FinishTime = nil
			p.SatisfiedConditions = make(map[core.ConditionID]bool)
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Phase transition", zap.String("Name", p.Name), zap.String("from", e.Src), zap.String("to", e.Dst))
			p.log.Debug("Phase state changed", zap.String("Name", p.Name), zap.String("state", p.CurrentState()))
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
func (p *Phase) OnConditionSatisfied(conditionID core.ConditionID) {
	p.log.Debug("Phase.OnConditionSatisfied")
	p.mu.Lock()
	p.SatisfiedConditions[conditionID] = true
	satisfied := p.checkConditionsSatisfied()
	if satisfied {
		p.IsClear = true
	}
	p.mu.Unlock()

	p.log.Debug("Phase.OnConditionSatisfied", zap.String("name", p.Name), zap.Bool("satisfied", satisfied), zap.Int64("condition_id", int64(conditionID)))
	if satisfied {
		err := p.Next(context.Background())
		if err != nil {
			p.log.Error("Failed to move to next state", zap.Error(err))
		}
	}
}

// checkConditionsSatisfied は条件が満たされているかチェックします
func (p *Phase) checkConditionsSatisfied() bool {
	if len(p.ConditionIDs) == 0 {
		return false
	}

	switch p.ConditionType {
	case ConditionTypeOr, ConditionTypeSingle:
		return len(p.SatisfiedConditions) > 0
	case ConditionTypeAnd:
		return len(p.SatisfiedConditions) == len(p.ConditionIDs)
	default:
		return false
	}
}

func (p *Phase) CurrentState() string {
	return p.fsm.Current()
}

func (p *Phase) GetStateInfo() *core.GameStateInfo {
	return core.GetGameStateInfo(p.CurrentState())
}

func (p *Phase) GetConditions() map[core.ConditionID]*Condition {
	return p.Conditions
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

	p.log.Debug("Phase.Reset: Resetting time information",
		zap.String("start_time", startTimeStr),
		zap.String("finish_time", finishTimeStr),
		zap.String("phase_name", p.Name),
		zap.Int("phase_order", p.Order))

	// 時間情報をリセット
	p.StartTime = nil
	p.FinishTime = nil

	// 条件とパーツをリセット
	for _, cond := range p.Conditions {
		if err := cond.Reset(ctx); err != nil {
			return fmt.Errorf("failed to reset condition: %w", err)
		}
	}

	return p.fsm.Event(ctx, core.EventReset)
}

// Phases はフェーズのコレクションを表す型です
type Phases []*Phase

func (p Phases) Current() *Phase {
	log := logger.DefaultLogger()
	log.Debug("Phases.Current")
	for _, phase := range p {
		if phase.isActive {
			log.Debug("Phases.Current", zap.String("name", phase.Name))
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

// ProcessAndActivateByNextOrder 次のフェーズに移行
func (p Phases) ProcessAndActivateByNextOrder(ctx context.Context) (*Phase, error) {
	log := logger.DefaultLogger()
	current := p.Current()

	if current == nil {
		if len(p) <= 0 {
			log.Error("Phases.ProcessAndActivateByNextOrder", zap.Error(fmt.Errorf("no phases available")))
			return nil, fmt.Errorf("no phases available")
		}

		log.Debug("Phases.ProcessAndActivateByNextOrder",
			zap.String("action", "Starting first phase"),
			zap.String("name", p[0].Name),
		)
		return p[0], p[0].Activate(ctx)
	}

	if err := current.Finish(ctx); err != nil {
		log.Error(current.Name, zap.Error(err))
		return nil, err
	}

	for _, phase := range p {
		log.Debug("Phases.ProcessAndActivateByNextOrder searching...")
		if current.Order+1 == phase.Order {
			log.Debug("Phases.ProcessAndActivateByNextOrder", zap.String("name", phase.Name), zap.String("action", "Activating next phase"))
			return phase, phase.Activate(ctx)
		}
	}

	return nil, nil
}
