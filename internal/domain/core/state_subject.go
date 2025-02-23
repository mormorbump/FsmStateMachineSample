package core

import (
	logger "state_sample/internal/lib"
	"sync"

	"go.uber.org/zap"
)

// StateSubject 状態遷移する対象のインターフェース
type StateSubject interface {
	AddObserver(observer StateObserver)
	RemoveObserver(observer StateObserver)
	NotifyStateChanged(state string)
}

type StateSubjectImpl struct {
	observers []StateObserver
	mu        sync.RWMutex
}

func NewStateSubjectImpl() *StateSubjectImpl {
	return &StateSubjectImpl{
		observers: make([]StateObserver, 0),
	}
}

func (s *StateSubjectImpl) AddObserver(observer StateObserver) {
	if observer == nil {
		return // nilオブザーバーは登録しない
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *StateSubjectImpl) RemoveObserver(observer StateObserver) {
	if observer == nil {
		return // nilオブザーバーは無視
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			return
		}
	}
}

func (s *StateSubjectImpl) NotifyStateChanged(state string) {
	log := logger.DefaultLogger()
	log.Debug("StateSubjectImpl.NotifyStateChanged", zap.String("state", state))
	s.mu.RLock()
	observers := make([]StateObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()

	for _, observer := range observers {
		observer.OnStateChanged(state)
	}
}
