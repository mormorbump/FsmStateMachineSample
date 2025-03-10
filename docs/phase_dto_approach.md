# PhaseのDTOアプローチ

## 概要

現在の実装では、`Phase`構造体に`MarshalJSON`メソッドを実装して循環参照問題を解決していますが、より良いアプローチとして、`Phase`構造体から`MarshalJSON`メソッドを削除し、UI層でDTOを構築してJSONシリアライズする方法が考えられます。

## メリット

1. **責務の分離**
   - ドメインモデル（`Phase`）はJSONシリアライズの詳細を知る必要がなくなります
   - UI層は表示に必要なデータ構造を自由に定義できます
   - 各層の関心事が明確に分離されます

2. **循環参照問題の解決**
   - UI層でDTOを構築する際に必要なデータだけを選択するため、循環参照が自然に解消されます
   - 複雑な`MarshalJSON`メソッドが不要になります

3. **柔軟性の向上**
   - UI層の要件変更に対して、ドメインモデルを変更する必要がなくなります
   - クライアント側の表示要件に合わせて、必要なデータだけを送信できます

4. **パフォーマンスの向上**
   - 不要なデータをシリアライズしないため、ネットワーク転送量が減少します
   - JSONの構造がシンプルになるため、クライアント側での処理も効率的になります

## 実装アプローチ

### 1. `Phase`構造体から`MarshalJSON`メソッドを削除

```go
// MarshalJSON メソッドを削除
```

### 2. UI層でDTOを構築

```go
// PhaseDTO はUI層で使用するフェーズのデータ転送オブジェクト
type PhaseDTO struct {
    ID          value.PhaseID     `json:"id"`
    ParentID    value.PhaseID     `json:"parent_id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Order       int               `json:"order"`
    State       string            `json:"state"`
    IsClear     bool              `json:"is_clear"`
    IsActive    bool              `json:"is_active"`
    HasChildren bool              `json:"has_children"`
    StartTime   *time.Time        `json:"start_time,omitempty"`
    FinishTime  *time.Time        `json:"finish_time,omitempty"`
    // 他の必要なフィールド
}

// ConvertPhaseToDTO はPhaseオブジェクトをDTOに変換する
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

// GetAllPhasesDTO は全てのフェーズをDTOに変換する
func GetAllPhasesDTO(phases entity.Phases) []PhaseDTO {
    result := make([]PhaseDTO, len(phases))
    for i, phase := range phases {
        result[i] = ConvertPhaseToDTO(phase)
    }
    return result
}
```

### 3. `server.go`での使用例

```go
func (s *StateServer) OnEntityChanged(entityObj interface{}) {
    // ...
    
    // 全てのフェーズを取得
    allPhases := s.gameFacade.GetController().GetPhases()
    
    // DTOに変換
    phaseDTOs := GetAllPhasesDTO(allPhases)
    
    // レスポンスを構築
    response := struct {
        Type    string      `json:"type"`
        Phases  []PhaseDTO  `json:"phases"`
        Current *PhaseDTO   `json:"current_phase,omitempty"`
    }{
        Type:   "state_change",
        Phases: phaseDTOs,
    }
    
    // 現在のフェーズがある場合は追加
    currentPhase := s.gameFacade.GetCurrentLeafPhase()
    if currentPhase != nil {
        currentDTO := ConvertPhaseToDTO(currentPhase)
        response.Current = &currentDTO
    }
    
    // クライアントに送信
    s.broadcastUpdate(response)
}
```

### 4. クライアント側での階層構造の構築

```javascript
// フェーズの階層構造を構築する
function buildPhaseHierarchy(phases) {
    const phaseMap = {};
    const rootPhases = [];
    
    // まず全てのフェーズをマップに格納
    phases.forEach(phase => {
        phaseMap[phase.id] = { ...phase, children: [] };
    });
    
    // 親子関係を構築
    phases.forEach(phase => {
        if (phase.parent_id === 0) {
            // ルートフェーズ
            rootPhases.push(phaseMap[phase.id]);
        } else {
            // 子フェーズ
            const parent = phaseMap[phase.parent_id];
            if (parent) {
                parent.children.push(phaseMap[phase.id]);
            }
        }
    });
    
    return rootPhases;
}

// 使用例
socket.onmessage = function(event) {
    const data = JSON.parse(event.data);
    if (data.type === 'state_change') {
        const phaseHierarchy = buildPhaseHierarchy(data.phases);
        renderPhaseTree(phaseHierarchy, data.current_phase);
    }
};
```

## 結論

`Phase`構造体から`MarshalJSON`メソッドを削除し、UI層でDTOを構築してJSONシリアライズする方法は、責務の分離、循環参照問題の解決、柔軟性の向上、パフォーマンスの向上など、多くのメリットがあります。この方法を採用することで、コードの保守性と拡張性が向上し、より堅牢なアプリケーションになると考えられます。