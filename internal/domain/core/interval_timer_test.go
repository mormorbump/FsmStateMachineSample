package core

import (
	"sync"
	"testing"
	"time"
)

func TestIntervalTimer_BasicOperations(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "タイマー開始と停止",
			fn: func(t *testing.T) {
				timer := NewIntervalTimer(100 * time.Millisecond)
				observer := NewMockTimeObserver()
				timer.AddObserver(observer)

				// タイマー開始
				timer.Start()

				// 最初のティックを待機
				if !observer.WaitForTick(200 * time.Millisecond) {
					t.Error("タイマーのティックが発生しませんでした")
				}

				// タイマー停止
				timer.Stop()

				// 停止後のティック数を記録
				initialCount := observer.GetTickCount()

				// 十分な時間待機
				time.Sleep(200 * time.Millisecond)

				// ティック数が増えていないことを確認
				if count := observer.GetTickCount(); count != initialCount {
					t.Errorf("停止後もティックが発生: count = %d, want %d", count, initialCount)
				}
			},
		},
		{
			name: "インターバル更新",
			fn: func(t *testing.T) {
				timer := NewIntervalTimer(1 * time.Second)
				observer := NewMockTimeObserver()
				timer.AddObserver(observer)

				timer.Start()
				timer.UpdateInterval(50 * time.Millisecond)

				// 更新後のインターバルでティックが発生することを確認
				if !observer.WaitForTick(100 * time.Millisecond) {
					t.Error("更新後のインターバルでティックが発生しませんでした")
				}

				timer.Stop()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestIntervalTimer_ObserverNotification(t *testing.T) {
	timer := NewIntervalTimer(50 * time.Millisecond)

	// 複数のオブザーバーを登録
	observers := make([]*MockTimeObserver, 3)
	for i := range observers {
		observers[i] = NewMockTimeObserver()
		timer.AddObserver(observers[i])
	}

	timer.Start()
	defer timer.Stop()

	// すべてのオブザーバーが通知を受け取ることを確認
	for i, observer := range observers {
		if !observer.WaitForTick(100 * time.Millisecond) {
			t.Errorf("Observer %d: ティック通知を受信できませんでした", i)
		}
	}
}

func TestIntervalTimer_ConcurrentAccess(t *testing.T) {
	timer := NewIntervalTimer(20 * time.Millisecond)
	observer := NewMockTimeObserver()
	timer.AddObserver(observer)

	const (
		goroutineCount = 10
		operationCount = 50
	)

	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	// 複数のゴルーチンで同時に操作
	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationCount; j++ {
				switch j % 4 {
				case 0:
					timer.Start()
				case 1:
					timer.Stop()
				case 2:
					timer.UpdateInterval(time.Duration(20+j) * time.Millisecond)
				case 3:
					timer.AddObserver(observer)
					timer.RemoveObserver(observer)
				}
				time.Sleep(time.Millisecond) // 競合の可能性を高める
			}
		}(i)
	}

	// タイムアウト付きで待機
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 正常終了
	case <-time.After(5 * time.Second):
		t.Fatal("タイムアウト: 並行アクセステストが完了しませんでした")
	}
}

func TestIntervalTimer_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "重複開始",
			fn: func(t *testing.T) {
				timer := NewIntervalTimer(50 * time.Millisecond)
				observer := NewMockTimeObserver()
				timer.AddObserver(observer)

				timer.Start()
				timer.Start() // 2回目の開始

				// 正常にティックが発生することを確認
				if !observer.WaitForTick(100 * time.Millisecond) {
					t.Error("ティックが発生しませんでした")
				}

				timer.Stop()
			},
		},
		{
			name: "停止済みタイマーの停止",
			fn: func(t *testing.T) {
				timer := NewIntervalTimer(50 * time.Millisecond)
				timer.Stop() // 開始前の停止
				timer.Stop() // 重複停止
			},
		},
		{
			name: "極端に短いインターバル",
			fn: func(t *testing.T) {
				timer := NewIntervalTimer(1 * time.Microsecond)
				observer := NewMockTimeObserver()
				timer.AddObserver(observer)

				timer.Start()

				// 短いインターバルでもティックが発生することを確認
				if !observer.WaitForTick(100 * time.Millisecond) {
					t.Error("極端に短いインターバルでティックが発生しませんでした")
				}

				timer.Stop()
			},
		},
		{
			name: "nilオブザーバー",
			fn: func(t *testing.T) {
				timer := NewIntervalTimer(50 * time.Millisecond)

				// nilオブザーバーの追加と削除
				timer.AddObserver(nil)
				timer.RemoveObserver(nil)

				// 正常に動作することを確認
				timer.Start()
				time.Sleep(100 * time.Millisecond)
				timer.Stop()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestIntervalTimer_ObserverManagement(t *testing.T) {
	timer := NewIntervalTimer(50 * time.Millisecond)
	observer1 := NewMockTimeObserver()
	observer2 := NewMockTimeObserver()

	// オブザーバーの追加
	timer.AddObserver(observer1)
	timer.AddObserver(observer2)

	timer.Start()

	// 両方のオブザーバーが通知を受け取ることを確認
	if !observer1.WaitForTick(100*time.Millisecond) || !observer2.WaitForTick(100*time.Millisecond) {
		t.Error("いずれかのオブザーバーがティック通知を受信できませんでした")
	}

	// observer1を削除
	timer.RemoveObserver(observer1)

	// ティックカウントをリセット
	initialCount1 := observer1.GetTickCount()
	initialCount2 := observer2.GetTickCount()

	// 十分な時間待機
	time.Sleep(100 * time.Millisecond)

	// observer1のカウントが変わっていないことを確認
	if count := observer1.GetTickCount(); count != initialCount1 {
		t.Errorf("削除されたobserver1がティック通知を受信: count = %d, want %d", count, initialCount1)
	}

	// observer2のカウントが増えていることを確認
	if count := observer2.GetTickCount(); count <= initialCount2 {
		t.Error("observer2がティック通知を受信していません")
	}

	timer.Stop()
}
