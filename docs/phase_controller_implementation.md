# PhaseControllerの実装詳細

## 概要

`PhaseController`は、複数のフェーズとその条件を管理するコンポーネントです。このドキュメントでは、`PhaseController`の実装詳細と、複数の条件を持つフェーズの管理方法について説明します。

## アーキテクチャ

`PhaseController`は、Observer/Subjectパターンを採用しており、フェーズの状態変化を監視し、適切なアクションを実行します。

### 主要コンポーネント

1. **PhaseController**: フェーズの状態を管理し、フェーズ間の遷移を制御します。
2. **Phase**: 個々のフェーズを表し、条件の集合を持ちます。
3. **Condition**: フェーズの条件を表し、条件パーツの集合を持ちます。
4. **ConditionPart**: 条件の一部を表し、特定の評価ロジックを持ちます。
5. **Strategy**: 条件パーツの評価ロジックを実装します（CounterStrategy、TimeStrategyなど）。

## 状態遷移

フェーズは以下の状態を持ちます：

- **ready**: 初期状態
- **active**: アクティブ状態
- **next**: 次のフェーズに進む準備ができた状態
- **finish**: 完了状態

条件は以下の状態を持ちます：

- **ready**: 初期状態
- **unsatisfied**: 未満足状態
- **satisfied**: 満足状態

条件パーツは以下の状態を持ちます：

- **ready**: 初期状態
- **unsatisfied**: 未満足状態
- **processing**: 処理中状態
- **satisfied**: 満足状態

## 条件の種類

`PhaseController`は、以下の種類の条件をサポートしています：

1. **AND条件**: すべての条件が満たされた場合に満足となります。
2. **OR条件**: いずれかの条件が満たされた場合に満足となります。

## 実装上の注意点

### フェーズの状態管理

フェーズの状態を適切に管理するためには、以下の点に注意する必要があります：

1. フェーズの状態遷移は、`PhaseController`を通じて行う必要があります。
2. 複数のフェーズが関連する場合、各フェーズの状態を明示的に制御する必要があります。
3. テスト実行の順序によって、前のテストの状態が次のテストに影響を与える可能性があります。

### テスト時の注意点

テストを実行する際には、以下の点に注意する必要があります：

1. テストの前提条件を明確に設定し、テスト間の独立性を確保します。
2. 各テストの前に明示的に状態をリセットします。
3. 複数のフェーズが関連するテストでは、各フェーズの状態を明示的に制御します。

例えば、`Phase2_OR_Condition`サブテストでは、以下のようにフェーズの状態を明示的に制御しています：

```go
// Phase2とPhase3の状態をリセットする
_ = phase2.Reset(ctx)
_ = phase3.Reset(ctx)

// Phase2を明示的にactive状態にする
_ = phase2.Activate(ctx)

// 現在のフェーズをPhase2に設定
controller.SetCurrentPhase(phase2.Name)
```

### 条件評価のタイミング

条件評価のタイミングには、以下の点に注意する必要があります：

1. 時間ベースの条件評価と手動トリガーの条件評価の組み合わせが複雑な状態遷移を引き起こす可能性があります。
2. 条件評価のタイミングによっては、予期しない状態遷移が発生する可能性があります。
3. 条件評価の順序によっては、異なる結果が得られる可能性があります。

## 実装例

### PhaseControllerの初期化

```go
func NewPhaseController(phases []*entity.Phase) *PhaseController {
    controller := &PhaseController{
        phases:        phases,
        currentPhase:  nil,
        stateObservers: make([]entity.StateObserver, 0),
        condObservers: make([]entity.ConditionObserver, 0),
        partObservers: make([]entity.ConditionPartObserver, 0),
        log:           logger.GetLogger(),
    }

    // フェーズの初期化
    for _, phase := range phases {
        phase.AddObserver(controller)
        for _, condition := range phase.GetConditions() {
            condition.AddConditionObserver(controller)
            for _, part := range condition.GetParts() {
                part.AddConditionPartObserver(controller)
            }
        }
    }

    // 最初のフェーズを現在のフェーズとして設定
    if len(phases) > 0 {
        controller.currentPhase = phases[0]
    }

    return controller
}
```

### フェーズの状態遷移

```go
func (pc *PhaseController) Start(ctx context.Context) error {
    pc.mu.Lock()
    defer pc.mu.Unlock()

    pc.log.Debug("PhaseController.Start", zap.String("action", "Starting phase sequence"))

    if pc.currentPhase == nil {
        return errors.New("no current phase set")
    }

    pc.log.Debug("PhaseController.Start", zap.String("current_phase", pc.currentPhase.Name), zap.Int("order", pc.currentPhase.Order), zap.String("state", pc.currentPhase.CurrentState()))

    // 現在のフェーズの状態に応じて処理を分岐
    switch pc.currentPhase.CurrentState() {
    case value.StateReady:
        // 最初のフェーズを開始
        return pc.ProcessAndActivateByNextOrder(ctx)
    case value.StateNext:
        // 次のフェーズに進む
        return pc.ProcessAndActivateByNextOrder(ctx)
    default:
        // 現在のフェーズが既にアクティブまたは完了している場合は何もしない
        return nil
    }
}
```

### 条件の評価

```go
func (pc *PhaseController) OnConditionChanged(condition *entity.Condition) {
    pc.log.Debug("PhaseController.OnConditionChanged", zap.Any("condition", condition))

    // 条件の変更を通知
    pc.NotifyConditionChanged(condition)
}
```

## 今後の改善点

1. **テストの強化**: 他のテストケースも同様の問題がないか確認し、必要に応じて修正します。
2. **テスト間の独立性の確保**: 各テストの前に明示的に状態をリセットするヘルパー関数の導入を検討します。
3. **ドキュメントの更新**: 複雑な状態遷移のテスト方法に関するドキュメントを追加します。
4. **コードリファクタリング**: 状態遷移のロジックをより明確にし、テスト容易性を向上させます。