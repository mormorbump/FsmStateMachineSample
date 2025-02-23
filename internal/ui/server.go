package ui

import (
	"log"
	"net/http"
	"state_sample/internal/domain/core"
	"state_sample/internal/usecase"
	"sync"

	"github.com/gorilla/websocket"
)

// StateServer はWebSocketを通じて状態変更を通知するサーバーです
type StateServer struct {
	stateFacade usecase.StateFacade
	clients     map[*websocket.Conn]bool
	upgrader    websocket.Upgrader
	mu          sync.RWMutex
}

// NewStateServer は新しいStateServerインスタンスを作成します
func NewStateServer(facade usecase.StateFacade) *StateServer {
	log.Println("Creating new state server instance")
	server := &StateServer{
		stateFacade: facade,
		clients:     make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 開発用に全てのオリジンを許可
			},
		},
	}

	// PhaseControllerの監視を設定
	facade.GetController().AddObserver(server)

	return server
}

// OnStateChanged は状態変更時に呼び出されます
func (s *StateServer) OnStateChanged(state string) {
	currentPhase := s.stateFacade.GetCurrentPhase()
	if currentPhase == nil {
		return
	}

	stateInfo := currentPhase.GetStateInfo()
	update := struct {
		Type    string              `json:"type"`
		State   string              `json:"state"`
		Info    *core.GameStateInfo `json:"info,omitempty"`
		Phase   string              `json:"phase"`
		Message string              `json:"message,omitempty"`
	}{
		Type:    "state_change",
		State:   state,
		Info:    stateInfo,
		Phase:   currentPhase.Type,
		Message: stateInfo.Message,
	}
	s.broadcastUpdate(update)
}

// OnError はエラー発生時に呼び出されます
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

// broadcastUpdate は全てのクライアントに更新を送信します
func (s *StateServer) broadcastUpdate(update interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for client := range s.clients {
		if err := client.WriteJSON(update); err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

// Close はサーバーのリソースを解放します
func (s *StateServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		client.Close()
	}
	s.clients = nil

	return nil
}
