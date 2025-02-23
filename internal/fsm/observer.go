package fsm

import (
	"sync"
)

// StateObserver は監視者のインターフェースを定義します
type StateObserver interface {
	OnStateChanged(state string)
	OnError(err error)
}

// ObserverImpl は StateObserver インターフェースの共通実装を提供します
type ObserverImpl struct {
	onStateChanged func(state string)
	onError        func(err error)
	mu             sync.RWMutex
}

// NewObserverImpl は新しい ObserverImpl インスタンスを作成します
func NewObserverImpl(onStateChanged func(state string), onError func(err error)) *ObserverImpl {
	return &ObserverImpl{
		onStateChanged: onStateChanged,
		onError:        onError,
	}
}

// OnStateChanged は状態変更を処理します
func (o *ObserverImpl) OnStateChanged(state string) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if o.onStateChanged != nil {
		o.onStateChanged(state)
	}
}

// OnError はエラーを処理します
func (o *ObserverImpl) OnError(err error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if o.onError != nil {
		o.onError(err)
	}
}

// UpdateHandlers はハンドラ関数を更新します
func (o *ObserverImpl) UpdateHandlers(onStateChanged func(state string), onError func(err error)) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.onStateChanged = onStateChanged
	o.onError = onError
}
