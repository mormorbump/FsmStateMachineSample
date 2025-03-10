# Phase.MarshalJSON メソッドの修正計画

## 問題の概要

現在、`Phase`構造体の`MarshalJSON`メソッドでスタックオーバーフローが発生しています。エラーログを見ると、以下のような無限再帰が発生しています：

```
encoding/json.Marshal -> Phase.MarshalJSON -> encoding/json.Marshal -> Phase.MarshalJSON -> ...
```

これは、`Phase`構造体の`MarshalJSON`メソッド内で、循環参照を適切に処理できていないことが原因です。現在の実装では、`Alias`型を使用して無限再帰を避けようとしていますが、`Children`フィールドの処理に問題があります。

## 現在の実装

現在の`MarshalJSON`メソッドは以下のようになっています：

```go
// MarshalJSON はJSONシリアライズ時の循環参照を回避するためのメソッドです
func (p *Phase) MarshalJSON() ([]byte, error) {
    type Alias Phase // 元の型のエイリアスを作成
    
    // 親子関係を持たない一時的な構造体を作成
    return json.Marshal(&struct {
        *Alias
        Parent   *value.PhaseID  `json:"parent,omitempty"`   // IDのみを含める
        Children []value.PhaseID `json:"children,omitempty"` // IDのみを含める
    }{
        Alias: (*Alias)(p),
        Parent: func() *value.PhaseID {
            if p.Parent != nil {
                id := p.Parent.ID
                return &id
            }
            return nil
        }(),
        Children: func() []value.PhaseID {
            ids := make([]value.PhaseID, len(p.Children))
            for i, child := range p.Children {
                ids[i] = child.ID
            }
            return ids
        }(),
    })
}
```

## 問題の原因

問題は、`Alias`型が`Phase`型のエイリアスであるため、`Alias`型にも`Children`フィールドと`Parent`フィールドが含まれていることです。そのため、`json.Marshal(&struct{ *Alias ... })`を呼び出すと、`Alias`型の`Children`フィールドと`Parent`フィールドもシリアライズされようとします。

これにより、`Children`フィールドの各要素（`*Phase`型）に対しても`MarshalJSON`メソッドが呼び出され、無限再帰が発生します。

## 解決策

解決策は、`Alias`型から`Children`フィールドと`Parent`フィールドを除外することです。これには、新しい構造体を定義して必要なフィールドだけをコピーする方法があります。

以下に修正案を示します：

```go
// MarshalJSON はJSONシリアライズ時の循環参照を回避するためのメソッドです
func (p *Phase) MarshalJSON() ([]byte, error) {
    // 循環参照を避けるために、必要なフィールドだけを持つ新しい構造体を定義
    type PhaseDTO struct {
        ID                          value.PhaseID     `json:"id"`
        Order                       int               `json:"order"`
        IsActive                    bool              `json:"is_active"`
        IsClear                     bool              `json:"is_clear"`
        Name                        string            `json:"name"`
        Description                 string            `json:"description"`
        Rule                        value.GameRule    `json:"rule"`
        ConditionType               value.ConditionType `json:"condition_type"`
        ConditionIDs                []value.ConditionID `json:"condition_ids"`
        SatisfiedConditions         map[value.ConditionID]bool `json:"satisfied_conditions"`
        Conditions                  map[value.ConditionID]*Condition `json:"conditions"`
        StartTime                   *time.Time        `json:"start_time,omitempty"`
        FinishTime                  *time.Time        `json:"finish_time,omitempty"`
        ParentID                    value.PhaseID     `json:"parent_id"`
        Parent                      *value.PhaseID    `json:"parent,omitempty"`
        Children                    []value.PhaseID   `json:"children,omitempty"`
        AutoProgressOnChildrenComplete bool           `json:"auto_progress_on_children_complete"`
    }
    
    // 新しい構造体にデータをコピー
    dto := PhaseDTO{
        ID:                          p.ID,
        Order:                       p.Order,
        IsActive:                    p.isActive,
        IsClear:                     p.IsClear,
        Name:                        p.Name,
        Description:                 p.Description,
        Rule:                        p.Rule,
        ConditionType:               p.ConditionType,
        ConditionIDs:                p.ConditionIDs,
        SatisfiedConditions:         p.SatisfiedConditions,
        Conditions:                  p.Conditions,
        StartTime:                   p.StartTime,
        FinishTime:                  p.FinishTime,
        ParentID:                    p.ParentID,
        AutoProgressOnChildrenComplete: p.AutoProgressOnChildrenComplete,
    }
    
    // Parent フィールドを ID のみに変換
    if p.Parent != nil {
        parentID := p.Parent.ID
        dto.Parent = &parentID
    }
    
    // Children フィールドを ID のリストに変換
    childrenIDs := make([]value.PhaseID, len(p.Children))
    for i, child := range p.Children {
        childrenIDs[i] = child.ID
    }
    dto.Children = childrenIDs
    
    // 新しい構造体をシリアライズ
    return json.Marshal(dto)
}
```

この修正では、`Alias`型を使用する代わりに、必要なフィールドだけを持つ新しい構造体（`PhaseDTO`）を定義しています。そして、`Phase`構造体のフィールドを`PhaseDTO`構造体にコピーし、`Parent`フィールドと`Children`フィールドを適切に変換しています。

これにより、循環参照が解消され、スタックオーバーフローが発生しなくなります。

## 注意点

1. `Condition`構造体にも循環参照がある場合は、同様の修正が必要になる可能性があります。
2. `PhaseDTO`構造体のフィールドは、`Phase`構造体のフィールドと一致している必要があります（`Parent`フィールドと`Children`フィールドを除く）。
3. `Phase`構造体に新しいフィールドが追加された場合は、`PhaseDTO`構造体にも追加する必要があります。

## 実装手順

1. `phase.go`ファイルを開きます。
2. `MarshalJSON`メソッドを上記の修正案に置き換えます。
3. コードをコンパイルして、エラーがないことを確認します。
4. アプリケーションを実行して、スタックオーバーフローが解消されたことを確認します。

## 代替案

もう一つの解決策として、`omitempty`タグを使用して、`Parent`フィールドと`Children`フィールドをJSONシリアライズから除外する方法もあります。ただし、この方法では、これらのフィールドの情報が完全に失われてしまいます。

```go
type Alias struct {
    // Phase 構造体のフィールドをコピー
    ID                          value.PhaseID     `json:"id"`
    // ... 他のフィールド ...
    
    // 循環参照を引き起こすフィールドを除外
    Parent                      *Phase            `json:"-"`
    Children                    []*Phase          `json:"-"`
}
```

しかし、この方法では、親子関係の情報が失われてしまうため、推奨されません。

## 結論

`MarshalJSON`メソッドを修正して、循環参照を適切に処理することで、スタックオーバーフローの問題を解決できます。上記の修正案を実装することで、`Phase`構造体のJSONシリアライズが正常に機能するようになります。