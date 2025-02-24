package core

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStateSubjectImpl_AddObserver(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*StateSubjectImpl, *MockConditionObserver)
		validate func(*testing.T, *StateSubjectImpl, *MockConditionObserver)
	}{
		{
			name: "正常なObserver登録",
			setup: func(s *StateSubjectImpl, m *MockConditionObserver) {
				s.AddObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockConditionObserver) {
				assert.Len(t, s.observers, 1, "Observer数が期待値と異なります")
			},
		},
		{
			name: "重複Observer登録",
			setup: func(s *StateSubjectImpl, m *MockConditionObserver) {
				s.AddObserver(m)
				s.AddObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockConditionObserver) {
				assert.Len(t, s.observers, 2, "Observer数が期待値と異なります（重複を許可）")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewStateSubjectImpl()
			observer := NewMockConditionObserver()

			tt.setup(subject, observer)
			tt.validate(t, subject, observer)
		})
	}
}

func TestStateSubjectImpl_RemoveObserver(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*StateSubjectImpl, *MockConditionObserver)
		validate func(*testing.T, *StateSubjectImpl, *MockConditionObserver)
	}{
		{
			name: "登録済みObserverの削除",
			setup: func(s *StateSubjectImpl, m *MockConditionObserver) {
				s.AddObserver(m)
				s.RemoveObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockConditionObserver) {
				assert.Empty(t, s.observers, "Observerが正しく削除されていません")
			},
		},
		{
			name: "未登録Observerの削除",
			setup: func(s *StateSubjectImpl, m *MockConditionObserver) {
				s.RemoveObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockConditionObserver) {
				assert.Empty(t, s.observers, "未登録Observerの削除で予期しない動作が発生しました")
			},
		},
		{
			name: "複数回削除",
			setup: func(s *StateSubjectImpl, m *MockConditionObserver) {
				s.AddObserver(m)
				s.RemoveObserver(m)
				s.RemoveObserver(m)
			},
			validate: func(t *testing.T, s *StateSubjectImpl, m *MockConditionObserver) {
				assert.Empty(t, s.observers, "複数回削除で予期しない動作が発生しました")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewStateSubjectImpl()
			observer := NewMockConditionObserver()

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
	}{
		{
			name:          "単一Observerへの通知",
			observerCount: 1,
			state:         "test_state",
		},
		{
			name:          "複数Observerへの通知",
			observerCount: 3,
			state:         "test_state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewStateSubjectImpl()
			observers := make([]*MockConditionObserver, tt.observerCount)

			// Observerの作成と登録
			for i := 0; i < tt.observerCount; i++ {
				observers[i] = NewMockConditionObserver()
				observers[i].On("OnStateChanged", tt.state).Return()
				subject.AddObserver(observers[i])
			}

			// 状態変更通知
			subject.NotifyStateChanged(tt.state)

			// 各Observerの呼び出しを検証
			for _, observer := range observers {
				observer.AssertCalled(t, "OnStateChanged", tt.state)
			}
		})
	}
}

func TestStateSubjectImpl_ConcurrentAccess(t *testing.T) {
	subject := NewStateSubjectImpl()
	observer := NewMockConditionObserver()
	observer.On("OnStateChanged", "test_state").Return()
	subject.AddObserver(observer)

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
		assert.Empty(t, subject.observers, "nilオブザーバーが登録されてしまいました")

		// nilオブザーバーの削除を試みる
		subject.RemoveObserver(nil)
		assert.Empty(t, subject.observers, "nilオブザーバーの削除で予期しない動作が発生しました")
	})

	t.Run("通知中のオブザーバー削除", func(t *testing.T) {
		subject := NewStateSubjectImpl()
		observer := NewMockConditionObserver()
		observer.On("OnStateChanged", "test_state").Run(func(args mock.Arguments) {
			subject.RemoveObserver(observer)
		}).Return()

		subject.AddObserver(observer)
		subject.NotifyStateChanged("test_state")

		assert.Empty(t, subject.observers, "通知中のオブザーバー削除が正しく動作していません")
		observer.AssertExpectations(t)
	})
}
