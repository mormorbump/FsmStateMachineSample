package condition_details

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
	"state_sample/internal/domain/entity"
	"time"
)

// TimeConditionStrategy 時間ベースの条件評価戦略
type TimeConditionStrategy struct {
	timer *core.IntervalTimer
}

func NewTimeConditionStrategy() *TimeConditionStrategy {
	return &TimeConditionStrategy{}
}

func (s *TimeConditionStrategy) Initialize(part *entity.ConditionPart) error {
	if part.ReferenceValueInt <= 0 {
		return fmt.Errorf("invalid time interval: %d", part.ReferenceValueInt)
	}

	// ConditionKindがTimeの時はReferenceValueIntが秒数
	duration := time.Duration(part.ReferenceValueInt) * time.Second
	s.timer = core.NewIntervalTimer(duration)
	s.timer.AddObserver(part)

	return nil
}

func (s *TimeConditionStrategy) Evaluate(ctx context.Context, part *entity.ConditionPart) error {
	if s.timer == nil {
		return fmt.Errorf("timer not initialized")
	}

	s.timer.Start()
	return nil
}

func (s *TimeConditionStrategy) Cleanup() error {
	if s.timer != nil {
		s.timer.Stop()
	}
	return nil
}
