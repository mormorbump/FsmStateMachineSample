# UI層の更新設計

## 1. StateServerの更新

### 1.1 依存関係の変更
```go
type StateServer struct {
    stateFacade fsm.StateFacade
    clients     map[*websocket.Conn]bool
    upgrader    websocket.Upgrader
    mu          sync.RWMutex
}
```

### 1.2 コンストラクタの更新
```go
func NewStateServer(facade fsm.StateFacade) *StateServer {
    server := &StateServer{
        stateFacade: facade,
        clients:     make(map[*websocket.Conn]bool),
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true
            },
        },
    }

    // PhaseControllerの監視を設定
    facade.GetController().AddObserver(server)

    return server
}
```

### 1.3 StateObserverインターフェースの実装
```go
// OnStateChanged は状態変更時に呼び出されます
func (s *StateServer) OnStateChanged(state string) {
    currentPhase := s.stateFacade.GetCurrentPhase()
    if currentPhase == nil {
        return
    }

    stateInfo := currentPhase.GetStateInfo()
    update := struct {
        Type    string           `json:"type"`
        State   string           `json:"state"`
        Info    *fsm.GameStateInfo `json:"info,omitempty"`
        Phase   string           `json:"phase"`
        Message string           `json:"message,omitempty"`
    }{
        Type:    "state_change",
        State:   state,
        Info:    stateInfo,
        Phase:   currentPhase.Type,
        Message: stateInfo.Message,
    }
    s.broadcastUpdate(update)
}
```

## 2. ハンドラーの更新

### 2.1 WebSocket接続ハンドラー
```go
func (s *StateServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
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
}
```

### 2.2 制御ハンドラー
```go
func (s *StateServer) handleStart(w http.ResponseWriter, r *http.Request) {
    if err := s.stateFacade.Start(r.Context()); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
}

func (s *StateServer) handleReset(w http.ResponseWriter, r *http.Request) {
    if err := s.stateFacade.Reset(r.Context()); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
}
```

## 3. メリット

1. FSM層との整合性
   - 新しいインターフェースの活用
   - 適切な状態情報の利用
   - 一貫した監視メカニズム

2. シンプルな実装
   - 明確な依存関係
   - 直接的な状態アクセス
   - 効率的な通知

3. 拡張性
   - 新しい状態情報の追加が容易
   - クライアント通知の柔軟性
   - エラーハンドリングの改善

## 4. 実装手順

1. 依存関係の更新
   - import文の修正
   - インターフェースの更新

2. StateServerの実装
   - コンストラクタの更新
   - StateObserver実装の追加
   - 通知メカニズムの調整

3. ハンドラーの更新
   - WebSocket処理の修正
   - 制御APIの更新
   - エラーハンドリングの改善