package condition

// ConditionPartObserver は条件パーツの状態変化を監視するインターフェースです
type ConditionPartObserver interface {
	OnPartSatisfied(partID ConditionPartID)
}

// ConditionObserver は条件の状態変化を監視するインターフェースです
type ConditionObserver interface {
	OnConditionSatisfied(conditionID ConditionID)
}
