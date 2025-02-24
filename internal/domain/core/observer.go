package core

// StateObserver 状態を監視するインターフェース
type StateObserver interface {
	OnStateChanged(state string)
}

// TimeObserver 時間を監視するインターフェース
type TimeObserver interface {
	OnTimeTicked()
}

// ConditionPartObserver 条件パーツの状態変化を監視するインターフェース
type ConditionPartObserver interface {
	OnPartSatisfied(partID ConditionPartID)
}

// ConditionObserver 条件の状態変化を監視するインターフェース
type ConditionObserver interface {
	OnConditionSatisfied(conditionID ConditionID)
}
