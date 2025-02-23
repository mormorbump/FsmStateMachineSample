package core

import (
	"sync"
	"testing"
	"time"
)

// TestStateTransitionWithObservers は状態遷移とObserver通知の連携をテストします
func TestStateTransitionWithObservers(t *testing.T) {
	helper := NewTestHelper(t)
	subject := NewStateSubjectImpl()
	observer1 := NewMockStateObserver()
	observer2 := NewMockStateObserver()

	// オブザーバーを登録
	subject.AddObserver(observer1)
	subject.AddObserver(observer2)

	// 状態遷移シーケンスのテスト
	states := []string{StateReady, StateActive, StateNext, StateFinish}
	for _, state := range states {
		subject.NotifyStateChanged(state)
	}

	// 各オブザーバーの状態変更履歴を検証
	helper.AssertStateSequence(observer1.GetStateChanges(), states)
	helper.AssertStateSequence(observer2.GetStateChanges(), states)
}

// TestTimerDrivenStateTransition はタイマーイベントによる状態遷移をテストします
func TestTimerDrivenStateTransition(t *testing.T) {
	timer := NewIntervalTimer(50 * time.Millisecond)
	subject := NewStateSubjectImpl()
	stateObserver := NewMockStateObserver()
	timeObserver := NewMockTimeObserver()

	subject.AddObserver(stateObserver)
	timer.AddObserver(timeObserver)

	// タイマー開始時の状態をReadyに設定
	subject.NotifyStateChanged(StateReady)

	// タイマーイベントで状態を変更するハンドラーを設定
	timeObserver.SetOnTimeTick(func() {
		currentState := stateObserver.GetStateChanges()[len(stateObserver.GetStateChanges())-1]
		switch currentState {
		case StateReady:
			subject.NotifyStateChanged(StateActive)
		case StateActive:
			subject.NotifyStateChanged(StateNext)
		case StateNext:
			subject.NotifyStateChanged(StateFinish)
		}
	})

	timer.Start()
	defer timer.Stop()

	// 全状態を遷移するまで待機
	helper := NewTestHelper(t)
	helper.AssertEventually(func() bool {
		states := stateObserver.GetStateChanges()
		return len(states) > 0 && states[len(states)-1] == StateFinish
	}, time.Second, "状態がFinishに到達しませんでした")

	// 期待される状態遷移シーケンスを検証
	expectedStates := []string{StateReady, StateActive, StateNext, StateFinish}
	helper.AssertStateSequence(stateObserver.GetStateChanges(), expectedStates)
}

// TestDeadlockPrevention はデッドロック防止機能をテストします
func TestDeadlockPrevention(t *testing.T) {
	subject := NewStateSubjectImpl()
	stateCounter := NewSafeCounter()

	// 複数のオブザーバーを作成
	const observerCount = 5
	var observers []*MockStateObserver
	var wg sync.WaitGroup
	wg.Add(observerCount)

	// 各オブザーバーは通知を受け取ったら新しい状態に遷移
	for i := 0; i < observerCount; i++ {
		observer := NewMockStateObserver()
		observers = append(observers, observer)
		subject.AddObserver(observer)

		observer.SetOnStateChange(func(state string) {
			defer wg.Done()
			// カウンターをインクリメント
			stateCounter.Increment()
			// 意図的に遅延を入れて競合の可能性を高める
			time.Sleep(time.Millisecond)
		})
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
	if count := stateCounter.GetCount(); count != observerCount {
		t.Errorf("期待される通知回数と異なります: got %d, want %d", count, observerCount)
	}
}

// TestComplexStateTransition は複雑な状態遷移シナリオをテストします
func TestComplexStateTransition(t *testing.T) {
	subject := NewStateSubjectImpl()
	timer := NewIntervalTimer(20 * time.Millisecond)
	observer := NewMockStateObserver()
	timeObserver := NewMockTimeObserver()

	subject.AddObserver(observer)
	timer.AddObserver(timeObserver)

	// 状態遷移の順序を制御するチャネル
	stateChan := make(chan string, 10)
	observer.SetOnStateChange(func(state string) {
		select {
		case stateChan <- state:
		default:
		}
	})

	// 並行して状態を変更
	go func() {
		states := []string{StateReady, StateActive, StateNext, StateFinish}
		for _, state := range states {
			subject.NotifyStateChanged(state)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	timer.Start()
	defer timer.Stop()

	// 一定時間後に結果を検証
	time.Sleep(100 * time.Millisecond)

	// 受信した状態遷移を収集
	var receivedStates []string
	close(stateChan)
	for state := range stateChan {
		receivedStates = append(receivedStates, state)
	}

	// 期待される状態遷移シーケンスを検証
	expectedStates := []string{StateReady, StateActive, StateNext, StateFinish}
	helper := NewTestHelper(t)
	helper.AssertStateSequence(receivedStates, expectedStates)
}
