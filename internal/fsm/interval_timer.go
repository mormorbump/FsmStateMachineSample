package fsm

import (
	"log"
	"sync"
	"time"
)

// IntervalTimer は時間間隔ベースのイベント通知を管理します
type IntervalTimer struct {
	interval          time.Duration
	isRunning         bool
	ticker            *time.Ticker
	stopChan          chan struct{}
	mu                sync.RWMutex
	nextTrigger       time.Time
	*StateSubjectImpl // 共通実装の埋め込み
}

// NewIntervalTimer は新しいIntervalTimerインスタンスを作成します
func NewIntervalTimer(interval time.Duration) *IntervalTimer {
	log.Printf("Creating new IntervalTimer with interval: %v", interval)
	return &IntervalTimer{
		interval:         interval,
		stopChan:         make(chan struct{}),
		StateSubjectImpl: NewStateSubjectImpl(),
	}
}

// Start は時間管理を開始します
func (t *IntervalTimer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isRunning {
		log.Println("IntervalTimer is already running")
		return
	}

	log.Println("Starting IntervalTimer")
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
		log.Println("IntervalTimer is not running")
		return
	}

	log.Println("Stopping IntervalTimer")
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
	log.Printf("Next event scheduled at: %v", t.nextTrigger)
}

// updateInterval はインターバルを更新します
func (t *IntervalTimer) updateInterval(newInterval time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.interval == newInterval {
		return
	}

	t.interval = newInterval
	log.Printf("Updating interval to: %v", newInterval)

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
	log.Println("Starting timer loop")
	for {
		select {
		case <-t.ticker.C:
			t.NotifyStateChanged("tick") // タイマーイベントを通知
		case <-t.stopChan:
			log.Println("Timer loop stopped")
			return
		}
	}
}
