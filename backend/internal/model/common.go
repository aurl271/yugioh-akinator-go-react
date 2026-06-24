package model

// AnswerValue はフロントエンドとAPIで共有する回答文字列。
type AnswerValue string

const (
	AnswerYes        AnswerValue = "yes"
	AnswerProbably   AnswerValue = "probably"
	AnswerUnknown    AnswerValue = "unknown"
	AnswerProbablyNo AnswerValue = "probably_no"
	AnswerNo         AnswerValue = "no"
)

// QuestionContent はフロントエンドへ返す質問の最小情報。
type QuestionContent struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// AnsweredQuestion はフロントエンドが保持し、リクエストごとに送る回答履歴。
type AnsweredQuestion struct {
	Question QuestionContent `json:"questionContent"`
	Answer   AnswerValue     `json:"answer"`
}

// Hyperparameters は推理の挙動を調整する値。
type Hyperparameters struct {
	// Beta はscore差を確率へ変換するときの鋭さ。
	Beta               float64 `json:"beta"`
	// AnswerThreshold は最上位候補を回答として提示する確率しきい値。
	AnswerThreshold    float64 `json:"answerThreshold"`
	// TopCandidatesCount はフロントエンドへ返す候補カード数。
	TopCandidatesCount int     `json:"topCandidatesCount"`
}
