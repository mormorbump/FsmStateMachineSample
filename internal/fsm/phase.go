package fsm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/looplab/fsm"
)

// Phase はゲームの各フェーズを表す構造体です
type Phase struct {
	Type     string
	Interval time.Duration
	Order    int

	fsm               *fsm.FSM
	timer             *IntervalTimer
	*StateSubjectImpl // Subject実装
	*ObserverImpl     // Observer実装
	mu                sync.RWMutex
}

// NewPhase は新しいPhaseインスタンスを作成します
func NewPhase(phaseType string, interval time.Duration, order int) *Phase {
	p := &Phase{
		Type:             phaseType,
		Interval:         interval,
		Order:            order,
		timer:            NewIntervalTimer(interval),
		StateSubjectImpl: NewStateSubjectImpl(),
	}

	// ObserverImplの初期化
	p.ObserverImpl = NewObserverImpl(
		func(state string) {
			if state == "tick" {
				p.OnTimerTick()
			}
		},
		func(err error) {
			p.NotifyError(err)
		},
	)

	callbacks := fsm.Callbacks{
		"enter_" + StateActive: func(ctx context.Context, e *fsm.Event) {
			p.mu.Lock()
			defer p.mu.Unlock()
			p.timer.updateInterval(p.Interval)
			p.NotifyStateChanged(StateActive)
		},
		"enter_" + StateNext: func(ctx context.Context, e *fsm.Event) {
			p.NotifyStateChanged(StateNext)
		},
		"enter_" + StateFinish: func(ctx context.Context, e *fsm.Event) {
			p.NotifyStateChanged(StateFinish)
		},
	}

	p.fsm = fsm.NewFSM(
		StateReady,
		fsm.Events{
			{Name: EventActivate, Src: []string{StateReady, StateNext}, Dst: StateActive},
			{Name: EventNext, Src: []string{StateActive}, Dst: StateNext},
			{Name: EventFinish, Src: []string{StateNext}, Dst: StateFinish},
			{Name: EventReset, Src: []string{StateFinish}, Dst: StateReady},
		},
		callbacks,
	)

	// タイマーの監視を開始
	p.timer.AddObserver(p)

	return p
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
func (p *Phase) GetStateInfo() *GameStateInfo {
	return GetGameStateInfo(p.CurrentState())
}

// Activate はフェーズをアクティブ状態に遷移させます
func (p *Phase) Activate(ctx context.Context) error {
	return p.fsm.Event(ctx, EventActivate)
}

// Next は次の状態に遷移させます
func (p *Phase) Next(ctx context.Context) error {
	return p.fsm.Event(ctx, EventNext)
}

// Finish はフェーズを完了状態に遷移させます
func (p *Phase) Finish(ctx context.Context) error {
	return p.fsm.Event(ctx, EventFinish)
}

// Reset はフェーズを初期状態にリセットします
func (p *Phase) Reset(ctx context.Context) error {
	return p.fsm.Event(ctx, EventReset)
}

// OnTimerTick はタイマーイベントを処理します
func (p *Phase) OnTimerTick() {
	p.NotifyStateChanged(p.CurrentState())
}

// Phases はフェーズのコレクションを表す型です
type Phases []*Phase

// Current は現在アクティブなフェーズを返します
func (p Phases) Current() *Phase {
	for _, phase := range p {
		if phase.CurrentState() == StateActive {
			return phase
		}
	}
	return nil
}

// MoveNext は次のフェーズに移行します
func (p Phases) MoveNext(ctx context.Context) error {
	current := p.Current()
	if current == nil {
		if len(p) > 0 {
			return p[0].Activate(ctx)
		}
		return fmt.Errorf("no phases available")
	}

	// 現在のフェーズを完了
	if err := current.Next(ctx); err != nil {
		return err
	}

	// 次のフェーズを探して活性化
	for i, phase := range p {
		if phase == current {
			nextIndex := (i + 1) % len(p)
			return p[nextIndex].Activate(ctx)
		}
	}

	return nil
}
