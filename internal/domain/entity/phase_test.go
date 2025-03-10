package entity

import (
	"context"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockPhaseStateObserver は PhaseObserver インターフェースのモック実装です
type MockPhaseStateObserver struct {
	Phases []*Phase
}

// インターフェースの実装を確認
var _ service.PhaseObserver = (*MockPhaseStateObserver)(nil)

// OnPhaseChanged は状態変更を記録します
func (m *MockPhaseStateObserver) OnPhaseChanged(phase interface{}) {
	if p, ok := phase.(*Phase); ok {
		m.Phases = append(m.Phases, p)
	}
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
	phase := NewPhase(1, name, order, conditions, conditionType, rule, 0, false)

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
	phase := NewPhase(1, "Test Phase", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, false)
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

func TestPhaseObserver(t *testing.T) {
	// テスト用のPhase
	phase := NewPhase(1, "Test Phase", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, false)

	// モックオブザーバーの作成
	mockObserver := &MockPhaseStateObserver{}

	// オブザーバーの追加
	phase.AddObserver(mockObserver)

	// 状態変更の通知
	phase.NotifyPhaseChanged()
	assert.Len(t, mockObserver.Phases, 1)
	assert.Equal(t, phase, mockObserver.Phases[0])

	// オブザーバーの削除
	phase.RemoveObserver(mockObserver)

	// 状態変更の通知（オブザーバーが削除されているので通知されない）
	mockObserver.Phases = nil
	phase.NotifyPhaseChanged()
	assert.Len(t, mockObserver.Phases, 0)
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用のPhase
			phase := NewPhase(1, "Test Phase", 1, []*Condition{condition1, condition2}, tc.conditionType, value.GameRule_Shooting, 0, false)
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
	phase := NewPhase(1, "Test Phase", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, false)
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
	phase1 := NewPhase(1, "Phase 1", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, false)
	phase2 := NewPhase(2, "Phase 2", 2, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, false)
	phase3 := NewPhase(3, "Phase 3", 3, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, false)

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
	phase := NewPhase(1, "Test Phase", 1, []*Condition{condition}, value.ConditionTypeOr, value.GameRule_Shooting, 0, false)
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

func TestPhaseHierarchy(t *testing.T) {
	// 親フェーズの作成
	parentPhase := NewPhase(1, "Parent Phase", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, true)

	// 子フェーズの作成
	childPhase1 := NewPhase(2, "Child Phase 1", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 1, false)
	childPhase2 := NewPhase(3, "Child Phase 2", 2, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 1, false)

	// 親子関係の設定
	parentPhase.AddChild(childPhase1)
	parentPhase.AddChild(childPhase2)

	// 親子関係の検証
	assert.Equal(t, 2, len(parentPhase.Children))
	assert.Equal(t, parentPhase, childPhase1.Parent)
	assert.Equal(t, parentPhase, childPhase2.Parent)
	assert.True(t, parentPhase.HasChildren())
	assert.False(t, childPhase1.HasChildren())

	// GetChildrenメソッドのテスト
	children := parentPhase.GetChildren()
	assert.Equal(t, 2, len(children))
	assert.Equal(t, childPhase1, children[0]) // Orderでソートされているので、Order=1のchildPhase1が先に来る
	assert.Equal(t, childPhase2, children[1])
}

func TestGroupPhasesByParentID(t *testing.T) {
	// フェーズの作成
	rootPhase1 := NewPhase(1, "Root Phase 1", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, true)
	rootPhase2 := NewPhase(2, "Root Phase 2", 2, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, true)

	childPhase1 := NewPhase(3, "Child Phase 1", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 1, false)
	childPhase2 := NewPhase(4, "Child Phase 2", 2, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 1, false)

	grandChildPhase := NewPhase(5, "Grand Child Phase", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 3, false)

	// フェーズのスライスを作成
	phases := Phases{rootPhase1, rootPhase2, childPhase1, childPhase2, grandChildPhase}

	// GroupPhasesByParentIDのテスト
	phaseMap := GroupPhasesByParentID(phases)

	// 結果の検証
	assert.Equal(t, 3, len(phaseMap))    // 3つの親ID（0, 1, 3）があるはず
	assert.Equal(t, 2, len(phaseMap[0])) // 親ID=0のフェーズは2つ
	assert.Equal(t, 2, len(phaseMap[1])) // 親ID=1のフェーズは2つ
	assert.Equal(t, 1, len(phaseMap[3])) // 親ID=3のフェーズは1つ

	// 各グループがOrderでソートされていることを確認
	assert.Equal(t, rootPhase1, phaseMap[0][0])
	assert.Equal(t, rootPhase2, phaseMap[0][1])
	assert.Equal(t, childPhase1, phaseMap[1][0])
	assert.Equal(t, childPhase2, phaseMap[1][1])
	assert.Equal(t, grandChildPhase, phaseMap[3][0])
}

func TestInitializePhaseHierarchy(t *testing.T) {
	// フェーズの作成（親子関係はまだ設定しない）
	rootPhase := NewPhase(1, "Root Phase", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 0, true)

	childPhase1 := NewPhase(2, "Child Phase 1", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 1, false)
	childPhase2 := NewPhase(3, "Child Phase 2", 2, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 1, false)

	grandChildPhase := NewPhase(4, "Grand Child Phase", 1, []*Condition{}, value.ConditionTypeOr, value.GameRule_Shooting, 2, false)

	// フェーズのスライスを作成
	phases := Phases{rootPhase, childPhase1, childPhase2, grandChildPhase}

	// 初期状態では親子関係が設定されていないことを確認
	assert.Nil(t, childPhase1.Parent)
	assert.Nil(t, childPhase2.Parent)
	assert.Nil(t, grandChildPhase.Parent)
	assert.Empty(t, rootPhase.Children)
	assert.Empty(t, childPhase1.Children)

	// InitializePhaseHierarchyのテスト
	InitializePhaseHierarchy(phases)

	// 親子関係が正しく設定されたことを確認
	assert.Equal(t, rootPhase, childPhase1.Parent)
	assert.Equal(t, rootPhase, childPhase2.Parent)
	assert.Equal(t, childPhase1, grandChildPhase.Parent)

	assert.Equal(t, 2, len(rootPhase.Children))
	assert.Contains(t, rootPhase.Children, childPhase1)
	assert.Contains(t, rootPhase.Children, childPhase2)

	assert.Equal(t, 1, len(childPhase1.Children))
	assert.Contains(t, childPhase1.Children, grandChildPhase)
}
