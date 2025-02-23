# 状態遷移の制御改善計画

## 1. 現状の問題

### 1.1 無限ループの発生
```go
// 現状のコード
func (pc *PhaseController) OnStateChanged(state string) {
    _ = pc.Start(context.Background()) // 新しい状態遷移を開始
    pc.NotifyStateChanged(state)       // 新しいOnStateChangedを呼び出す
}
```

- 状態変更を受け取るたびにStart()を呼び出す
- Start()が新しい状態遷移を引き起こす
- その状態遷移が再びOnStateChangedを呼び出す
- これが無限ループとなる

### 1.2 問題の影響
- サーバーログが大量に出力される
- システムリソースの無駄な消費
- 状態遷移の制御が不安定

## 2. 修正方針

### 2.1 状態遷移の制御改善
```go
// 修正案1: 状態に基づく制御
func (pc *PhaseController) OnStateChanged(state string) {
    // 特定の状態の場合のみ次のフェーズに進む
    if state == "next" {
        _ = pc.Start(context.Background())
    }
    pc.NotifyStateChanged(state)
}

// 修正案2: フラグによる制御
type PhaseController struct {
    // 既存のフィールド
    isTransitioning bool
    mu              sync.RWMutex
}

func (pc *PhaseController) OnStateChanged(state string) {
    pc.mu.Lock()
    if pc.isTransitioning {
        pc.mu.Unlock()
        return
    }
    pc.isTransitioning = true
    pc.mu.Unlock()

    defer func() {
        pc.mu.Lock()
        pc.isTransitioning = false
        pc.mu.Unlock()
    }()

    _ = pc.Start(context.Background())
    pc.NotifyStateChanged(state)
}
```

### 2.2 期待される効果
1. 状態遷移の制御
   - 適切なタイミングでのみ遷移を実行
   - 不要な遷移の防止
   - リソース使用の最適化

2. ログ出力の改善
   - 必要な情報のみを出力
   - デバッグのしやすさ向上
   - システム状態の把握が容易に

## 3. 実装手順

1. PhaseControllerの修正
   - isTransitioningフラグの追加
   - OnStateChanged関数の改善
   - ロック制御の追加

2. ログ出力の最適化
   - 重要なイベントのみをログ出力
   - デバッグレベルの調整
   - コンテキスト情報の追加

## 4. 検証項目

1. 基本機能
   - 状態遷移の正常動作
   - フェーズの順序通りの進行
   - エラー時の適切な処理

2. パフォーマンス
   - ログ出力量の確認
   - メモリ使用量の監視
   - CPU使用率の確認

3. エッジケース
   - 高速な状態変更
   - 同時実行時の動作
   - エラー発生時の挙動