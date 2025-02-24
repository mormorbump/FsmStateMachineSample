package core

import (
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockConditionObserver ConditionObserver インターフェースのモック実装
type MockConditionObserver struct {
	mock.Mock
}

func NewMockConditionObserver() *MockConditionObserver {
	return &MockConditionObserver{}
}

func (m *MockConditionObserver) OnConditionSatisfied(conditionID ConditionID) {
	m.Called(conditionID)
}

func (m *MockConditionObserver) OnStateChanged(state string) {
	m.Called(state)
}

// MockConditionPartObserver ConditionPartObserver インターフェースのモック実装
type MockConditionPartObserver struct {
	mock.Mock
}

func NewMockConditionPartObserver() *MockConditionPartObserver {
	return &MockConditionPartObserver{}
}

func (m *MockConditionPartObserver) OnPartSatisfied(partID ConditionPartID) {
	m.Called(partID)
}

func (m *MockConditionPartObserver) OnStateChanged(state string) {
	m.Called(state)
}

// MockTimeObserver TimeObserver インターフェースのモック実装
type MockTimeObserver struct {
	mock.Mock
	waitForTick chan struct{}
	tickCount   int
	mu          sync.Mutex
}

func NewMockTimeObserver() *MockTimeObserver {
	return &MockTimeObserver{
		waitForTick: make(chan struct{}, 1),
	}
}

func (m *MockTimeObserver) OnTimeTicked() {
	m.Called()
	m.mu.Lock()
	m.tickCount++
	m.mu.Unlock()
	select {
	case m.waitForTick <- struct{}{}:
	default:
	}
}

func (m *MockTimeObserver) GetTickCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tickCount
}

func (m *MockTimeObserver) WaitForTick(timeout time.Duration) bool {
	select {
	case <-m.waitForTick:
		return true
	case <-time.After(timeout):
		return false
	}
}

// MockTimer タイマーのモック実装
type MockTimer struct {
	C       chan time.Time
	stopped bool
	mu      sync.Mutex
}

func NewMockTimer() *MockTimer {
	return &MockTimer{
		C: make(chan time.Time, 1),
	}
}

func (m *MockTimer) Stop() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return false
	}
	m.stopped = true
	return true
}

func (m *MockTimer) Reset(d time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	wasActive := !m.stopped
	m.stopped = false
	return wasActive
}

func (m *MockTimer) Tick() {
	select {
	case m.C <- time.Now():
	default:
	}
}

// SafeCounter スレッドセーフなカウンター
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func NewSafeCounter() *SafeCounter {
	return &SafeCounter{}
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounter) GetCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}
