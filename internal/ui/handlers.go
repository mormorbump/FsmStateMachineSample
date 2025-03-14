package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"strconv"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// handleWebSocket WebSocket接続を処理
func (s *StateServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log := logger.DefaultLogger()
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error upgrading connection", zap.Error(err))
		return
	}

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	// WebSocket接続時には初期状態を送信しない
	// 初期状態はクライアント側で/api/initial-stateエンドポイントから取得する

	go func() { _ = s.recvWsMessage(conn) }()
}

func (s *StateServer) recvWsMessage(conn *websocket.Conn) error {
	log := logger.DefaultLogger()
	defer func() {
		log.Debug("recvWsMessage: Closing connection")
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		err := conn.Close()
		if err != nil {
			log.Error("Error closing connection", zap.Error(err))
			return
		}
	}()

	for {
		var msg struct {
			Event string `json:"event"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Error("Error reading message", zap.Error(err))
			return err
		}

		log.Debug("WS: Received message", zap.String("event", msg.Event))
		err := s.handleActionRequest(msg.Event)
		if err != nil {
			log.Error("Error handling action request", zap.Error(err))
			return err
		}
	}
}

func (s *StateServer) handleActionRequest(action string) error {
	log := logger.DefaultLogger()
	var err error
	switch action {
	case "start", "activate":
		err = s.stateFacade.Start(context.Background())
	case "stop":
		err = s.stateFacade.Reset(context.Background())
	case "reset", "finish":
		err = s.stateFacade.Reset(context.Background())
	default:
		log.Error("Invalid action", zap.String("action", action))
	}
	return err
}

// handleAutoTransition 自動遷移の制御を処理
func (s *StateServer) handleAutoTransition(w http.ResponseWriter, r *http.Request) {
	log := logger.DefaultLogger()
	action := r.URL.Query().Get("action")
	log.Debug("Received auto-transition control request", zap.String("action", action))

	log.Debug("HTTP: Received message", zap.String("event", action))
	err := s.handleActionRequest(action)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleConditionPartEvaluate カウンター条件の評価を処理
func (s *StateServer) handleConditionPartEvaluate(w http.ResponseWriter, r *http.Request) {
	log := logger.DefaultLogger()
	vars := mux.Vars(r)

	currentPhase := s.stateFacade.GetCurrentLeafPhase()
	if currentPhase == nil {
		http.Error(w, "no active phase", http.StatusBadRequest)
		return
	}

	if currentPhase.CurrentState() == value.StateReady {
		http.Error(w, "state is ready", http.StatusBadRequest)
		return
	}

	// URLパラメータの取得と検証
	conditionID := vars["condition_id"]
	partID := vars["part_id"]
	if conditionID == "" || partID == "" {
		http.Error(w, "Missing condition_id or part_id", http.StatusBadRequest)
		return
	}

	log.Debug("Received condition part evaluation request")
	// リクエストボディの解析
	var request struct {
		Increment int64 `json:"increment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// IDの変換
	condIDInt, err := strconv.ParseInt(conditionID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid condition_id", http.StatusBadRequest)
		return
	}
	partIDInt, err := strconv.ParseInt(partID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid part_id", http.StatusBadRequest)
		return
	}

	// 条件パーツの取得と評価
	part, err := s.stateFacade.GetConditionPart(condIDInt, partIDInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get condition part: %v", err), http.StatusInternalServerError)
		return
	}

	err = part.Process(r.Context(), request.Increment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to evaluate condition: %v", err), http.StatusInternalServerError)
		return
	}

	// レスポンスの作成
	response := struct {
		CurrentValue interface{} `json:"current_value"`
		TargetValue  int64       `json:"target_value"`
		IsSatisfied  bool        `json:"is_satisfied"`
	}{
		CurrentValue: part.GetCurrentValue(),
		TargetValue:  part.GetReferenceValueInt(),
		IsSatisfied:  part.IsSatisfied(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleInitialState 初期状態を取得するAPIエンドポイント
func (s *StateServer) handleInitialState(w http.ResponseWriter, r *http.Request) {
	log := logger.DefaultLogger()

	// すべてのフェーズを取得
	allPhases := s.stateFacade.GetController().GetPhases()
	phaseDTOs := GetAllPhasesDTO(allPhases)

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
		Conditions: allConditions,
	}

	// 現在のルートフェーズを取得（親ID=0のフェーズ）
	currentRootPhase := s.stateFacade.GetCurrentPhase(0)
	if currentRootPhase != nil {
		// 現在のルートフェーズが存在する場合のみ、関連情報を設定
		stateInfo := s.getGameStateInfo(currentRootPhase)
		response.State = currentRootPhase.CurrentState()
		response.Info = stateInfo
		response.Message = fmt.Sprintf("order: %v, message: %v", currentRootPhase.Order, stateInfo.Message)

		// 現在のルートフェーズをDTOに変換
		currentDTO := ConvertPhaseToDTO(currentRootPhase)
		response.CurrentPhase = &currentDTO
	} else {
		// 現在のルートフェーズが存在しない場合は、デフォルト値を設定
		response.State = "ready"
		response.Message = "初期状態です。フェーズを開始してください。"
	}

	// レスポンスを送信
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *StateServer) Start(addr string) error {
	log := logger.DefaultLogger()
	r := mux.NewRouter()

	r.HandleFunc("/ws", s.handleWebSocket)
	r.HandleFunc("/api/auto-transition", s.handleAutoTransition).Methods("POST")
	r.HandleFunc("/api/condition/{condition_id}/part/{part_id}/evaluate", s.handleConditionPartEvaluate).Methods("POST")
	r.HandleFunc("/api/initial-state", s.handleInitialState).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("internal/ui/static")))

	log.Debug("Starting server on", zap.String("addr", addr))
	return http.ListenAndServe(addr, r)
}
