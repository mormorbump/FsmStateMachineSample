package condition

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
	"time"
)

// TimeConditionStrategy は時間ベースの条件評価戦略です
type TimeConditionStrategy struct {
	timer *core.IntervalTimer
}

// NewTimeConditionStrategy は新しいTimeConditionStrategyを作成します
func NewTimeConditionStrategy() *TimeConditionStrategy {
	return &TimeConditionStrategy{}
}

// Initialize は時間条件の戦略を初期化します
func (s *TimeConditionStrategy) Initialize(part Part) error {
	if part.GetReferenceValueInt() <= 0 {
		return fmt.Errorf("invalid time interval: %d", part.GetReferenceValueInt())
	}

	duration := time.Duration(part.GetReferenceValueInt()) * time.Millisecond
	s.timer = core.NewIntervalTimer(duration)
	s.timer.AddObserver(part)

	return nil
}

// Evaluate は時間条件を評価します
func (s *TimeConditionStrategy) Evaluate(ctx context.Context, part Part) error {
	if s.timer == nil {
		return fmt.Errorf("timer not initialized")
	}

	s.timer.Start()
	return nil
}

// Cleanup はタイマーリソースを解放します
func (s *TimeConditionStrategy) Cleanup() error {
	if s.timer != nil {
		s.timer.Stop()
	}
	return nil
}
