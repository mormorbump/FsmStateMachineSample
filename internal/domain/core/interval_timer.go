package core

import (
	"go.uber.org/zap"
	logger "state_sample/internal/lib"
	"sync"
	"time"
)

type TimeSubject interface {
	AddObserver(observer TimeObserver)
	RemoveObserver(observer TimeObserver)
	NotifyTimeTicker()
}

// IntervalTimer は時間間隔ベースのイベント通知を管理します
type IntervalTimer struct {
	observers   []TimeObserver
	interval    time.Duration
	isRunning   bool
	ticker      *time.Ticker
	stopChan    chan struct{}
	mu          sync.RWMutex
	nextTrigger time.Time
	log         *zap.Logger
}

// NewIntervalTimer は新しいIntervalTimerインスタンスを作成します
func NewIntervalTimer(interval time.Duration) *IntervalTimer {
	log := logger.DefaultLogger()
	log.Debug("Creating new IntervalTimer with interval: %v", zap.Duration("interval", interval))
	return &IntervalTimer{
		interval: interval,
		stopChan: make(chan struct{}),
		log:      log,
	}
}

// Start は時間管理を開始します
func (t *IntervalTimer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isRunning {
		t.log.Debug("IntervalTimer is already running")
		return
	}

	t.log.Debug("Starting IntervalTimer")
	t.isRunning = true
	t.ticker = time.NewTicker(t.interval)
	t.updateNextTrigger()

	go t.run()
}

// Stop は時間管理を停止します
func (t *IntervalTimer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isRunning {
		t.log.Debug("IntervalTimer is not running")
		return
	}

	t.log.Debug("Stopping IntervalTimer")
	t.isRunning = false
	if t.ticker != nil {
		t.ticker.Stop()
	}
	close(t.stopChan)
	t.stopChan = make(chan struct{})
}

// IsRunning は時間管理が実行中かどうかを返します
func (t *IntervalTimer) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isRunning
}

// GetNextTrigger は次のイベント予定時刻を返します
func (t *IntervalTimer) GetNextTrigger() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nextTrigger
}

// updateNextTrigger は次のイベント予定時刻を更新します
func (t *IntervalTimer) updateNextTrigger() {
	t.nextTrigger = time.Now().Add(t.interval)
	t.log.Debug("Next event scheduled at", zap.Time("next_trigger", t.nextTrigger))
}

// UpdateInterval はインターバルを更新します
func (t *IntervalTimer) UpdateInterval(newInterval time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.interval == newInterval {
		return
	}

	t.interval = newInterval
	t.log.Debug("Updating interval to", zap.Duration("new_interval", newInterval))

	// タイマーが実行中の場合は再起動
	if t.isRunning {
		if t.ticker != nil {
			t.ticker.Stop()
		}
		t.ticker = time.NewTicker(t.interval)
		t.updateNextTrigger()
	}
}

// GetInterval は現在のインターバルを返します
func (t *IntervalTimer) GetInterval() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.interval
}

// run は時間管理のメインループを実行します
func (t *IntervalTimer) run() {
	t.log.Debug("Starting timer loop")
	for {
		select {
		case <-t.ticker.C:
			t.log.Debug("Timer tick")
			t.NotifyTimeTicker()
		case <-t.stopChan:
			t.log.Debug("Timer loop stopped")
			return
		}
	}
}

// AddObserver は監視者を追加します
func (t *IntervalTimer) AddObserver(observer TimeObserver) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.observers = append(t.observers, observer)
}

// RemoveObserver は監視者を削除します
func (t *IntervalTimer) RemoveObserver(observer TimeObserver) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, obs := range t.observers {
		if obs == observer {
			t.observers = append(t.observers[:i], t.observers[i+1:]...)
			break
		}
	}
}

func (t *IntervalTimer) NotifyTimeTicker() {
	t.mu.RLock()
	observers := make([]TimeObserver, len(t.observers))
	copy(observers, t.observers)
	t.mu.RUnlock()

	for _, observer := range observers {
		observer.OnTimeTicked()
	}
}
