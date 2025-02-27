package entity

import (
	"context"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockPhaseStateObserver は StateObserver インターフェースのモック実装です
type MockPhaseStateObserver struct {
	States []string
}

// インターフェースの実装を確認
var _ service.StateObserver = (*MockPhaseStateObserver)(nil)

// OnStateChanged は状態変更を記録します
func (m *MockPhaseStateObserver) OnStateChanged(state string) {
	m.States = append(m.States, state)
}

func TestNewPhase(t *testing.T) {
	// テスト用のパラメータ
	name := "Test Phase"
	order := 1
	conditions := []*Condition{
		NewCondition(1, "Condition 1", value.KindCounter),
		NewCondition(2, "Condition 2", value.KindTime),
	}
	conditionType := value.ConditionTypeAnd
	rule := value.GameRule_Shooting

	// Phaseの作成
	phase := NewPhase(name, order, conditions, conditionType, rule)

	// 初期状態の検証
	assert.Equal(t, name, phase.Name)
	assert.Equal(t, order, phase.Order)
	assert.Equal(t, conditionType, phase.ConditionType)
	assert.Equal(t, rule, phase.Rule)
	assert.False(t, phase.isActive)
	assert.False(t, phase.IsClear)
	assert.Nil(t, phase.StartTime)
	assert.Nil(t, phase.FinishTime)
	assert.Equal(t, value.StateReady, phase.CurrentState())
	assert.Len(t, phase.Conditions, 2)
	assert.Len(t, phase.ConditionIDs, 2)
	assert.Contains(t, phase.ConditionIDs, value.ConditionID(1))
	assert.Contains(t, phase.ConditionIDs, value.ConditionID(2))
}

func TestPhaseStateTransitions(t *testing.T) {
	// テスト用のPhase
	phase := NewPhase("Test Phase", 1, []*Condition{}, value.ConditionTypeSingle, value.GameRule_Shooting)
	ctx := context.Background()

	// Activate: Ready -> Active
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateActive, phase.CurrentState())
	assert.True(t, phase.isActive)
	assert.NotNil(t, phase.StartTime)

	// Next: Active -> Next
	err = phase.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateNext, phase.CurrentState())
	// 注意: 現在の実装ではNext状態でもisActiveはtrueのまま
	assert.True(t, phase.isActive)

	// Finish: Next -> Finish
	err = phase.Finish(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateFinish, phase.CurrentState())
	assert.False(t, phase.isActive)
	assert.NotNil(t, phase.FinishTime)

	// Reset: Finish -> Ready
	err = phase.Reset(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateReady, phase.CurrentState())
	assert.False(t, phase.isActive)
	assert.Nil(t, phase.StartTime)
	assert.Nil(t, phase.FinishTime)
}

func TestPhaseWithConditions(t *testing.T) {
	// テスト用の条件
	condition1 := NewCondition(1, "Condition 1", value.KindCounter)
	condition2 := NewCondition(2, "Condition 2", value.KindCounter)

	// テスト用のPhase
	phase := NewPhase("Test Phase", 1, []*Condition{condition1, condition2}, value.ConditionTypeAnd, value.GameRule_Shooting)
	ctx := context.Background()

	// Activate
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateActive, phase.CurrentState())
	assert.Equal(t, value.StateUnsatisfied, condition1.CurrentState())
	assert.Equal(t, value.StateUnsatisfied, condition2.CurrentState())

	// 条件が満たされた場合
	phase.OnConditionChanged(condition1)
	assert.True(t, phase.SatisfiedConditions[condition1.ID])
	assert.Equal(t, value.StateActive, phase.CurrentState()) // まだ全ての条件が満たされていない

	phase.OnConditionChanged(condition2)
	assert.True(t, phase.SatisfiedConditions[condition2.ID])
	assert.Equal(t, value.StateNext, phase.CurrentState()) // 全ての条件が満たされた
	assert.True(t, phase.IsClear)
}

func TestPhaseObserver(t *testing.T) {
	// テスト用のPhase
	phase := NewPhase("Test Phase", 1, []*Condition{}, value.ConditionTypeSingle, value.GameRule_Shooting)

	// モックオブザーバーの作成
	mockObserver := &MockPhaseStateObserver{}

	// オブザーバーの追加
	phase.AddObserver(mockObserver)

	// 状態変更の通知
	phase.NotifyStateChanged("test_state")
	assert.Len(t, mockObserver.States, 1)
	assert.Equal(t, "test_state", mockObserver.States[0])

	// オブザーバーの削除
	phase.RemoveObserver(mockObserver)

	// 状態変更の通知（オブザーバーが削除されているので通知されない）
	mockObserver.States = nil
	phase.NotifyStateChanged("another_state")
	assert.Len(t, mockObserver.States, 0)
}

func TestPhaseConditionTypes(t *testing.T) {
	// テスト用の条件
	condition1 := NewCondition(1, "Condition 1", value.KindCounter)
	condition2 := NewCondition(2, "Condition 2", value.KindCounter)

	// テストケース
	testCases := []struct {
		name          string
		conditionType value.ConditionType
		satisfyFirst  bool
		satisfySecond bool
		expectedClear bool
	}{
		{
			name:          "AND_None",
			conditionType: value.ConditionTypeAnd,
			satisfyFirst:  false,
			satisfySecond: false,
			expectedClear: false,
		},
		{
			name:          "AND_First",
			conditionType: value.ConditionTypeAnd,
			satisfyFirst:  true,
			satisfySecond: false,
			expectedClear: false,
		},
		{
			name:          "AND_Both",
			conditionType: value.ConditionTypeAnd,
			satisfyFirst:  true,
			satisfySecond: true,
			expectedClear: true,
		},
		{
			name:          "OR_None",
			conditionType: value.ConditionTypeOr,
			satisfyFirst:  false,
			satisfySecond: false,
			expectedClear: false,
		},
		{
			name:          "OR_First",
			conditionType: value.ConditionTypeOr,
			satisfyFirst:  true,
			satisfySecond: false,
			expectedClear: true,
		},
		{
			name:          "OR_Second",
			conditionType: value.ConditionTypeOr,
			satisfyFirst:  false,
			satisfySecond: true,
			expectedClear: true,
		},
		{
			name:          "Single_First",
			conditionType: value.ConditionTypeSingle,
			satisfyFirst:  true,
			satisfySecond: false,
			expectedClear: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用のPhase
			phase := NewPhase("Test Phase", 1, []*Condition{condition1, condition2}, tc.conditionType, value.GameRule_Shooting)
			ctx := context.Background()

			// Activate
			err := phase.Activate(ctx)
			assert.NoError(t, err)

			// 条件を満たす
			if tc.satisfyFirst {
				phase.SatisfiedConditions[condition1.ID] = true
			}
			if tc.satisfySecond {
				phase.SatisfiedConditions[condition2.ID] = true
			}

			// 条件が満たされているかチェック
			satisfied := phase.checkConditionsSatisfied()
			assert.Equal(t, tc.expectedClear, satisfied)
		})
	}
}

func TestPhaseGetStateInfo(t *testing.T) {
	// テスト用のPhase
	phase := NewPhase("Test Phase", 1, []*Condition{}, value.ConditionTypeSingle, value.GameRule_Shooting)
	ctx := context.Background()

	// Ready状態
	stateInfo := phase.GetStateInfo()
	assert.NotNil(t, stateInfo)
	assert.Equal(t, "Ready", stateInfo.Name)
	assert.Contains(t, stateInfo.AllowedNext, value.EventActivate)

	// Active状態
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	stateInfo = phase.GetStateInfo()
	assert.NotNil(t, stateInfo)
	assert.Equal(t, "Active", stateInfo.Name)
	assert.Contains(t, stateInfo.AllowedNext, value.EventNext)

	// Next状態
	err = phase.Next(ctx)
	assert.NoError(t, err)
	stateInfo = phase.GetStateInfo()
	assert.NotNil(t, stateInfo)
	assert.Equal(t, "Next", stateInfo.Name)
	assert.Contains(t, stateInfo.AllowedNext, value.EventFinish)

	// Finish状態
	err = phase.Finish(ctx)
	assert.NoError(t, err)
	stateInfo = phase.GetStateInfo()
	assert.NotNil(t, stateInfo)
	assert.Equal(t, "Finish", stateInfo.Name)
	assert.Contains(t, stateInfo.AllowedNext, value.EventReset)
}

func TestPhasesCollection(t *testing.T) {
	// テスト用のPhase
	phase1 := NewPhase("Phase 1", 1, []*Condition{}, value.ConditionTypeSingle, value.GameRule_Shooting)
	phase2 := NewPhase("Phase 2", 2, []*Condition{}, value.ConditionTypeSingle, value.GameRule_Shooting)
	phase3 := NewPhase("Phase 3", 3, []*Condition{}, value.ConditionTypeSingle, value.GameRule_Shooting)

	// Phasesコレクションの作成
	phases := Phases{phase1, phase2, phase3}
	ctx := context.Background()

	// 初期状態ではアクティブなフェーズはない
	assert.Nil(t, phases.Current())

	// Phase1をアクティブにする
	err := phase1.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phase1, phases.Current())

	// ResetAllのテスト
	err = phases.ResetAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateReady, phase1.CurrentState())
	assert.Equal(t, value.StateReady, phase2.CurrentState())
	assert.Equal(t, value.StateReady, phase3.CurrentState())
	assert.Nil(t, phases.Current())

	// ProcessAndActivateByNextOrderのテスト（初期状態）
	nextPhase, err := phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phase1, nextPhase)
	assert.Equal(t, value.StateActive, phase1.CurrentState())

	// 次のフェーズに進むには、現在のフェーズをNextに遷移させる必要がある
	err = phase1.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateNext, phase1.CurrentState())

	// ProcessAndActivateByNextOrderのテスト（次のフェーズへ）
	nextPhase, err = phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phase2, nextPhase)
	assert.Equal(t, value.StateFinish, phase1.CurrentState())
	assert.Equal(t, value.StateActive, phase2.CurrentState())

	// 次のフェーズに進むには、現在のフェーズをNextに遷移させる必要がある
	err = phase2.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateNext, phase2.CurrentState())

	// ProcessAndActivateByNextOrderのテスト（最後のフェーズへ）
	nextPhase, err = phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phase3, nextPhase)
	assert.Equal(t, value.StateFinish, phase2.CurrentState())
	assert.Equal(t, value.StateActive, phase3.CurrentState())

	// 次のフェーズに進むには、現在のフェーズをNextに遷移させる必要がある
	err = phase3.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateNext, phase3.CurrentState())

	// ProcessAndActivateByNextOrderのテスト（全てのフェーズが終了）
	nextPhase, err = phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Nil(t, nextPhase)
	assert.Equal(t, value.StateFinish, phase3.CurrentState())
}

func TestPhaseReset(t *testing.T) {
	// テスト用の条件
	condition := NewCondition(1, "Test Condition", value.KindCounter)
	part := NewConditionPart(1, "Test Part")
	condition.AddPart(part)

	// テスト用のPhase
	phase := NewPhase("Test Phase", 1, []*Condition{condition}, value.ConditionTypeSingle, value.GameRule_Shooting)
	ctx := context.Background()

	// Activate
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateActive, phase.CurrentState())
	assert.Equal(t, value.StateUnsatisfied, condition.CurrentState())

	// 時間情報を設定
	now := time.Now()
	phase.StartTime = &now
	phase.FinishTime = &now

	// 条件を満たす
	phase.SatisfiedConditions[condition.ID] = true
	phase.IsClear = true

	// Reset
	err = phase.Reset(ctx)
	assert.NoError(t, err)
	assert.Equal(t, value.StateReady, phase.CurrentState())
	// リセット後はIsClearがfalseになることを確認
	assert.False(t, phase.IsClear, "IsClear should be false after reset")
	assert.Nil(t, phase.StartTime)
	assert.Nil(t, phase.FinishTime)
	assert.Empty(t, phase.SatisfiedConditions)
	assert.Equal(t, value.StateReady, condition.CurrentState())
}
