package state

//
//// TestStateFacadeWithMultipleConditions は複数の条件を持つStateFacadeの動作をテストします
//func TestStateFacadeWithMultipleConditions(t *testing.T) {
//	// 本来のNewStateFacadeを使用すると、テストが複雑になるため、
//	// テスト用のStateFacadeを作成します
//	ctx := context.Background()
//	facade := createTestStateFacade()
//
//	// 現在のPhaseを取得
//	currentPhase := facade.GetCurrentPhase()
//	assert.NotNil(t, currentPhase)
//	assert.Equal(t, "PHASE1", currentPhase.Name)
//	assert.Equal(t, value.StateReady, currentPhase.CurrentState())
//
//	// Phaseをアクティブにする
//	err := facade.Start(ctx)
//	assert.NoError(t, err)
//	currentPhase = facade.GetCurrentPhase()
//	assert.Equal(t, value.StateActive, currentPhase.CurrentState())
//
//	// 条件の数を確認
//	conditions := currentPhase.GetConditions()
//	assert.Len(t, conditions, 2, "Phase1 should have 2 conditions")
//
//	// 条件のIDを確認
//	var conditionIDs []value.ConditionID
//	for id := range conditions {
//		conditionIDs = append(conditionIDs, id)
//	}
//	assert.Contains(t, conditionIDs, value.ConditionID(1))
//	assert.Contains(t, conditionIDs, value.ConditionID(2))
//
//	// 条件1（カウンター条件）を満たす
//	part1, err := facade.GetConditionPart(1, 1)
//	assert.NoError(t, err)
//	assert.NotNil(t, part1)
//	err = part1.Process(ctx, 1) // カウンターを1増やす
//	assert.NoError(t, err)
//	time.Sleep(100 * time.Millisecond) // 状態更新を待つ
//
//	// まだPhaseは完了していない（AND条件なので両方必要）
//	currentPhase = facade.GetCurrentPhase()
//	assert.Equal(t, value.StateActive, currentPhase.CurrentState())
//
//	// 条件2（時間条件）を満たす
//	time.Sleep(1 * time.Second) // 時間条件を満たすために待機
//
//	// Phaseが完了したことを確認（Next状態に遷移）
//	currentPhase = facade.GetCurrentPhase()
//	assert.Equal(t, value.StateNext, currentPhase.CurrentState())
//
//	// リセットして初期状態に戻す
//	err = facade.Reset(ctx)
//	assert.NoError(t, err)
//	currentPhase = facade.GetCurrentPhase()
//	assert.Equal(t, value.StateReady, currentPhase.CurrentState())
//}
//
//// createTestStateFacade はテスト用のStateFacadeを作成します
//func createTestStateFacade() GameFacade {
//	// テスト用のストラテジーファクトリを作成
//	factory := strategy.NewStrategyFactory()
//
//	// Phase1: AND条件（2つの条件）
//	part1_1 := entity.NewConditionPart(1, "Counter_Part_1")
//	part1_1.ReferenceValueInt = 1
//	part1_1.ComparisonOperator = value.ComparisonOperatorGTE
//	cond1_1 := entity.NewCondition(1, "Counter_Condition_1", value.KindCounter)
//	cond1_1.AddPart(part1_1)
//
//	part1_2 := entity.NewConditionPart(2, "Time_Part_1")
//	part1_2.ReferenceValueInt = 1
//	cond1_2 := entity.NewCondition(2, "Time_Condition_1", value.KindTime)
//	cond1_2.AddPart(part1_2)
//
//	if err := cond1_1.InitializePartStrategies(factory); err != nil {
//		panic(err)
//	}
//	if err := cond1_2.InitializePartStrategies(factory); err != nil {
//		panic(err)
//	}
//
//	phase1 := entity.NewPhase("PHASE1", 1, []*entity.Condition{cond1_1, cond1_2}, value.ConditionTypeAnd, value.GameRule_Animation)
//	part1_1.AddConditionPartObserver(cond1_1)
//	part1_2.AddConditionPartObserver(cond1_2)
//	cond1_1.AddConditionObserver(phase1)
//	cond1_2.AddConditionObserver(phase1)
//
//	// Phase2: OR条件（2つの条件）
//	part2_1 := entity.NewConditionPart(3, "Counter_Part_2")
//	part2_1.ReferenceValueInt = 1
//	part2_1.ComparisonOperator = value.ComparisonOperatorGTE
//	cond2_1 := entity.NewCondition(3, "Counter_Condition_2", value.KindCounter)
//	cond2_1.AddPart(part2_1)
//
//	part2_2 := entity.NewConditionPart(4, "Time_Part_2")
//	part2_2.ReferenceValueInt = 3
//	cond2_2 := entity.NewCondition(4, "Time_Condition_2", value.KindTime)
//	cond2_2.AddPart(part2_2)
//
//	if err := cond2_1.InitializePartStrategies(factory); err != nil {
//		panic(err)
//	}
//	if err := cond2_2.InitializePartStrategies(factory); err != nil {
//		panic(err)
//	}
//
//	phase2 := entity.NewPhase("PHASE2", 2, []*entity.Condition{cond2_1, cond2_2}, value.ConditionTypeOr, value.GameRule_Animation)
//	part2_1.AddConditionPartObserver(cond2_1)
//	part2_2.AddConditionPartObserver(cond2_2)
//	cond2_1.AddConditionObserver(phase2)
//	cond2_2.AddConditionObserver(phase2)
//
//	// PhaseControllerを作成
//	phases := entity.Phases{phase1, phase2}
//	controller := NewPhaseController(phases)
//
//	// StateFacadeを作成
//	return &GameFacade{
//		controller: controller,
//	}
//}
