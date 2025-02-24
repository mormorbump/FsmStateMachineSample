package core

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestStateTransitionWithObservers は状態遷移とObserver通知の連携をテストします
func TestStateTransitionWithObservers(t *testing.T) {
	subject := NewStateSubjectImpl()
	observer1 := NewMockConditionObserver()
	observer2 := NewMockConditionObserver()

	// 期待される状態遷移を設定
	states := []string{StateReady, StateActive, StateNext, StateFinish}
	for _, state := range states {
		observer1.On("OnStateChanged", state).Return()
		observer2.On("OnStateChanged", state).Return()
	}

	// オブザーバーを登録
	subject.AddObserver(observer1)
	subject.AddObserver(observer2)

	// 状態遷移シーケンスのテスト
	for _, state := range states {
		subject.NotifyStateChanged(state)
	}

	// 各オブザーバーの呼び出しを検証
	observer1.AssertExpectations(t)
	observer2.AssertExpectations(t)
}

// TestTimerDrivenStateTransition はタイマーイベントによる状態遷移をテストします
func TestTimerDrivenStateTransition(t *testing.T) {
	timer := NewIntervalTimer(50 * time.Millisecond)
	subject := NewStateSubjectImpl()
	stateObserver := NewMockConditionObserver()
	timeObserver := NewMockTimeObserver()

	// 期待される状態遷移を設定
	stateObserver.On("OnStateChanged", StateReady).Return()
	stateObserver.On("OnStateChanged", StateActive).Return()
	stateObserver.On("OnStateChanged", StateNext).Return()
	stateObserver.On("OnStateChanged", StateFinish).Return()

	// タイマーイベントの期待値を設定
	timeObserver.On("OnTimeTicked").Return()

	subject.AddObserver(stateObserver)
	timer.AddObserver(timeObserver)

	// タイマー開始時の状態をReadyに設定
	subject.NotifyStateChanged(StateReady)

	// タイマーを開始
	timer.Start()
	defer timer.Stop()

	// 状態遷移を実行
	subject.NotifyStateChanged(StateActive)
	subject.NotifyStateChanged(StateNext)
	subject.NotifyStateChanged(StateFinish)

	// 期待される呼び出しを検証
	require.Eventually(t, func() bool {
		return stateObserver.AssertExpectations(t)
	}, time.Second, 10*time.Millisecond, "状態遷移が期待通りに実行されませんでした")
}

// TestDeadlockPrevention はデッドロック防止機能をテストします
func TestDeadlockPrevention(t *testing.T) {
	subject := NewStateSubjectImpl()
	stateCounter := NewSafeCounter()

	// 複数のオブザーバーを作成
	const observerCount = 5
	var observers []*MockConditionObserver
	var wg sync.WaitGroup
	wg.Add(observerCount)

	// 各オブザーバーは通知を受け取ったら新しい状態に遷移
	for i := 0; i < observerCount; i++ {
		observer := NewMockConditionObserver()
		observer.On("OnStateChanged", StateReady).Run(func(args mock.Arguments) {
			defer wg.Done()
			stateCounter.Increment()
			time.Sleep(time.Millisecond)
		}).Return()

		observers = append(observers, observer)
		subject.AddObserver(observer)
	}

	// 初期状態を設定
	subject.NotifyStateChanged(StateReady)

	// タイムアウト付きで待機
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 正常終了
	case <-time.After(time.Second):
		t.Fatal("デッドロックが発生した可能性があります")
	}

	// すべてのオブザーバーが通知を受け取ったことを確認
	assert.Equal(t, observerCount, stateCounter.GetCount(), "期待される通知回数と異なります")
}

// TestComplexStateTransition は複雑な状態遷移シナリオをテストします
func TestComplexStateTransition(t *testing.T) {
	subject := NewStateSubjectImpl()
	timer := NewIntervalTimer(20 * time.Millisecond)
	observer := NewMockConditionObserver()
	timeObserver := NewMockTimeObserver()

	// 期待される状態遷移を設定
	states := []string{StateReady, StateActive, StateNext, StateFinish}
	for _, state := range states {
		observer.On("OnStateChanged", state).Times(1).Return()
	}

	// タイマーイベントの期待値を設定
	timeObserver.On("OnTimeTicked").Return()

	subject.AddObserver(observer)
	timer.AddObserver(timeObserver)

	timer.Start()
	defer timer.Stop()

	// 状態遷移を実行
	for _, state := range states {
		subject.NotifyStateChanged(state)
		time.Sleep(10 * time.Millisecond)
	}

	// 期待される呼び出しを検証
	require.Eventually(t, func() bool {
		return observer.AssertExpectations(t)
	}, time.Second, 10*time.Millisecond, "状態遷移が期待通りに実行されませんでした")
}
