# WebSocket通信のブロッキング問題解決

## 問題の分析

### 現状の問題点
1. WebSocketメッセージ処理が同期的
2. handleActionRequestがメインゴルーチンをブロック
3. Start/Reset処理が同期的に実行

### 影響
- 他のメッセージ処理が遅延
- UIのレスポンス性が低下
- システム全体のパフォーマンスが低下

## 解決策

### 1. メッセージ処理の非同期化
```go
// WebSocketメッセージ受信処理
func (s *StateServer) recvWsMessage(conn *websocket.Conn) error {
    for {
        var msg struct {
            Event string `json:"event"`
        }
        if err := conn.ReadJSON(&msg); err != nil {
            return err
        }

        // 非同期でアクションを処理
        go s.processAction(msg.Event)
    }
}

// アクション処理を別ゴルーチンで実行
func (s *StateServer) processAction(event string) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := s.handleActionRequest(ctx, event); err != nil {
        s.OnError(err)
    }
}
```

### 2. コンテキスト制御
```go
func (s *StateServer) handleActionRequest(ctx context.Context, action string) error {
    switch action {
    case "start", "activate":
        return s.stateFacade.Start(ctx)
    case "stop", "reset", "finish":
        return s.stateFacade.Reset(ctx)
    default:
        return fmt.Errorf("invalid action: %s", action)
    }
}
```

## 期待される効果

### 1. パフォーマンス改善
- WebSocketメッセージ処理のブロッキング解消
- UIのレスポンス性向上
- 並行処理の効率化

### 2. 安定性向上
- タイムアウト制御による信頼性向上
- エラーハンドリングの改善
- デッドロックリスクの軽減

### 3. 保守性向上
- 処理の分離による可読性向上
- エラー追跡の容易化
- コードの構造化

## 実装手順

1. 非同期処理の導入
   - recvWsMessageの修正
   - processActionの実装
   - goroutineの適切な管理

2. コンテキスト制御の実装
   - タイムアウト設定の追加
   - エラーハンドリングの改善
   - リソース管理の最適化

3. テストと検証
   - 並行処理のテスト
   - エラーケースの確認
   - パフォーマンス測定

## 注意点

1. goroutineの管理
   - 適切なエラーハンドリング
   - リソースリークの防止
   - 終了処理の確実な実行

2. エラー処理
   - エラーの適切な伝播
   - ユーザーへの通知
   - ログ出力の充実

3. パフォーマンス
   - goroutineの適切な数の管理
   - メモリ使用量の監視
   - 負荷テストの実施