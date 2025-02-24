package condition_details

import (
	"context"
	"fmt"
	"state_sample/internal/domain/entity"
)

type ConditionStrategy interface {
	Initialize(part *entity.ConditionPart) error
	// Evaluate 条件を評価
	Evaluate(ctx context.Context, part *entity.ConditionPart) error
	Cleanup() error
}

type ConditionStrategyFactory interface {
	CreateStrategy(kind entity.ConditionKind) (ConditionStrategy, error)
}

type DefaultConditionStrategyFactory struct{}

func NewDefaultConditionStrategyFactory() *DefaultConditionStrategyFactory {
	return &DefaultConditionStrategyFactory{}
}

// CreateStrategy は条件の種類に応じた評価戦略を作成します
func (f *DefaultConditionStrategyFactory) CreateStrategy(kind entity.ConditionKind) (ConditionStrategy, error) {
	switch kind {
	case entity.ConditionKindTime:
		return NewTimeConditionStrategy(), nil
	case entity.ConditionKindScore:
		// TODO: スコア条件の戦略を実装
		return nil, fmt.Errorf("score condition strategy not implemented yet")
	default:
		return nil, fmt.Errorf("unknown condition kind: %v", kind)
	}
}
