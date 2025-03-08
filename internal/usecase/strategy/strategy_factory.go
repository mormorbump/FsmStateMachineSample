package strategy

import (
	"fmt"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
)

// StrategyFactory は戦略を作成するファクトリの実装です
type StrategyFactory struct{}

// NewStrategyFactory は新しいStrategyFactoryを作成します
func NewStrategyFactory() *StrategyFactory {
	return &StrategyFactory{}
}

// CreateStrategy は指定された種類の戦略を作成します
func (f *StrategyFactory) CreateStrategy(kind value.ConditionKind) (service.PartStrategy, error) {
	switch kind {
	case value.KindTime:
		return NewTimeStrategy(), nil
	case value.KindCounter:
		return NewCounterStrategy(), nil
	default:
		return nil, fmt.Errorf("unknown condition kind: %v", kind)
	}
}
