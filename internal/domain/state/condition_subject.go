package state

import (
	"state_sample/internal/domain/core"
	logger "state_sample/internal/lib"
	"sync"

	"go.uber.org/zap"
)

type ConditionSubject interface {
	AddConditionObserver(observer core.ConditionObserver)
	RemoveConditionObserver(observer core.ConditionObserver)
	NotifyConditionSatisfied(conditionID core.ConditionID)
}

type ConditionPartSubject interface {
	AddConditionPartObserver(observer core.ConditionPartObserver)
	RemoveConditionPartObserver(observer core.ConditionPartObserver)
	NotifyPartSatisfied(partID core.ConditionPartID)
}

type ConditionSubjectImpl struct {
	observers []core.ConditionObserver
	mu        sync.RWMutex
}

func NewConditionSubjectImpl() *ConditionSubjectImpl {
	return &ConditionSubjectImpl{
		observers: make([]core.ConditionObserver, 0),
	}
}

func (s *ConditionSubjectImpl) AddConditionObserver(observer core.ConditionObserver) {
	if observer == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *ConditionSubjectImpl) RemoveConditionObserver(observer core.ConditionObserver) {
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

func (s *ConditionSubjectImpl) NotifyConditionSatisfied(conditionID core.ConditionID) {
	log := logger.DefaultLogger()
	log.Debug("ConditionSubjectImpl.NotifyStateChanged", zap.Int("conditionID", int(conditionID)))
	s.mu.RLock()
	observers := make([]core.ConditionObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()
	for _, observer := range observers {
		observer.OnConditionSatisfied(conditionID)
	}
}

type ConditionPartSubjectImpl struct {
	observers []core.ConditionPartObserver
	mu        sync.RWMutex
}

func NewConditionPartSubjectImpl() *ConditionPartSubjectImpl {
	return &ConditionPartSubjectImpl{
		observers: make([]core.ConditionPartObserver, 0),
	}
}

func (s *ConditionPartSubjectImpl) AddConditionPartObserver(observer core.ConditionPartObserver) {
	if observer == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *ConditionPartSubjectImpl) RemoveConditionPartObserver(observer core.ConditionPartObserver) {
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

func (s *ConditionPartSubjectImpl) NotifyPartSatisfied(partID core.ConditionPartID) {
	log := logger.DefaultLogger()
	log.Debug("PartSubjectImpl.NotifyStateChanged", zap.Int("partID", int(partID)))
	s.mu.RLock()
	observers := make([]core.ConditionPartObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()
	for _, observer := range observers {
		observer.OnPartSatisfied(partID)
	}
}
