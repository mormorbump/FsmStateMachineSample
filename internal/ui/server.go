package ui

import (
	"fmt"
	"net/http"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"state_sample/internal/usecase/state"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

// ConditionInfo は条件の状態情報を表す構造体です
type ConditionInfo struct {
	ID          value.ConditionID   `json:"id"`
	Label       string              `json:"label"`
	State       string              `json:"state"`
	Kind        value.ConditionKind `json:"kind"`
	IsClear     bool                `json:"is_clear"`
	Description string              `json:"description"`
	Parts       []ConditionPartInfo `json:"parts"`
}

// ConditionPartInfo は条件パーツの状態情報を表す構造体です
type ConditionPartInfo struct {
	ID                   value.ConditionPartID    `json:"id"`
	Label                string                   `json:"label"`
	State                string                   `json:"state"`
	ComparisonOperator   value.ComparisonOperator `json:"comparison_operator"`
	IsClear              bool                     `json:"is_clear"`
	TargetEntityType     string                   `json:"target_entity_type"`
	TargetEntityID       int64                    `json:"target_entity_id"`
	ReferenceValueInt    int64                    `json:"reference_value_int"`
	ReferenceValueFloat  float64                  `json:"reference_value_float"`
	ReferenceValueString string                   `json:"reference_value_string"`
	MinValue             int64                    `json:"min_value"`
	MaxValue             int64                    `json:"max_value"`
	Priority             int32                    `json:"priority"`
	CurrentValue         interface{}              `json:"current_value"`
}

type StateServer struct {
	stateFacade state.StateFacade
	clients     map[*websocket.Conn]bool
	upgrader    websocket.Upgrader
	mu          sync.RWMutex
}

func NewStateServer(facade state.StateFacade) *StateServer {
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

	// オブザーバーとして登録
	controller := facade.GetController()
	log.Debug("Got controller from facade", zap.String("controller", fmt.Sprintf("%p", controller)))

	controller.AddControllerObserver(server)
	log.Debug("Added StateServer as StateObserver", zap.String("server", fmt.Sprintf("%p", server)))

	return server
}

// GameStateInfo は状態情報を表す構造体です
type GameStateInfo struct {
	CurrentState string `json:"current_state"`
	Message      string `json:"message"`
}

// getGameStateInfo はGameStateInfoを取得します
func (s *StateServer) getGameStateInfo(phase *entity.Phase) *GameStateInfo {
	if phase == nil {
		return &GameStateInfo{
			CurrentState: value.StateReady,
			Message:      "No active phase",
		}
	}

	return &GameStateInfo{
		CurrentState: phase.CurrentState(),
		Message:      fmt.Sprintf("Phase: %s, Order: %d", phase.Name, phase.Order),
	}
}

type Response struct {
	Type  string         `json:"type"`
	State string         `json:"state"`
	Info  *GameStateInfo `json:"info,omitempty"`
	Phase struct {
		Name        string     `json:"name"`
		Description string     `json:"description"`
		Order       int        `json:"order"`
		IsClear     bool       `json:"is_clear"`
		StartTime   *time.Time `json:"start_time"`
		FinishTime  *time.Time `json:"finish_time"`
	} `json:"phase"`
	Message    string          `json:"message,omitempty"`
	Conditions []ConditionInfo `json:"conditions"`
}

func (s *StateServer) EditResponse(stateName string, currentPhase *entity.Phase, stateInfo *GameStateInfo) Response {
	var currentState string
	if stateName == "" {
		currentState = currentPhase.CurrentState()
	} else {
		currentState = stateName
	}

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
		for _, part := range condition.GetParts() {
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
				CurrentValue:         part.GetCurrentValue(), // strategy経由で現在値を取得
			}
			condInfo.Parts = append(condInfo.Parts, partInfo)
			// 重複して追加していた2回目のappendを削除
		}

		conditions = append(conditions, condInfo)
	}

	update := Response{
		Type:  "state_change",
		State: currentState,
		Info:  stateInfo,
		Phase: struct {
			Name        string     `json:"name"`
			Description string     `json:"description"`
			Order       int        `json:"order"`
			IsClear     bool       `json:"is_clear"`
			StartTime   *time.Time `json:"start_time"`
			FinishTime  *time.Time `json:"finish_time"`
		}{
			Name:        currentPhase.Name,
			Description: currentPhase.Description,
			Order:       currentPhase.Order,
			IsClear:     currentPhase.IsClear,
			StartTime:   currentPhase.StartTime,
			FinishTime:  currentPhase.FinishTime,
		},
		Message:    fmt.Sprintf("order: %v, message: %v", currentPhase.Order, stateInfo.Message),
		Conditions: conditions,
	}
	return update
}

func (s *StateServer) OnEntityChanged(entityObj interface{}) {
	log := logger.DefaultLogger()
	var currentPhase *entity.Phase
	var stateInfo *GameStateInfo
	if entityObj != nil {
		switch e := entityObj.(type) {
		case *entity.Phase:
			log.Debug("StateServer.OnEntityChanged", zap.Any("entity", e))
		case *entity.Condition:
			log.Debug("StateServer.OnEntityChanged", zap.Any("entity", e))
		case *entity.ConditionPart:
			log.Debug("StateServer.OnEntityChanged", zap.Any("entity", e))
		default:
			log.Debug("StateServer.OnEntityChanged", zap.Any("entity", e))
		}
		currentPhase = s.stateFacade.GetCurrentPhase()
		stateInfo = s.getGameStateInfo(currentPhase)
	} else {
		// nilの場合は終了なので、最後の情報を取得
		currentPhase = s.stateFacade.GetController().GetPhases()[:1][0]
		stateInfo = s.getGameStateInfo(currentPhase)
	}

	update := s.EditResponse(currentPhase.CurrentState(), currentPhase, stateInfo)
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
			log.Error("Error sending message to client", zap.Error(err))
			err := client.Close()
			if err != nil {
				log.Error("Error closing client connection", zap.Error(err))
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
