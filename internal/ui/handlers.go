package ui

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	logger "state_sample/internal/lib"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// handleWebSocket WebSocket接続を処理
func (s *StateServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log := logger.DefaultLogger()
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error upgrading connection: %v", zap.Error(err))
		return
	}

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	// 初期状態を送信
	currentPhase := s.stateFacade.GetCurrentPhase()
	if currentPhase != nil {
		s.OnStateChanged(currentPhase.CurrentState())
	}

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
			log.Error("Error closing connection: %v", zap.Error(err))
			return
		}
	}()

	for {
		var msg struct {
			Event string `json:"event"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Error("Error reading message: %v", zap.Error(err))
			return err
		}

		log.Debug("WS: Received message: %v", zap.String("event", msg.Event))
		err := s.handleActionRequest(msg.Event)
		if err != nil {
			log.Error("Error reading message: %v", zap.Error(err))
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
		log.Error("Invalid action: %v", zap.String("action", action))
	}
	return err
}

// handleAutoTransition 自動遷移の制御を処理
func (s *StateServer) handleAutoTransition(w http.ResponseWriter, r *http.Request) {
	log := logger.DefaultLogger()
	action := r.URL.Query().Get("action")
	log.Debug("Received auto-transition control request: ", zap.String("action", action))

	log.Debug("HTTP: Received message: %v", zap.String("event", action))
	err := s.handleActionRequest(action)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *StateServer) Start(addr string) error {
	log := logger.DefaultLogger()
	r := mux.NewRouter()

	r.HandleFunc("/ws", s.handleWebSocket)
	r.HandleFunc("/api/auto-transition", s.handleAutoTransition).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("internal/ui/static")))

	log.Debug("Starting server on %s", zap.String("addr", addr))
	return http.ListenAndServe(addr, r)
}
