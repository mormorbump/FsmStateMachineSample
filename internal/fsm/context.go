package fsm

import (
	"context"
	"sync"

	"github.com/looplab/fsm"
)

// FSMContext は状態遷移を管理するコンテキストです
type FSMContext struct {
	fsm       *fsm.FSM
	observers []StateObserver
	mu        sync.RWMutex
}

// NewFSMContext は新しいFSMContextインスタンスを作成します
func NewFSMContext() *FSMContext {
	f := &FSMContext{
		observers: make([]StateObserver, 0),
	}

	// FSMの初期化
	f.fsm = fsm.NewFSM(
		StateReady, // 初期状態
		fsm.Events{
			// StateReadyとStateNextからStateActiveへの遷移を許可
			fsm.EventDesc{Name: EventActivate, Src: []string{StateReady, StateNext}, Dst: StateActive},
			fsm.EventDesc{Name: EventNext, Src: []string{StateActive}, Dst: StateNext},
			fsm.EventDesc{Name: EventFinish, Src: []string{StateNext}, Dst: StateFinish},
			fsm.EventDesc{Name: EventReset, Src: []string{StateFinish}, Dst: StateReady}, // リセットイベントを追加
		},
		fsm.Callbacks{
			"before_event": func(ctx context.Context, e *fsm.Event) {
				if !IsValidTransition(e.Src, e.Event) {
					e.Cancel(ErrInvalidTransition)
				}
			},
			"enter_state": func(ctx context.Context, e *fsm.Event) {
				f.notifyStateChanged(e.Dst)
			},
			"after_event": func(ctx context.Context, e *fsm.Event) {
				if e.Err != nil {
					f.notifyError(e.Err)
				}
			},
		},
	)

	return f
}

// CurrentState は現在の状態を返します
func (f *FSMContext) CurrentState() string {
	return f.fsm.Current()
}

// Transition は指定されたイベントによる状態遷移を実行します
func (f *FSMContext) Transition(ctx context.Context, event string) error {
	err := f.fsm.Event(ctx, event)
	if err != nil {
		// エラーの種類を判定して適切なエラーを返す
		switch err.(type) {
		case fsm.InvalidEventError:
			return &StateError{
				Code:    "INVALID_EVENT",
				Message: err.Error(),
			}
		case fsm.CanceledError:
			if stateErr, ok := err.(*StateError); ok {
				return stateErr
			}
			return &StateError{
				Code:    "CANCELED",
				Message: err.Error(),
			}
		default:
			return &StateError{
				Code:    "TRANSITION_ERROR",
				Message: err.Error(),
			}
		}
	}
	return nil
}

// Reset は状態をリセットします
func (f *FSMContext) Reset(ctx context.Context) error {
	if f.CurrentState() != StateFinish {
		return &StateError{
			Code:    "INVALID_RESET",
			Message: "リセットは終了状態からのみ可能です",
		}
	}
	return f.Transition(ctx, EventReset)
}

// AddObserver はオブザーバーを追加します
func (f *FSMContext) AddObserver(observer StateObserver) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.observers = append(f.observers, observer)
}

// RemoveObserver はオブザーバーを削除します
func (f *FSMContext) RemoveObserver(observer StateObserver) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, obs := range f.observers {
		if obs == observer {
			f.observers = append(f.observers[:i], f.observers[i+1:]...)
			break
		}
	}
}

// notifyStateChanged は状態変更をオブザーバーに通知します
func (f *FSMContext) notifyStateChanged(newState string) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, observer := range f.observers {
		observer.OnStateChanged(newState)
	}
}

// notifyError はエラーをオブザーバーに通知します
func (f *FSMContext) notifyError(err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, observer := range f.observers {
		observer.OnError(err)
	}
}

// GetCurrentStateInfo は現在の状態の情報を返します
func (f *FSMContext) GetCurrentStateInfo() *StateInfo {
	return GetStateInfo(f.CurrentState())
}

// CanTransition は指定されたイベントによる遷移が可能かどうかを返します
func (f *FSMContext) CanTransition(event string) bool {
	return IsValidTransition(f.CurrentState(), event)
}
