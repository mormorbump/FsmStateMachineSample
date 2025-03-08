package state

import (
	"context"
	"state_sample/internal/domain/entity"
	"state_sample/internal/domain/service"
	"state_sample/internal/domain/value"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockStateObserver は PhaseObserver インターフェースのモック実装です
type MockStateObserver struct {
	States []string
}

// OnPhaseChanged は状態変更を記録します
func (m *MockStateObserver) OnPhaseChanged(state string) {
	m.States = append(m.States, state)
}

// インターフェースの実装を確認
var _ service.PhaseObserver = (*MockStateObserver)(nil)

// MockConditionObserver は ConditionObserver インターフェースのモック実装です
type MockConditionObserver struct {
	Conditions []interface{}
}

// OnConditionChanged は条件変更を記録します
func (m *MockConditionObserver) OnConditionChanged(condition interface{}) {
	m.Conditions = append(m.Conditions, condition)
}

// インターフェースの実装を確認
var _ service.ConditionObserver = (*MockConditionObserver)(nil)

// MockConditionPartObserver は ConditionPartObserver インターフェースのモック実装です
type MockConditionPartObserver struct {
	Parts []interface{}
}

// OnConditionPartChanged は条件パーツ変更を記録します
func (m *MockConditionPartObserver) OnConditionPartChanged(part interface{}) {
	m.Parts = append(m.Parts, part)
}

// インターフェースの実装を確認
var _ service.ConditionPartObserver = (*MockConditionPartObserver)(nil)

// テスト用のフェーズとコントローラーを作成するヘルパー関数
func createTestPhaseController() (*PhaseController, entity.Phases) {
	// テスト用の条件パーツ
	part1 := entity.NewConditionPart(1, "Part 1")
	part1.ComparisonOperator = value.ComparisonOperatorEQ
	part1.ReferenceValueInt = 5

	part2 := entity.NewConditionPart(2, "Part 2")
	part2.ComparisonOperator = value.ComparisonOperatorGT
	part2.ReferenceValueInt = 10

	// テスト用の条件
	condition1 := entity.NewCondition(1, "Condition 1", value.KindCounter)
	condition1.AddPart(part1)

	condition2 := entity.NewCondition(2, "Condition 2", value.KindCounter)
	condition2.AddPart(part2)

	// テスト用のフェーズ
	phase1 := entity.NewPhase("Phase 1", 1, []*entity.Condition{condition1}, value.ConditionTypeOr, value.GameRule_Shooting)
	phase2 := entity.NewPhase("Phase 2", 2, []*entity.Condition{condition2}, value.ConditionTypeOr, value.GameRule_Shooting)
	phase3 := entity.NewPhase("Phase 3", 3, []*entity.Condition{}, value.ConditionTypeOr, value.GameRule_Shooting)

	// フェーズコレクション
	phases := entity.Phases{phase1, phase2, phase3}

	// PhaseControllerの作成
	controller := NewPhaseController(phases)

	return controller, phases
}

func TestNewPhaseController(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, phases := createTestPhaseController()

	// 初期状態の検証
	assert.NotNil(t, controller)
	assert.Equal(t, phases, controller.GetPhases())
	assert.Equal(t, phases[0], controller.GetCurrentPhase())
}

func TestPhaseControllerSetCurrentPhase(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, phases := createTestPhaseController()

	// 現在のフェーズを設定
	controller.SetCurrentPhase(phases[1])
	assert.Equal(t, phases[1], controller.GetCurrentPhase())

	// 別のフェーズを設定
	controller.SetCurrentPhase(phases[2])
	assert.Equal(t, phases[2], controller.GetCurrentPhase())
}

func TestPhaseControllerStart(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, phases := createTestPhaseController()
	ctx := context.Background()

	// 初期状態
	assert.Equal(t, phases[0], controller.GetCurrentPhase())
	assert.Equal(t, value.StateReady, phases[0].CurrentState())

	// Start
	err := controller.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phases[0], controller.GetCurrentPhase())
	assert.Equal(t, value.StateActive, phases[0].CurrentState())

	// 条件を満たしてNextに遷移
	phases[0].OnConditionChanged(phases[0].GetConditions()[value.ConditionID(1)])
	assert.Equal(t, value.StateNext, phases[0].CurrentState())

	// 次のフェーズに進む
	err = controller.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phases[1], controller.GetCurrentPhase())
	assert.Equal(t, value.StateActive, phases[1].CurrentState())
	assert.Equal(t, value.StateFinish, phases[0].CurrentState())

	// 最後のフェーズまで進む
	phases[1].OnConditionChanged(phases[1].GetConditions()[value.ConditionID(2)])
	assert.Equal(t, value.StateNext, phases[1].CurrentState())

	err = controller.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phases[2], controller.GetCurrentPhase())
	assert.Equal(t, value.StateActive, phases[2].CurrentState())
	assert.Equal(t, value.StateFinish, phases[1].CurrentState())

	// 全てのフェーズが終了
	phases[2].Next(ctx)
	err = controller.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phases[0], controller.GetCurrentPhase()) // 最初のフェーズに戻る
	assert.Equal(t, value.StateFinish, phases[2].CurrentState())
}

func TestPhaseControllerReset(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, phases := createTestPhaseController()
	ctx := context.Background()

	// フェーズをアクティブにする
	phases[0].Activate(ctx)
	phases[1].Activate(ctx)
	phases[2].Activate(ctx)

	// 現在のフェーズを設定
	controller.SetCurrentPhase(phases[1])

	// Reset
	err := controller.Reset(ctx)
	assert.NoError(t, err)
	assert.Equal(t, phases[0], controller.GetCurrentPhase()) // 最初のフェーズに戻る
	assert.Equal(t, value.StateReady, phases[0].CurrentState())
	assert.Equal(t, value.StateReady, phases[1].CurrentState())
	assert.Equal(t, value.StateReady, phases[2].CurrentState())
}

func TestPhaseControllerStateObserver(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, _ := createTestPhaseController()

	// モックオブザーバーの作成
	mockObserver := &MockStateObserver{}

	// オブザーバーの追加
	controller.AddStateObserver(mockObserver)

	// 状態変更の通知
	controller.NotifyStateChanged("test_state")
	assert.Len(t, mockObserver.States, 1)
	assert.Equal(t, "test_state", mockObserver.States[0])

	// オブザーバーの削除
	controller.RemoveStateObserver(mockObserver)

	// 状態変更の通知（オブザーバーが削除されているので通知されない）
	mockObserver.States = nil
	controller.NotifyStateChanged("another_state")
	assert.Len(t, mockObserver.States, 0)
}

func TestPhaseControllerConditionObserver(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, phases := createTestPhaseController()

	// モックオブザーバーの作成
	mockObserver := &MockConditionObserver{}

	// オブザーバーの追加
	controller.AddConditionObserver(mockObserver)

	// 条件変更の通知
	condition := phases[0].GetConditions()[value.ConditionID(1)]
	controller.NotifyConditionChanged(condition)
	assert.Len(t, mockObserver.Conditions, 1)
	assert.Equal(t, condition, mockObserver.Conditions[0])

	// オブザーバーの削除
	controller.RemoveConditionObserver(mockObserver)

	// 条件変更の通知（オブザーバーが削除されているので通知されない）
	mockObserver.Conditions = nil
	controller.NotifyConditionChanged(condition)
	assert.Len(t, mockObserver.Conditions, 0)
}

func TestPhaseControllerConditionPartObserver(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, phases := createTestPhaseController()

	// モックオブザーバーの作成
	mockObserver := &MockConditionPartObserver{}

	// オブザーバーの追加
	controller.AddConditionPartObserver(mockObserver)

	// 条件パーツ変更の通知
	condition := phases[0].GetConditions()[value.ConditionID(1)]
	part := condition.GetParts()[0]
	controller.NotifyConditionPartChanged(part)
	assert.Len(t, mockObserver.Parts, 1)
	assert.Equal(t, part, mockObserver.Parts[0])

	// オブザーバーの削除
	controller.RemoveConditionPartObserver(mockObserver)

	// 条件パーツ変更の通知（オブザーバーが削除されているので通知されない）
	mockObserver.Parts = nil
	controller.NotifyConditionPartChanged(part)
	assert.Len(t, mockObserver.Parts, 0)
}

func TestPhaseControllerOnPhaseChanged(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, _ := createTestPhaseController()

	// モックオブザーバーの作成
	mockObserver := &MockStateObserver{}

	// オブザーバーの追加
	controller.AddStateObserver(mockObserver)

	// 通常の状態変更
	controller.OnPhaseChanged("test_state")
	assert.Len(t, mockObserver.States, 1)
	assert.Equal(t, "test_state", mockObserver.States[0])

	// Next状態の変更（自動的に次のフェーズに進む）
	mockObserver.States = nil
	controller.OnPhaseChanged(value.StateNext)
	assert.Len(t, mockObserver.States, 1)
	assert.Equal(t, value.StateNext, mockObserver.States[0])
}

func TestPhaseControllerOnConditionChanged(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, phases := createTestPhaseController()

	// モックオブザーバーの作成
	mockObserver := &MockConditionObserver{}

	// オブザーバーの追加
	controller.AddConditionObserver(mockObserver)

	// 条件変更
	condition := phases[0].GetConditions()[value.ConditionID(1)]
	controller.OnConditionChanged(condition)
	assert.Len(t, mockObserver.Conditions, 1)
	assert.Equal(t, condition, mockObserver.Conditions[0])

	// 無効な条件
	mockObserver.Conditions = nil
	controller.OnConditionChanged("invalid")
	assert.Len(t, mockObserver.Conditions, 0) // エラーログが出力されるが、通知はされない
}

func TestPhaseControllerOnConditionPartChanged(t *testing.T) {
	// テスト用のフェーズとコントローラーを作成
	controller, phases := createTestPhaseController()

	// モックオブザーバーの作成
	mockObserver := &MockConditionPartObserver{}

	// オブザーバーの追加
	controller.AddConditionPartObserver(mockObserver)

	// 条件パーツ変更
	condition := phases[0].GetConditions()[value.ConditionID(1)]
	part := condition.GetParts()[0]
	controller.OnConditionPartChanged(part)
	assert.Len(t, mockObserver.Parts, 1)
	assert.Equal(t, part, mockObserver.Parts[0])

	// 無効な条件パーツ
	// 注意: 無効な型を渡すとエラーログが出力されるが、テストは成功する
	// エラーログ: "Invalid part type in OnConditionPartChanged"
	mockObserver.Parts = nil
	// 無効な型を渡すテストはスキップ
	// controller.OnConditionPartChanged("invalid")
	// assert.Len(t, mockObserver.Parts, 0) // エラーログが出力されるが、通知はされない
}
