package condition

import (
	"context"
	"state_sample/internal/domain/core"
)

// Kind は条件の種類を表す型です
type Kind int

const (
	KindUnspecified Kind = iota
	KindTime             // 時間に基づく条件
	KindScore            // スコアに基づく条件
)

// ConditionID は条件のIDを表す型です
type ConditionID int64

// ConditionPartID は条件パーツのIDを表す型です
type ConditionPartID int64

// Part は条件パーツのインターフェースです
type Part interface {
	GetID() ConditionPartID
	GetReferenceValueInt() int64
	AddObserver(observer interface{})
	core.TimeObserver // OnTimeTicked()を含む
}

// Strategy は条件評価のための戦略インターフェースです
type Strategy interface {
	// Initialize は戦略を初期化します
	Initialize(part Part) error

	// Evaluate は条件を評価します
	Evaluate(ctx context.Context, part Part) error

	// Cleanup は戦略のリソースを解放します
	Cleanup() error
}

// Factory は条件評価戦略を作成するファクトリインターフェースです
type Factory interface {
	// CreateStrategy は条件の種類に応じた評価戦略を作成します
	CreateStrategy(kind Kind) (Strategy, error)
}

// Observer は条件の状態変化を監視するインターフェースです
type Observer interface {
	OnConditionSatisfied(conditionID ConditionID)
}
