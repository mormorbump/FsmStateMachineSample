package ui

import (
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/value"
	"time"
)

// PhaseDTO はUI層で使用するフェーズのデータ転送オブジェクト
type PhaseDTO struct {
	ID          value.PhaseID `json:"id"`
	ParentID    value.PhaseID `json:"parent_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Order       int           `json:"order"`
	State       string        `json:"state"`
	IsClear     bool          `json:"is_clear"`
	IsActive    bool          `json:"is_active"`
	HasChildren bool          `json:"has_children"`
	StartTime   *time.Time    `json:"start_time,omitempty"`
	FinishTime  *time.Time    `json:"finish_time,omitempty"`
}

// ConvertPhaseToDTO はPhaseオブジェクトをDTOに変換する
func ConvertPhaseToDTO(phase *entity.Phase) PhaseDTO {
	return PhaseDTO{
		ID:          phase.ID,
		ParentID:    phase.ParentID,
		Name:        phase.Name,
		Description: phase.Description,
		Order:       phase.Order,
		State:       phase.CurrentState(),
		IsClear:     phase.IsClear,
		IsActive:    phase.IsActive(),
		HasChildren: phase.HasChildren(),
		StartTime:   phase.StartTime,
		FinishTime:  phase.FinishTime,
	}
}

// GetAllPhasesDTO は全てのフェーズをDTOに変換する
func GetAllPhasesDTO(phases entity.Phases) []PhaseDTO {
	result := make([]PhaseDTO, len(phases))
	for i, phase := range phases {
		result[i] = ConvertPhaseToDTO(phase)
	}
	return result
}
