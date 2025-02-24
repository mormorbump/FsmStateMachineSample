# 状態管理機能の拡張計画

## 1. エンティティの拡張

### 1.1 Phase
```go
type Phase struct {
    ID          PhaseID
    Name        string
    Description string
    Rule        GameRule
    StartTime   *time.Time
    FinishTime  *time.Time
    // 他のフィールド...
}
```

#### 変更点
- `enter_active`でStartTimeを設定
- `enter_finish`でFinishTimeを設定
- Reset時に両方を初期化

### 1.2 Condition
```go
type Condition struct {
    Name        string
    Description string
    StartTime   *time.Time
    FinishTime  *time.Time
    // 他のフィールド...
}
```

#### 変更点
- `enter_unsatisfied`でStartTimeを設定
- `enter_satisfied`でFinishTimeを設定
- Reset時に両方を初期化

### 1.3 ConditionPart
```go
type ConditionPart struct {
    StartTime   *time.Time
    FinishTime  *time.Time
    // 他のフィールド...
}
```

#### 変更点
- `enter_unsatisfied`でStartTimeを設定
- `enter_satisfied`でFinishTimeを設定
- Reset時に両方を初期化

## 2. state_facade.goの修正

### 2.1 Phase生成時の初期化
```go
func NewPhase(...) *Phase {
    return &Phase{
        ID:          id,
        Name:        name,
        Description: description,
        Rule:        rule,
        // 他のフィールド...
    }
}
```

### 2.2 Condition生成時の初期化
```go
func NewCondition(...) *Condition {
    return &Condition{
        Name:        name,
        Description: description,
        // 他のフィールド...
    }
}
```

## 3. テストケースの追加

### 3.1 Phase時間管理のテスト
```go
func TestPhaseTimeManagement(t *testing.T) {
    // 1. Activate時にStartTimeが設定されることを確認
    // 2. Finish時にFinishTimeが設定されることを確認
    // 3. Reset時に両方がnilになることを確認
}
```

### 3.2 Condition時間管理のテスト
```go
func TestConditionTimeManagement(t *testing.T) {
    // 1. Activate時にStartTimeが設定されることを確認
    // 2. Complete時にFinishTimeが設定されることを確認
    // 3. Reset時に両方がnilになることを確認
}
```

### 3.3 ConditionPart時間管理のテスト
```go
func TestConditionPartTimeManagement(t *testing.T) {
    // 1. Activate時にStartTimeが設定されることを確認
    // 2. Complete/Timeout時にFinishTimeが設定されることを確認
    // 3. Reset時に両方がnilになることを確認
}
```

## 4. 実装手順

1. 各エンティティのcallbacksを修正
   - StartTime設定の追加
   - FinishTime設定の追加
   - Reset処理の修正

2. state_facade.goの修正
   - インスタンス生成時の初期化処理追加

3. テストの実装
   - 各エンティティの時間管理テスト追加
   - 状態遷移のテスト追加
   - Reset処理のテスト追加

4. 動作確認
   - 各状態遷移での時間設定を確認
   - Reset時の初期化を確認