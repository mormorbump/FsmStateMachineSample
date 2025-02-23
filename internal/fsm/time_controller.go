package fsm

import (
	"log"
	"sync"
	"time"
)

// TimeController は時間ベースのイベント通知を管理します
type TimeController struct {
	interval    time.Duration
	isRunning   bool
	ticker      *time.Ticker
	stopChan    chan struct{}
	mu          sync.RWMutex
	nextTrigger time.Time
	onTick      func() // 時間経過時のコールバック
}

// NewTimeController は新しいTimeControllerインスタンスを作成します
func NewTimeController(interval time.Duration, onTick func()) *TimeController {
	log.Printf("Creating new TimeController with interval: %v", interval)
	return &TimeController{
		interval: interval,
		onTick:  onTick,
		stopChan: make(chan struct{}),
	}
}

// Start は時間管理を開始します
func (tc *TimeController) Start() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.isRunning {
		log.Println("TimeController is already running")
		return
	}

	log.Println("Starting TimeController")
	tc.isRunning = true
	tc.ticker = time.NewTicker(tc.interval)
	tc.updateNextTrigger()

	go tc.run()
}

// Stop は時間管理を停止します
func (tc *TimeController) Stop() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if !tc.isRunning {
		log.Println("TimeController is not running")
		return
	}

	log.Println("Stopping TimeController")
	tc.isRunning = false
	if tc.ticker != nil {
		tc.ticker.Stop()
	}
	close(tc.stopChan)
	tc.stopChan = make(chan struct{})
}

// IsRunning は時間管理が実行中かどうかを返します
func (tc *TimeController) IsRunning() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.isRunning
}

// GetNextTrigger は次のイベント予定時刻を返します
func (tc *TimeController) GetNextTrigger() time.Time {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.nextTrigger
}

// updateNextTrigger は次のイベント予定時刻を更新します
func (tc *TimeController) updateNextTrigger() {
	tc.nextTrigger = time.Now().Add(tc.interval)
	log.Printf("Next event scheduled at: %v", tc.nextTrigger)
}

// run は時間管理のメインループを実行します
func (tc *TimeController) run() {
	log.Println("Starting timer loop")
	for {
		select {
		case <-tc.ticker.C:
			if tc.onTick != nil {
				tc.onTick()
			}
			tc.updateNextTrigger()
		case <-tc.stopChan:
			log.Println("Timer loop stopped")
			return
		}
	}
}
