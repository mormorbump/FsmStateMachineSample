package entity

import (
	"state_sample/internal/domain/value"
)

// GameState はゲームの状態を表す構造体です
type GameState struct {
	CurrentState string
	StateInfo    *value.GameStateInfo
	Phases       Phases
	CurrentPhase *Phase
}

// NewGameState は新しいGameStateインスタンスを作成します
func NewGameState(phases Phases) *GameState {
	var currentPhase *Phase
	if len(phases) > 0 {
		currentPhase = phases[0]
	}

	return &GameState{
		CurrentState: value.StateReady,
		StateInfo:    value.GetGameStateInfo(value.StateReady),
		Phases:       phases,
		CurrentPhase: currentPhase,
	}
}

// SetCurrentPhase は現在のフェーズを設定します
func (g *GameState) SetCurrentPhase(phase *Phase) {
	g.CurrentPhase = phase
	g.CurrentState = phase.CurrentState()
	g.StateInfo = value.GetGameStateInfo(g.CurrentState)
}

// GetCurrentPhase は現在のフェーズを返します
func (g *GameState) GetCurrentPhase() *Phase {
	return g.CurrentPhase
}

// GetPhases は全てのフェーズを返します
func (g *GameState) GetPhases() Phases {
	return g.Phases
}

// GetStateInfo は現在の状態情報を返します
func (g *GameState) GetStateInfo() *value.GameStateInfo {
	return g.StateInfo
}

// UpdateState は状態を更新します
func (g *GameState) UpdateState(state string) {
	g.CurrentState = state
	g.StateInfo = value.GetGameStateInfo(state)
}
