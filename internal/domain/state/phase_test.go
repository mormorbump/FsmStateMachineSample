package state

import (
	"context"
	"fmt"
	"state_sample/internal/domain/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createCondition(order int, interval int64) *Condition {
	cond := NewCondition(
		core.ConditionID(order),
		fmt.Sprintf("test_phase_%d_timer", order),
		core.KindTime,
	)

	part := NewConditionPart(
		core.ConditionPartID(order),
		fmt.Sprintf("test_phase_%d_timer_part", order),
	)
	part.ReferenceValueInt = interval
	cond.AddPart(part)

	return cond
}

func createMultipleConditions(count int) []*Condition {
	conditions := make([]*Condition, count)
	for i := 0; i < count; i++ {
		conditions[i] = createCondition(i+1, int64(i+1))
	}
	return conditions
}

func TestPhase_ConditionTypeAnd(t *testing.T) {
	// Arrange
	conditions := createMultipleConditions(3)
	phase := NewPhase("test_phase", 1, conditions)
	phase.ConditionType = ConditionTypeAnd
	ctx := context.Background()

	// Act & Assert: 初期状態の確認
	assert.Equal(t, core.StateReady, phase.CurrentState())
	assert.False(t, phase.IsClear)

	// フェーズをアクティブ化
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateActive, phase.CurrentState())

	// 1つ目の条件を満たす
	phase.OnConditionSatisfied(core.ConditionID(1))
	assert.Equal(t, core.StateActive, phase.CurrentState(), "一つの条件だけでは次の状態に進まないはず")
	assert.False(t, phase.IsClear)

	// 2つ目の条件を満たす
	phase.OnConditionSatisfied(core.ConditionID(2))
	assert.Equal(t, core.StateActive, phase.CurrentState(), "二つの条件でも次の状態に進まないはず")
	assert.False(t, phase.IsClear)

	// 3つ目の条件を満たす
	phase.OnConditionSatisfied(core.ConditionID(3))
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.Equal(t, core.StateNext, phase.CurrentState(), "全ての条件が満たされたので次の状態に進むはず")
	assert.True(t, phase.IsClear)
}

func TestPhase_NewPhase(t *testing.T) {
	// Arrange
	name := "test_phase"
	order := 1
	cond := createCondition(order, 1)

	// Act
	phase := NewPhase(name, order, []*Condition{cond})

	// Assert
	assert.NotNil(t, phase)
	assert.Equal(t, name, phase.Name)
	assert.Equal(t, order, phase.Order)
	assert.Equal(t, core.StateReady, phase.CurrentState())
	assert.Equal(t, ConditionTypeSingle, phase.ConditionType)
	assert.Len(t, phase.ConditionIDs, 1)
	assert.Len(t, phase.Conditions, 1)
}

func TestPhase_IsClear(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, []*Condition{createCondition(1, 1)})
	ctx := context.Background()

	// Assert initial state
	assert.False(t, phase.IsClear, "IsClear should be false initially")

	// Act: activate phase and satisfy condition
	_ = phase.Activate(ctx)
	phase.OnConditionSatisfied(core.ConditionID(1))

	// Assert: IsClear should be true after condition is satisfied
	assert.True(t, phase.IsClear, "IsClear should be true after condition is satisfied")
}

func TestPhase_ConditionTypeOr(t *testing.T) {
	// Arrange
	conditions := createMultipleConditions(3)
	phase := NewPhase("test_phase", 1, conditions)
	phase.ConditionType = ConditionTypeOr
	ctx := context.Background()

	// Act & Assert: 初期状態の確認
	assert.Equal(t, core.StateReady, phase.CurrentState())
	assert.False(t, phase.IsClear)

	// フェーズをアクティブ化
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateActive, phase.CurrentState())

	// 1つ目の条件を満たす
	phase.OnConditionSatisfied(core.ConditionID(1))
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.Equal(t, core.StateNext, phase.CurrentState(), "OR条件なので1つの条件で次の状態に進むはず")
	assert.True(t, phase.IsClear)
}

func TestPhase_StateTransitions(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, []*Condition{createCondition(1, 1)})
	ctx := context.Background()

	// Act & Assert
	assert.Equal(t, core.StateReady, phase.CurrentState())

	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateActive, phase.CurrentState())

	err = phase.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateNext, phase.CurrentState())

	err = phase.Finish(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateFinish, phase.CurrentState())
}

func TestPhase_ConditionSatisfaction(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, []*Condition{createCondition(1, 1)})
	ctx := context.Background()
	mockObserver := &mockPhaseObserver{
		stateChanges: make([]string, 0),
	}
	phase.AddObserver(mockObserver)

	// Act
	_ = phase.Activate(ctx)
	phase.OnConditionSatisfied(core.ConditionID(1))

	// Assert
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.Equal(t, core.StateNext, phase.CurrentState())
	assert.Contains(t, mockObserver.stateChanges, core.StateNext)
}

func TestPhases_ProcessAndActivateByNextOrder(t *testing.T) {
	// Arrange
	phases := Phases{
		NewPhase("phase1", 1, []*Condition{createCondition(1, 1)}),
		NewPhase("phase2", 2, []*Condition{createCondition(2, 2)}),
		NewPhase("phase3", 3, []*Condition{createCondition(3, 3)}),
	}
	ctx := context.Background()

	// Act & Assert
	// 最初のフェーズを開始
	nextPhase, err := phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "phase1", nextPhase.Name)
	assert.Equal(t, core.StateActive, nextPhase.CurrentState())

	// 次のフェーズに移行
	nextPhase.OnConditionSatisfied(core.ConditionID(1))
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機

	nextPhase, err = phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "phase2", nextPhase.Name)
	assert.Equal(t, core.StateActive, nextPhase.CurrentState())
}

func TestPhases_ResetAll(t *testing.T) {
	// Arrange
	phases := Phases{
		NewPhase("phase1", 1, []*Condition{createCondition(1, 1)}),
		NewPhase("phase2", 2, []*Condition{createCondition(2, 2)}),
	}
	ctx := context.Background()

	// 最初のフェーズを開始
	_, _ = phases.ProcessAndActivateByNextOrder(ctx)

	// Act
	err := phases.ResetAll(ctx)

	// Assert
	assert.NoError(t, err)
	for _, phase := range phases {
		assert.Equal(t, core.StateReady, phase.CurrentState())
	}
}

type mockPhaseObserver struct {
	stateChanges []string
}

func (m *mockPhaseObserver) OnStateChanged(state string) {
	m.stateChanges = append(m.stateChanges, state)
}

func TestPhase_ConditionTypeUnspecified(t *testing.T) {
	// Arrange
	conditions := createMultipleConditions(2)
	phase := NewPhase("test_phase", 1, conditions)
	phase.ConditionType = ConditionTypeUnspecified
	ctx := context.Background()

	// Act & Assert: 初期状態の確認
	assert.Equal(t, core.StateReady, phase.CurrentState())
	assert.False(t, phase.IsClear)

	// フェーズをアクティブ化
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateActive, phase.CurrentState())

	// 条件を満たしても状態が変化しないことを確認
	phase.OnConditionSatisfied(core.ConditionID(1))
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.Equal(t, core.StateActive, phase.CurrentState(), "未指定の条件タイプでは状態が変化しないはず")
	assert.False(t, phase.IsClear)
}

func TestPhase_InvalidStateTransition(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, []*Condition{createCondition(1, 1)})
	ctx := context.Background()

	// Ready → Next (無効)
	err := phase.Next(ctx)
	assert.Error(t, err, "Ready状態からNextへの遷移は失敗するはず")
	assert.Equal(t, core.StateReady, phase.CurrentState())

	// Ready → Finish (無効)
	err = phase.Finish(ctx)
	assert.Error(t, err, "Ready状態からFinishへの遷移は失敗するはず")
	assert.Equal(t, core.StateReady, phase.CurrentState())

	// Activate → Finish (無効)
	_ = phase.Activate(ctx)
	err = phase.Finish(ctx)
	assert.Error(t, err, "Active状態からFinishへの遷移は失敗するはず")
	assert.Equal(t, core.StateActive, phase.CurrentState())
}

func TestPhase_NoConditions(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, []*Condition{})
	ctx := context.Background()

	// Act & Assert
	assert.Equal(t, core.StateReady, phase.CurrentState())
	assert.False(t, phase.IsClear)

	// フェーズをアクティブ化
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.Equal(t, core.StateActive, phase.CurrentState())

	// 条件がないので満たされることはない
	assert.False(t, phase.checkConditionsSatisfied(), "条件がない場合はfalseを返すはず")
	assert.False(t, phase.IsClear)
}

func TestPhase_ConcurrentConditionSatisfaction(t *testing.T) {
	// Arrange
	conditions := createMultipleConditions(5)
	phase := NewPhase("test_phase", 1, conditions)
	phase.ConditionType = ConditionTypeOr
	ctx := context.Background()

	// フェーズをアクティブ化
	err := phase.Activate(ctx)
	assert.NoError(t, err)

	// 複数の条件を同時に満たす
	done := make(chan bool)
	for i := 1; i <= 5; i++ {
		go func(id int) {
			phase.OnConditionSatisfied(core.ConditionID(id))
			done <- true
		}(i)
	}

	// すべてのゴルーチンの完了を待つ
	for i := 0; i < 5; i++ {
		<-done
	}

	// 状態遷移が正しく行われたことを確認
	time.Sleep(100 * time.Millisecond) // 非同期通知の待機
	assert.Equal(t, core.StateNext, phase.CurrentState(), "OR条件なので1つの条件で次の状態に進むはず")
	assert.True(t, phase.IsClear)
}

func TestPhase_TimeManagement(t *testing.T) {
	// Arrange
	phase := NewPhase("test_phase", 1, []*Condition{createCondition(1, 1)})
	ctx := context.Background()

	// 初期状態の確認
	assert.Nil(t, phase.StartTime, "初期状態ではStartTimeはnilのはず")
	assert.Nil(t, phase.FinishTime, "初期状態ではFinishTimeはnilのはず")

	// Activate時のStartTime設定を確認
	err := phase.Activate(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, phase.StartTime, "Activate後はStartTimeが設定されているはず")
	assert.Nil(t, phase.FinishTime, "Activate後もFinishTimeはnilのはず")
	activateTime := *phase.StartTime

	// Next状態への遷移
	err = phase.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, activateTime, *phase.StartTime, "Next後もStartTimeは変更されないはず")
	assert.Nil(t, phase.FinishTime, "Next後もFinishTimeはnilのはず")

	// Finish時のFinishTime設定を確認
	err = phase.Finish(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, phase.FinishTime, "Finish後はFinishTimeが設定されているはず")
	assert.Equal(t, activateTime, *phase.StartTime, "Finish後もStartTimeは変更されないはず")

	// Reset時の時間情報初期化を確認
	err = phase.Reset(ctx)
	assert.NoError(t, err)
	assert.Nil(t, phase.StartTime, "Reset後はStartTimeがnilになるはず")
	assert.Nil(t, phase.FinishTime, "Reset後はFinishTimeがnilになるはず")
}

func TestPhases_LastPhase(t *testing.T) {
	// Arrange
	phases := Phases{
		NewPhase("phase1", 1, []*Condition{createCondition(1, 1)}),
		NewPhase("phase2", 2, []*Condition{createCondition(2, 2)}),
	}
	ctx := context.Background()

	// 最初のフェーズを開始して完了
	nextPhase, err := phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "phase1", nextPhase.Name)
	nextPhase.OnConditionSatisfied(core.ConditionID(1))
	time.Sleep(100 * time.Millisecond)

	// 2番目のフェーズを開始して完了
	nextPhase, err = phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "phase2", nextPhase.Name)
	nextPhase.OnConditionSatisfied(core.ConditionID(2))
	time.Sleep(100 * time.Millisecond)

	// 最後のフェーズの後は nil が返されることを確認
	nextPhase, err = phases.ProcessAndActivateByNextOrder(ctx)
	assert.NoError(t, err)
	assert.Nil(t, nextPhase, "最後のフェーズの後はnilが返されるはず")
}
