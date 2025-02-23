package core

import (
	"sync"
	"testing"
	"time"
)

func TestStateSubjectImpl_AddObserver(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*StateSubjectImpl, *MockStateObserver)
		validate func(*testing.T, *StateSubjectImpl, *MockStateObserver)
	}{
		{
			name: "正常なObserver登録",
			setup: func(s *StateSubjectImpl, m *MockStateObserver) {
				s.AddObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockStateObserver) {
				if len(s.observers) != 1 {
					t.Errorf("Observer数 = %d, want 1", len(s.observers))
				}
			},
		},
		{
			name: "重複Observer登録",
			setup: func(s *StateSubjectImpl, m *MockStateObserver) {
				s.AddObserver(m)
				s.AddObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockStateObserver) {
				if len(s.observers) != 2 {
					t.Errorf("Observer数 = %d, want 2 (重複を許可)", len(s.observers))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewStateSubjectImpl()
			observer := NewMockStateObserver()
			
			tt.setup(subject, observer)
			tt.validate(t, subject, observer)
		})
	}
}

func TestStateSubjectImpl_RemoveObserver(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*StateSubjectImpl, *MockStateObserver)
		validate func(*testing.T, *StateSubjectImpl, *MockStateObserver)
	}{
		{
			name: "登録済みObserverの削除",
			setup: func(s *StateSubjectImpl, m *MockStateObserver) {
				s.AddObserver(m)
				s.RemoveObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockStateObserver) {
				if len(s.observers) != 0 {
					t.Errorf("Observer数 = %d, want 0", len(s.observers))
				}
			},
		},
		{
			name: "未登録Observerの削除",
			setup: func(s *StateSubjectImpl, m *MockStateObserver) {
				s.RemoveObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockStateObserver) {
				if len(s.observers) != 0 {
					t.Errorf("Observer数 = %d, want 0", len(s.observers))
				}
			},
		},
		{
			name: "複数回削除",
			setup: func(s *StateSubjectImpl, m *MockStateObserver) {
				s.AddObserver(m)
				s.RemoveObserver(m)
				s.RemoveObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockStateObserver) {
				if len(s.observers) != 0 {
					t.Errorf("Observer数 = %d, want 0", len(s.observers))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewStateSubjectImpl()
			observer := NewMockStateObserver()
			
			tt.setup(subject, observer)
			tt.validate(t, subject, observer)
		})
	}
}

func TestStateSubjectImpl_NotifyStateChanged(t *testing.T) {
	tests := []struct {
		name          string
		observerCount int
		state         string
		wantState     string
	}{
		{
			name:          "単一Observerへの通知",
			observerCount: 1,
			state:         "test_state",
			wantState:     "test_state",
		},
		{
			name:          "複数Observerへの通知",
			observerCount: 3,
			state:         "test_state",
			wantState:     "test_state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewStateSubjectImpl()
			observers := make([]*MockStateObserver, tt.observerCount)
			
			// Observerの作成と登録
			for i := 0; i < tt.observerCount; i++ {
				observers[i] = NewMockStateObserver()
				subject.AddObserver(observers[i])
			}

			// 状態変更通知
			subject.NotifyStateChanged(tt.state)

			// 各Observerの状態確認
			for i, observer := range observers {
				states := observer.GetStateChanges()
				if len(states) != 1 {
					t.Errorf("Observer %d: 状態変更回数 = %d, want 1", i, len(states))
				}
				if len(states) > 0 && states[0] != tt.wantState {
					t.Errorf("Observer %d: 状態 = %s, want %s", i, states[0], tt.wantState)
				}
			}
		})
	}
}

func TestStateSubjectImpl_ConcurrentAccess(t *testing.T) {
	subject := NewStateSubjectImpl()
	observer := NewMockStateObserver()
	subject.AddObserver(observer)

	// 並行アクセスのテスト
	const (
		goroutineCount = 10
		operationCount = 100
	)

	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	// 複数のゴルーチンで同時に操作
	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationCount; j++ {
				switch j % 3 {
				case 0:
					subject.AddObserver(observer)
				case 1:
					subject.RemoveObserver(observer)
				case 2:
					subject.NotifyStateChanged("test_state")
				}
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

func TestStateSubjectImpl_EdgeCases(t *testing.T) {
	t.Run("nilオブザーバーの登録と削除", func(t *testing.T) {
		subject := NewStateSubjectImpl()
		
		// nilオブザーバーの登録を試みる
		subject.AddObserver(nil)
		if len(subject.observers) != 0 {
			t.Error("nilオブザーバーが登録されてしまいました")
		}

		// nilオブザーバーの削除を試みる
		subject.RemoveObserver(nil)
		if len(subject.observers) != 0 {
			t.Error("nilオブザーバーの削除で予期しない動作が発生しました")
		}
	})

	t.Run("通知中のオブザーバー削除", func(t *testing.T) {
		subject := NewStateSubjectImpl()
		observer := NewMockStateObserver()
		
		// 通知中に自身を削除するオブザーバー
		observer.SetOnStateChange(func(state string) {
			subject.RemoveObserver(observer)
		})

		subject.AddObserver(observer)
		subject.NotifyStateChanged("test_state")

		// 通知後にオブザーバーが正しく削除されていることを確認
		if len(subject.observers) != 0 {
			t.Error("通知中のオブザーバー削除が正しく動作していません")
		}
	})
}