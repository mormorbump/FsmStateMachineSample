package core

import (
	logger "state_sample/internal/lib"
	"sync"

	"go.uber.org/zap"
)

type ConditionSubject interface {
	AddConditionObserver(observer ConditionObserver)
	RemoveConditionObserver(observer ConditionObserver)
	NotifyConditionSatisfied(conditionID ConditionID)
}

type ConditionPartSubject interface {
	AddConditionPartObserver(observer ConditionPartObserver)
	RemoveConditionPartObserver(observer ConditionPartObserver)
	NotifyPartSatisfied(partID ConditionPartID)
}

type ConditionSubjectImpl struct {
	observers []ConditionObserver
	mu        sync.RWMutex
}

func NewConditionSubjectImpl() *ConditionSubjectImpl {
	return &ConditionSubjectImpl{
		observers: make([]ConditionObserver, 0),
	}
}

func (s *ConditionSubjectImpl) AddConditionObserver(observer ConditionObserver) {
	if observer == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *ConditionSubjectImpl) RemoveConditionObserver(observer ConditionObserver) {
	if observer == nil {
		return
	}
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
	log := logger.DefaultLogger()
	log.Debug("ConditionSubjectImpl.NotifyStateChanged", zap.Int("conditionID", int(conditionID)))
	s.mu.RLock()
	observers := make([]ConditionObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()
	for _, observer := range observers {
		observer.OnConditionSatisfied(conditionID)
	}
}

type ConditionPartSubjectImpl struct {
	observers []ConditionPartObserver
	mu        sync.RWMutex
}

func NewConditionPartSubjectImpl() *ConditionPartSubjectImpl {
	return &ConditionPartSubjectImpl{
		observers: make([]ConditionPartObserver, 0),
	}
}

func (s *ConditionPartSubjectImpl) AddConditionPartObserver(observer ConditionPartObserver) {
	if observer == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *ConditionPartSubjectImpl) RemoveConditionPartObserver(observer ConditionPartObserver) {
	if observer == nil {
		return
	}
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
	log := logger.DefaultLogger()
	log.Debug("PartSubjectImpl.NotifyStateChanged", zap.Int("partID", int(partID)))
	s.mu.RLock()
	observers := make([]ConditionPartObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()
	for _, observer := range observers {
		observer.OnPartSatisfied(partID)
	}
}
