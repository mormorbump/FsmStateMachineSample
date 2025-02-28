# 複数条件テスト計画

## 背景

現在のアプリケーションでは、Phaseに複数のConditionを紐づけることができますが、同じIDを持つConditionが上書きされる問題がありました。この問題は修正されましたが、複数のConditionが正しく動作することを確認するためのテストが必要です。

## テスト目的

1. 複数の条件を持つPhaseが正しく動作することを確認する
2. 異なる`ConditionType`（AND, OR, Single）を持つPhaseが正しく動作することを確認する
3. 複数の条件がUIに正しく表示されることを確認する

## テスト計画

### 1. PhaseControllerの複数条件テスト

`internal/usecase/state/phase_controller_multiple_conditions_test.go`に以下のテストを実装します：

```go
func TestPhaseControllerWithMultipleConditions(t *testing.T) {
    // 異なるタイプの条件を持つPhaseを作成
    // - Phase1: AND条件（2つの条件が両方満たされる必要がある）
    // - Phase2: OR条件（2つの条件のいずれかが満たされればよい）
    // - Phase3: Single条件（1つの条件のみ）
    
    // PhaseControllerを作成
    
    // 各Phaseをアクティブにして、条件が正しく評価されることを確認
    
    // Phase1（AND条件）のテスト
    // - 1つ目の条件のみ満たす → Phaseは完了しない
    // - 2つ目の条件も満たす → Phaseは完了する
    
    // Phase2（OR条件）のテスト
    // - 1つ目の条件のみ満たす → Phaseは完了する
    
    // Phase3（Single条件）のテスト
    // - 条件を満たす → Phaseは完了する
}
```

### 2. 複数条件のUIテスト

`internal/ui/server_multiple_conditions_test.go`に以下のテストを実装します：

```go
func TestServerWithMultipleConditions(t *testing.T) {
    // 複数の条件を持つPhaseを作成
    
    // StateFacadeとStateServerを作成
    
    // EditResponseメソッドを呼び出して、レスポンスを取得
    
    // レスポンスに複数の条件が含まれていることを確認
    // - 条件の数が正しいこと
    // - 各条件のIDが正しいこと
    // - 各条件のラベルが正しいこと
    // - 各条件のタイプが正しいこと
}
```

### 3. 統合テスト

`internal/usecase/state/state_facade_multiple_conditions_test.go`に以下のテストを実装します：

```go
func TestStateFacadeWithMultipleConditions(t *testing.T) {
    // StateFacadeを作成（複数の条件を持つPhaseを含む）
    
    // 各Phaseをアクティブにして、条件が正しく評価されることを確認
    
    // GetCurrentPhaseメソッドを呼び出して、現在のPhaseを取得
    
    // 現在のPhaseから条件を取得し、複数の条件が存在することを確認
    
    // 条件を満たし、Phaseが正しく次の状態に遷移することを確認
}
```

## 実装手順

1. Codeモードに切り替える
2. 上記のテストファイルを作成する
3. テストを実行して、複数条件が正しく動作することを確認する

## 期待される結果

- すべてのテストが成功する
- 複数の条件を持つPhaseが正しく動作する
- 異なる`ConditionType`を持つPhaseが正しく動作する
- 複数の条件がUIに正しく表示される