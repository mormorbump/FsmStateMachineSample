package fsm

import (
	"context"
	"fmt"
)

// State は状態の定数を定義します
const (
	StateReady  = "ready"
	StateActive = "active"
	StateNext   = "next"
	StateFinish = "finish"
)

// Event はイベントの定数を定義します
const (
	EventActivate = "activate"
	EventNext     = "next"
	EventFinish   = "finish"
	EventReset    = "reset"
)

// StateError は状態遷移に関するエラーを表現します
type StateError struct {
	Code    string
	Message string
	Details interface{}
}

func (e *StateError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// 定義済みエラー
var (
	ErrInvalidTransition = &StateError{
		Code:    "INVALID_TRANSITION",
		Message: "指定された状態遷移は許可されていません",
	}
	ErrInvalidState = &StateError{
		Code:    "INVALID_STATE",
		Message: "無効な状態です",
	}
)

// StateObserver は状態変更を監視するインターフェースです
type StateObserver interface {
	OnStateChanged(newState string)
	OnError(err error)
}

// StateMachine は状態遷移を管理するインターフェースです
type StateMachine interface {
	CurrentState() string
	Transition(ctx context.Context, event string) error
	AddObserver(observer StateObserver)
	RemoveObserver(observer StateObserver)
}

// StateInfo は状態に関する情報を保持します
type StateInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	AllowedNext []string `json:"allowed_next"`
	Message     string   `json:"message,omitempty"`
}

// GetStateInfo は指定された状態の情報を返します
func GetStateInfo(state string) *StateInfo {
	stateInfoMap := map[string]*StateInfo{
		StateReady: {
			Name:        "Ready",
			Description: "初期状態",
			AllowedNext: []string{EventActivate},
			Message:     "開始待ち",
		},
		StateActive: {
			Name:        "Active",
			Description: "アクティブ状態",
			AllowedNext: []string{EventNext},
			Message:     "処理中...",
		},
		StateNext: {
			Name:        "Next",
			Description: "次状態への準備",
			AllowedNext: []string{EventActivate, EventFinish},
			Message:     "次の状態を判定中...",
		},
		StateFinish: {
			Name:        "Finish",
			Description: "終了状態",
			AllowedNext: []string{EventReset},
			Message:     "処理が完了しました。リセットして再開できます。",
		},
	}

	if info, exists := stateInfoMap[state]; exists {
		return info
	}
	return nil
}

// IsValidTransition は指定された状態遷移が有効かどうかを確認します
func IsValidTransition(currentState, event string) bool {
	// リセットイベントは終了状態からのみ許可
	if event == EventReset {
		return currentState == StateFinish
	}

	// 状態ごとの許可されたイベントを定義
	allowedEvents := map[string][]string{
		StateReady:  {EventActivate},
		StateActive: {EventNext},
		StateNext:   {EventActivate, EventFinish}, // StateNextからStateActiveへの遷移を許可
		StateFinish: {EventReset},
	}

	if events, exists := allowedEvents[currentState]; exists {
		for _, allowed := range events {
			if allowed == event {
				return true
			}
		}
	}
	return false
}
