package ui

import (
	"fmt"
	"net/http"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/entity"
	logger "state_sample/internal/lib"
	"state_sample/internal/usecase"
	"sync"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

// ConditionInfo は条件の状態情報を表す構造体です
type ConditionInfo struct {
	ID          core.ConditionID    `json:"id"`
	Label       string              `json:"label"`
	State       string              `json:"state"`
	Kind        core.ConditionKind  `json:"kind"`
	IsClear     bool                `json:"is_clear"`
	Description string              `json:"description"`
	Parts       []ConditionPartInfo `json:"parts"`
}

// ConditionPartInfo は条件パーツの状態情報を表す構造体です
type ConditionPartInfo struct {
	ID                   core.ConditionPartID      `json:"id"`
	Label                string                    `json:"label"`
	State                string                    `json:"state"`
	ComparisonOperator   entity.ComparisonOperator `json:"comparison_operator"`
	IsClear              bool                      `json:"is_clear"`
	TargetEntityType     string                    `json:"target_entity_type"`
	TargetEntityID       int64                     `json:"target_entity_id"`
	ReferenceValueInt    int64                     `json:"reference_value_int"`
	ReferenceValueFloat  float64                   `json:"reference_value_float"`
	ReferenceValueString string                    `json:"reference_value_string"`
	MinValue             float64                   `json:"min_value"`
	MaxValue             float64                   `json:"max_value"`
	Priority             int32                     `json:"priority"`
}

type StateServer struct {
	stateFacade usecase.StateFacade
	clients     map[*websocket.Conn]bool
	upgrader    websocket.Upgrader
	mu          sync.RWMutex
}

func NewStateServer(facade usecase.StateFacade) *StateServer {
	log := logger.DefaultLogger()
	log.Debug("Creating new state server instance")
	server := &StateServer{
		stateFacade: facade,
		clients:     make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	facade.GetController().AddObserver(server)
	return server
}

func (s *StateServer) OnStateChanged(state string) {
	log := logger.DefaultLogger()
	log.Debug("StateServer.OnStateChanged", zap.String("state", state))
	currentPhase := s.stateFacade.GetCurrentPhase()
	stateInfo := currentPhase.GetStateInfo()

	// Conditionの情報を収集
	conditions := make([]ConditionInfo, 0)
	for _, condition := range currentPhase.GetConditions() {
		condInfo := ConditionInfo{
			ID:          condition.ID,
			Label:       condition.Label,
			State:       condition.CurrentState(),
			Kind:        condition.Kind,
			IsClear:     condition.IsClear,
			Description: condition.Description,
			Parts:       make([]ConditionPartInfo, 0),
		}

		// ConditionPartの情報を収集
		for _, part := range condition.Parts {
			partInfo := ConditionPartInfo{
				ID:                   part.ID,
				Label:                part.Label,
				State:                part.CurrentState(),
				ComparisonOperator:   part.ComparisonOperator,
				IsClear:              part.IsClear,
				TargetEntityType:     part.TargetEntityType,
				TargetEntityID:       part.TargetEntityID,
				ReferenceValueInt:    part.ReferenceValueInt,
				ReferenceValueFloat:  part.ReferenceValueFloat,
				ReferenceValueString: part.ReferenceValueString,
				MinValue:             part.MinValue,
				MaxValue:             part.MaxValue,
				Priority:             part.Priority,
			}
			condInfo.Parts = append(condInfo.Parts, partInfo)
		}

		conditions = append(conditions, condInfo)
	}

	update := struct {
		Type       string              `json:"type"`
		State      string              `json:"state"`
		Info       *core.GameStateInfo `json:"info,omitempty"`
		Phase      string              `json:"phase"`
		Message    string              `json:"message,omitempty"`
		Conditions []ConditionInfo     `json:"conditions"`
	}{
		Type:       "state_change",
		State:      state,
		Info:       stateInfo,
		Phase:      currentPhase.Type,
		Message:    fmt.Sprintf("order: %v, message: %v", currentPhase.Order, stateInfo.Message),
		Conditions: conditions,
	}
	s.broadcastUpdate(update)
}

func (s *StateServer) OnError(err error) {
	update := struct {
		Type  string `json:"type"`
		Error string `json:"error"`
	}{
		Type:  "error",
		Error: err.Error(),
	}
	s.broadcastUpdate(update)
}

func (s *StateServer) broadcastUpdate(update interface{}) {
	log := logger.DefaultLogger()
	s.mu.RLock()
	defer s.mu.RUnlock()
	log.Debug("Broadcasting update to clients", zap.Any("update", update))
	for client := range s.clients {
		if err := client.WriteJSON(update); err != nil {
			log.Error("Error sending message to client: %v", zap.Error(err))
			err := client.Close()
			if err != nil {
				log.Error("Error closing client connection: %v", zap.Error(err))
				return
			}
			delete(s.clients, client)
		}
	}
}

func (s *StateServer) Close() error {
	log := logger.DefaultLogger()
	log.Debug("Closing state server")
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		if err := client.Close(); err != nil {
			log.Error("Error closing client", zap.Error(err))
		}
	}
	s.clients = nil
	return nil
}
