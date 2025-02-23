package entity

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

// Phase はゲームの各フェーズを表す構造体です
type Phase struct {
	Type                   string
	Interval               time.Duration
	Order                  int
	isActive               bool
	fsm                    *fsm.FSM
	timer                  *core.IntervalTimer
	*core.StateSubjectImpl // Subject実装
	mu                     sync.RWMutex
	log                    *zap.Logger
}

// NewPhase は新しいPhaseインスタンスを作成します
func NewPhase(phaseType string, interval time.Duration, order int) *Phase {
	log := logger.DefaultLogger()
	p := &Phase{
		Type:             phaseType,
		Interval:         interval,
		isActive:         false,
		Order:            order,
		timer:            core.NewIntervalTimer(interval),
		StateSubjectImpl: core.NewStateSubjectImpl(),
		log:              log,
	}

	callbacks := fsm.Callbacks{
		"enter_" + core.StateActive: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.isActive = true
			p.timer.UpdateInterval(p.Interval)
			p.timer.Start() // タイマーを開始
			p.log.Debug("Phase transition", zap.String("from", e.Src), zap.String("to", core.StateActive))
		},
		"enter_" + core.StateNext: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.timer.Stop() // タイマーを停止
			p.log.Debug("Phase transition", zap.String("from", e.Src), zap.String("to", core.StateNext))
		},
		"enter_" + core.StateFinish: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.isActive = false
			p.timer.Stop() // タイマーを停止
			p.log.Debug("Phase transition", zap.String("from", e.Src), zap.String("to", core.StateFinish))
		},
		"after_event": func(ctx context.Context, e *fsm.Event) {
			p.log.Debug("Phase state changed", zap.String("state", p.CurrentState()))
			//if p.isActive {
			p.NotifyStateChanged(p.CurrentState())
			//}
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

	// タイマーの監視を開始
	p.timer.AddObserver(p)

	return p
}

func (p *Phase) OnTimeTicked() {
	_ = p.Next(context.Background())
}

// CurrentState は現在の状態を返します
func (p *Phase) CurrentState() string {
	return p.fsm.Current()
}

// GetInterval はインターバルを返します
func (p *Phase) GetInterval() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Interval
}

// GetOrder は順序を返します
func (p *Phase) GetOrder() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Order
}

// GetStateInfo は現在の状態の情報を返します
func (p *Phase) GetStateInfo() *core.GameStateInfo {
	return core.GetGameStateInfo(p.CurrentState())
}

// Activate はフェーズをアクティブ状態に遷移させます
func (p *Phase) Activate(ctx context.Context) error {
	return p.fsm.Event(ctx, core.EventActivate)
}

// Next は次の状態に遷移させます
func (p *Phase) Next(ctx context.Context) error {
	return p.fsm.Event(ctx, core.EventNext)
}

// Finish はフェーズを完了状態に遷移させます
func (p *Phase) Finish(ctx context.Context) error {
	return p.fsm.Event(ctx, core.EventFinish)
}

// Reset はフェーズを初期状態にリセットします
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
	for i, phase := range p {
		if phase == current {
			nextIndex := (i + 1) % len(p)
			nextPhase := p[nextIndex]
			log.Debug("Phase action", zap.String("type", nextPhase.Type), zap.String("action", "Activating next phase"))
			return nextPhase, nextPhase.Activate(ctx)
		}
	}

	return nil, fmt.Errorf("no phases available")
}
