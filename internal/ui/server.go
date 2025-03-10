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
	PhaseID     value.PhaseID       `json:"phase_id"`
	PhaseName   string              `json:"phase_name"`
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
	stateFacade *state.GameFacade
	clients     map[*websocket.Conn]bool
	upgrader    websocket.Upgrader
	mu          sync.RWMutex
	updateChan  chan interface{} // 更新メッセージを送信するためのチャネル
	done        chan struct{}    // サーバー終了を通知するためのチャネル
}

func NewStateServer(facade *state.GameFacade) *StateServer {
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
		updateChan: make(chan interface{}, 100), // バッファ付きチャネルを作成
		done:       make(chan struct{}),
	}

	// オブザーバーとして登録
	controller := facade.GetController()
	log.Debug("Got controller from facade", zap.String("controller", fmt.Sprintf("%p", controller)))

	controller.AddControllerObserver(server)
	log.Debug("Added StateServer as StateObserver", zap.String("server", fmt.Sprintf("%p", server)))

	// 更新メッセージを処理するゴルーチンを起動
	go server.processUpdates()

	return server
}

// processUpdates は更新メッセージを処理するゴルーチン
func (s *StateServer) processUpdates() {
	log := logger.DefaultLogger()
	log.Debug("Starting update processor goroutine")

	for {
		select {
		case update := <-s.updateChan:
			// 実際の更新処理を行う
			s.sendUpdateToClients(update)
		case <-s.done:
			log.Debug("Update processor goroutine shutting down")
			return
		}
	}
}

// sendUpdateToClients は実際にクライアントに更新を送信する
func (s *StateServer) sendUpdateToClients(update interface{}) {
	log := logger.DefaultLogger()
	s.mu.RLock()
	defer s.mu.RUnlock()

	log.Debug("Sending update to clients", zap.Any("update", update))
	for client := range s.clients {
		if err := client.WriteJSON(update); err != nil {
			log.Error("Error sending message to client", zap.Error(err))
			err := client.Close()
			if err != nil {
				log.Error("Error closing client connection", zap.Error(err))
			}
			delete(s.clients, client)
		}
	}
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

// PhaseInfo は階層構造のPhase情報を表す構造体です
type PhaseInfo struct {
	ID          value.PhaseID `json:"id"`
	ParentID    value.PhaseID `json:"parent_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Order       int           `json:"order"`
	IsClear     bool          `json:"is_clear"`
	IsActive    bool          `json:"is_active"`
	State       string        `json:"state"`
	HasChildren bool          `json:"has_children"`
}

type Response struct {
	Type  string         `json:"type"`
	State string         `json:"state"`
	Info  *GameStateInfo `json:"info,omitempty"`
	Phase struct {
		ID          value.PhaseID `json:"id"`
		ParentID    value.PhaseID `json:"parent_id"`
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Order       int           `json:"order"`
		IsClear     bool          `json:"is_clear"`
		StartTime   *time.Time    `json:"start_time"`
		FinishTime  *time.Time    `json:"finish_time"`
		HasChildren bool          `json:"has_children"`
	} `json:"phase"`
	ParentPhase *PhaseInfo      `json:"parent_phase,omitempty"`
	ChildPhases []*PhaseInfo    `json:"child_phases,omitempty"`
	Message     string          `json:"message,omitempty"`
	Conditions  []ConditionInfo `json:"conditions"`
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

	// 基本的なレスポンス構造を作成
	update := Response{
		Type:  "state_change",
		State: currentState,
		Info:  stateInfo,
		Phase: struct {
			ID          value.PhaseID `json:"id"`
			ParentID    value.PhaseID `json:"parent_id"`
			Name        string        `json:"name"`
			Description string        `json:"description"`
			Order       int           `json:"order"`
			IsClear     bool          `json:"is_clear"`
			StartTime   *time.Time    `json:"start_time"`
			FinishTime  *time.Time    `json:"finish_time"`
			HasChildren bool          `json:"has_children"`
		}{
			ID:          currentPhase.ID,
			ParentID:    currentPhase.ParentID,
			Name:        currentPhase.Name,
			Description: currentPhase.Description,
			Order:       currentPhase.Order,
			IsClear:     currentPhase.IsClear,
			StartTime:   currentPhase.StartTime,
			FinishTime:  currentPhase.FinishTime,
			HasChildren: currentPhase.HasChildren(),
		},
		Message:    fmt.Sprintf("order: %v, message: %v", currentPhase.Order, stateInfo.Message),
		Conditions: conditions,
	}

	// 親フェーズの情報を追加（存在する場合）
	if currentPhase.Parent != nil {
		parentPhase := currentPhase.Parent
		update.ParentPhase = &PhaseInfo{
			ID:          parentPhase.ID,
			ParentID:    parentPhase.ParentID,
			Name:        parentPhase.Name,
			Description: parentPhase.Description,
			Order:       parentPhase.Order,
			IsClear:     parentPhase.IsClear,
			IsActive:    parentPhase.CurrentState() == value.StateActive,
			State:       parentPhase.CurrentState(),
			HasChildren: true, // 現在のフェーズが子なので必ずtrue
		}
	}

	// 子フェーズの情報を追加（存在する場合）
	if currentPhase.HasChildren() {
		childPhases := currentPhase.GetChildren()
		update.ChildPhases = make([]*PhaseInfo, len(childPhases))
		for i, childPhase := range childPhases {
			update.ChildPhases[i] = &PhaseInfo{
				ID:          childPhase.ID,
				ParentID:    childPhase.ParentID,
				Name:        childPhase.Name,
				Description: childPhase.Description,
				Order:       childPhase.Order,
				IsClear:     childPhase.IsClear,
				IsActive:    childPhase.CurrentState() == value.StateActive,
				State:       childPhase.CurrentState(),
				HasChildren: childPhase.HasChildren(),
			}
		}
	}

	return update
}

func (s *StateServer) OnEntityChanged(entityObj interface{}) {
	log := logger.DefaultLogger()
	var currentPhase *entity.Phase
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
		// ルートフェーズを取得（親ID=0のフェーズ）
		currentPhase = s.stateFacade.GetCurrentPhase(0)
		// ルートフェーズが存在しない場合は最下層のフェーズを取得
		if currentPhase == nil {
			currentPhase = s.stateFacade.GetCurrentLeafPhase()
		}
	} else {
		// nilの場合は終了なので、最後の情報を取得
		currentPhase = s.stateFacade.GetController().GetPhases()[:1][0]
	}

	// currentPhaseがnilの場合の対処
	if currentPhase == nil {
		log.Debug("OnEntityChanged: No active phase found, using default state")
		// デフォルトの状態情報を送信
		defaultUpdate := struct {
			Type    string `json:"type"`
			State   string `json:"state"`
			Message string `json:"message"`
		}{
			Type:    "state_info",
			State:   value.StateReady,
			Message: "No active phase. System is initializing or in transition.",
		}
		s.broadcastUpdate(defaultUpdate)
		return
	}

	// DTOアプローチを使用して全てのフェーズを取得
	allPhases := s.stateFacade.GetController().GetPhases()
	phaseDTOs := GetAllPhasesDTO(allPhases)

	// 現在のルートフェーズを取得
	currentRootPhase := s.stateFacade.GetCurrentPhase(0)

	// 現在のフェーズを決定（ルートフェーズを優先）
	displayPhase := currentPhase
	if currentRootPhase != nil && currentRootPhase.CurrentState() != value.StateFinish {
		displayPhase = currentRootPhase
	}

	// すべてのフェーズの条件を取得
	var allConditions []ConditionInfo
	for _, phase := range allPhases {
		conditions := s.getConditionInfos(phase)
		allConditions = append(allConditions, conditions...)
	}

	// レスポンスを構築
	response := struct {
		Type         string          `json:"type"`
		Phases       []PhaseDTO      `json:"phases"`
		CurrentPhase *PhaseDTO       `json:"current_phase,omitempty"`
		State        string          `json:"state"`
		Info         *GameStateInfo  `json:"info,omitempty"`
		Message      string          `json:"message,omitempty"`
		Conditions   []ConditionInfo `json:"conditions"`
	}{
		Type:       "state_change",
		Phases:     phaseDTOs,
		State:      displayPhase.CurrentState(),
		Info:       s.getGameStateInfo(displayPhase),
		Message:    fmt.Sprintf("order: %v, message: %v", displayPhase.Order, s.getGameStateInfo(displayPhase).Message),
		Conditions: allConditions,
	}

	// 現在のフェーズをDTOに変換
	currentDTO := ConvertPhaseToDTO(displayPhase)
	response.CurrentPhase = &currentDTO

	s.broadcastUpdate(response)
}

// getConditionInfos は条件情報を取得する
func (s *StateServer) getConditionInfos(phase *entity.Phase) []ConditionInfo {
	conditions := make([]ConditionInfo, 0)
	for _, condition := range phase.GetConditions() {
		condInfo := ConditionInfo{
			ID:          condition.ID,
			Label:       condition.Label,
			State:       condition.CurrentState(),
			Kind:        condition.Kind,
			IsClear:     condition.IsClear,
			Description: condition.Description,
			PhaseID:     phase.ID,
			PhaseName:   phase.Name,
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
				CurrentValue:         part.GetCurrentValue(),
			}
			condInfo.Parts = append(condInfo.Parts, partInfo)
		}
		conditions = append(conditions, condInfo)
	}
	return conditions
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
	log.Debug("Queueing update for broadcast", zap.Any("update", update))

	// 更新メッセージをチャネルに送信（非ブロッキング）
	select {
	case s.updateChan <- update:
		// メッセージが正常にキューに入った
	default:
		// チャネルがいっぱいの場合
		log.Warn("Update channel is full, dropping message")
	}
}

func (s *StateServer) Close() error {
	log := logger.DefaultLogger()
	log.Debug("Closing state server")

	// 更新処理ゴルーチンを終了
	close(s.done)
	log.Debug("Sent shutdown signal to update processor")

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
