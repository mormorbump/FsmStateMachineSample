package ui

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// handleWebSocket はWebSocket接続を処理します
func (s *StateServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
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

	// クライアントからのメッセージを処理
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			conn.Close()
		}()

		for {
			var msg struct {
				Event string `json:"event"`
			}
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Error reading message: %v", err)
				}
				return
			}

			// イベントを処理
			if err := s.stateFacade.Start(context.Background()); err != nil {
				s.OnError(err)
			} else {
				currentPhase := s.stateFacade.GetCurrentPhase()
				if currentPhase != nil {
					s.OnStateChanged(currentPhase.CurrentState())
				}
			}
		}
	}()
}

// handleAutoTransition は自動遷移の制御を処理します
func (s *StateServer) handleAutoTransition(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	log.Printf("Received auto-transition control request: %s", action)

	var err error
	switch action {
	case "start":
		err = s.stateFacade.Start(r.Context())
	case "stop":
		err = s.stateFacade.Stop()
	case "reset":
		err = s.stateFacade.Reset(r.Context())
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Start はサーバーを起動します
func (s *StateServer) Start(addr string) error {
	r := mux.NewRouter()

	r.HandleFunc("/ws", s.handleWebSocket)
	r.HandleFunc("/api/auto-transition", s.handleAutoTransition).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("internal/ui/static")))

	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, r)
}
