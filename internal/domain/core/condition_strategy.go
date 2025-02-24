package core

import (
	"context"
	"fmt"
)

// PartStrategy 条件評価のための戦略インターフェース
type PartStrategy interface {
	Initialize(part ConditionPart) error
	// Evaluate 条件を評価
	Evaluate(ctx context.Context, part ConditionPart) error
	Cleanup() error
}

type DefaultConditionStrategyFactory struct{}

func NewDefaultConditionStrategyFactory() *DefaultConditionStrategyFactory {
	return &DefaultConditionStrategyFactory{}
}

func (f *DefaultConditionStrategyFactory) CreateStrategy(kind ConditionKind) (PartStrategy, error) {
	switch kind {
	case KindTime:
		return NewTimeConditionStrategy(), nil
	case KindScore:
		// TODO: スコア条件の戦略を実装
		return nil, fmt.Errorf("score condition strategy not implemented yet")
	default:
		return nil, fmt.Errorf("unknown condition kind: %v", kind)
	}
}
