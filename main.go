package main

import (
	"log"

	"state_sample/internal/ui"
)

func main() {
	// サーバーの初期化
	server := ui.NewStateServer()

	// サーバーの起動（ポート8080で待ち受け）
	log.Println("Starting server on :8080")
	if err := server.Start(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}