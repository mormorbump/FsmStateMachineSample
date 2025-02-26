package service

// StateObserver 状態を監視するインターフェース
type StateObserver interface {
	OnStateChanged(state string)
}

// StrategyObserver 戦略を監視するインターフェース
type StrategyObserver interface {
	OnUpdated(event string)
}

// ConditionPartObserver 条件パーツの状態変化を監視するインターフェース
type ConditionPartObserver interface {
	OnConditionPartChanged(part interface{})
}

// ConditionObserver 条件の状態変化を監視するインターフェース
type ConditionObserver interface {
	OnConditionChanged(condition interface{})
}

// Subject 監視対象のインターフェース
type Subject interface {
	AddObserver(observer interface{})
	RemoveObserver(observer interface{})
	Notify(event string, data interface{})
}

// StateSubject 状態変更を通知するインターフェース
type StateSubject interface {
	AddObserver(observer StateObserver)
	RemoveObserver(observer StateObserver)
	NotifyStateChanged(state string)
}

// StrategySubject 戦略の更新を通知するインターフェース
type StrategySubject interface {
	AddObserver(observer StrategyObserver)
	RemoveObserver(observer StrategyObserver)
	NotifyUpdate(event string)
}

// ConditionSubject 条件の変更を通知するインターフェース
type ConditionSubject interface {
	AddConditionObserver(observer ConditionObserver)
	RemoveConditionObserver(observer ConditionObserver)
	NotifyConditionChanged(condition interface{})
}

// ConditionPartSubject 条件パーツの変更を通知するインターフェース
type ConditionPartSubject interface {
	AddConditionPartObserver(observer ConditionPartObserver)
	RemoveConditionPartObserver(observer ConditionPartObserver)
	NotifyPartChanged(part interface{})
}
