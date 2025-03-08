# PhaseControllerのデバッグ記録

## 問題の概要

`TestPhaseControllerWithMultipleConditions`テストの`Phase2_OR_Condition`サブテストが失敗していました。

## 原因

テストログの分析から、以下の問題が特定されました：

1. `Phase2`が既に`finish`状態になっていた可能性があり、`controller.Start(ctx)`が失敗していました。
2. `Phase3`が`active`状態になっていたため、テストの期待する状態遷移が発生しませんでした。

## 修正内容

`internal/usecase/state/phase_controller_multiple_conditions_test.go`ファイルの`Phase2_OR_Condition`サブテストを以下のように修正しました：

1. 最初の修正では、`Phase2`が`finish`状態の場合にリセットして`active`状態に戻す処理を追加しましたが、これだけでは不十分でした。
2. 最終的な修正では、以下の処理を実装しました：
   - `Phase2`と`Phase3`の状態を明示的にリセット
   - `Phase2`を明示的に`active`状態に設定
   - 現在のフェーズを`Phase2`に設定

```go
// Phase2とPhase3の状態をリセットする
_ = phase2.Reset(ctx)
_ = phase3.Reset(ctx)

// Phase2を明示的にactive状態にする
_ = phase2.Activate(ctx)

// 現在のフェーズをPhase2に設定
controller.SetCurrentPhase(phase2.Name)
```

## 教訓

1. 複数のフェーズが関連するテストでは、各フェーズの状態を明示的に制御することが重要です。
2. テスト実行の順序によって、前のテストの状態が次のテストに影響を与える可能性があります。
3. テストの前提条件を明確に設定し、テスト間の独立性を確保することが重要です。

## 関連するコンポーネント

- `PhaseController`: フェーズの状態を管理するコントローラー
- `Phase`: 個々のフェーズを表すエンティティ
- `Condition`: フェーズの条件を表すエンティティ
- `ConditionPart`: 条件の一部を表すエンティティ