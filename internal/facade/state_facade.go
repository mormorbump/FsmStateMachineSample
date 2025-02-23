package facade

import (
	"context"
	"log"
	"time"

	"state_sample/internal/fsm"
)

// StateObserver はFacadeの状態変更を監視するインターフェースです
type StateObserver interface {
	OnStateChanged(state StateInfo)
	OnError(err error)
}

// StateInfo は状態に関する情報を保持します
type StateInfo struct {
	State          string         `json:"state"`
	Info           *fsm.StateInfo `json:"info,omitempty"`
	LastTransition time.Time      `json:"lastTransition"`
	NextTransition time.Time      `json:"nextTransition"`
}

// StateFacade はFSMの操作を提供するインターフェースです
type StateFacade interface {
	// 状態の取得と監視
	GetCurrentState() StateInfo
	AddObserver(observer StateObserver)
	RemoveObserver(observer StateObserver)

	// 状態遷移の制御
	Start(ctx context.Context) error
	Stop() error
	Reset(ctx context.Context) error

	// リソース管理
	Close() error
}

// stateFacadeImpl はStateFacadeの実装です
type stateFacadeImpl struct {
	fsm            *fsm.FSMContext
	timeController *fsm.TimeController
	observers      []StateObserver
	nextCounter    int
}

// NewStateFacade は新しいStateFacadeインスタンスを作成します
func NewStateFacade() StateFacade {
	fsmContext := fsm.NewFSMContext()
	facade := &stateFacadeImpl{
		fsm:       fsmContext,
		observers: make([]StateObserver, 0),
	}

	// タイマーイベントのハンドラを設定
	facade.timeController = fsm.NewTimeController(10*time.Second, facade.onTimerTick)

	return facade
}

// GetCurrentState は現在の状態情報を返します
func (f *stateFacadeImpl) GetCurrentState() StateInfo {
	return StateInfo{
		State:          f.fsm.CurrentState(),
		Info:           f.fsm.GetCurrentStateInfo(),
		LastTransition: time.Now(),
		NextTransition: f.timeController.GetNextTrigger(),
	}
}

// AddObserver はオブザーバーを追加します
func (f *stateFacadeImpl) AddObserver(observer StateObserver) {
	f.observers = append(f.observers, observer)
}

// RemoveObserver はオブザーバーを削除します
func (f *stateFacadeImpl) RemoveObserver(observer StateObserver) {
	for i, obs := range f.observers {
		if obs == observer {
			f.observers = append(f.observers[:i], f.observers[i+1:]...)
			break
		}
	}
}

// Start は自動遷移を開始します
func (f *stateFacadeImpl) Start(ctx context.Context) error {
	if f.fsm.CurrentState() != fsm.StateReady {
		return &fsm.StateError{
			Code:    "INVALID_STATE",
			Message: "Can only start from ready state",
		}
	}

	// ready -> activeの即時遷移
	if err := f.fsm.Transition(ctx, fsm.EventActivate); err != nil {
		return err
	}

	// active状態になったらタイマーを開始
	f.timeController.Start()
	f.notifyStateChanged()
	return nil
}

// Stop は自動遷移を停止します
func (f *stateFacadeImpl) Stop() error {
	f.timeController.Stop()
	return nil
}

// Reset は状態をリセットします
func (f *stateFacadeImpl) Reset(ctx context.Context) error {
	f.timeController.Stop()
	f.nextCounter = 0
	log.Printf("Reset next counter: %d", f.nextCounter)

	if err := f.fsm.Reset(ctx); err != nil {
		return err
	}

	f.notifyStateChanged()
	return nil
}

// Close はリソースを解放します
func (f *stateFacadeImpl) Close() error {
	f.timeController.Stop()
	return nil
}

// onTimerTick はタイマーイベントを処理します
func (f *stateFacadeImpl) onTimerTick() {
	currentState := f.fsm.CurrentState()
	log.Printf("Timer event received in state: %s", currentState)

	var event string
	switch currentState {
	case fsm.StateActive:
		event = fsm.EventNext
	case fsm.StateNext:
		if f.nextCounter < 2 {
			event = fsm.EventActivate // 2回までactiveに戻る
			f.nextCounter++
			log.Printf("Incrementing next counter: %d", f.nextCounter)
		} else {
			event = fsm.EventFinish // 3回目でfinishに遷移
			log.Printf("Next counter reached limit: %d, transitioning to finish", f.nextCounter)
		}
	case fsm.StateFinish:
		f.timeController.Stop()
		return
	default:
		return
	}

	if err := f.fsm.Transition(context.Background(), event); err != nil {
		log.Printf("Error during state transition: %v", err)
		f.notifyError(err)
	} else {
		f.notifyStateChanged()
	}
}

// notifyStateChanged は状態変更をオブザーバーに通知します
func (f *stateFacadeImpl) notifyStateChanged() {
	state := f.GetCurrentState()
	for _, observer := range f.observers {
		observer.OnStateChanged(state)
	}
}

// notifyError はエラーをオブザーバーに通知します
func (f *stateFacadeImpl) notifyError(err error) {
	for _, observer := range f.observers {
		observer.OnError(err)
	}
}