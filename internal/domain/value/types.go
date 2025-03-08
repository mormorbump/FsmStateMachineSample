package value

// ConditionKind は条件の種類を表す型です
type ConditionKind int

const (
	KindUnspecified ConditionKind = iota
	KindTime                      // 時間に基づく条件
	KindCounter                   // カウンターに基づく条件
)

// ComparisonOperator は比較演算子を表す型です
type ComparisonOperator int

const (
	ComparisonOperatorUnspecified ComparisonOperator = iota
	ComparisonOperatorEQ
	ComparisonOperatorNEQ
	ComparisonOperatorGT
	ComparisonOperatorGTE
	ComparisonOperatorLT
	ComparisonOperatorLTE
	ComparisonOperatorBetween
	ComparisonOperatorIn
	ComparisonOperatorNotIn
)

// ConditionID は条件のIDを表す型です
type ConditionID int64

// ConditionPartID は条件パーツのIDを表す型です
type ConditionPartID int64

// PhaseID はフェーズのIDを表す型です
type PhaseID int

// GameRule はゲームルールを表す型です
type GameRule int

const (
	GameRule_Shooting GameRule = iota
	GameRule_PushSwitch
	GameRule_Animation
)

// ConditionType は条件の組み合わせ方を表す型です
type ConditionType int

const (
	ConditionTypeUnspecified ConditionType = iota
	ConditionTypeAnd                       // すべての条件を満たす必要がある
	ConditionTypeOr                        // いずれかの条件を満たせばよい
)

// ゲーム状態の定義
const (
	StateReady  = "ready"
	StateActive = "active"
	StateNext   = "next"
	StateFinish = "finish"
)

// ゲームイベントの定義
const (
	EventActivate = "activate"
	EventNext     = "next"
	EventFinish   = "finish"
	EventReset    = "reset"
)

// 条件状態の定義
const (
	StateUnsatisfied = "unsatisfied" // 条件未達成
	StateProcessing  = "processing"  // 処理中
	StateSatisfied   = "satisfied"   // 条件達成
)

// 条件イベントの定義
const (
	EventProcess  = "process"  // 処理開始
	EventComplete = "complete" // 条件達成で完了
	EventTimeout  = "timeout"  // (時間系の)条件達成で完了
	EventRevert   = "revert"   // 条件未達で差し戻し
)
