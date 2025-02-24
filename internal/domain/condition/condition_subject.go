package condition

import (
	"sync"
)

// ConditionSubject は条件の状態変化を通知するインターフェースです
type ConditionSubject interface {
	AddConditionObserver(observer ConditionObserver)
	RemoveConditionObserver(observer ConditionObserver)
	NotifyConditionSatisfied(conditionID ConditionID)
}

// ConditionPartSubject は条件パーツの状態変化を通知するインターフェースです
type ConditionPartSubject interface {
	AddConditionPartObserver(observer ConditionPartObserver)
	RemoveConditionPartObserver(observer ConditionPartObserver)
	NotifyPartSatisfied(partID ConditionPartID)
}

// ConditionSubjectImpl は条件の状態変化を通知する実装です
type ConditionSubjectImpl struct {
	observers []ConditionObserver
	mu        sync.RWMutex
}

// NewConditionSubjectImpl は新しいConditionSubjectImplを作成します
func NewConditionSubjectImpl() *ConditionSubjectImpl {
	return &ConditionSubjectImpl{
		observers: make([]ConditionObserver, 0),
	}
}

func (s *ConditionSubjectImpl) AddConditionObserver(observer ConditionObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *ConditionSubjectImpl) RemoveConditionObserver(observer ConditionObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

func (s *ConditionSubjectImpl) NotifyConditionSatisfied(conditionID ConditionID) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, observer := range s.observers {
		observer.OnConditionSatisfied(conditionID)
	}
}

// ConditionPartSubjectImpl は条件パーツの状態変化を通知する実装です
type ConditionPartSubjectImpl struct {
	observers []ConditionPartObserver
	mu        sync.RWMutex
}

// NewConditionPartSubjectImpl は新しいConditionPartSubjectImplを作成します
func NewConditionPartSubjectImpl() *ConditionPartSubjectImpl {
	return &ConditionPartSubjectImpl{
		observers: make([]ConditionPartObserver, 0),
	}
}

func (s *ConditionPartSubjectImpl) AddConditionPartObserver(observer ConditionPartObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *ConditionPartSubjectImpl) RemoveConditionPartObserver(observer ConditionPartObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

func (s *ConditionPartSubjectImpl) NotifyPartSatisfied(partID ConditionPartID) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, observer := range s.observers {
		observer.OnPartSatisfied(partID)
	}
}
