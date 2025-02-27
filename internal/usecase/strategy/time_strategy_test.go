package strategy

import (
	"context"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/value"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockStrategyObserver は StrategyObserver インターフェースのモック実装です
type MockTimeStrategyObserver struct {
	Events []string
}

// OnUpdated はイベントを記録します
func (m *MockTimeStrategyObserver) OnUpdated(event string) {
	m.Events = append(m.Events, event)
}

func TestNewTimeStrategy(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// 初期状態の検証
	assert.Empty(t, strategy.observers)
	assert.False(t, strategy.isRunning)
	assert.Nil(t, strategy.ticker)
	assert.Nil(t, strategy.stopChan)
}

func TestTimeStrategyInitialize(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")
	part.ReferenceValueInt = 5 // 5秒

	// 初期化
	err := strategy.Initialize(part)
	assert.NoError(t, err)

	// 設定が正しく行われていることを確認
	assert.Equal(t, 5*time.Second, strategy.interval)
	assert.NotNil(t, strategy.stopChan)
	assert.Len(t, strategy.observers, 1)

	// 無効なパラメータでの初期化
	err = strategy.Initialize("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid part type")

	// 無効な時間間隔での初期化
	invalidPart := entity.NewConditionPart(2, "Invalid Part")
	invalidPart.ReferenceValueInt = 0 // 0秒は無効
	err = strategy.Initialize(invalidPart)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid time interval")
}

func TestTimeStrategyGetCurrentValue(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// 現在値の取得（常にnilを返す）
	value := strategy.GetCurrentValue()
	assert.Nil(t, value)
}

func TestTimeStrategyStart(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")
	part.ReferenceValueInt = 1 // 1秒

	// 初期化（これによりlogが設定される）
	err := strategy.Initialize(part)
	assert.NoError(t, err)

	// 開始
	ctx := context.Background()
	err = strategy.Start(ctx, part)
	assert.NoError(t, err)
	assert.True(t, strategy.isRunning)
	assert.NotNil(t, strategy.ticker)

	// 既に実行中の場合
	err = strategy.Start(ctx, part)
	assert.NoError(t, err) // エラーは返さない

	// クリーンアップ
	strategy.Cleanup()
}

func TestTimeStrategyEvaluate(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")
	part.ReferenceValueInt = 1 // 1秒
	
	// 初期化（これによりlogが設定される）
	err := strategy.Initialize(part)
	assert.NoError(t, err)

	// 評価（現在は何もしない）
	ctx := context.Background()
	err = strategy.Evaluate(ctx, part, nil)
	assert.NoError(t, err)
}

func TestTimeStrategyCleanup(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")
	part.ReferenceValueInt = 1 // 1秒

	// 初期化（これによりlogが設定される）
	err := strategy.Initialize(part)
	assert.NoError(t, err)

	strategy.isRunning = true
	strategy.ticker = time.NewTicker(1 * time.Second)

	// クリーンアップ
	err = strategy.Cleanup()
	assert.NoError(t, err)
	assert.False(t, strategy.isRunning)
	assert.NotNil(t, strategy.stopChan) // 新しいチャネルが作成される

	// 実行中でない場合
	strategy.isRunning = false
	err = strategy.Cleanup()
	assert.NoError(t, err) // エラーは返さない
}

func TestTimeStrategyObserver(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// モックオブザーバーの作成
	mockObserver1 := &MockTimeStrategyObserver{}
	mockObserver2 := &MockTimeStrategyObserver{}

	// オブザーバーの追加
	strategy.AddObserver(mockObserver1)
	strategy.AddObserver(mockObserver2)
	assert.Len(t, strategy.observers, 2)

	// 通知
	strategy.NotifyUpdate("test_event")
	assert.Len(t, mockObserver1.Events, 1)
	assert.Equal(t, "test_event", mockObserver1.Events[0])
	assert.Len(t, mockObserver2.Events, 1)
	assert.Equal(t, "test_event", mockObserver2.Events[0])

	// オブザーバーの削除
	strategy.RemoveObserver(mockObserver1)
	assert.Len(t, strategy.observers, 1)

	// 通知（削除したオブザーバーには通知されない）
	mockObserver1.Events = nil
	mockObserver2.Events = nil
	strategy.NotifyUpdate("another_event")
	assert.Len(t, mockObserver1.Events, 0)
	assert.Len(t, mockObserver2.Events, 1)
	assert.Equal(t, "another_event", mockObserver2.Events[0])
}

func TestTimeStrategyUpdateNextTrigger(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")
	part.ReferenceValueInt = 5 // 5秒

	// 初期化（これによりlogが設定される）
	err := strategy.Initialize(part)
	assert.NoError(t, err)

	// 次のトリガー時間を更新
	now := time.Now()
	strategy.updateNextTrigger()

	// 次のトリガー時間が現在時刻から5秒後に設定されていることを確認
	// 厳密な時間比較は難しいので、前後1秒の範囲内であることを確認
	expectedTime := now.Add(5 * time.Second)
	assert.True(t, strategy.nextTrigger.After(expectedTime.Add(-1*time.Second)))
	assert.True(t, strategy.nextTrigger.Before(expectedTime.Add(1*time.Second)))
}

// タイマーループのテストは複雑なので、簡易的なテストにとどめる
func TestTimeStrategyRun(t *testing.T) {
	// 新しいTimeStrategyを作成
	strategy := NewTimeStrategy()

	// テスト用のConditionPartを作成
	part := entity.NewConditionPart(1, "Test Part")
	part.ReferenceValueInt = 1 // 1秒

	// 初期化（これによりlogが設定される）
	err := strategy.Initialize(part)
	assert.NoError(t, err)

	strategy.interval = 100 * time.Millisecond
	strategy.ticker = time.NewTicker(strategy.interval)

	// モックオブザーバーの作成
	mockObserver := &MockTimeStrategyObserver{}
	strategy.AddObserver(mockObserver)

	// タイマーループを開始（ゴルーチンで実行）
	go strategy.run()

	// 少し待機してタイマーイベントが発生するのを待つ
	time.Sleep(150 * time.Millisecond)

	// タイマーループを停止
	close(strategy.stopChan)

	// タイマーイベントが発生したことを確認
	assert.GreaterOrEqual(t, len(mockObserver.Events), 1)
	if len(mockObserver.Events) > 0 {
		assert.Equal(t, value.EventTimeout, mockObserver.Events[0])
	}
}
