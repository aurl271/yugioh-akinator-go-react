package model

type AnswerValue string

const (
	AnswerYes        AnswerValue = "yes"
	AnswerProbably   AnswerValue = "probably"
	AnswerUnknown    AnswerValue = "unknown"
	AnswerProbablyNo AnswerValue = "probably_no"
	AnswerNo         AnswerValue = "no"
)

type QuestionContent struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type AnsweredQuestion struct {
	Question QuestionContent `json:"questionContent"`
	Answer   AnswerValue     `json:"answer"`
}

type Hyperparameters struct {
	Beta               float64 `json:"beta"`
	AnswerThreshold    float64 `json:"answerThreshold"`
	TopCandidatesCount int     `json:"topCandidatesCount"`
}
