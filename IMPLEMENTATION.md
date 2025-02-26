# ステートマシン実装のリファクタリング計画

## 現状の課題

1. **ディレクトリ構造の問題**
   - `domain/state`に state, value, type, strategy, observer, subject が混在している
   - 循環参照が発生しやすい構造になっている
   - 責務の分離が不十分

2. **拡張性の問題**
   - 新機能追加時に複数箇所の変更が必要
   - PartStrategyの追加が複雑

3. **アーキテクチャの問題**
   - オニオンアーキテクチャの原則とステートマシンの特性の両立が難しい
   - インターフェースと実装の分離が不十分

## リファクタリングの目標

1. **責務の明確な分離**
   - ドメイン層にはエンティティと値オブジェクトを配置
   - 実装の詳細はユースケース層に移動

2. **拡張性の向上**
   - PartStrategyの追加だけで新機能を実装できるようにする
   - Factory/Builderパターンの強化

3. **循環参照の解消**
   - 依存関係を一方向に整理
   - インターフェースを活用した依存性逆転の原則の適用

## 新しいディレクトリ構造

```
internal/
  ├── domain/
  │   ├── entity/           # エンティティの定義
  │   │   ├── condition.go      # 条件エンティティ
  │   │   ├── condition_part.go # 条件パーツエンティティ
  │   │   ├── phase.go          # フェーズエンティティ
  │   │   └── game_state.go     # ゲーム状態エンティティ
  │   ├── value/            # 値オブジェクト
  │   │   ├── game_state.go     # ゲーム状態の値オブジェクト
  │   │   └── types.go          # 共通の型定義
  │   └── service/          # ドメインサービスインターフェース
  │       ├── observer.go       # オブザーバーインターフェース
  │       └── strategy.go       # 戦略インターフェース
  ├── usecase/
  │   ├── state/            # 状態管理ユースケース
  │   │   ├── phase_controller.go # フェーズ制御
  │   │   └── state_facade.go     # 状態管理ファサード
  │   └── strategy/         # 戦略実装
  │       ├── time_strategy.go    # 時間ベース戦略
  │       ├── counter_strategy.go # カウンターベース戦略
  │       └── strategy_factory.go # 戦略ファクトリ
  └── ui/                   # UI層（変更なし）
      ├── handlers.go
      ├── server.go
      └── static/
```

## 実装計画

### フェーズ1: ドメイン層の再構築

1. **エンティティの定義**
   - `domain/entity/condition_part.go`: 条件パーツエンティティ
   - `domain/entity/condition.go`: 条件エンティティ
   - `domain/entity/phase.go`: フェーズエンティティ
   - `domain/entity/game_state.go`: ゲーム状態の定義

2. **値オブジェクトの整理**
   - `domain/value/game_state.go`: ゲーム状態の値オブジェクト
   - `domain/value/types.go`: 共通の型定義（ConditionID, ConditionPartID, ConditionKind, ComparisonOperator等）

3. **サービスインターフェースの定義**
   - `domain/service/observer.go`: オブザーバーパターンのインターフェース
   - `domain/service/strategy.go`: 戦略パターンのインターフェース

### フェーズ2: ユースケース層の実装

1. **状態管理ユースケース**
   - `usecase/state/phase_controller.go`: フェーズ制御ロジック
   - `usecase/state/state_facade.go`: 状態管理ファサード

2. **戦略実装**
   - `usecase/strategy/time_strategy.go`: 時間ベース戦略の実装
   - `usecase/strategy/counter_strategy.go`: カウンターベース戦略の実装
   - `usecase/strategy/strategy_factory.go`: 戦略ファクトリの実装

### フェーズ3: UI層の調整

1. **既存UI層との連携**
   - 新しいインターフェースに合わせてUI層のコードを調整

## 主要なインターフェース設計

### 戦略パターンのインターフェース

```go
// domain/service/strategy.go
package service

import (
	"context"
)

// PartStrategy 条件評価のための戦略インターフェース
type PartStrategy interface {
	Initialize(part interface{}) error
	GetCurrentValue() interface{}
	Evaluate(ctx context.Context, part interface{}, params interface{}) error
	Cleanup() error
}

// StrategyFactory 戦略を作成するファクトリインターフェース
type StrategyFactory interface {
	CreateStrategy(kind interface{}) (PartStrategy, error)
}
```

### オブザーバーパターンのインターフェース

```go
// domain/service/observer.go
package service

// StateObserver 状態を監視するインターフェース
type StateObserver interface {
	OnStateChanged(state string)
}

// ConditionObserver 条件の状態変化を監視するインターフェース
type ConditionObserver interface {
	OnConditionChanged(condition interface{})
}

// ConditionPartObserver 条件パーツの状態変化を監視するインターフェース
type ConditionPartObserver interface {
	OnConditionPartChanged(part interface{})
}
```

## 実装の詳細

### PartStrategyの拡張方法

新しい戦略を追加する場合、以下の手順で実装します：

1. `usecase/strategy/` に新しい戦略の実装ファイルを作成
2. `domain/service/strategy.go` で定義された `PartStrategy` インターフェースを実装
3. `usecase/strategy/strategy_factory.go` に新しい戦略の作成ロジックを追加

例：新しいランダム条件戦略の追加

```go
// usecase/strategy/random_strategy.go
package strategy

import (
	"context"
	"math/rand"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	"time"
)

// RandomConditionStrategy はランダム値に基づく条件評価戦略
type RandomConditionStrategy struct {
	currentValue int64
	maxValue     int64
	observers    []service.StrategyObserver
}

// NewRandomConditionStrategy は新しいRandomConditionStrategyを作成
func NewRandomConditionStrategy() *RandomConditionStrategy {
	rand.Seed(time.Now().UnixNano())
	return &RandomConditionStrategy{}
}

// Initialize は戦略の初期化を行う
func (s *RandomConditionStrategy) Initialize(part interface{}) error {
	// 型アサーションでConditionPartの情報を取得
	condPart := part.(*ConditionPart)
	s.maxValue = condPart.GetReferenceValueInt()
	return nil
}

// GetCurrentValue は現在の値を返す
func (s *RandomConditionStrategy) GetCurrentValue() interface{} {
	return s.currentValue
}

// Evaluate はランダム条件を評価する
func (s *RandomConditionStrategy) Evaluate(ctx context.Context, part interface{}, params interface{}) error {
	// 型アサーションでConditionPartの情報を取得
	condPart := part.(*ConditionPart)
	
	// ランダム値の生成と評価ロジック
	s.currentValue = rand.Int63n(s.maxValue + 1)
	
	// 条件の評価
	if s.currentValue >= condPart.GetReferenceValueInt() {
		s.NotifyUpdate(value.EventComplete)
		return nil
	}
	return nil
}

// Cleanup は戦略のリソースを解放する
func (s *RandomConditionStrategy) Cleanup() error {
	s.currentValue = 0
	s.observers = nil
	return nil
}

// AddObserver オブザーバーを追加
func (s *RandomConditionStrategy) AddObserver(observer service.StrategyObserver) {
	s.observers = append(s.observers, observer)
}

// RemoveObserver オブザーバーを削除
func (s *RandomConditionStrategy) RemoveObserver(observer service.StrategyObserver) {
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// NotifyUpdate オブザーバーに更新を通知
func (s *RandomConditionStrategy) NotifyUpdate(event string) {
	for _, observer := range s.observers {
		observer.OnUpdated(event)
	}
}
```

そして、ファクトリに追加：

```go
// usecase/strategy/strategy_factory.go の CreateStrategy メソッドに追加
func (f *StrategyFactoryImpl) CreateStrategy(kind interface{}) (service.PartStrategy, error) {
	condKind := kind.(value.ConditionKind)
	switch condKind {
	case value.KindTime:
		return NewTimeConditionStrategy(), nil
	case value.KindCounter:
		return NewCounterConditionStrategy(), nil
	case value.KindRandom:  // 新しい種類を追加
		return NewRandomConditionStrategy(), nil
	default:
		return nil, fmt.Errorf("unknown condition kind: %v", condKind)
	}
}
```

## 移行計画

1. 新しいディレクトリ構造を作成
2. ドメイン層のエンティティと値オブジェクトを定義
3. サービスインターフェースを定義
4. 既存コードを新しい構造に徐々に移行
5. テストを実行して機能が正しく動作することを確認
6. 古いコードを削除

## 期待される効果

1. **拡張性の向上**
   - 新しい戦略の追加が容易になる
   - 変更箇所が限定される

2. **保守性の向上**
   - 責務が明確に分離される
   - テストが書きやすくなる

3. **理解しやすさの向上**
   - コードの構造が整理される
   - 依存関係が明確になる

## リスクと対策

1. **リファクタリングによる機能退行**
   - 対策: 段階的な移行と継続的なテスト

2. **パフォーマンスへの影響**
   - 対策: パフォーマンステストの実施
   - 対策: ボトルネックの特定と最適化

3. **開発工数の増加**
   - 対策: 優先順位の高い部分から段階的に実施