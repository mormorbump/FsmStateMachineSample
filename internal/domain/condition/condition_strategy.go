package condition

import (
	"fmt"
)

// DefaultConditionStrategyFactory はデフォルトの条件評価戦略ファクトリです
type DefaultConditionStrategyFactory struct{}

// NewDefaultConditionStrategyFactory は新しいDefaultConditionStrategyFactoryを作成します
func NewDefaultConditionStrategyFactory() *DefaultConditionStrategyFactory {
	return &DefaultConditionStrategyFactory{}
}

// CreateStrategy は条件の種類に応じた評価戦略を作成します
func (f *DefaultConditionStrategyFactory) CreateStrategy(kind Kind) (Strategy, error) {
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
