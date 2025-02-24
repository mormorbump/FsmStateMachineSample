package state

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
)

// ConditionCounterStrategy はカウンターベースの条件評価戦略です
type ConditionCounterStrategy struct {
	currentValue int64
	observers    []ConditionPart
}

// NewCounterConditionStrategy は新しいConditionCounterStrategyを作成します
func NewCounterConditionStrategy() *ConditionCounterStrategy {
	return &ConditionCounterStrategy{
		currentValue: 0,
	}
}

// Initialize は戦略の初期化を行います
func (s *ConditionCounterStrategy) Initialize(part *ConditionPart) error {
	s.currentValue = 0
	return nil
}

// Evaluate はカウンター条件を評価します
func (s *ConditionCounterStrategy) Evaluate(ctx context.Context, part *ConditionPart, params interface{}) error {
	// パラメータから増分値を取得
	increment, ok := params.(int64)
	if !ok {
		return fmt.Errorf("invalid increment value")
	}

	// カウンター値を更新
	s.currentValue += increment

	// ComparisonOperatorを使用して条件を評価
	satisfied := false
	switch part.GetComparisonOperator() {
	case core.ComparisonOperatorEQ:
		satisfied = s.currentValue == part.GetReferenceValueInt()
	case core.ComparisonOperatorNEQ:
		satisfied = s.currentValue != part.GetReferenceValueInt()
	case core.ComparisonOperatorGT:
		satisfied = s.currentValue > part.GetReferenceValueInt()
	case core.ComparisonOperatorGTE:
		satisfied = s.currentValue >= part.GetReferenceValueInt()
	case core.ComparisonOperatorLT:
		satisfied = s.currentValue < part.GetReferenceValueInt()
	case core.ComparisonOperatorLTE:
		satisfied = s.currentValue <= part.GetReferenceValueInt()
	case core.ComparisonOperatorBetween:
		satisfied = s.currentValue >= part.GetMinValue() && s.currentValue <= part.GetMaxValue()
	default:
		return fmt.Errorf("unsupported comparison operator: %v", part.GetComparisonOperator())
	}

	if satisfied {
		return part.Complete(ctx)
	}
	return nil
}

// Cleanup は戦略のリソースを解放します
func (s *ConditionCounterStrategy) Cleanup() error {
	s.currentValue = 0
	s.observers = nil
	return nil
}

// GetCurrentValue は現在のカウンター値を返します
func (s *ConditionCounterStrategy) GetCurrentValue() int64 {
	return s.currentValue
}
