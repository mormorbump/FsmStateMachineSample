package service

import "context"

// StateMachine は状態マシンの操作を定義するインターフェース
type StateMachine interface {
	// CurrentState 現在の状態を返す
	CurrentState() string
	
	// CanTransition 指定されたイベントで遷移可能か確認
	CanTransition(event string) bool
	
	// Transition 指定されたイベントで状態遷移を実行
	Transition(ctx context.Context, event string) error
	
	// AddCallback 状態遷移時のコールバックを登録
	AddCallback(callbackType string, state string, callback func(ctx context.Context, event string, fromState string, toState string) error)
}

// StateMachineFactory は状態マシンを生成するファクトリインターフェース
type StateMachineFactory interface {
	// CreateStateMachine 新しい状態マシンを作成
	CreateStateMachine(initialState string, transitions []StateTransition) (StateMachine, error)
}

// StateTransition は状態遷移の定義
type StateTransition struct {
	Event string   // 遷移イベント名
	From  []string // 遷移元の状態
	To    string   // 遷移先の状態
}