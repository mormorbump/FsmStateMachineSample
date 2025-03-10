# 状態遷移システム詳細解説

## はじめに

本文書では、状態遷移の可視化サンプルアプリケーションの設計思想、アーキテクチャ、実装詳細について解説します。このシステムは、looplab/fsmライブラリを基盤とし、Observer/Subjectパターンを採用することで効率的な状態管理と通知を実現しています。さらに、階層構造を持つフェーズ管理システムにより、複雑な状態遷移を柔軟に表現できる設計となっています。

## システムの目的と背景

### 目的

このシステムの主な目的は以下の通りです：

1. 複雑な状態遷移を視覚的に表現し、理解を容易にする
2. 階層構造を持つフェーズ管理により、複雑なビジネスロジックを表現する
3. リアルタイムな状態更新をクライアントに提供する
4. 拡張性と保守性の高いアーキテクチャを実現する
5. 効率的な状態遷移と通知メカニズムを提供する

### 背景

多くのアプリケーションでは、状態遷移の管理が複雑になりがちです。特にゲームやワークフローシステムなどでは、複数の条件が絡み合い、階層的な状態構造を持つことがあります。このような複雑な状態遷移を管理するためには、単純なステートマシンだけでなく、階層構造や条件評価の仕組みが必要となります。

本システムは、このような複雑な状態遷移を効率的に管理し、視覚化するためのサンプル実装として開発されました。特に、以下の課題に対応することを目指しています：

- 複数の条件が組み合わさった状態遷移の管理
- 親子関係を持つ階層的なフェーズの管理
- リアルタイムな状態変更の通知
- 拡張性の高い設計による新しい条件タイプや戦略の追加のしやすさ

## アーキテクチャ概要

本システムは、クリーンアーキテクチャの考え方を取り入れ、以下の3つの主要な層で構成されています：

1. **ドメイン層**：ビジネスロジックの中核となるエンティティと値オブジェクトを定義
2. **ユースケース層**：ドメイン層のエンティティを操作するユースケースを実装
3. **UI層**：ユーザーインターフェースとWebSocket通信を担当

各層の責務を明確に分離することで、テスト容易性と保守性を高めています。また、依存関係は内側に向かうように設計されており、ドメイン層は他の層に依存せず、ユースケース層はドメイン層にのみ依存し、UI層はユースケース層とドメイン層に依存するという構造になっています。

### ドメイン層

ドメイン層は、システムの中核となるビジネスロジックを表現するエンティティと値オブジェクトで構成されています。主要なコンポーネントは以下の通りです：

- **Phase**：フェーズを表すエンティティ。階層構造を持ち、条件の集合を管理します。
- **Condition**：条件を表すエンティティ。複数の条件パーツから構成されます。
- **ConditionPart**：条件の最小単位を表すエンティティ。戦略パターンを用いて条件の評価方法を実装します。
- **PhaseFacade**：フェーズの階層構造を管理するファサード。

ドメイン層では、以下の設計パターンを活用しています：

- **オブザーバーパターン**：状態変更の通知を効率的に行うために使用
- **戦略パターン**：条件の評価方法を柔軟に切り替えるために使用
- **ファサードパターン**：複雑な階層構造の管理を簡素化するために使用

### ユースケース層

ユースケース層は、ドメイン層のエンティティを操作するユースケースを実装しています。主要なコンポーネントは以下の通りです：

- **PhaseController**：フェーズの制御を担当するコントローラー。フェーズの状態変更を監視し、適切な処理を行います。
- **GameFacade**：ゲーム状態全体を管理するファサード。クライアントからのリクエストを処理し、適切なフェーズやコントローラーに委譲します。
- **StrategyFactory**：条件評価の戦略を生成するファクトリー。

ユースケース層では、以下の設計パターンを活用しています：

- **ファクトリーパターン**：戦略オブジェクトの生成を担当
- **コントローラーパターン**：フェーズの制御ロジックをカプセル化
- **ファサードパターン**：複雑なユースケースの操作を簡素化

### UI層

UI層は、ユーザーインターフェースとWebSocket通信を担当しています。主要なコンポーネントは以下の通りです：

- **StateServer**：WebSocketサーバーを実装し、クライアントとの通信を管理します。
- **DTO**：データ転送オブジェクト。ドメインオブジェクトとクライアントの間でデータを変換します。
- **Handlers**：HTTPリクエストを処理するハンドラー。

UI層では、以下の設計パターンを活用しています：

- **DTOパターン**：ドメインオブジェクトとクライアント間のデータ変換を担当
- **オブザーバーパターン**：エンティティの変更をリアルタイムにクライアントに通知
- **非同期処理パターン**：WebSocketメッセージの効率的な処理を実現

## 階層構造を持つフェーズ管理システム

本システムの特徴的な機能の一つが、階層構造を持つフェーズ管理システムです。この仕組みにより、複雑なビジネスロジックを階層的に表現し、管理することが可能になっています。

### フェーズの階層構造

フェーズは以下のような階層構造を持ちます：

1. **ルートフェーズ**：最上位のフェーズ。親IDが0のフェーズがルートフェーズとなります。
2. **子フェーズ**：ルートフェーズの子となるフェーズ。親IDがルートフェーズのIDとなります。
3. **孫フェーズ**：子フェーズの子となるフェーズ。親IDが子フェーズのIDとなります。

この階層構造により、以下のような複雑なシナリオを表現することが可能です：

- ゲームのステージとそのサブステージ
- ワークフローの大きなフェーズとその詳細なステップ
- 複数の条件が絡み合う複雑なビジネスロジック

### フェーズの親子関係の管理

フェーズの親子関係は、`PhaseFacade`クラスによって管理されています。主要な機能は以下の通りです：

1. **階層構造の初期化**：`InitializePhaseHierarchy`メソッドにより、フェーズの親子関係を初期化します。
2. **親IDによるグループ化**：`GroupPhasesByParentID`メソッドにより、フェーズを親IDごとにグループ化します。
3. **現在のフェーズマップの管理**：`CurrentPhaseMap`により、各階層レベルでの現在アクティブなフェーズを管理します。
4. **最下層のフェーズの検索**：`GetCurrentLeafPhase`メソッドにより、現在アクティブな最下層のフェーズを取得します。

### フェーズの再帰的なアクティブ化

フェーズの階層構造を活かした重要な機能として、フェーズの再帰的なアクティブ化があります。これは、`PhaseController`クラスの`ActivatePhaseRecursively`メソッドによって実装されています。

このメソッドは、以下のような処理を行います：

1. 指定されたフェーズをアクティブ化
2. そのフェーズを現在のフェーズとして設定
3. フェーズに子フェーズがある場合、最初の子フェーズを再帰的にアクティブ化

この仕組みにより、親フェーズがアクティブになると、自動的にその最初の子フェーズもアクティブになるという階層的な状態遷移が実現されています。

### 自動進行の仕組み

階層構造を持つフェーズ管理のもう一つの特徴は、子フェーズの完了に基づく親フェーズの自動進行です。これは、`AutoProgressOnChildrenComplete`フラグによって制御されています。

このフラグがtrueに設定されている親フェーズの場合、最後の子フェーズが完了すると、親フェーズは自動的に次の状態に進みます。これにより、子フェーズの完了を条件とする親フェーズの状態遷移が可能になります。

具体的な処理は、`PhaseController`クラスの`OnPhaseChanged`メソッド内で行われています。子フェーズが`next`状態になった際に、次の子フェーズがない場合、親フェーズの`AutoProgressOnChildrenComplete`フラグをチェックし、trueであれば親フェーズを次の状態に進めます。

## オブザーバーパターンの実装

本システムでは、状態変更の通知を効率的に行うために、オブザーバーパターンを広範囲に活用しています。これにより、状態変更が発生した際に、関連するコンポーネントに自動的に通知することが可能になっています。

### オブザーバーインターフェース

システムには、以下の主要なオブザーバーインターフェースが定義されています：

1. **PhaseObserver**：フェーズの状態変更を監視するインターフェース
   - `OnPhaseChanged(phase)`：フェーズの状態が変更された際に呼び出されるメソッド

2. **ConditionObserver**：条件の状態変更を監視するインターフェース
   - `OnConditionChanged(condition)`：条件の状態が変更された際に呼び出されるメソッド

3. **ConditionPartObserver**：条件パーツの状態変更を監視するインターフェース
   - `OnConditionPartChanged(part)`：条件パーツの状態が変更された際に呼び出されるメソッド

4. **ControllerObserver**：コントローラーからのエンティティ変更を監視するインターフェース
   - `OnEntityChanged(entity)`：エンティティが変更された際に呼び出されるメソッド

5. **StrategyObserver**：戦略からの更新を監視するインターフェース
   - `OnUpdated(event)`：戦略が更新された際に呼び出されるメソッド

### 通知の流れ

オブザーバーパターンによる通知の流れは以下の通りです：

1. **条件パーツの変更**：
   - 条件パーツの状態が変更されると、`NotifyPartChanged`メソッドが呼び出されます。
   - 登録されている`ConditionPartObserver`（通常は条件オブジェクト）に通知されます。

2. **条件の変更**：
   - 条件パーツからの通知を受けた条件オブジェクトは、状態を更新します。
   - 条件が満たされると、`NotifyConditionChanged`メソッドが呼び出されます。
   - 登録されている`ConditionObserver`（通常はフェーズオブジェクト）に通知されます。

3. **フェーズの変更**：
   - 条件からの通知を受けたフェーズオブジェクトは、状態を更新します。
   - フェーズの状態が変更されると、`NotifyPhaseChanged`メソッドが呼び出されます。
   - 登録されている`PhaseObserver`（通常はPhaseController）に通知されます。

4. **コントローラーからの通知**：
   - フェーズからの通知を受けたPhaseControllerは、適切な処理を行います。
   - エンティティの変更があると、`NotifyEntityChanged`メソッドが呼び出されます。
   - 登録されている`ControllerObserver`（通常はStateServer）に通知されます。

5. **クライアントへの通知**：
   - コントローラーからの通知を受けたStateServerは、WebSocketを通じてクライアントに通知します。

### スレッドセーフな実装

オブザーバーパターンの実装では、複数のゴルーチンからの同時アクセスに対応するため、スレッドセーフな実装が行われています。具体的には、以下の対策が取られています：

1. **ミューテックスの使用**：各オブザーバーリストへのアクセスは、ミューテックス（`sync.RWMutex`）によって保護されています。
2. **オブザーバーリストのコピー**：通知時には、オブザーバーリストのコピーを作成してから処理を行うことで、通知中にリストが変更されても問題が発生しないようにしています。
3. **非同期通知**：WebSocketクライアントへの通知は、チャネルを使用した非同期処理によって行われ、通知処理がブロックされないようにしています。

## 戦略パターンの実装

本システムでは、条件の評価方法を柔軟に切り替えるために、戦略パターンを活用しています。これにより、異なる種類の条件（時間ベース、カウンターベースなど）を統一的なインターフェースで扱うことが可能になっています。

### 戦略インターフェース

戦略パターンの中核となるのは、`PartStrategy`インターフェースです。このインターフェースは、以下のメソッドを定義しています：

```go
type PartStrategy interface {
    Initialize(part *ConditionPart) error
    GetCurrentValue() interface{}
    Start(ctx context.Context, part *ConditionPart) error
    Evaluate(ctx context.Context, part *ConditionPart, params interface{}) error
    Cleanup() error
    AddObserver(observer StrategyObserver)
    RemoveObserver(observer StrategyObserver)
    NotifyUpdate(event interface{})
}
```

このインターフェースを実装することで、異なる種類の条件評価戦略を作成することができます。

### 具体的な戦略実装

システムには、以下の具体的な戦略実装が含まれています：

1. **CounterStrategy**：カウンターベースの条件評価を行う戦略
   - カウンターの値を管理し、指定された閾値と比較して条件の満足を判定します。
   - `Evaluate`メソッドでは、カウンターの値を増加させ、条件を評価します。

2. **TimeStrategy**：時間ベースの条件評価を行う戦略
   - タイマーを使用して、指定された時間が経過したかどうかを判定します。
   - `Start`メソッドでタイマーを開始し、指定時間後に条件を満たしたと判定します。

### 戦略ファクトリー

戦略オブジェクトの生成は、`StrategyFactory`クラスによって行われます。このファクトリーは、条件の種類に応じて適切な戦略オブジェクトを生成する役割を担っています。

```go
func (f *StrategyFactory) CreateStrategy(kind value.ConditionKind) (service.PartStrategy, error) {
    switch kind {
    case value.KindCounter:
        return strategy.NewCounterStrategy(), nil
    case value.KindTime:
        return strategy.NewTimeStrategy(), nil
    default:
        return nil, fmt.Errorf("unknown condition kind: %v", kind)
    }
}
```

この仕組みにより、新しい種類の条件評価戦略を追加する際には、新しい戦略クラスを実装し、ファクトリーに登録するだけで済むようになっています。

### 戦略の使用方法

条件パーツは、初期化時に戦略ファクトリーを使用して適切な戦略オブジェクトを取得し、`SetStrategy`メソッドで設定します。

```go
func (c *Condition) InitializePartStrategies(factory service.StrategyFactory) error {
    for _, part := range c.Parts {
        strategy, err := factory.CreateStrategy(c.Kind)
        if err != nil {
            return err
        }
        if err := part.SetStrategy(strategy); err != nil {
            return err
        }
    }
    return nil
}
```

条件パーツは、戦略オブジェクトを使用して条件の評価を行います。例えば、`Process`メソッドでは、戦略の`Evaluate`メソッドを呼び出して条件を評価します。

```go
func (p *ConditionPart) Process(ctx context.Context, increment interface{}) error {
    if p.strategy == nil {
        return fmt.Errorf("strategy is not set")
    }
    return p.strategy.Evaluate(ctx, p, increment)
}
```

## DTOパターンの実装

本システムでは、ドメインオブジェクトとクライアント間のデータ変換を効率的に行うために、DTOパターン（Data Transfer Object）を採用しています。これにより、ドメイン層の内部構造をクライアントに露出させることなく、必要な情報のみを転送することが可能になっています。

### DTOの定義

システムには、以下の主要なDTOが定義されています：

1. **PhaseDTO**：フェーズの情報をクライアントに送信するためのDTO
   ```go
   type PhaseDTO struct {
       ID          value.PhaseID `json:"id"`
       ParentID    value.PhaseID `json:"parent_id"`
       Name        string        `json:"name"`
       Description string        `json:"description"`
       Order       int           `json:"order"`
       State       string        `json:"state"`
       IsClear     bool          `json:"is_clear"`
       IsActive    bool          `json:"is_active"`
       HasChildren bool          `json:"has_children"`
       StartTime   *time.Time    `json:"start_time,omitempty"`
       FinishTime  *time.Time    `json:"finish_time,omitempty"`
   }
   ```

2. **ConditionInfo**：条件の情報をクライアントに送信するためのDTO
   ```go
   type ConditionInfo struct {
       ID          value.ConditionID   `json:"id"`
       Label       string              `json:"label"`
       State       string              `json:"state"`
       Kind        value.ConditionKind `json:"kind"`
       IsClear     bool                `json:"is_clear"`
       Description string              `json:"description"`
       PhaseID     value.PhaseID       `json:"phase_id"`
       PhaseName   string              `json:"phase_name"`
       Parts       []ConditionPartInfo `json:"parts"`
   }
   ```

3. **ConditionPartInfo**：条件パーツの情報をクライアントに送信するためのDTO
   ```go
   type ConditionPartInfo struct {
       ID                   value.ConditionPartID    `json:"id"`
       Label                string                   `json:"label"`
       State                string                   `json:"state"`
       ComparisonOperator   value.ComparisonOperator `json:"comparison_operator"`
       IsClear              bool                     `json:"is_clear"`
       TargetEntityType     string                   `json:"target_entity_type"`
       TargetEntityID       int64                    `json:"target_entity_id"`
       ReferenceValueInt    int64                    `json:"reference_value_int"`
       ReferenceValueFloat  float64                  `json:"reference_value_float"`
       ReferenceValueString string                   `json:"reference_value_string"`
       MinValue             int64                    `json:"min_value"`
       MaxValue             int64                    `json:"max_value"`
       Priority             int32                    `json:"priority"`
       CurrentValue         interface{}              `json:"current_value"`
   }
   ```

### エンティティからDTOへの変換

エンティティからDTOへの変換は、専用の変換関数によって行われます。例えば、`ConvertPhaseToDTO`関数は、Phaseエンティティを対応するDTOに変換します。

```go
func ConvertPhaseToDTO(phase *entity.Phase) PhaseDTO {
    return PhaseDTO{
        ID:          phase.ID,
        ParentID:    phase.ParentID,
        Name:        phase.Name,
        Description: phase.Description,
        Order:       phase.Order,
        State:       phase.CurrentState(),
        IsClear:     phase.IsClear,
        IsActive:    phase.IsActive(),
        HasChildren: phase.HasChildren(),
        StartTime:   phase.StartTime,
        FinishTime:  phase.FinishTime,
    }
}
```

同様に、`getConditionInfos`メソッドは、条件エンティティを対応するDTOに変換します。

```go
func (s *StateServer) getConditionInfos(phase *entity.Phase) []ConditionInfo {
    conditions := make([]ConditionInfo, 0)
    for _, condition := range phase.GetConditions() {
        condInfo := ConditionInfo{
            ID:          condition.ID,
            Label:       condition.Label,
            State:       condition.CurrentState(),
            Kind:        condition.Kind,
            IsClear:     condition.IsClear,
            Description: condition.Description,
            PhaseID:     phase.ID,
            PhaseName:   phase.Name,
            Parts:       make([]ConditionPartInfo, 0),
        }
        for _, part := range condition.GetParts() {
            partInfo := ConditionPartInfo{
                ID:                   part.ID,
                Label:                part.Label,
                State:                part.CurrentState(),
                ComparisonOperator:   part.ComparisonOperator,
                IsClear:              part.IsClear,
                TargetEntityType:     part.TargetEntityType,
                TargetEntityID:       part.TargetEntityID,
                ReferenceValueInt:    part.ReferenceValueInt,
                ReferenceValueFloat:  part.ReferenceValueFloat,
                ReferenceValueString: part.ReferenceValueString,
                MinValue:             part.MinValue,
                MaxValue:             part.MaxValue,
                Priority:             part.Priority,
                CurrentValue:         part.GetCurrentValue(),
            }
            condInfo.Parts = append(condInfo.Parts, partInfo)
        }
        conditions = append(conditions, condInfo)
    }
    return conditions
}
```

### DTOの利点

DTOパターンを採用することで、以下のような利点があります：

1. **ドメイン層の保護**：ドメイン層の内部構造をクライアントに露出させることなく、必要な情報のみを転送できます。
2. **データ転送の最適化**：クライアントに必要な情報のみを含むDTOを作成することで、データ転送量を削減できます。
3. **バージョニングの容易さ**：APIの変更が必要な場合、ドメインモデルを変更せずにDTOのみを変更することができます。
4. **クライアント固有の表現**：クライアントに適した形式でデータを表現することができます。

## WebSocket通信の実装

本システムでは、クライアントとのリアルタイム通信を実現するために、WebSocketを採用しています。これにより、状態変更をリアルタイムにクライアントに通知することが可能になっています。

### WebSocketサーバーの実装

WebSocket通信は、`StateServer`クラスによって実装されています。このクラスは、以下の主要なコンポーネントで構成されています：

1. **クライアント管理**：接続中のWebSocketクライアントを管理するためのマップ
2. **アップグレーダー**：HTTPリクエストをWebSocket接続にアップグレードするためのコンポーネント
3. **更新チャネル**：非同期で更新メッセージを処理するためのチャネル
4. **終了チャネル**：サーバー終了を通知するためのチャネル

```go
type StateServer struct {
    stateFacade *state.GameFacade
    clients     map[*websocket.Conn]bool
    upgrader    websocket.Upgrader
    mu          sync.RWMutex
    updateChan  chan interface{}
    done        chan struct{}
}
```

### 非同期メッセージ処理

WebSocketメッセージの処理は、非同期で行われます。これにより、メッセージ送信がブロックされることなく、効率的な処理が可能になっています。

具体的には、以下のような仕組みが実装されています：

1. **更新チャネル**：エンティティの変更通知を受け取ると、更新メッセージが更新チャネルに送信されます。
2. **更新プロセッサ**：別のゴルーチンで動作する更新プロセッサが、更新チャネルからメッセージを取り出し、クライアントに送信します。

```go
func (s *StateServer) processUpdates() {
    log := logger.DefaultLogger()
    log.Debug("Starting update processor goroutine")

    for {
        select {
        case update := <-s.updateChan:
            // 実際の更新処理を行う
            s.sendUpdateToClients(update)
        case <-s.done:
            log.Debug("Update processor goroutine shutting down")
            return
        }
    }
}
```

### クライアントへの通知

クライアントへの通知は、`broadcastUpdate`メソッドによって行われます。このメソッドは、更新メッセージを更新チャネルに送信します。

```go
func (s *StateServer) broadcastUpdate(update interface{}) {
    log := logger.DefaultLogger()
    log.Debug("Queueing update for broadcast", zap.Any("update", update))

    // 更新メッセージをチャネルに送信（非ブロッキング）
    select {
    case s.updateChan <- update:
        // メッセージが正常にキューに入った
    default:
        // チャネルがいっぱいの場合
        log.Warn("Update channel is full, dropping message")
    }
}
```

実際のクライアントへの送信は、`sendUpdateToClients`メソッドによって行われます。このメソッドは、すべてのクライアントに対してメッセージを送信します。

```go
func (s *StateServer) sendUpdateToClients(update interface{}) {
    log := logger.DefaultLogger()
    s.mu.RLock()
    defer s.mu.RUnlock()

    log.Debug("Sending update to clients", zap.Any("update", update))
    for client := range s.clients {
        if err := client.WriteJSON(update); err != nil {
            log.Error("Error sending message to client", zap.Error(err))
            err := client.Close()
            if err != nil {
                log.Error("Error closing client connection", zap.Error(err))
            }
            delete(s.clients, client)
        }
    }
}
```

### エンティティ変更の通知

`StateServer`は、`ControllerObserver`インターフェースを実装しており、エンティティの変更通知を受け取ることができます。エンティティの変更通知を受け取ると、適切なレスポンスを構築し、クライアントに通知します。

```go
func (s *StateServer) OnEntityChanged(entityObj interface{}) {
    // エンティティの変更通知を処理
    // ...

    // レスポンスを構築
    response := struct {
        Type         string          `json:"type"`
        Phases       []PhaseDTO      `json:"phases"`
        CurrentPhase *PhaseDTO       `json:"current_phase,omitempty"`
        State        string          `json:"state"`
        Info         *GameStateInfo  `json:"info,omitempty"`
        Message      string          `json:"message,omitempty"`
        Conditions   []ConditionInfo `json:"conditions"`
    }{
        // ...
    }

    // クライアントに通知
    s.broadcastUpdate(response)
}
```

## 状態遷移の最適化

本システムでは、状態遷移の効率化と最適化のために、いくつかの工夫が施されています。これにより、不要な状態遷移を防止し、リソース使用を最適化しています。

### 不要な状態遷移の防止

不要な状態遷移を防止するために、以下の対策が取られています：

1. **状態チェック**：状態遷移を行う前に、現在の状態をチェックし、不要な遷移を防止しています。

```go
func (p *Phase) Reset(ctx context.Context) error {
    if p.CurrentState() == value.StateReady {
        return nil
    }
    // ...
}
```

2. **条件の満足状態の管理**：条件が満たされたかどうかを管理し、同じ条件が複数回満たされないようにしています。

```go
func (p *Phase) OnConditionChanged(condition interface{}) {
    // ...
    p.mu.Lock()
    p.SatisfiedConditions[cond.ID] = true
    satisfied := p.checkConditionsSatisfied()
    // ...
}
```

### イベント通知の効率化

イベント通知の効率化のために、以下の対策が取られています：

1. **オブザーバーリストのコピー**：通知時には、オブザーバーリストのコピーを作成してから処理を行うことで、通知中にリストが変更されても問題が発生しないようにしています。

```go
func (p *Phase) NotifyPhaseChanged() {
    // ...
    p.mu.RLock()
    observers := make([]service.PhaseObserver, len(p.observers))
    copy(observers, p.observers)
    p.mu.RUnlock()

    for i, observer := range observers {
        // ...
        observer.OnPhaseChanged(p)
    }
}
```

2. **非同期通知**：WebSocketクライアントへの通知は、チャネルを使用した非同期処理によって行われ、通知処理がブロックされないようにしています。

```go
func (s *StateServer) broadcastUpdate(update interface{}) {
    // ...
    select {
    case s.updateChan <- update:
        // メッセージが正常にキューに入った
    default:
        // チャネルがいっぱいの場合
        log.Warn("Update channel is full, dropping message")
    }
}
```

### リソース使用の最適化

リソース使用の最適化のために、以下の対策が取られています：

1. **リソースのクリーンアップ**：不要になったリソースは適切にクリーンアップされ、メモリリークを防止しています。

```go
func (s *StateServer) Close() error {
    // ...
    close(s.done)
    // ...
    for client := range s.clients {
        if err := client.Close(); err != nil {
            log.Error("Error closing client", zap.Error(err))
        }
    }
    s.clients = nil
    return nil
}
```

2. **タイマーの適切な管理**：タイマーは使用後に適切に停止され、リソースの無駄遣いを防止しています。

```go
func (t *TimeStrategy) Cleanup() error {
    t.mu.Lock()
    defer t.mu.Unlock()

    if t.isRunning {
        t.stopChan <- struct{}{}
        t.isRunning = false
    }
    return nil
}
```

### デバウンス処理の実装

状態変更が頻繁に発生する場合に、クライアントへの通知が過剰にならないようにするために、デバウンス処理が実装されています。

具体的には、更新チャネルのバッファサイズを制限し、チャネルがいっぱいの場合は新しいメッセージをドロップすることで、クライアントへの通知頻度を制限しています。

```go
func NewStateServer(facade *state.GameFacade) *StateServer {
    // ...
    server := &StateServer{
        // ...
        updateChan: make(chan interface{}, 100), // バッファ付きチャネルを作成
        // ...
    }
    // ...
}
```

## エラーハンドリング

本システムでは、様々な状況でのエラーを適切に処理するために、包括的なエラーハンドリングの仕組みが実装されています。これにより、システムの安定性と信頼性を確保しています。

### サーバーサイドのエラーハンドリング

サーバーサイドでは、以下のようなエラーハンドリングが行われています：

1. **不正な状態遷移の防止**：FSMの状態遷移ルールに従わない遷移が要求された場合、適切なエラーが返されます。

```go
func (p *Phase) Next(ctx context.Context) error {
    return p.fsm.Event(ctx, value.EventNext)
}
```

2. **WebSocket接続エラーの処理**：WebSocket接続でエラーが発生した場合、接続を閉じ、クライアントリストから削除します。

```go
func (s *StateServer) sendUpdateToClients(update interface{}) {
    // ...
    for client := range s.clients {
        if err := client.WriteJSON(update); err != nil {
            log.Error("Error sending message to client", zap.Error(err))
            err := client.Close()
            if err != nil {
                log.Error("Error closing client connection", zap.Error(err))
            }
            delete(s.clients, client)
        }
    }
}
```

3. **リソース管理の最適化**：エラーが発生した場合でも、リソースが適切に解放されるようにしています。

```go
func (s *StateServer) Close() error {
    // ...
    for client := range s.clients {
        if err := client.Close(); err != nil {
            log.Error("Error closing client", zap.Error(err))
        }
    }
    // ...
}
```

4. **エラー状態からの復帰**：エラーが発生した場合でも、システムが適切に復帰できるようにしています。

```go
func (pc *PhaseController) OnPhaseChanged(phaseEntity interface{}) {
    // ...
    // 現在のフェーズを終了
    if err := phase.Finish(ctx); err != nil {
        pc.log.Error("Failed to finish current phase", zap.Error(err))
        // エラーが発生しても次のフェーズに進む試みをする
    }
    // ...
}
```

### クライアントサイドのエラーハンドリング

クライアントサイドでは、以下のようなエラーハンドリングが行われています：

1. **接続エラーの処理**：WebSocket接続でエラーが発生した場合、再接続を試みます。

```javascript
function connectWebSocket() {
    const socket = new WebSocket(wsUrl);
    
    socket.onclose = function(event) {
        console.log('WebSocket connection closed');
        // 再接続を試みる
        setTimeout(function() {
            connectWebSocket();
        }, 1000);
    };
    
    socket.onerror = function(error) {
        console.error('WebSocket error:', error);
        socket.close();
    };
    
    // ...
}
```

2. **エラー表示の実装**：サーバーからエラーメッセージを受信した場合、ユーザーに適切に表示します。

```javascript
socket.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    if (data.type === 'error') {
        // エラーメッセージを表示
        showError(data.error);
    } else {
        // 通常のメッセージ処理
        // ...
    }
};
```

### エラーログの記録

システム全体で、エラーが発生した場合には適切にログに記録されるようになっています。これにより、問題の診断と解決が容易になります。

```go
func (pc *PhaseController) OnPhaseChanged(phaseEntity interface{}) {
    // ...
    if err := phase.Finish(ctx); err != nil {
        pc.log.Error("Failed to finish current phase", zap.Error(err))
        // ...
    }
    // ...
}
```

## パフォーマンス最適化

本システムでは、高いパフォーマンスを実現するために、様々な最適化が施されています。これにより、リソース使用を最小限に抑えつつ、効率的な処理が可能になっています。

### 状態遷移の最適化

状態遷移の最適化のために、以下の対策が取られています：

1. **不要な状態遷移の防止**：状態遷移を行う前に、現在の状態をチェックし、不要な遷移を防止しています。

```go
func (p *Phase) Reset(ctx context.Context) error {
    if p.CurrentState() == value.StateReady {
        return nil
    }
    // ...
}
```

2. **イベント通知の効率化**：オブザーバーリストのコピーを作成してから処理を行うことで、通知中にリストが変更されても問題が発生しないようにしています。

```go
func (p *Phase) NotifyPhaseChanged() {
    // ...
    p.mu.RLock()
    observers := make([]service.PhaseObserver, len(p.observers))
    copy(observers, p.observers)
    p.mu.RUnlock()
    // ...
}
```

### メモリ管理の最適化

メモリ管理の最適化のために、以下の対策が取られています：

1. **オブザーバーの適切な解放**：不要になったオブザーバーは適切に解放され、メモリリークを防止しています。

```go
func (p *Phase) RemoveObserver(observer service.PhaseObserver) {
    if observer == nil {
        return
    }
    p.mu.Lock()
    defer p.mu.Unlock()
    for i, obs := range p.observers {
        if obs == observer {
            p.observers = append(p.observers[:i], p.observers[i+1:]...)
            return
        }
    }
}
```

2. **WebSocket接続の管理**：WebSocket接続は適切に管理され、不要になった接続は閉じられます。

```go
func (s *StateServer) sendUpdateToClients(update interface{}) {
    // ...
    for client := range s.clients {
        if err := client.WriteJSON(update); err != nil {
            // ...
            err := client.Close()
            // ...
            delete(s.clients, client)
        }
    }
}
```

3. **リソースのクリーンアップ**：タイマーなどのリソースは使用後に適切にクリーンアップされます。

```go
func (t *TimeStrategy) Cleanup() error {
    t.mu.Lock()
    defer t.mu.Unlock()

    if t.isRunning {
        t.stopChan <- struct{}{}
        t.isRunning = false
    }
    return nil
}
```

### 非同期処理の活用

パフォーマンスを向上させるために、非同期処理が積極的に活用されています：

1. **WebSocketメッセージの非同期処理**：WebSocketメッセージの処理は、チャネルを使用した非同期処理によって行われ、メッセージ送信がブロックされないようにしています。

```go
func (s *StateServer) broadcastUpdate(update interface{}) {
    // ...
    select {
    case s.updateChan <- update:
        // メッセージが正常にキューに入った
    default:
        // チャネルがいっぱいの場合
        log.Warn("Update channel is full, dropping message")
    }
}
```

2. **タイマー処理の非同期実行**：タイマー処理は別のゴルーチンで実行され、メインスレッドがブロックされないようにしています。

```go
func (t *TimeStrategy) Start(ctx context.Context, part *entity.ConditionPart) error {
    // ...
    go t.run()
    // ...
}
```

### ロック範囲の最小化

パフォーマンスを向上させるために、ロック範囲を最小限に抑える工夫が施されています：

1. **読み取りロックの使用**：読み取り専用の操作には、読み取りロックを使用しています。

```go
func (p *Phase) NotifyPhaseChanged() {
    // ...
    p.mu.RLock()
    observers := make([]service.PhaseObserver, len(p.observers))
    copy(observers, p.observers)
    p.mu.RUnlock()
    // ...
}
```

2. **ロック範囲の限定**：ロックが必要な範囲を最小限に抑えています。

```go
func (p *Phase) OnConditionChanged(condition interface{}) {
    // ...
    p.mu.Lock()
    p.SatisfiedConditions[cond.ID] = true
    satisfied := p.checkConditionsSatisfied()
    if satisfied {
        p.IsClear = true
    }
    currentState := p.CurrentState()
    p.mu.Unlock()
    // ...
}
```

## 今後の展開と拡張性

本システムは、拡張性を考慮した設計となっており、今後の機能追加や改善が容易に行えるようになっています。以下では、今後の展開と拡張性について解説します。

### テスト強化

システムの信頼性をさらに高めるために、以下のようなテスト強化が計画されています：

1. **単体テストの拡充**：各コンポーネントの単体テストを拡充し、コードカバレッジを向上させます。
2. **統合テストの追加**：複数のコンポーネントが連携する統合テストを追加し、システム全体の動作を検証します。
3. **パフォーマンステスト**：高負荷時のシステムの挙動を検証するパフォーマンステストを実施します。

### 機能拡張

システムの機能をさらに拡充するために、以下のような機能拡張が計画されています：

1. **認証機能の追加**：WebSocket接続に認証機能を追加し、セキュリティを向上させます。
2. **セッション管理**：ユーザーセッションを管理し、複数のユーザーが同時に利用できるようにします。
3. **UI機能の強化**：より直感的で使いやすいUIを提供するために、クライアントサイドの機能を強化します。
4. **新しい条件タイプの追加**：新しい種類の条件評価戦略を追加し、より多様な条件を表現できるようにします。

### パフォーマンス改善

システムのパフォーマンスをさらに向上させるために、以下のような改善が計画されています：

1. **さらなる最適化**：状態遷移やイベント通知の処理をさらに最適化し、パフォーマンスを向上させます。
2. **スケーラビリティの向上**：システムがより多くのユーザーや複雑な状態遷移に対応できるように、スケーラビリティを向上させます。
3. **モニタリングの強化**：システムの動作状況をリアルタイムに監視するためのモニタリング機能を強化します。

### 拡張性の確保

システムの拡張性を確保するために、以下のような設計上の工夫が施されています：

1. **インターフェースの活用**：主要なコンポーネントはインターフェースを通じて連携しており、実装の詳細を隠蔽しています。これにより、コンポーネントの差し替えや拡張が容易になっています。

```go
type PartStrategy interface {
    Initialize(part *ConditionPart) error
    GetCurrentValue() interface{}
    Start(ctx context.Context, part *ConditionPart) error
    Evaluate(ctx context.Context, part *ConditionPart, params interface{}) error
    Cleanup() error
    AddObserver(observer StrategyObserver)
    RemoveObserver(observer StrategyObserver)
    NotifyUpdate(event interface{})
}
```

2. **ファクトリーパターンの活用**：オブジェクトの生成はファクトリーパターンを通じて行われており、新しい種類のオブジェクトを追加する際には、ファクトリーに登録するだけで済むようになっています。

```go
func (f *StrategyFactory) CreateStrategy(kind value.ConditionKind) (service.PartStrategy, error) {
    switch kind {
    case value.KindCounter:
        return strategy.NewCounterStrategy(), nil
    case value.KindTime:
        return strategy.NewTimeStrategy(), nil
    default:
        return nil, fmt.Errorf("unknown condition kind: %v", kind)
    }
}
```

3. **DTOパターンの活用**：クライアントとの通信はDTOパターンを通じて行われており、APIの変更が必要な場合でも、ドメインモデルを変更せずにDTOのみを変更することができます。

```go
func ConvertPhaseToDTO(phase *entity.Phase) PhaseDTO {
    return PhaseDTO{
        ID:          phase.ID,
        ParentID:    phase.ParentID,
        Name:        phase.Name,
        Description: phase.Description,
        Order:       phase.Order,
        State:       phase.CurrentState(),
        IsClear:     phase.IsClear,
        IsActive:    phase.IsActive(),
        HasChildren: phase.HasChildren(),
        StartTime:   phase.StartTime,
        FinishTime:  phase.FinishTime,
    }
}
```

## まとめ

本文書では、状態遷移の可視化サンプルアプリケーションの設計思想、アーキテクチャ、実装詳細について解説しました。このシステムは、階層構造を持つフェーズ管理、オブザーバーパターン、戦略パターン、DTOパターンなどの設計パターンを活用することで、複雑な状態遷移を効率的に管理し、視覚化する機能を提供しています。

また、WebSocketを使用したリアルタイム通信、非同期処理の活用、スレッドセーフな実装など、パフォーマンスと安定性を確保するための工夫も施されています。

今後は、テスト強化、機能拡張、パフォーマンス改善などを通じて、さらに使いやすく、信頼性の高いシステムへと発展させていく予定です。

このシステムが、複雑な状態遷移を管理する必要のあるアプリケーション開発の参考になれば幸いです。