package value

// 状態定義
const (
	StateReady       = "ready"       // 初期状態
	StateUnsatisfied = "unsatisfied" // 条件未達成
	StateProcessing  = "processing"  // 処理中
	StateSatisfied   = "satisfied"   // 条件達成
)

// イベント定義
const (
	EventActivate     = "activate"      // 有効化
	EventStartProcess = "start_process" // 処理開始
	EventComplete     = "complete"      // 条件達成で完了
	EventRevert       = "revert"        // 条件未達で差し戻し
)
