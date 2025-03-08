package strategy

import (
	"context"
	"fmt"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TimeStrategy は時間ベースの条件評価戦略です
type TimeStrategy struct {
	observers   []service.StrategyObserver
	interval    time.Duration
	isRunning   bool
	ticker      *time.Ticker
	stopChan    chan struct{}
	mu          sync.RWMutex
	nextTrigger time.Time
	log         *zap.Logger
}

// NewTimeStrategy は新しいTimeStrategyを作成します
func NewTimeStrategy() *TimeStrategy {
	return &TimeStrategy{
		observers: make([]service.StrategyObserver, 0),
	}
}

// Initialize は戦略の初期化を行います
func (s *TimeStrategy) Initialize(part interface{}) error {
	condPart, ok := part.(*entity.ConditionPart)
	if !ok {
		return fmt.Errorf("invalid part type: expected *entity.ConditionPart, got %T", part)
	}

	if condPart.GetReferenceValueInt() <= 0 {
		return fmt.Errorf("invalid time interval: %d", condPart.GetReferenceValueInt())
	}

	s.log = logger.DefaultLogger()
	duration := time.Duration(condPart.GetReferenceValueInt()) * time.Second
	s.interval = duration
	s.AddObserver(condPart)
	s.stopChan = make(chan struct{})

	return nil
}

// GetCurrentValue は現在の値を返します
func (s *TimeStrategy) GetCurrentValue() interface{} {
	return nil
}

// Start は時間条件の評価を開始します
func (s *TimeStrategy) Start(ctx context.Context, part interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		s.log.Debug("IntervalTimer is already running")
		return nil
	}

	// インターバルが設定されていることを確認
	if s.interval <= 0 {
		s.log.Error("Invalid interval in Start", zap.Duration("interval", s.interval))
		return fmt.Errorf("invalid interval: %v", s.interval)
	}

	s.log.Debug("Starting IntervalTimer", zap.Duration("interval", s.interval))
	s.isRunning = true
	s.ticker = time.NewTicker(s.interval)
	s.updateNextTrigger()

	// stopChanが閉じられていないことを確認
	select {
	case <-s.stopChan:
		// stopChanが閉じられている場合は再作成
		s.log.Debug("Recreating stopChan")
		s.stopChan = make(chan struct{})
	default:
		// 問題なし
	}

	go s.run()
	return nil
}

func (s *TimeStrategy) Evaluate(ctx context.Context, part interface{}, params interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO 再生、停止機能を追加するならここ
	//s.isRunning = true
	//s.ticker = time.NewTicker(s.interval)
	//s.updateNextTrigger()

	return nil
}

// Cleanup はタイマーリソースを解放します
func (s *TimeStrategy) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		s.log.Debug("IntervalTimer is not running")
		return nil
	}

	s.log.Debug("Stopping IntervalTimer")
	s.isRunning = false

	// タイマーを停止
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
		s.log.Debug("Ticker stopped and set to nil")
	}

	// stopChanを閉じる（すでに閉じられていないことを確認）
	select {
	case <-s.stopChan:
		// すでに閉じられている
		s.log.Debug("stopChan was already closed")
	default:
		// まだ閉じられていない
		s.log.Debug("Closing stopChan")
		close(s.stopChan)
	}

	// 新しいstopChanを作成
	s.stopChan = make(chan struct{})
	s.log.Debug("Created new stopChan")

	// observersをクリア
	s.observers = make([]service.StrategyObserver, 0)
	s.log.Debug("Cleared observers")

	return nil
}

// 次のトリガー時間を更新します
func (s *TimeStrategy) updateNextTrigger() {
	s.nextTrigger = time.Now().Add(s.interval)
	s.log.Debug("Next event scheduled at", zap.Time("next_trigger", s.nextTrigger))
}

// タイマーループを実行します
func (s *TimeStrategy) run() {
	s.log.Debug("Starting timer loop")
	defer s.log.Debug("Timer loop exited")

	// タイマーが初期化されていることを確認
	s.mu.RLock()
	ticker := s.ticker
	stopChan := s.stopChan
	interval := s.interval
	s.mu.RUnlock()

	if ticker == nil {
		s.log.Error("Timer is nil in run()")
		return
	}

	s.log.Debug("Timer loop started with interval", zap.Duration("interval", interval))

	// タイマーループ
	for {
		select {
		case <-ticker.C:
			s.log.Debug("Timer tick received")
			// タイマーが停止されていないことを確認
			s.mu.RLock()
			isRunning := s.isRunning
			s.mu.RUnlock()

			if !isRunning {
				s.log.Debug("Timer is no longer running, exiting loop")
				return
			}

			s.log.Debug("Notifying observers about timeout event")
			s.NotifyUpdate(value.EventTimeout)
			s.log.Debug("Notification complete")

		case <-stopChan:
			s.log.Debug("Stop signal received, timer loop stopped")
			return
		}
	}
}

// AddObserver オブザーバーを追加します
func (s *TimeStrategy) AddObserver(observer service.StrategyObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

// RemoveObserver オブザーバーを削除します
func (s *TimeStrategy) RemoveObserver(observer service.StrategyObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// NotifyUpdate オブザーバーに更新を通知します
func (s *TimeStrategy) NotifyUpdate(event string) {
	// ロガーが初期化されていない場合は初期化する
	if s.log == nil {
		s.log = logger.DefaultLogger()
	}
	
	s.log.Debug("TimeStrategy.NotifyUpdate", zap.String("event", event))
	s.mu.RLock()
	observers := make([]service.StrategyObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()

	for _, observer := range observers {
		observer.OnUpdated(event)
	}
}
