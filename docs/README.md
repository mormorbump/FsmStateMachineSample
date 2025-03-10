# State Machine Visualization Sample

このプロジェクトは、looplab/fsmを使用した状態遷移の可視化サンプルアプリケーションです。Observer/Subjectパターンを採用し、効率的な状態管理と通知を実現しています。階層構造を持つフェーズ管理システムにより、複雑な状態遷移を柔軟に表現できます。

## 機能

- 現在の状態をリアルタイムに表示
- 状態間の遷移を視覚的に表現
- シンプルなUIによる状態遷移の制御
- WebSocketを使用したリアルタイム更新
- 最適化された状態遷移制御
- 構造化ログによるデバッグ支援
- 階層構造を持つフェーズ管理
- DTOパターンによるUI層とドメイン層の分離

## プロジェクト構造

```
state_sample/
├── main.go                # エントリーポイント
├── go.mod                # モジュール定義
├── docs/                 # ドキュメント
├── internal/
│   ├── domain/          # ドメイン層
│   │   ├── entity/     # エンティティ
│   │   │   ├── condition_part.go  # 条件パーツ実装
│   │   │   ├── condition.go       # 条件実装
│   │   │   ├── game_state.go      # ゲーム状態定義
│   │   │   ├── phase_facade.go    # フェーズ階層管理
│   │   │   └── phase.go           # フェーズ実装
│   │   ├── service/    # サービス
│   │   │   ├── observer.go        # オブザーバー定義
│   │   │   └── strategy.go        # 戦略パターン定義
│   │   └── value/      # 値オブジェクト
│   │       ├── game_state.go      # 状態値定義
│   │       └── types.go           # 型定義
│   ├── usecase/        # ユースケース層
│   │   ├── state/      # 状態管理
│   │   │   ├── game_facade.go     # ゲーム状態ファサード
│   │   │   └── phase_controller.go # フェーズ制御
│   │   └── strategy/   # 戦略実装
│   │       ├── counter_strategy.go # カウンター戦略
│   │       ├── strategy_factory.go # 戦略ファクトリ
│   │       └── time_strategy.go    # タイマー戦略
│   ├── lib/            # 共通ライブラリ
│   │   └── logger.go   # ロギング機能
│   └── ui/             # UI層
│       ├── dto.go       # データ転送オブジェクト
│       ├── handlers.go  # リクエストハンドラ
│       ├── server.go    # WebSocketサーバー
│       └── static/      # 静的ファイル
│           ├── index.html  # メインページ
│           ├── style.css   # スタイル
│           └── script.js   # クライアントサイドロジック
```

## アーキテクチャ

### 全体アーキテクチャ

```mermaid
graph TD
    Client[クライアント] <-->|WebSocket| UI[UI層]
    UI <-->|DTO| Usecase[ユースケース層]
    Usecase <-->|エンティティ| Domain[ドメイン層]
    
    subgraph "ドメイン層"
        Entity[エンティティ] <--> Service[サービス]
        Entity <--> Value[値オブジェクト]
    end
    
    subgraph "ユースケース層"
        StateFacade[GameFacade] <--> PhaseController[PhaseController]
        PhaseController <--> Strategy[戦略パターン]
    end
    
    subgraph "UI層"
        WebSocket[WebSocketサーバー] <--> Handler[ハンドラー]
        Handler <--> DTO[DTOマッピング]
    end
```

### コアコンポーネント

```mermaid
graph TD
    Core[コアコンポーネント] -->|contains| Observer[オブザーバーパターン]
    Core -->|contains| Strategy[戦略パターン]
    Core -->|contains| FSM[有限状態機械]
    Core -->|contains| Hierarchy[階層構造]
    
    Observer -->|implements| Phase[フェーズ管理]
    Strategy -->|implements| Condition[条件評価]
    FSM -->|controls| StateTransition[状態遷移]
    Hierarchy -->|organizes| PhaseStructure[フェーズ階層]
    
    Phase -->|notifies| Controller[PhaseController]
    Condition -->|notifies| Phase
    StateTransition -->|affects| Phase
    PhaseStructure -->|managed by| PhaseFacade[PhaseFacade]
    
    Controller -->|manages| StateFacade[GameFacade]
    PhaseFacade -->|used by| Controller
```

## 階層構造

### フェーズ階層構造

```mermaid
graph TD
    Root1[ルートフェーズ1] -->|parent-child| Child1[子フェーズ1]
    Root1 -->|parent-child| Child2[子フェーズ2]
    Root2[ルートフェーズ2] -->|parent-child| Child3[子フェーズ3]
    Child2 -->|parent-child| GrandChild1[孫フェーズ1]
    
    subgraph "階層レベル1"
        Root1
        Root2
    end
    
    subgraph "階層レベル2"
        Child1
        Child2
        Child3
    end
    
    subgraph "階層レベル3"
        GrandChild1
    end
```

### フェーズ管理構造

```mermaid
graph TD
    PhaseFacade[PhaseFacade] -->|manages| PhaseMap[PhaseMap]
    PhaseFacade -->|tracks| CurrentPhaseMap[CurrentPhaseMap]
    PhaseFacade -->|stores| AllPhases[AllPhases]
    
    PhaseMap -->|groups by| ParentID[ParentID]
    CurrentPhaseMap -->|tracks active phase for| ParentID
    
    PhaseFacade -->|provides| GetCurrentLeafPhase[GetCurrentLeafPhase]
    PhaseFacade -->|provides| GetPhasesByParentID[GetPhasesByParentID]
    
    PhaseController[PhaseController] -->|uses| PhaseFacade
    PhaseController -->|activates| ActivatePhaseRecursively[ActivatePhaseRecursively]
```

## 状態遷移制御

### 基本シーケンス図

```mermaid
sequenceDiagram
    participant C as Client
    participant S as StateServer
    participant F as GameFacade
    participant PC as PhaseController
    participant PF as PhaseFacade
    participant P as Phase
    participant Cond as Condition
    participant CP as ConditionPart
    participant Strat as Strategy

    C->>S: WebSocket Connect
    S->>F: Get Current State
    F->>PC: Get Current Phase
    PC->>PF: Get Current Phase(0)
    PF-->>PC: Root Phase
    PC-->>F: Phase State
    F-->>S: Current State
    S-->>C: Initial State

    C->>S: Send Command
    S->>F: Execute Action
    F->>PC: Process Command
    PC->>PF: Get Phase By ParentID
    PF-->>PC: Phase
    PC->>P: Activate Phase
    P->>Cond: Activate Condition
    Cond->>CP: Activate ConditionPart
    CP->>Strat: Start Strategy
    Strat-->>CP: Strategy Started
    CP-->>Cond: ConditionPart Activated
    Cond-->>P: Condition Activated
    P-->>PC: Phase Changed
    PC-->>F: Entity Changed
    F-->>S: State Changed
    S-->>C: New State
```

### 階層フェーズ遷移シーケンス図

```mermaid
sequenceDiagram
    participant C as Client
    participant S as StateServer
    participant F as GameFacade
    participant PC as PhaseController
    participant PF as PhaseFacade
    participant RP as RootPhase
    participant CP as ChildPhase

    C->>S: Start Command
    S->>F: Start()
    F->>PC: ActivatePhaseRecursively(rootPhase)
    PC->>RP: Activate()
    RP-->>PC: Phase Activated
    PC->>PF: SetCurrentPhase(rootPhase)
    PC->>RP: HasChildren()
    RP-->>PC: true
    PC->>RP: GetChildren()
    RP-->>PC: [childPhase1, childPhase2]
    PC->>CP: Activate()
    CP-->>PC: Phase Activated
    PC->>PF: SetCurrentPhase(childPhase)
    PC->>CP: HasChildren()
    CP-->>PC: false
    PC-->>F: Activation Complete
    F-->>S: State Changed
    S-->>C: New State with Hierarchy
```

### 条件評価プロセス

```mermaid
sequenceDiagram
    participant C as Client
    participant S as StateServer
    participant F as GameFacade
    participant PC as PhaseController
    participant P as Phase
    participant Cond as Condition
    participant CP as ConditionPart
    participant Strat as Strategy

    C->>S: Send Increment Command
    S->>F: Process Increment
    F->>PC: Forward to Current Phase
    PC->>P: Forward to Condition
    P->>Cond: Forward to ConditionPart
    Cond->>CP: Process Increment
    CP->>Strat: Evaluate Condition
    Strat-->>CP: Condition Satisfied
    CP-->>Cond: ConditionPart Changed
    Cond-->>P: Condition Satisfied
    P->>P: Move to Next State
    P-->>PC: Phase State Changed
    PC-->>F: Entity Changed
    F-->>S: State Changed
    S-->>C: New State
```

## クラス図

### コアエンティティ

```mermaid
classDiagram
    class Phase {
        +ID PhaseID
        +Order int
        -isActive bool
        +IsClear bool
        +Name string
        +Description string
        +Rule GameRule
        +ConditionType ConditionType
        +ConditionIDs []ConditionID
        +SatisfiedConditions map[ConditionID]bool
        +Conditions map[ConditionID]*Condition
        +StartTime *time.Time
        +FinishTime *time.Time
        -fsm *fsm.FSM
        -observers []PhaseObserver
        -mu sync.RWMutex
        -log *zap.Logger
        +ParentID PhaseID
        +Parent *Phase
        +Children []*Phase
        +AutoProgressOnChildrenComplete bool
        +NewPhase(name, order, conditions, conditionType, rule, parentID, autoProgress)
        +OnConditionChanged(condition)
        -checkConditionsSatisfied() bool
        +CurrentState() string
        +GetStateInfo() *GameStateInfo
        +GetConditions() map[ConditionID]*Condition
        +Activate(ctx) error
        +Next(ctx) error
        +Finish(ctx) error
        +Reset(ctx) error
        +AddObserver(observer)
        +RemoveObserver(observer)
        +NotifyPhaseChanged()
        +AddChild(child)
        +GetChildren() Phases
        +HasChildren() bool
        +IsActive() bool
    }

    class PhaseFacade {
        -allPhases Phases
        -phaseMap PhaseMap
        -currentPhaseMap CurrentPhaseMap
        -mu sync.RWMutex
        -log *zap.Logger
        +NewPhaseFacade(phases)
        +GetAllPhases() Phases
        +GetPhaseMap() PhaseMap
        +GetCurrentPhaseMap() CurrentPhaseMap
        +GetCurrentPhase(parentID) *Phase
        +GetCurrentLeafPhase() *Phase
        -FindCurrentLeafPhase(phase) *Phase
        +GetPhasesByParentID(parentID) Phases
        +SetCurrentPhase(phase)
        +ResetCurrentPhaseMap()
    }

    class Phases {
        +Current() *Phase
        +SortByOrder() Phases
        +GetByOrder(order) *Phase
        +GetNextByOrder(currentOrder) *Phase
        +ResetAll(ctx) error
        +ProcessAndActivateByNextOrder(ctx) (*Phase, error)
    }

    class PhaseMap {
        <<map[PhaseID]Phases>>
    }

    class CurrentPhaseMap {
        <<map[PhaseID]*Phase>>
    }

    class Condition {
        +ID ConditionID
        +Label string
        +Kind ConditionKind
        +Parts map[ConditionPartID]*ConditionPart
        +Name string
        +Description string
        +IsClear bool
        +StartTime *time.Time
        +FinishTime *time.Time
        -fsm *fsm.FSM
        -stateObservers []StateObserver
        -condObservers []ConditionObserver
        -mu sync.RWMutex
        -log *zap.Logger
        -satisfiedParts map[ConditionPartID]bool
        +NewCondition(id, label, kind)
        +GetParts() []*ConditionPart
        +OnConditionPartChanged(part)
        -checkAllPartsSatisfied() bool
        +Validate() error
        +CurrentState() string
        +Activate(ctx) error
        +Complete(ctx) error
        +Revert(ctx) error
        +Reset(ctx) error
        +AddPart(part)
        +InitializePartStrategies(factory) error
        +AddObserver(observer)
        +RemoveObserver(observer)
        +NotifyStateChanged()
        +AddConditionObserver(observer)
        +RemoveConditionObserver(observer)
        +NotifyConditionChanged()
    }

    class ConditionPart {
        +ID ConditionPartID
        +Label string
        +ComparisonOperator ComparisonOperator
        +IsClear bool
        +TargetEntityType string
        +TargetEntityID int64
        +ReferenceValueInt int64
        +ReferenceValueFloat float64
        +ReferenceValueString string
        +MinValue int64
        +MaxValue int64
        +Priority int32
        +StartTime *time.Time
        +FinishTime *time.Time
        -fsm *fsm.FSM
        -mu sync.RWMutex
        -log *zap.Logger
        -strategy PartStrategy
        -partObservers []ConditionPartObserver
        +NewConditionPart(id, label)
        +GetReferenceValueInt() int64
        +GetComparisonOperator() ComparisonOperator
        +GetMaxValue() int64
        +GetMinValue() int64
        +IsSatisfied() bool
        +GetCurrentValue() interface
        +OnUpdated(event)
        +Validate() error
        +CurrentState() string
        +Activate(ctx) error
        +Process(ctx, increment) error
        +Complete(ctx) error
        +Timeout(ctx) error
        +Revert(ctx) error
        +Reset(ctx) error
        +SetStrategy(strategy) error
        +AddConditionPartObserver(observer)
        +RemoveConditionPartObserver(observer)
        +NotifyPartChanged()
    }

    Phase "1" *-- "*" Condition
    Condition "1" *-- "*" ConditionPart
    Phase "1" *-- "*" Phase : parent-children
    PhaseFacade "1" *-- "1" Phases : allPhases
    PhaseFacade "1" *-- "1" PhaseMap : phaseMap
    PhaseFacade "1" *-- "1" CurrentPhaseMap : currentPhaseMap
```

### ユースケース層

```mermaid
classDiagram
    class PhaseController {
        -phaseFacade *PhaseFacade
        -observers []ControllerObserver
        -mu sync.RWMutex
        -log *zap.Logger
        +NewPhaseController(phases)
        +OnPhaseChanged(phaseEntity)
        +OnConditionChanged(condition)
        +OnConditionPartChanged(part)
        +GetPhases() Phases
        +ActivatePhaseRecursively(ctx, phase) error
        +Reset(ctx) error
        +AddControllerObserver(observer)
        +RemoveControllerObserver(observer)
        +NotifyEntityChanged(entity)
    }

    class GameFacade {
        -controller *PhaseController
        +NewStateFacade()
        +Start(ctx) error
        +Reset(ctx) error
        +GetCurrentPhase(parentID) *Phase
        +GetCurrentLeafPhase() *Phase
        +GetController() *PhaseController
        +GetConditionPart(conditionID, partID) (*ConditionPart, error)
    }

    class PartStrategy {
        <<interface>>
        +Initialize(part) error
        +GetCurrentValue() interface
        +Start(ctx, part) error
        +Evaluate(ctx, part, params) error
        +Cleanup() error
        +AddObserver(observer)
        +RemoveObserver(observer)
        +NotifyUpdate(event)
    }

    class CounterStrategy {
        -currentValue int64
        -observers []StrategyObserver
        -mu sync.RWMutex
        +NewCounterStrategy()
        +Initialize(part) error
        +GetCurrentValue() interface
        +Start(ctx, part) error
        +Evaluate(ctx, part, params) error
        +Cleanup() error
        +AddObserver(observer)
        +RemoveObserver(observer)
        +NotifyUpdate(event)
    }

    class TimeStrategy {
        -observers []StrategyObserver
        -interval time.Duration
        -isRunning bool
        -ticker *time.Ticker
        -stopChan chan struct
        -mu sync.RWMutex
        -nextTrigger time.Time
        -log *zap.Logger
        +NewTimeStrategy()
        +Initialize(part) error
        +GetCurrentValue() interface
        +Start(ctx, part) error
        +Evaluate(ctx, part, params) error
        +Cleanup() error
        -updateNextTrigger()
        -run()
        +AddObserver(observer)
        +RemoveObserver(observer)
        +NotifyUpdate(event)
    }

    class StrategyFactory {
        +NewStrategyFactory()
        +CreateStrategy(kind) (PartStrategy, error)
    }

    GameFacade "1" *-- "1" PhaseController
    PhaseController "1" *-- "1" PhaseFacade
    CounterStrategy ..|> PartStrategy
    TimeStrategy ..|> PartStrategy
    StrategyFactory ..> PartStrategy : creates
```

### UI層とDTO

```mermaid
classDiagram
    class StateServer {
        -stateFacade *GameFacade
        -clients map[*websocket.Conn]bool
        -upgrader websocket.Upgrader
        -mu sync.RWMutex
        -updateChan chan interface
        -done chan struct{}
        +NewStateServer(facade)
        -processUpdates()
        -sendUpdateToClients(update)
        -getGameStateInfo(phase) *GameStateInfo
        +EditResponse(stateName, currentPhase, stateInfo) Response
        +OnEntityChanged(entityObj)
        -getConditionInfos(phase) []ConditionInfo
        +OnError(err)
        +broadcastUpdate(update)
        +Close() error
    }

    class PhaseDTO {
        +ID PhaseID
        +ParentID PhaseID
        +Name string
        +Description string
        +Order int
        +State string
        +IsClear bool
        +IsActive bool
        +HasChildren bool
        +StartTime *time.Time
        +FinishTime *time.Time
    }

    class ConditionInfo {
        +ID ConditionID
        +Label string
        +State string
        +Kind ConditionKind
        +IsClear bool
        +Description string
        +PhaseID PhaseID
        +PhaseName string
        +Parts []ConditionPartInfo
    }

    class ConditionPartInfo {
        +ID ConditionPartID
        +Label string
        +State string
        +ComparisonOperator ComparisonOperator
        +IsClear bool
        +TargetEntityType string
        +TargetEntityID int64
        +ReferenceValueInt int64
        +ReferenceValueFloat float64
        +ReferenceValueString string
        +MinValue int64
        +MaxValue int64
        +Priority int32
        +CurrentValue interface
    }

    class Response {
        +Type string
        +State string
        +Info *GameStateInfo
        +Phase struct
        +ParentPhase *PhaseInfo
        +ChildPhases []*PhaseInfo
        +Message string
        +Conditions []ConditionInfo
    }

    StateServer ..> PhaseDTO : uses
    StateServer ..> ConditionInfo : uses
    StateServer ..> ConditionPartInfo : uses
    StateServer ..> Response : creates
```

### オブザーバーパターン

```mermaid
classDiagram
    class PhaseObserver {
        <<interface>>
        +OnPhaseChanged(phase)
    }

    class ConditionObserver {
        <<interface>>
        +OnConditionChanged(condition)
    }

    class ConditionPartObserver {
        <<interface>>
        +OnConditionPartChanged(part)
    }

    class ControllerObserver {
        <<interface>>
        +OnEntityChanged(entity)
    }

    class StrategyObserver {
        <<interface>>
        +OnUpdated(event)
    }

    class PhaseController {
        +OnPhaseChanged(phase)
        +OnConditionChanged(condition)
        +OnConditionPartChanged(part)
        +NotifyEntityChanged(entity)
    }

    class StateServer {
        +OnEntityChanged(entity)
    }

    class Phase {
        +AddObserver(observer)
        +RemoveObserver(observer)
        +NotifyPhaseChanged()
    }

    class Condition {
        +AddConditionObserver(observer)
        +RemoveConditionObserver(observer)
        +NotifyConditionChanged()
    }

    class ConditionPart {
        +AddConditionPartObserver(observer)
        +RemoveConditionPartObserver(observer)
        +NotifyPartChanged()
    }

    PhaseController ..|> PhaseObserver
    PhaseController ..|> ConditionObserver
    PhaseController ..|> ConditionPartObserver
    StateServer ..|> ControllerObserver
    ConditionPart ..|> StrategyObserver
```

## 状態遷移図

### フェーズ状態遷移図

```mermaid
stateDiagram-v2
    [*] --> ready: 初期状態
    ready --> active: activate / OnPhaseChanged(StateActive)
    active --> next: next / OnPhaseChanged(StateNext)
    next --> finish: finish / OnPhaseChanged(StateFinish)
    finish --> ready: reset / OnPhaseChanged(StateReady)
    finish --> [*]
```

### 条件状態遷移図

```mermaid
stateDiagram-v2
    [*] --> ready: 初期状態
    ready --> active: activate / OnConditionChanged(StateActive)
    active --> satisfied: evaluate_condition / OnConditionChanged(StateSatisfied)
    satisfied --> [*]
    active --> ready: reset / OnConditionChanged(StateReady)
```

### 条件パート状態遷移図

```mermaid
stateDiagram-v2
    [*] --> ready: 初期状態
    ready --> active: activate / OnConditionPartChanged(StateActive)
    active --> satisfied: evaluate_condition / OnConditionPartChanged(StateSatisfied)
    satisfied --> [*]
    active --> ready: reset / OnConditionPartChanged(StateReady)
```

## 階層フェーズ管理

### 階層構造の状態遷移

```mermaid
stateDiagram-v2
    [*] --> RootReady: 初期状態
    
    state RootReady {
        [*] --> ChildrenInactive
    }
    
    state RootActive {
        [*] --> Child1Active
        Child1Active --> Child2Active: Child1完了
        Child2Active --> [*]: Child2完了
    }
    
    RootReady --> RootActive: activate
    RootActive --> RootNext: すべての子フェーズ完了
    RootNext --> RootFinish: finish
    RootFinish --> [*]
```

### 親子フェーズの連動

```mermaid
graph TD
    Start[開始] --> ActivateRoot[ルートフェーズをアクティブ化]
    ActivateRoot --> HasChildren{子フェーズがある?}
    HasChildren -->|Yes| ActivateChild[最初の子フェーズをアクティブ化]
    HasChildren -->|No| WaitCondition[条件が満たされるのを待つ]
    
    ActivateChild --> ChildHasChildren{子フェーズがある?}
    ChildHasChildren -->|Yes| ActivateGrandchild[最初の孫フェーズをアクティブ化]
    ChildHasChildren -->|No| WaitChildCondition[子フェーズの条件が満たされるのを待つ]
    
    WaitChildCondition --> ChildConditionMet{条件満たされた?}
    ChildConditionMet -->|Yes| NextChild[次の子フェーズへ]
    ChildConditionMet -->|No| WaitChildCondition
    
    NextChild --> HasNextChild{次の子フェーズがある?}
    HasNextChild -->|Yes| ActivateNextChild[次の子フェーズをアクティブ化]
    HasNextChild -->|No| AutoProgress{自動進行設定?}
    
    AutoProgress -->|Yes| ParentNext[親フェーズをnext状態へ]
    AutoProgress -->|No| WaitParentCondition[親フェーズの条件が満たされるのを待つ]
    
    ParentNext --> NextParent[次の親フェーズへ]
```

## DTOとエンティティのマッピング

```mermaid
graph TD
    Entity[ドメインエンティティ] -->|変換| DTO[データ転送オブジェクト]
    DTO -->|シリアライズ| JSON[JSONレスポンス]
    JSON -->|WebSocket| Client[クライアント]
    
    subgraph "ドメイン層"
        Phase[Phase]
        Condition[Condition]
        ConditionPart[ConditionPart]
    end
    
    subgraph "UI層"
        PhaseDTO[PhaseDTO]
        ConditionInfo[ConditionInfo]
        ConditionPartInfo[ConditionPartInfo]
        Response[Response]
    end
    
    Phase -->|ConvertPhaseToDTO| PhaseDTO
    Condition -->|変換| ConditionInfo
    ConditionPart -->|変換| ConditionPartInfo
    
    PhaseDTO -->|組み込み| Response
    ConditionInfo -->|組み込み| Response
    ConditionPartInfo -->|組み込み| Response
```

## WebSocket通信

```mermaid
sequenceDiagram
    participant Client as クライアント
    participant Server as WebSocketサーバー
    participant UpdateChan as 更新チャネル
    participant Processor as 更新プロセッサ
    participant Facade as GameFacade
    
    Client->>Server: WebSocket接続
    Server->>Facade: 初期状態取得
    Facade-->>Server: 現在の状態
    Server-->>Client: 初期状態送信
    
    Client->>Server: コマンド送信
    Server->>Facade: コマンド処理
    Facade->>Facade: 状態更新
    Facade-->>Server: OnEntityChanged通知
    Server->>UpdateChan: 更新メッセージ送信
    UpdateChan->>Processor: メッセージ取得
    Processor->>Server: クライアントに送信
    Server-->>Client: 更新状態送信
    
    Note over Server,Processor: 非同期処理による効率化
```

## 実装詳細

### Observer Pattern

- PhaseObserver, ConditionObserver, ConditionPartObserver, ControllerObserver インターフェース
  - 状態変更通知の受信
  - エラー通知の処理
  - 階層構造の変更通知

- Subject実装
  - オブザーバーの管理
  - 状態変更の通知
  - スレッドセーフな実装

### 階層フェーズ管理

- PhaseFacade
  - 階層構造の初期化と管理
  - 親子関係の構築
  - 現在のフェーズマップの管理

- PhaseController
  - 再帰的なフェーズのアクティブ化
  - 階層間の状態遷移制御
  - 自動進行の管理

### DTOパターン

- PhaseDTO
  - UI層とドメイン層の分離
  - 必要な情報のみの転送
  - 階層構造の表現

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
  "action": "start|reset|increment",
  "payload": {
    "condition_id": 1,
    "part_id": 1,
    "value": 1
  }
}
```

### イベントタイプ

1. コマンド
- start: フェーズの開始
- reset: 状態のリセット
- increment: カウンターの増加

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