package model

type SettingHyperParameter struct {
	Meta Hyperparameters `json:"meta"`
}

type GameRequest struct {
	Meta              Hyperparameters    `json:"meta"`
	AnsweredQuestions []AnsweredQuestion `json:"questions"`
}

type ConfirmAnswerRequest struct {
	Meta              Hyperparameters    `json:"meta"`
	AnsweredQuestions []AnsweredQuestion `json:"questions"`
	AnswerCard        CardData           `json:"answerCard"`
	IsCorrect         bool               `json:"isCorrect"`
}
