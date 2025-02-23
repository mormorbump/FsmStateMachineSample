package core

import (
	"sync"
)

// StateSubject 状態遷移する対象のインターフェース
type StateSubject interface {
	AddObserver(observer StateObserver)
	RemoveObserver(observer StateObserver)
	NotifyStateChanged()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

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

func (s *StateSubjectImpl) NotifyStateChanged(state string) {
	s.mu.RLock()
	observers := make([]StateObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()

	for _, observer := range observers {
		observer.OnStateChanged(state)
	}
}
