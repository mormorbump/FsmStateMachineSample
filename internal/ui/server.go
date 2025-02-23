package ui

import (
	"fmt"
	"net/http"
	"state_sample/internal/domain/core"
	logger "state_sample/internal/lib"
	"state_sample/internal/usecase"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

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

	update := struct {
		Type           string              `json:"type"`
		State          string              `json:"state"`
		Info           *core.GameStateInfo `json:"info,omitempty"`
		Phase          string              `json:"phase"`
		NextTransition time.Duration       `json:"next_transition"`
		Message        string              `json:"message,omitempty"`
	}{
		Type:           "state_change",
		State:          state,
		Info:           stateInfo,
		Phase:          currentPhase.Type,
		NextTransition: currentPhase.Interval,
		Message:        fmt.Sprintf("interval: %v, order: %v, message: %v", currentPhase.Interval, currentPhase.Order, stateInfo.Message),
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
