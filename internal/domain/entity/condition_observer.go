package entity

// ConditionObserver 条件の状態変化を監視するインターフェース
type ConditionObserver interface {
	OnConditionSatisfied(conditionID ConditionID)
}

// ConditionPartObserver 条件パーツの状態変化を監視するインターフェース
type ConditionPartObserver interface {
	OnPartSatisfied(partID ConditionPartID)
}
