# リファクタリング計画

本ドキュメントでは、状態遷移システムのコードベースを分析し、リファクタリングすべきポイントとその計画を記述します。

## リファクタリングポイント

### 1. 責務の分離

#### 1.1 PhaseControllerの責務過多

**問題点**:
`PhaseController`クラスが複数の責務を持っています。フェーズの制御、オブザーバーの管理、通知の処理など、多くの責務が一つのクラスに集中しています。

```go
// PhaseController はフェーズの制御を担当するコントローラーです
type PhaseController struct {
    phaseFacade *entity.PhaseFacade
    observers   []service.ControllerObserver
    mu          sync.RWMutex
    log         *zap.Logger
}
```

**改善案**:
- フェーズの制御とオブザーバーの管理を別々のクラスに分離する
- 通知処理を専用のNotifierクラスに移動する

#### 1.2 StateServerの責務過多

**問題点**:
`StateServer`クラスがWebSocket通信、クライアント管理、エンティティ変更の通知処理など、多くの責務を持っています。

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

**改善案**:
- WebSocket通信を担当するConnectionManagerクラスを作成
- エンティティ変更の通知処理を担当するNotificationServiceクラスを作成
- クライアント管理を担当するClientManagerクラスを作成

### 2. コードの重複

#### 2.1 条件情報の取得処理の重複

**問題点**:
`EditResponse`メソッドと`getConditionInfos`メソッドで、条件情報の取得処理が重複しています。

```go
// EditResponseメソッド内
conditions := make([]ConditionInfo, 0)
for _, condition := range currentPhase.GetConditions() {
    condInfo := ConditionInfo{
        ID:          condition.ID,
        Label:       condition.Label,
        State:       condition.CurrentState(),
        Kind:        condition.Kind,
        IsClear:     condition.IsClear,
        Description: condition.Description,
        Parts:       make([]ConditionPartInfo, 0),
    }
    // ...
}

// getConditionInfosメソッド
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
        // ...
    }
    // ...
}
```

**改善案**:
- 共通の処理を`createConditionInfo`などのヘルパーメソッドに抽出する

#### 2.2 オブザーバー管理コードの重複

**問題点**:
オブザーバーの追加・削除・通知のコードが複数のクラスで重複しています。

```go
// Phaseクラス
func (p *Phase) AddObserver(observer service.PhaseObserver) {
    if observer == nil {
        return
    }
    p.mu.Lock()
    defer p.mu.Unlock()
    p.observers = append(p.observers, observer)
}

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

// Conditionクラス
func (c *Condition) AddConditionObserver(observer service.ConditionObserver) {
    if observer == nil {
        return
    }
    c.mu.Lock()
    defer c.mu.Unlock()
    c.condObservers = append(c.condObservers, observer)
}

func (c *Condition) RemoveConditionObserver(observer service.ConditionObserver) {
    if observer == nil {
        return
    }
    c.mu.Lock()
    defer c.mu.Unlock()
    for i, obs := range c.condObservers {
        if obs == observer {
            c.condObservers = append(c.condObservers[:i], c.condObservers[i+1:]...)
            return
        }
    }
}
```

**改善案**:
- ジェネリックなObservableトレイトまたは基底クラスを作成し、共通のオブザーバー管理コードを提供する

### 3. エラーハンドリングの一貫性

#### 3.1 エラーハンドリングの不一致

**問題点**:
エラーハンドリングの方法が一貫していません。一部の場所ではエラーをログに記録して無視し、他の場所ではエラーを返しています。

```go
// エラーをログに記録して無視する例
if err := phase.Finish(ctx); err != nil {
    pc.log.Error("Failed to finish current phase", zap.Error(err))
    // エラーが発生しても次のフェーズに進む試みをする
}

// エラーを返す例
if err := phase.Activate(ctx); err != nil {
    return err
}
```

**改善案**:
- エラーハンドリングのポリシーを定義し、一貫して適用する
- 重大なエラーと非重大なエラーを区別し、適切に処理する
- エラーラッピングを活用して、エラーの文脈を保持する

#### 3.2 エラー型の不足

**問題点**:
カスタムエラー型が不足しており、エラーの種類を区別するのが困難です。

**改善案**:
- ドメイン固有のエラー型を定義する（例: `PhaseNotFoundError`, `InvalidStateTransitionError`など）
- エラー型に基づいて適切な処理を行う

### 4. インターフェースの設計

#### 4.1 インターフェースの粒度

**問題点**:
一部のインターフェースが大きすぎる、または小さすぎる可能性があります。

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

**改善案**:
- インターフェースを責務ごとに分割する（例: `Strategy`, `Observable`など）
- インターフェース分離の原則に従って、クライアントが必要としないメソッドを強制しないようにする

#### 4.2 インターフェースの一貫性

**問題点**:
類似した機能を持つインターフェースの命名や構造が一貫していません。

```go
type PhaseObserver interface {
    OnPhaseChanged(phase interface{})
}

type ConditionObserver interface {
    OnConditionChanged(condition interface{})
}

type ControllerObserver interface {
    OnEntityChanged(entity interface{})
}
```

**改善案**:
- 命名規則を統一する（例: すべて`EntityObserver`と`OnEntityChanged`のような形式にする）
- パラメータの型を一貫させる（可能であれば具体的な型を使用する）

### 5. メモリ管理の最適化

#### 5.1 リソースの解放

**問題点**:
一部のリソース（特にタイマーやチャネル）の解放が明示的に行われていない可能性があります。

```go
func (t *TimeStrategy) Start(ctx context.Context, part *entity.ConditionPart) error {
    t.mu.Lock()
    defer t.mu.Unlock()

    if t.isRunning {
        return nil
    }

    t.isRunning = true
    t.stopChan = make(chan struct{})
    t.ticker = time.NewTicker(t.interval)

    go t.run()

    return nil
}
```

**改善案**:
- `defer`を使用して、関数終了時にリソースを確実に解放する
- コンテキストのキャンセルを適切に処理する
- ファイナライザーまたはクリーンアップメソッドを実装する

#### 5.2 メモリリークの防止

**問題点**:
オブザーバーパターンの実装でメモリリークが発生する可能性があります。オブザーバーが適切に削除されない場合、参照が残り続けます。

**改善案**:
- 弱参照を使用してオブザーバーを保持する
- オブザーバーの登録解除を確実に行う仕組みを提供する
- 定期的にデッドオブザーバーをクリーンアップする

### 6. 並行処理の改善

#### 6.1 ロックの最適化

**問題点**:
ロックの範囲が広すぎる場所があり、パフォーマンスに影響を与える可能性があります。

```go
func (p *Phase) NotifyPhaseChanged() {
    p.mu.RLock()
    observers := make([]service.PhaseObserver, len(p.observers))
    copy(observers, p.observers)
    p.mu.RUnlock()

    for i, observer := range observers {
        observer.OnPhaseChanged(p)
    }
}
```

**改善案**:
- ロックの範囲を最小限に抑える
- 読み取り操作と書き込み操作を明確に分離する
- 必要に応じて、より細かい粒度のロックを使用する

#### 6.2 コンテキストの活用

**問題点**:
コンテキストが一貫して使用されていない、またはキャンセルが適切に処理されていない可能性があります。

```go
func (pc *PhaseController) OnPhaseChanged(phaseEntity interface{}) {
    // ...
    ctx := context.Background()
    // ...
}
```

**改善案**:
- コンテキストを一貫して使用し、適切に伝播させる
- コンテキストのキャンセルを監視し、リソースを適切に解放する
- タイムアウトやデッドラインを設定する

### 7. テスト容易性の向上

#### 7.1 依存性の注入

**問題点**:
一部のクラスが依存オブジェクトを直接作成しており、テストが困難になっています。

```go
func NewStateFacade() *GameFacade {
    log := logger.DefaultLogger()
    factory := strategy.NewStrategyFactory()
    // ...
}
```

**改善案**:
- 依存オブジェクトをコンストラクタの引数として受け取る
- インターフェースを使用して依存関係を抽象化する
- モックやスタブを使用しやすい設計にする

#### 7.2 テスト用のフックの追加

**問題点**:
テスト時に内部状態を検証したり、振る舞いを制御したりするためのフックが不足しています。

**改善案**:
- テスト用のフックメソッドを追加する（ビルドタグで本番環境では除外可能）
- 内部状態を検証するためのゲッターを追加する
- 振る舞いを制御するためのセッターを追加する

### 8. 設定の外部化

#### 8.1 ハードコードされた値

**問題点**:
タイムアウト値、バッファサイズ、再試行回数などの設定値がハードコードされています。

```go
updateChan: make(chan interface{}, 100), // バッファ付きチャネルを作成
```

**改善案**:
- 設定値を構造体にまとめる
- 設定を外部ファイル（JSON, YAML, TOMLなど）から読み込む
- 環境変数を使用して設定を上書きできるようにする

#### 8.2 初期化コードの集中

**問題点**:
`NewStateFacade`メソッドに多くの初期化コードが集中しており、テストや拡張が困難です。

```go
func NewStateFacade() *GameFacade {
    // 多くの初期化コード
    // ...
}
```

**改善案**:
- ファクトリーメソッドを分割する
- ビルダーパターンを使用して初期化を段階的に行う
- 設定を外部から注入できるようにする

### 9. ロギングの改善

#### 9.1 ログレベルの一貫性

**問題点**:
ログレベル（Debug, Info, Warn, Error）の使用が一貫していません。

```go
// Debugレベルで重要な情報をログ
p.log.Debug("Phase.OnConditionChanged: Moving to next state",
    zap.String("phase", p.Name),
    zap.String("from_state", currentState))

// Errorレベルで非重大なエラーをログ
pc.log.Error("Failed to finish current phase", zap.Error(err))
```

**改善案**:
- ログレベルの使用ガイドラインを定義する
- 各レベルの使用例を提供する
- ログメッセージの形式を統一する

#### 9.2 構造化ロギングの活用

**問題点**:
構造化ロギングの機能が十分に活用されていない可能性があります。

**改善案**:
- 関連する情報をフィールドとして常に含める
- コンテキスト情報（リクエストID、セッションIDなど）を一貫して記録する
- ログの検索や分析を容易にするためのフィールドを追加する

### 10. ドキュメントの改善

#### 10.1 コードコメントの充実

**問題点**:
一部のコードにコメントが不足しており、意図や動作が分かりにくくなっています。

**改善案**:
- 公開APIには必ずドキュメントコメントを追加する
- 複雑なロジックには説明コメントを追加する
- 非自明な決定や制約には理由を記述する

#### 10.2 アーキテクチャドキュメントの更新

**問題点**:
アーキテクチャドキュメントが実装と完全に一致していない可能性があります。

**改善案**:
- コードから自動的にドキュメントを生成する仕組みを導入する
- 定期的にドキュメントをレビューし、更新する
- アーキテクチャの決定記録（ADR）を作成し、維持する

## リファクタリング計画

以上のリファクタリングポイントを踏まえ、以下のリファクタリング計画を提案します。

### フェーズ1: 基盤の改善（2週間）

1. **責務の分離**
   - PhaseControllerの責務を分離
   - StateServerの責務を分離

2. **コードの重複の排除**
   - 共通のヘルパーメソッドの抽出
   - ジェネリックなObservableトレイトの作成

3. **エラーハンドリングの一貫性確保**
   - エラーハンドリングポリシーの定義
   - カスタムエラー型の実装

### フェーズ2: 設計の改善（2週間）

4. **インターフェースの最適化**
   - インターフェースの粒度の見直し
   - インターフェースの命名と構造の統一

5. **メモリ管理の最適化**
   - リソース解放の確認と改善
   - メモリリーク防止策の実装

6. **並行処理の改善**
   - ロックの最適化
   - コンテキストの一貫した使用

### フェーズ3: テスト容易性と保守性の向上（2週間）

7. **テスト容易性の向上**
   - 依存性注入の導入
   - テスト用フックの追加

8. **設定の外部化**
   - 設定構造体の作成
   - 外部設定ファイルのサポート

9. **ロギングの改善**
   - ログレベルの使用ガイドラインの作成
   - 構造化ロギングの活用

10. **ドキュメントの改善**
    - コードコメントの充実
    - アーキテクチャドキュメントの更新

### フェーズ4: 検証と最適化（1週間）

11. **パフォーマンステスト**
    - リファクタリング前後のパフォーマンス比較
    - ボトルネックの特定と最適化

12. **コードレビュー**
    - リファクタリングの成果の評価
    - 残りの技術的負債の特定

## 優先順位

リファクタリングの優先順位は以下の通りです：

1. **高優先度**
   - 責務の分離
   - エラーハンドリングの一貫性確保
   - メモリ管理の最適化

2. **中優先度**
   - コードの重複の排除
   - インターフェースの最適化
   - 並行処理の改善

3. **低優先度**
   - テスト容易性の向上
   - 設定の外部化
   - ロギングの改善
   - ドキュメントの改善

## リスク管理

リファクタリング中のリスクを管理するために、以下の対策を講じます：

1. **テストカバレッジの確保**
   - リファクタリング前にテストカバレッジを向上させる
   - リファクタリング中に継続的にテストを実行する

2. **段階的なリファクタリング**
   - 大きな変更を小さな変更に分割する
   - 各変更後にテストを実行する

3. **コードレビュー**
   - すべての変更に対してコードレビューを実施する
   - 複数の視点からリファクタリングの品質を確認する

4. **ロールバック計画**
   - 問題が発生した場合のロールバック手順を準備する
   - 重要なマイルストーンでスナップショットを作成する

## 結論

本リファクタリング計画は、状態遷移システムのコードベースを改善し、より保守性が高く、拡張性のあるシステムにすることを目的としています。計画を実行することで、技術的負債を減らし、将来の機能追加や変更をより容易にすることができます。

リファクタリングは継続的なプロセスであり、この計画は最初のステップに過ぎません。システムの進化に合わせて、定期的にコードベースを評価し、必要に応じてリファクタリングを行うことが重要です。