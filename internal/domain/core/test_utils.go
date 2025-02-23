package core

import (
	"sync"
	"testing"
	"time"
)

// MockStateObserver は StateObserver インターフェースのモック実装です
type MockStateObserver struct {
	mu            sync.Mutex
	stateChanges  []string
	onStateChange func(state string)
}

func NewMockStateObserver() *MockStateObserver {
	return &MockStateObserver{
		stateChanges: make([]string, 0),
	}
}

func (m *MockStateObserver) OnStateChanged(state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stateChanges = append(m.stateChanges, state)
	if m.onStateChange != nil {
		m.onStateChange(state)
	}
}

func (m *MockStateObserver) GetStateChanges() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.stateChanges))
	copy(result, m.stateChanges)
	return result
}

func (m *MockStateObserver) SetOnStateChange(f func(state string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onStateChange = f
}

// MockTimeObserver は TimeObserver インターフェースのモック実装です
type MockTimeObserver struct {
	mu          sync.Mutex
	tickCount   int
	onTimeTick  func()
	waitForTick chan struct{}
}

func NewMockTimeObserver() *MockTimeObserver {
	return &MockTimeObserver{
		waitForTick: make(chan struct{}, 1),
	}
}

func (m *MockTimeObserver) OnTimeTicked() {
	m.mu.Lock()
	m.tickCount++
	if m.onTimeTick != nil {
		m.onTimeTick()
	}
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

func (m *MockTimeObserver) SetOnTimeTick(f func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onTimeTick = f
}

func (m *MockTimeObserver) WaitForTick(timeout time.Duration) bool {
	select {
	case <-m.waitForTick:
		return true
	case <-time.After(timeout):
		return false
	}
}

// TestHelper はテストで共通して使用するヘルパー関数を提供します
type TestHelper struct {
	t *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// AssertStateSequence は状態遷移のシーケンスが期待通りかを検証します
func (h *TestHelper) AssertStateSequence(got []string, want []string) {
	h.t.Helper()
	if len(got) != len(want) {
		h.t.Errorf("状態遷移回数が異なります。got %d, want %d", len(got), len(want))
		return
	}
	for i, state := range got {
		if state != want[i] {
			h.t.Errorf("状態遷移が異なります。index %d: got %s, want %s", i, state, want[i])
		}
	}
}

// AssertEventually は指定された条件が一定時間内に満たされることを検証します
func (h *TestHelper) AssertEventually(condition func() bool, timeout time.Duration, message string) {
	h.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	h.t.Errorf("タイムアウト: %s", message)
}

// WaitForCondition は指定された条件が満たされるまで待機します
func (h *TestHelper) WaitForCondition(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// MockTimer はtime.Timerのモック実装です
type MockTimer struct {
	C      chan time.Time
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

// SafeCounter はスレッドセーフなカウンターを提供します
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