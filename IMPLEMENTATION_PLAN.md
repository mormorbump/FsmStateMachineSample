# ObserverManagerのリネーム実装計画

## 1. 変更対象

### 1.1 condition_observer.go
- ObserverManagerをConditionSubjectImplにリネーム
- インターフェース名とメソッド名の整理
- コメントの更新

### 1.2 condition.go
- ObserverManagerの参照をConditionSubjectImplに変更
- 構造体の埋め込みを更新
- コメントの更新

### 1.3 condition_part.go
- ObserverManagerの参照をConditionSubjectImplに変更
- 構造体の埋め込みを更新
- コメントの更新

## 2. 実装の詳細

### 2.1 ConditionSubjectImpl
```go
// ConditionSubjectImpl は条件の状態変化を通知する機能を提供します
type ConditionSubjectImpl struct {
    conditionObservers     []ConditionObserver
    conditionPartObservers []ConditionPartObserver
}
```

### 2.2 メソッド名の整理
- AddConditionObserver -> 変更なし
- RemoveConditionObserver -> 変更なし
- NotifyConditionSatisfied -> 変更なし
- AddConditionPartObserver -> 変更なし
- RemoveConditionPartObserver -> 変更なし
- NotifyPartSatisfied -> 変更なし

## 3. 移行手順

1. condition_observer.goの修正
   - 型名の変更
   - コメントの更新
   - インターフェースの整理

2. condition.goの修正
   - 構造体の埋め込み更新
   - NewCondition関数の更新
   - コメントの更新

3. condition_part.goの修正
   - 構造体の埋め込み更新
   - NewConditionPart関数の更新
   - コメントの更新

## 4. 影響範囲

- 型名の変更のみで、インターフェースや機能は変更なし
- 既存のテストコードへの影響も最小限
- 外部からの利用方法は変更なし

## 5. 期待される改善

- Subject/Observerパターンの意図がより明確に
- コードの意図が理解しやすく
- 将来の拡張がしやすい設計に