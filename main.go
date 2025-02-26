package main

import (
	logger "state_sample/internal/lib"
	"state_sample/internal/ui"
	"state_sample/internal/usecase/state"

	"go.uber.org/zap"
)

func main() {
	log := logger.DefaultLogger()
	facade := state.NewStateFacade()
	// サーバーの初期化
	server := ui.NewStateServer(facade)

	// サーバーの起動（ポート8080で待ち受け）
	log.Debug("Starting server on :8080")
	if err := server.Start(":8080"); err != nil {
		log.Error("Server error", zap.Error(err))
	}
}
