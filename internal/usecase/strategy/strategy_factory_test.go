package strategy

import (
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStrategyFactory(t *testing.T) {
	// 新しいStrategyFactoryを作成
	factory := NewStrategyFactory()

	// ファクトリが正しく作成されていることを確認
	assert.NotNil(t, factory)
}

func TestCreateStrategy(t *testing.T) {
	// テストケース
	testCases := []struct {
		name          string
		kind          value.ConditionKind
		expectedType  string
		expectedError bool
	}{
		{
			name:          "KindTime",
			kind:          value.KindTime,
			expectedType:  "*strategy.TimeStrategy",
			expectedError: false,
		},
		{
			name:          "KindCounter",
			kind:          value.KindCounter,
			expectedType:  "*strategy.CounterStrategy",
			expectedError: false,
		},
		{
			name:          "KindUnspecified",
			kind:          value.KindUnspecified,
			expectedType:  "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 新しいStrategyFactoryを作成
			factory := NewStrategyFactory()

			// 戦略の作成
			strategy, err := factory.CreateStrategy(tc.kind)

			// 結果の確認
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, strategy)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, strategy)
				assert.Equal(t, tc.expectedType, getTypeName(strategy))
			}
		})
	}
}

// getTypeName は与えられたオブジェクトの型名を返します
func getTypeName(obj interface{}) string {
	if obj == nil {
		return ""
	}
	
	switch obj.(type) {
	case *TimeStrategy:
		return "*strategy.TimeStrategy"
	case *CounterStrategy:
		return "*strategy.CounterStrategy"
	default:
		return ""
	}
}

// StrategyFactoryがservice.StrategyFactoryインターフェースを実装していることを確認
func TestStrategyFactoryImplementsInterface(t *testing.T) {
	// 新しいStrategyFactoryを作成
	factory := NewStrategyFactory()

	// インターフェースを実装していることを確認
	var _ service.StrategyFactory = factory
}