package value

// GameStateInfo はゲーム状態の情報を表す構造体です
type GameStateInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	AllowedNext []string `json:"allowed_next"`
	Message     string   `json:"message,omitempty"`
}

// GetGameStateInfo は状態名に対応するGameStateInfoを返します
func GetGameStateInfo(state string) *GameStateInfo {
	stateInfoMap := map[string]*GameStateInfo{
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
