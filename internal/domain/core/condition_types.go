package core

type ConditionKind int

const (
	KindUnspecified ConditionKind = iota
	KindTime                      // 時間に基づく条件
	KindCounter                   // カウンターに基づく条件
)

// ComparisonOperator は比較演算子を表す型です
type ComparisonOperator int

const (
	ComparisonOperatorUnspecified ComparisonOperator = iota
	ComparisonOperatorEQ
	ComparisonOperatorNEQ
	ComparisonOperatorGT
	ComparisonOperatorGTE
	ComparisonOperatorLT
	ComparisonOperatorLTE
	ComparisonOperatorBetween
	ComparisonOperatorIn
	ComparisonOperatorNotIn
)

type ConditionID int64
type ConditionPartID int64
