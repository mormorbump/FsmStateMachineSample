package core

// StateObserver 状態を監視するインターフェース
type StateObserver interface {
	OnStateChanged(state string)
}

// TimeObserver 時間を監視するインターフェース
type TimeObserver interface {
	OnTimeTicked()
}
