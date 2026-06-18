package model

type CandidateResponse struct {
	Rank        int     `json:"rank"`
	CardID      int64   `json:"cardId"`
	CardName    string  `json:"cardName"`
	Probability float64 `json:"probability"`
}

type CardData struct {
	CardID   int64  `json:"cardId"`
	CardName string `json:"cardName"`
}

type GameResponse struct {
	Meta       Hyperparameters     `json:"meta"`
	Question   *QuestionContent    `json:"question"`
	History    []AnsweredQuestion  `json:"history"`
	Candidates []CandidateResponse `json:"candidates"`
	IsAnswer   bool                `json:"isAnswer"`
	IsCorrect  *bool               `json:"isCorrect"`
	AnswerCard *CardData           `json:"answerCard"`
}
