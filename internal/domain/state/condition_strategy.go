package state

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
)

// PartStrategy 条件評価のための戦略インターフェース
type PartStrategy interface {
	Initialize(part *ConditionPart) error
	// Evaluate 条件を評価
	Evaluate(ctx context.Context, part *ConditionPart, params interface{}) error
	Cleanup() error
}

type DefaultConditionStrategyFactory struct{}

func NewDefaultConditionStrategyFactory() *DefaultConditionStrategyFactory {
	return &DefaultConditionStrategyFactory{}
}

func (f *DefaultConditionStrategyFactory) CreateStrategy(kind core.ConditionKind) (PartStrategy, error) {
	switch kind {
	case core.KindTime:
		return NewTimeConditionStrategy(), nil
	case core.KindCounter:
		return NewCounterConditionStrategy(), nil
	default:
		return nil, fmt.Errorf("unknown condition kind: %v", kind)
	}
}
