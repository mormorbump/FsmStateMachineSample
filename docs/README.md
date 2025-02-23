# State Machine Visualization Sample

このプロジェクトは、looplab/fsmを使用した状態遷移の可視化サンプルアプリケーションです。Observer/Subjectパターンを採用し、効率的な状態管理と通知を実現しています。

## 機能

- 現在の状態をリアルタイムに表示
- 状態間の遷移を視覚的に表現
- シンプルなUIによる状態遷移の制御
- WebSocketを使用したリアルタイム更新
- 最適化された状態遷移制御
- 構造化ログによるデバッグ支援

## プロジェクト構造

```
state_sample/
├── main.go                # エントリーポイント
├── go.mod                # モジュール定義
├── docs/                 # ドキュメント
│   └── state_machines_prd/ # 状態遷移の仕様
├── internal/
│   ├── domain/          # ドメイン層
│   │   ├── core/       # コア機能
│   │   │   ├── observer.go    # Observer定義
│   │   │   ├── subject.go     # Subject定義
│   │   │   └── interval_timer.go # タイマー実装
│   │   └── entity/    # エンティティ
│   │       ├── game_state.go  # 状態定義
│   │       └── phase.go       # フェーズ実装
│   ├── usecase/       # ユースケース層
│   │   ├── phase_controller.go # フェーズ制御
│   │   └── state_facade.go    # システムインターフェース
│   ├── lib/           # 共通ライブラリ
│   │   └── logger.go  # ロギング機能
│   └── ui/            # UI層
│       ├── server.go   # WebSocketサーバー
│       ├── handlers.go # リクエストハンドラ
│       └── static/     # 静的ファイル
│           ├── index.html  # メインページ
│           ├── style.css   # スタイル
│           └── script.js   # クライアントサイドロジック
```

## アーキテクチャ

### コアコンポーネント

```mermaid
graph TD
    Core[Core Components] -->|contains| Observer[Observer Pattern]
    Core -->|contains| Timer[Interval Timer]
    Observer -->|implements| Phase[Phase Management]
    Timer -->|notifies| Phase
    Phase -->|notifies| Controller[Phase Controller]
    Controller -->|manages| StateFacade[State Facade]
```

### 状態遷移制御

```mermaid
sequenceDiagram
    participant C as Client
    participant S as StateServer
    participant F as StateFacade
    participant P as Phase

    C->>S: WebSocket Connect
    S->>F: Get Current State
    F->>P: Get State Info
    P-->>S: State Info
    S-->>C: Initial State

    C->>S: Send Command
    S->>F: Execute Action
    F->>P: Update State
    P-->>S: State Changed
    S-->>C: New State
```

## 実装詳細

### Observer Pattern

- StateObserverインターフェース
  - 状態変更通知の受信
  - エラー通知の処理

- Subjectインターフェース
  - オブザーバーの管理
  - 状態変更の通知
  - スレッドセーフな実装

### フェーズ管理

- IntervalTimer
  - 時間間隔の管理
  - イベント通知
  - スレッドセーフな操作

- Phase
  - FSMとの統合
  - タイマーイベントの処理
  - 状態遷移の制御

### 状態遷移の最適化

- 不要な状態遷移の防止
- イベント通知の効率化
- リソース使用の最適化
- デバウンス処理の実装

### ロギング機能

- 構造化ログの採用
- 環境別設定（開発/本番）
- エラートレースの改善
- デバッグ情報の最適化

## 依存関係

- github.com/looplab/fsm
- github.com/gorilla/websocket
- github.com/gorilla/mux
- go.uber.org/zap

## 使用方法

1. サーバーの起動
```bash
go run main.go
```

2. ブラウザでアクセス
```
http://localhost:8080
```

## WebSocket API

### メッセージフォーマット

```json
{
  "type": "command",
  "action": "start|reset",
  "payload": {}
}
```

### イベントタイプ

1. コマンド
- start: フェーズの開始
- reset: 状態のリセット

2. 通知
- state_change: 状態変更の通知
- error: エラーの通知

## エラーハンドリング

### サーバーサイド

- 不正な状態遷移の防止
- WebSocket接続エラーの処理
- リソース管理の最適化
- エラー状態からの復帰

### クライアントサイド

- 接続エラーの処理
- 再接続ロジック
- エラー表示の実装

## パフォーマンス最適化

### 状態遷移

- イベント通知の効率化
- 不要な遷移の防止
- リソース使用の最適化

### メモリ管理

- オブザーバーの適切な解放
- WebSocket接続の管理
- リソースのクリーンアップ

## 今後の展開

1. テスト強化
- 単体テストの拡充
- 統合テストの追加
- パフォーマンステスト

2. 機能拡張
- 認証機能の追加
- セッション管理
- UI機能の強化

3. パフォーマンス改善
- さらなる最適化
- スケーラビリティの向上
- モニタリングの強化