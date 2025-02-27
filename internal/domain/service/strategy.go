package service

import (
	"context"
	"state_sample/internal/domain/value"
)

// PartStrategy 条件評価のための戦略インターフェース
type PartStrategy interface {
	StrategySubject
	Initialize(part interface{}) error
	GetCurrentValue() interface{}
	Start(ctx context.Context, part interface{}) error
	Evaluate(ctx context.Context, part interface{}, params interface{}) error
	Cleanup() error
}

// StrategyFactory 戦略を作成するファクトリインターフェース
type StrategyFactory interface {
	CreateStrategy(kind value.ConditionKind) (PartStrategy, error)
}
