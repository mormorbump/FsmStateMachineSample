package strategy

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"
)

// CounterStrategy はカウンターベースの条件評価戦略です
type CounterStrategy struct {
	currentValue int64
	observers    []service.StrategyObserver
	mu           sync.RWMutex
}

// NewCounterStrategy は新しいCounterStrategyを作成します
func NewCounterStrategy() *CounterStrategy {
	return &CounterStrategy{
		currentValue: 0,
		observers:    make([]service.StrategyObserver, 0),
	}
}

// Initialize は戦略の初期化を行います
func (s *CounterStrategy) Initialize(part interface{}) error {
	condPart, ok := part.(*entity.ConditionPart)
	if !ok {
		return fmt.Errorf("invalid part type: expected *entity.ConditionPart, got %T", part)
	}

	s.currentValue = 0
	s.AddObserver(condPart)
	return nil
}

func (s *CounterStrategy) GetCurrentValue() interface{} {
	return s.currentValue
}

func (s *CounterStrategy) Start(ctx context.Context, part interface{}) error {
	return nil
}

// Evaluate はカウンター条件を評価します
func (s *CounterStrategy) Evaluate(ctx context.Context, part interface{}, params interface{}) error {
	log := logger.DefaultLogger()
	log.Debug("Counter Evaluate")

	if params == nil {
		return fmt.Errorf("invalid nil params: %v", params)
	}

	condPart, ok := part.(*entity.ConditionPart)
	if !ok {
		return fmt.Errorf("invalid part type: expected *entity.ConditionPart, got %T", part)
	}
	increment := params.(int64)

	// カウンター値を更新
	s.mu.Lock()
	s.currentValue += increment
	s.mu.Unlock()
	log.Debug("currentValue", zap.Int64("currentValue", s.currentValue))

	// ComparisonOperatorを使用して条件を評価
	satisfied := false
	switch condPart.GetComparisonOperator() {
	case value.ComparisonOperatorEQ:
		satisfied = s.currentValue == condPart.GetReferenceValueInt()
	case value.ComparisonOperatorNEQ:
		satisfied = s.currentValue != condPart.GetReferenceValueInt()
	case value.ComparisonOperatorGT:
		satisfied = s.currentValue > condPart.GetReferenceValueInt()
	case value.ComparisonOperatorGTE:
		satisfied = s.currentValue >= condPart.GetReferenceValueInt()
	case value.ComparisonOperatorLT:
		satisfied = s.currentValue < condPart.GetReferenceValueInt()
	case value.ComparisonOperatorLTE:
		satisfied = s.currentValue <= condPart.GetReferenceValueInt()
	case value.ComparisonOperatorBetween:
		satisfied = s.currentValue >= condPart.GetMinValue() && s.currentValue <= condPart.GetMaxValue()
	default:
		return fmt.Errorf("unsupported comparison operator: %v", condPart.GetComparisonOperator())
	}

	log.Debug("Counter Evaluate", zap.Bool("satisfied", satisfied))
	if satisfied {
		s.NotifyUpdate(value.EventComplete)
	} else {
		s.NotifyUpdate(value.EventProcess)
	}
	return nil
}

// Cleanup は戦略のリソースを解放します
func (s *CounterStrategy) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentValue = 0
	s.observers = nil
	return nil
}

// AddObserver オブザーバーを追加します
func (s *CounterStrategy) AddObserver(observer service.StrategyObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

// RemoveObserver オブザーバーを削除します
func (s *CounterStrategy) RemoveObserver(observer service.StrategyObserver) {
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
func (s *CounterStrategy) NotifyUpdate(event string) {
	log := logger.DefaultLogger()
	log.Debug("CounterStrategy.NotifyUpdate", zap.String("event", event))
	s.mu.RLock()
	observers := make([]service.StrategyObserver, len(s.observers))
	copy(observers, s.observers)
	s.mu.RUnlock()

	for _, observer := range observers {
		observer.OnUpdated(event)
	}
}
