package core

// StateObserver は監視者のインターフェースを定義します
type StateObserver interface {
	OnStateChanged(state string)
}

type TimeObserver interface {
	OnTimeTicked()
}
