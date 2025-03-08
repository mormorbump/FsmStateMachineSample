package service

// PhaseObserver 状態を監視するインターフェース
type PhaseObserver interface {
	OnPhaseChanged(phase interface{})
}

// ConditionPartObserver 条件パーツの状態変化を監視するインターフェース
type ConditionPartObserver interface {
	OnConditionPartChanged(part interface{})
}

// ConditionObserver 条件の状態変化を監視するインターフェース
type ConditionObserver interface {
	OnConditionChanged(condition interface{})
}

type ControllerObserver interface {
	OnEntityChanged(entity interface{})
}

// PhaseSubject 状態変更を通知するインターフェース
type PhaseSubject interface {
	AddObserver(observer PhaseObserver)
	RemoveObserver(observer PhaseObserver)
	NotifyPhaseChanged()
}

// ConditionSubject 条件の変更を通知するインターフェース
type ConditionSubject interface {
	AddConditionObserver(observer ConditionObserver)
	RemoveConditionObserver(observer ConditionObserver)
	NotifyConditionChanged()
}

// ConditionPartSubject 条件パーツの変更を通知するインターフェース
type ConditionPartSubject interface {
	AddConditionPartObserver(observer ConditionPartObserver)
	RemoveConditionPartObserver(observer ConditionPartObserver)
	NotifyPartChanged()
}

type ControllerSubject interface {
	AddControllerObserver(observer ControllerObserver)
	RemoveControllerObserver(observer ControllerObserver)
	NotifyEntityChanged(entity interface{})
}
