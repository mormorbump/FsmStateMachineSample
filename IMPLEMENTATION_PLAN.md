# Core Domain Test Implementation Plan

## テストユーティリティの改善

### 削除可能な関数（testifyで代替）
1. AssertStateSequence -> assert.Equal
2. AssertEventually -> require.Eventually
3. WaitForCondition -> require.Eventually

### モック実装の簡略化
```go
type mockStateObserver struct {
    stateChanges []string
    mock.Mock
}

func (m *mockStateObserver) OnStateChanged(state string) {
    m.Called(state)
    m.stateChanges = append(m.stateChanges, state)
}

type mockTimeObserver struct {
    mock.Mock
}

func (m *mockTimeObserver) OnTimeTicked() {
    m.Called()
}
```

## テストケースの実装方針

### 基本方針
- testifyのassertパッケージを使用
- テーブル駆動テストを活用
- 並行処理のテストにはrequire.Eventuallyを使用

### テストケース例
```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "case 1",
            input:    "input1",
            expected: "expected1",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := someFunction(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 並行処理のテスト例
```go
func TestConcurrent(t *testing.T) {
    require.Eventually(t, func() bool {
        // テスト条件
        return true
    }, time.Second, 10*time.Millisecond, "timeout waiting for condition")
}
```

## 実装順序

1. test_utils.goの更新
   - 不要な関数の削除
   - モック実装の簡略化

2. 既存のテストの修正
   - testifyの使用に合わせてテストを更新
   - テーブル駆動テストの導入
   - アサーションの書き換え

3. 新規テストの実装
   - condition_subject_test.go
   - condition_strategy_test.go
   - その他必要なテスト

## テスト実行方法

```bash
# 通常のテスト実行
go test ./internal/domain/core/... -v

# レースディテクタを有効にしてテスト実行
go test -race ./internal/domain/core/... -v
```

## 注意点

1. アサーションの使い分け
   - assert: 通常のアサーション
   - require: テストを即座に終了させる必要がある場合

2. モックの使用
   - testify/mockを活用
   - 必要最小限のモック実装に留める

3. テストの可読性
   - テストケース名は意図が明確に分かるように
   - テーブル駆動テストで類似のケースをまとめる