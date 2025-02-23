package core

import (
	"testing"
)

func TestGetGameStateInfo(t *testing.T) {
	// テーブル駆動テスト用のテストケース
	tests := []struct {
		name          string
		state         string
		wantName      string
		wantDesc      string
		wantAllowed   []string
		wantMessage   string
		wantNilResult bool
	}{
		{
			name:        "Ready状態の情報を取得",
			state:       StateReady,
			wantName:    "Ready",
			wantDesc:    "初期状態",
			wantAllowed: []string{EventActivate},
			wantMessage: "開始待ち",
		},
		{
			name:        "Active状態の情報を取得",
			state:       StateActive,
			wantName:    "Active",
			wantDesc:    "アクティブ状態",
			wantAllowed: []string{EventNext},
			wantMessage: "処理中...",
		},
		{
			name:        "Next状態の情報を取得",
			state:       StateNext,
			wantName:    "Next",
			wantDesc:    "次状態への準備",
			wantAllowed: []string{EventActivate, EventFinish},
			wantMessage: "次の状態を判定中...",
		},
		{
			name:        "Finish状態の情報を取得",
			state:       StateFinish,
			wantName:    "Finish",
			wantDesc:    "終了状態",
			wantAllowed: []string{EventReset},
			wantMessage: "処理が完了しました。リセットして再開できます。",
		},
		{
			name:          "存在しない状態の場合はnilを返却",
			state:         "invalid_state",
			wantNilResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト対象の関数を実行
			got := GetGameStateInfo(tt.state)

			// nilチェック
			if tt.wantNilResult {
				if got != nil {
					t.Errorf("GetGameStateInfo() = %v, want nil", got)
				}
				return
			}

			// 結果の検証
			if got == nil {
				t.Fatal("GetGameStateInfo() returned nil, want non-nil")
			}

			// 各フィールドの検証
			if got.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", got.Name, tt.wantName)
			}

			if got.Description != tt.wantDesc {
				t.Errorf("Description = %v, want %v", got.Description, tt.wantDesc)
			}

			if got.Message != tt.wantMessage {
				t.Errorf("Message = %v, want %v", got.Message, tt.wantMessage)
			}

			// AllowedNextの検証
			if len(got.AllowedNext) != len(tt.wantAllowed) {
				t.Errorf("AllowedNext length = %v, want %v", len(got.AllowedNext), len(tt.wantAllowed))
			} else {
				for i, event := range got.AllowedNext {
					if event != tt.wantAllowed[i] {
						t.Errorf("AllowedNext[%d] = %v, want %v", i, event, tt.wantAllowed[i])
					}
				}
			}
		})
	}
}

func TestStateTransitions(t *testing.T) {
	// 状態遷移のテストケース
	tests := []struct {
		name          string
		currentState  string
		event        string
		wantAllowed  bool
	}{
		{
			name:         "Ready状態からActivateイベント",
			currentState: StateReady,
			event:       EventActivate,
			wantAllowed: true,
		},
		{
			name:         "Ready状態から不正なイベント",
			currentState: StateReady,
			event:       EventNext,
			wantAllowed: false,
		},
		{
			name:         "Active状態からNextイベント",
			currentState: StateActive,
			event:       EventNext,
			wantAllowed: true,
		},
		{
			name:         "Next状態からActivateイベント",
			currentState: StateNext,
			event:       EventActivate,
			wantAllowed: true,
		},
		{
			name:         "Next状態からFinishイベント",
			currentState: StateNext,
			event:       EventFinish,
			wantAllowed: true,
		},
		{
			name:         "Finish状態からResetイベント",
			currentState: StateFinish,
			event:       EventReset,
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 状態情報を取得
			stateInfo := GetGameStateInfo(tt.currentState)
			if stateInfo == nil {
				t.Fatalf("状態情報の取得に失敗: %s", tt.currentState)
			}

			// イベントが許可されているか確認
			isAllowed := false
			for _, allowed := range stateInfo.AllowedNext {
				if allowed == tt.event {
					isAllowed = true
					break
				}
			}

			if isAllowed != tt.wantAllowed {
				t.Errorf("イベント %s は状態 %s で %v, want %v",
					tt.event, tt.currentState, isAllowed, tt.wantAllowed)
			}
		})
	}
}