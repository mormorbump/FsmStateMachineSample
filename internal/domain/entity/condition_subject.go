package entity

// ConditionSubjectImpl 条件の状態変化を通知する機能を提供
type ConditionSubjectImpl struct {
	conditionObservers     []ConditionObserver
	conditionPartObservers []ConditionPartObserver
}

func NewConditionSubjectImpl() *ConditionSubjectImpl {
	return &ConditionSubjectImpl{
		conditionObservers:     make([]ConditionObserver, 0),
		conditionPartObservers: make([]ConditionPartObserver, 0),
	}
}

func (s *ConditionSubjectImpl) AddConditionObserver(observer ConditionObserver) {
	s.conditionObservers = append(s.conditionObservers, observer)
}

func (s *ConditionSubjectImpl) RemoveConditionObserver(observer ConditionObserver) {
	for i, o := range s.conditionObservers {
		if o == observer {
			s.conditionObservers = append(s.conditionObservers[:i], s.conditionObservers[i+1:]...)
			break
		}
	}
}

// NotifyConditionSatisfied 条件が満たされたことを通知
func (s *ConditionSubjectImpl) NotifyConditionSatisfied(conditionID ConditionID) {
	for _, observer := range s.conditionObservers {
		observer.OnConditionSatisfied(conditionID)
	}
}

func (s *ConditionSubjectImpl) AddConditionPartObserver(observer ConditionPartObserver) {
	s.conditionPartObservers = append(s.conditionPartObservers, observer)
}

func (s *ConditionSubjectImpl) RemoveConditionPartObserver(observer ConditionPartObserver) {
	for i, o := range s.conditionPartObservers {
		if o == observer {
			s.conditionPartObservers = append(s.conditionPartObservers[:i], s.conditionPartObservers[i+1:]...)
			break
		}
	}
}

// NotifyPartSatisfied 条件パーツが満たされたことを通知
func (s *ConditionSubjectImpl) NotifyPartSatisfied(partID ConditionPartID) {
	for _, observer := range s.conditionPartObservers {
		observer.OnPartSatisfied(partID)
	}
}
