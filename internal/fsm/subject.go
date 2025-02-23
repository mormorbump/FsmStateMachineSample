package fsm

import (
	"sync"
)

// Subject は観察対象のインターフェースを定義します
type Subject interface {
	AddObserver(observer StateObserver)
	RemoveObserver(observer StateObserver)
	NotifyStateChanged()
}

// StateSubjectImpl は Subject インターフェースの共通実装を提供します
type StateSubjectImpl struct {
	observers []StateObserver
	mu        sync.RWMutex
}

// NewStateSubjectImpl は新しい StateSubjectImpl インスタンスを作成します
func NewStateSubjectImpl() *StateSubjectImpl {
	return &StateSubjectImpl{
		observers: make([]StateObserver, 0),
	}
}

// AddObserver は監視者を追加します
func (s *StateSubjectImpl) AddObserver(observer StateObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

// RemoveObserver は監視者を削除します
func (s *StateSubjectImpl) RemoveObserver(observer StateObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// NotifyStateChanged は状態変更を通知します
func (s *StateSubjectImpl) NotifyStateChanged(state string) {
	s.mu.RLock()
	observers := make([]StateObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()

	for _, observer := range observers {
		observer.OnStateChanged(state)
	}
}

// NotifyError はエラーを通知します
func (s *StateSubjectImpl) NotifyError(err error) {
	s.mu.RLock()
	observers := make([]StateObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()

	for _, observer := range observers {
		observer.OnError(err)
	}
}
