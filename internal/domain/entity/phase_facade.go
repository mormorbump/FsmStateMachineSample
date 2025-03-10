package entity

import (
	"state_sample/internal/domain/value"
	logger "state_sample/internal/lib"
	"sync"

	"go.uber.org/zap"
)

// PhaseFacade はフェーズの検索と管理を担当する構造体です
type PhaseFacade struct {
	allPhases       Phases
	phaseMap        PhaseMap
	currentPhaseMap CurrentPhaseMap
	mu              sync.RWMutex
	log             *zap.Logger
}

// NewPhaseFacade は新しいPhaseFacadeを作成します
func NewPhaseFacade(phases Phases) *PhaseFacade {
	log := logger.DefaultLogger()

	// 親子関係を初期化
	InitializePhaseHierarchy(phases)

	// ParentIDごとにグループ化
	phaseMap := GroupPhasesByParentID(phases)

	return &PhaseFacade{
		allPhases:       phases,
		phaseMap:        phaseMap,
		currentPhaseMap: make(CurrentPhaseMap),
		log:             log,
	}
}

// GetAllPhases は全フェーズを取得します
func (pf *PhaseFacade) GetAllPhases() Phases {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.allPhases
}

// GetPhaseMap はPhaseMapを取得します
func (pf *PhaseFacade) GetPhaseMap() PhaseMap {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.phaseMap
}

// GetCurrentPhaseMap はCurrentPhaseMapを取得します
func (pf *PhaseFacade) GetCurrentPhaseMap() CurrentPhaseMap {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.currentPhaseMap
}

// GetCurrentPhase は指定された親IDに対する現在のフェーズを返します
func (pf *PhaseFacade) GetCurrentPhase(parentID value.PhaseID) *Phase {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.currentPhaseMap[parentID]
}

// GetCurrentLeafPhase は現在アクティブな最下層のフェーズを返します
func (pf *PhaseFacade) GetCurrentLeafPhase() *Phase {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	// 現在のルートフェーズを取得
	rootPhase := pf.currentPhaseMap[0]
	if rootPhase == nil {
		return nil
	}

	// 子フェーズがある場合は再帰的に最下層のフェーズを探す
	return pf.FindCurrentLeafPhase(rootPhase)
}

// FindCurrentLeafPhase は再帰的に最下層のフェーズを探します
func (pf *PhaseFacade) FindCurrentLeafPhase(phase *Phase) *Phase {
	if phase == nil {
		pf.log.Error("FindCurrentLeafPhase: phase is nil")
		return nil
	}

	if !phase.HasChildren() {
		return phase
	}

	childPhase := pf.currentPhaseMap[phase.ID]
	if childPhase == nil {
		return phase
	}

	// 再帰呼び出しの結果がnilの場合は現在のフェーズを返す
	result := pf.FindCurrentLeafPhase(childPhase)
	if result == nil {
		return phase
	}

	return result
}

// GetPhasesByParentID は指定された親IDに対するフェーズのスライスを返します
func (pf *PhaseFacade) GetPhasesByParentID(parentID value.PhaseID) Phases {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.phaseMap[parentID]
}

// SetCurrentPhase は指定された親IDに対する現在のフェーズを設定します
func (pf *PhaseFacade) SetCurrentPhase(phase *Phase) {
	if phase == nil {
		pf.log.Error("SetCurrentPhase: phase is nil")
		return
	}

	pf.mu.Lock()
	defer pf.mu.Unlock()

	oldPhase := pf.currentPhaseMap[phase.ParentID]
	oldPhaseName := ""
	if oldPhase != nil {
		oldPhaseName = oldPhase.Name
	}

	pf.currentPhaseMap[phase.ParentID] = phase
	pf.log.Debug("PhaseFacade.SetCurrentPhase",
		zap.String("old phase", oldPhaseName),
		zap.String("new phase", phase.Name),
		zap.Int64("parent_id", int64(phase.ParentID)))
}

// ResetCurrentPhaseMap は現在のフェーズマップをリセットします
func (pf *PhaseFacade) ResetCurrentPhaseMap() {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.currentPhaseMap = make(CurrentPhaseMap)
}
