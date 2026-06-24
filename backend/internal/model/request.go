package model

// SettingHyperParameter はゲーム開始時のリクエスト。
type SettingHyperParameter struct {
	Meta Hyperparameters `json:"meta"`
}

// GameRequest は質問回答時のリクエスト。
// frontendが持つ回答履歴をすべて送ることで、backendはステートレスに推理状態を復元する。
type GameRequest struct {
	Meta              Hyperparameters    `json:"meta"`
	AnsweredQuestions []AnsweredQuestion `json:"questions"`
}

// ConfirmAnswerRequest は回答確認時のリクエスト。
type ConfirmAnswerRequest struct {
	Meta              Hyperparameters    `json:"meta"`
	AnsweredQuestions []AnsweredQuestion `json:"questions"`
	AnswerCard        CardData           `json:"answerCard"`
	IsCorrect         bool               `json:"isCorrect"`
}
