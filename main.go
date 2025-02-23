package main

import (
	"log"
	"state_sample/internal/fsm"
	"time"

	"state_sample/internal/ui"
)

func main() {
	phases := fsm.Phases{
		fsm.NewPhase("BUILD_PHASE", 5*time.Second, 1),
		fsm.NewPhase("COMBAT_PHASE", 3*time.Second, 2),
		fsm.NewPhase("RESOLUTION_PHASE", 2*time.Second, 3),
	}
	facade := fsm.NewStateFacade(phases)
	// サーバーの初期化
	server := ui.NewStateServer(facade)

	// サーバーの起動（ポート8080で待ち受け）
	log.Println("Starting server on :8080")
	if err := server.Start(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
