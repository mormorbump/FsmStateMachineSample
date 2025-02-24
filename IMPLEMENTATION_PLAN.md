# 状態表示機能の拡張計画

## 概要
現在のUIをPhase、Condition、ConditionPartの状態を全て表示できるように拡張する。

## 実装手順

### 1. サーバーサイドの拡張

#### 1.1 状態情報構造体の拡張
```go
type StateUpdate struct {
    Type    string              `json:"type"`
    State   string              `json:"state"`
    Info    *core.GameStateInfo `json:"info,omitempty"`
    Phase   string              `json:"phase"`
    Message string              `json:"message,omitempty"`
    // 追加する情報
    Conditions []ConditionInfo `json:"conditions,omitempty"`
}

type ConditionInfo struct {
    ID    string `json:"id"`
    State string `json:"state"`
    Parts []ConditionPartInfo `json:"parts"`
}

type ConditionPartInfo struct {
    ID    string `json:"id"`
    State string `json:"state"`
}
```

#### 1.2 OnStateChangedメソッドの拡張
- CurrentPhaseからConditionとConditionPartの情報を取得
- 新しい構造体に情報を格納してクライアントに送信

### 2. クライアントサイドの拡張

#### 2.1 HTMLの拡張
- Condition状態表示セクションの追加
- ConditionPart状態表示セクションの追加
- 階層構造を視覚的に表現するレイアウトの実装

#### 2.2 CSSの拡張
- 新しいセクションのスタイル定義
- 状態に応じた視覚的フィードバック
- レスポンシブデザインの対応

#### 2.3 JavaScriptの拡張
- 状態更新処理の拡張
- Condition/ConditionPart状態の表示処理追加
- 状態変更時のアニメーション実装

## 実装の流れ

1. サーバーサイドの実装
   - 新しい構造体の追加
   - OnStateChangedメソッドの拡張
   - テストの追加

2. クライアントサイドの実装
   - HTML構造の更新
   - CSS定義の追加
   - JavaScript処理の拡張

3. テストと動作確認
   - 各状態遷移時の表示確認
   - エラーケースの確認
   - パフォーマンスの確認

## 期待される結果

- Phase、Condition、ConditionPartの状態が一目で確認可能
- 状態変更がリアルタイムに反映
- 直感的なUI/UXの提供