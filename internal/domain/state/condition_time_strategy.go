package state

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
	"time"
)

// ConditionTimeStrategy は時間ベースの条件評価戦略です
type ConditionTimeStrategy struct {
	timer *core.IntervalTimer
}

// NewTimeConditionStrategy は新しいTimeConditionStrategyを作成します
func NewTimeConditionStrategy() *ConditionTimeStrategy {
	return &ConditionTimeStrategy{}
}

func (s *ConditionTimeStrategy) Initialize(part *ConditionPart) error {
	if part.GetReferenceValueInt() <= 0 {
		return fmt.Errorf("invalid time interval: %d", part.GetReferenceValueInt())
	}

	duration := time.Duration(part.GetReferenceValueInt()) * time.Second
	s.timer = core.NewIntervalTimer(duration)
	s.timer.AddObserver(part)

	return nil
}

// Evaluate は時間条件を評価します
func (s *ConditionTimeStrategy) Evaluate(ctx context.Context, part *ConditionPart, params interface{}) error {
	if s.timer == nil {
		return fmt.Errorf("timer not initialized")
	}

	s.timer.Start()
	return nil
}

// Cleanup はタイマーリソースを解放します
func (s *ConditionTimeStrategy) Cleanup() error {
	if s.timer != nil {
		s.timer.Stop()
	}
	return nil
}
