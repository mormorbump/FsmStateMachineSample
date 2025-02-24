package core

type ConditionKind int

const (
	KindUnspecified ConditionKind = iota
	KindTime                      // 時間に基づく条件
	KindScore                     // スコアに基づく条件
)

type ConditionID int64
type ConditionPartID int64

type ConditionPart interface {
	GetReferenceValueInt() int64
	ConditionPartSubject
	TimeObserver
}
