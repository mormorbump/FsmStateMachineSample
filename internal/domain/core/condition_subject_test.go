package core

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConditionSubjectImpl(t *testing.T) {
	t.Run("AddConditionObserver", func(t *testing.T) {
		subject := NewConditionSubjectImpl()
		observer := NewMockConditionObserver()

		// nilオブザーバーの追加を試みる
		subject.AddConditionObserver(nil)
		assert.Empty(t, subject.observers, "nilオブザーバーが登録されてしまいました")

		// オブザーバーを追加
		subject.AddConditionObserver(observer)
		assert.Len(t, subject.observers, 1, "Observer数が期待値と異なります")

		// 同じオブザーバーを再度追加
		subject.AddConditionObserver(observer)
		assert.Len(t, subject.observers, 2, "Observer数が期待値と異なります（重複を許可）")
	})

	t.Run("RemoveConditionObserver", func(t *testing.T) {
		subject := NewConditionSubjectImpl()
		observer := NewMockConditionObserver()

		// オブザーバーを追加して削除
		subject.AddConditionObserver(observer)
		subject.RemoveConditionObserver(observer)
		assert.Empty(t, subject.observers, "Observerが正しく削除されていません")

		// 存在しないオブザーバーを削除
		subject.RemoveConditionObserver(observer)
		assert.Empty(t, subject.observers, "未登録Observerの削除で予期しない動作が発生しました")

		// nilオブザーバーの削除を試みる
		subject.RemoveConditionObserver(nil)
		assert.Empty(t, subject.observers, "nilオブザーバーの削除で予期しない動作が発生しました")
	})

	t.Run("NotifyConditionSatisfied", func(t *testing.T) {
		subject := NewConditionSubjectImpl()
		observer := NewMockConditionObserver()

		observer.On("OnConditionSatisfied", ConditionID(1)).Return()
		subject.AddConditionObserver(observer)

		// 通知を送信
		subject.NotifyConditionSatisfied(1)
		observer.AssertCalled(t, "OnConditionSatisfied", ConditionID(1))
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		subject := NewConditionSubjectImpl()
		observer := NewMockConditionObserver()
		var wg sync.WaitGroup
		const numGoroutines = 10

		observer.On("OnConditionSatisfied", ConditionID(1)).Return()
		subject.AddConditionObserver(observer)

		// 複数のゴルーチンから同時にアクセス
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				subject.NotifyConditionSatisfied(1)
			}()
		}

		// 完了を待機
		wg.Wait()
		observer.AssertNumberOfCalls(t, "OnConditionSatisfied", numGoroutines)
	})
}

func TestConditionPartSubjectImpl(t *testing.T) {
	t.Run("AddConditionPartObserver", func(t *testing.T) {
		subject := NewConditionPartSubjectImpl()
		observer := NewMockConditionPartObserver()

		// nilオブザーバーの追加を試みる
		subject.AddConditionPartObserver(nil)
		assert.Empty(t, subject.observers, "nilオブザーバーが登録されてしまいました")

		// オブザーバーを追加
		subject.AddConditionPartObserver(observer)
		assert.Len(t, subject.observers, 1, "Observer数が期待値と異なります")

		// 同じオブザーバーを再度追加
		subject.AddConditionPartObserver(observer)
		assert.Len(t, subject.observers, 2, "Observer数が期待値と異なります（重複を許可）")
	})

	t.Run("RemoveConditionPartObserver", func(t *testing.T) {
		subject := NewConditionPartSubjectImpl()
		observer := NewMockConditionPartObserver()

		// オブザーバーを追加して削除
		subject.AddConditionPartObserver(observer)
		subject.RemoveConditionPartObserver(observer)
		assert.Empty(t, subject.observers, "Observerが正しく削除されていません")

		// 存在しないオブザーバーを削除
		subject.RemoveConditionPartObserver(observer)
		assert.Empty(t, subject.observers, "未登録Observerの削除で予期しない動作が発生しました")

		// nilオブザーバーの削除を試みる
		subject.RemoveConditionPartObserver(nil)
		assert.Empty(t, subject.observers, "nilオブザーバーの削除で予期しない動作が発生しました")
	})

	t.Run("NotifyPartSatisfied", func(t *testing.T) {
		subject := NewConditionPartSubjectImpl()
		observer := NewMockConditionPartObserver()

		observer.On("OnPartSatisfied", ConditionPartID(1)).Return()
		subject.AddConditionPartObserver(observer)

		// 通知を送信
		subject.NotifyPartSatisfied(1)
		observer.AssertCalled(t, "OnPartSatisfied", ConditionPartID(1))
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		subject := NewConditionPartSubjectImpl()
		observer := NewMockConditionPartObserver()
		var wg sync.WaitGroup
		const numGoroutines = 10

		observer.On("OnPartSatisfied", ConditionPartID(1)).Return()
		subject.AddConditionPartObserver(observer)

		// 複数のゴルーチンから同時にアクセス
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				subject.NotifyPartSatisfied(1)
			}()
		}

		// 完了を待機
		wg.Wait()
		observer.AssertNumberOfCalls(t, "OnPartSatisfied", numGoroutines)
	})
}
