# 実装計画

## 1. 問題の特定
現在のNewStateFacade()には以下の問題があります：
- Phaseの作成時にConditionが適切に初期化されていない
- ConditionPartの設定が完全に欠落している
- 各コンポーネント間の関係性が正しく構築されていない

## 2. 具体的な要件

### 2.1 時間条件の設定
- すべてのConditionKindはTime
- 各Phaseに対して1つのConditionPartを設定
- 時間設定：
  * Phase1: 1秒
  * Phase2: 2秒
  * Phase3: 3秒

### 2.2 コンポーネントの階層構造
```
Phase
  └── Condition (Kind: Time)
        └── ConditionPart (ReferenceValueInt: 1-3秒)
```

## 3. 実装手順

1. ConditionPartの作成
   - TimeStrategyを使用
   - ReferenceValueIntに各フェーズの時間を設定
   - ComparisonOperatorはEQ（等価比較）を使用

2. Conditionの作成
   - Kind: Time
   - ConditionTypeはSingle（単一条件）
   - 各フェーズに対応するConditionPartを設定

3. Phaseの修正
   - 各フェーズに対応するConditionを設定
   - ObserverパターンでConditionの状態を監視

4. NewStateFacade()の修正
```go
func NewStateFacade() StateFacade {
    // Phase1 (1秒)
    part1 := NewConditionPart(1, "Phase1_Part")
    part1.ReferenceValueInt = 1
    cond1 := NewCondition(1, "Phase1_Condition", condition.KindTime)
    cond1.AddPart(part1)
    phase1 := NewPhase("PHASE1", 1, cond1)

    // Phase2 (2秒)
    part2 := NewConditionPart(2, "Phase2_Part")
    part2.ReferenceValueInt = 2
    cond2 := NewCondition(2, "Phase2_Condition", condition.KindTime)
    cond2.AddPart(part2)
    phase2 := NewPhase("PHASE2", 2, cond2)

    // Phase3 (3秒)
    part3 := NewConditionPart(3, "Phase3_Part")
    part3.ReferenceValueInt = 3
    cond3 := NewCondition(3, "Phase3_Condition", condition.KindTime)
    cond3.AddPart(part3)
    phase3 := NewPhase("PHASE3", 3, cond3)

    phases := entity.Phases{phase1, phase2, phase3}
    controller := NewPhaseController(phases)

    return &stateFacadeImpl{
        controller: controller,
    }
}
```

## 4. テスト計画

1. ConditionPartのテスト
   - 時間経過による状態変化の検証
   - ReferenceValueIntの正しい設定確認

2. Conditionのテスト
   - Time条件の評価検証
   - 状態遷移の確認

3. Phaseのテスト
   - 時間経過による遷移の検証
   - フェーズ間の連携確認

4. 統合テスト
   - 全フェーズの順次実行確認
   - タイミングの正確性検証