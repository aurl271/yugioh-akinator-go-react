package model

// CandidateResponse はフロントエンドへ表示する候補カード1件。
type CandidateResponse struct {
	Rank        int     `json:"rank"`
	CardID      int64   `json:"cardId"`
	CardName    string  `json:"cardName"`
	Probability float64 `json:"probability"`
}

// CardData は最終回答として提示するカード情報。
type CardData struct {
	CardID   int64  `json:"cardId"`
	CardName string `json:"cardName"`
}

// GameResponse はゲーム系APIで共通して返すレスポンス。
// 質問中・回答提示中・フィードバック後で使わないフィールドはnilや空配列になる。
type GameResponse struct {
	Meta       Hyperparameters     `json:"meta"`
	Question   *QuestionContent    `json:"question"`
	History    []AnsweredQuestion  `json:"history"`
	Candidates []CandidateResponse `json:"candidates"`
	IsAnswer   bool                `json:"isAnswer"`
	IsCorrect  *bool               `json:"isCorrect"`
	AnswerCard *CardData           `json:"answerCard"`
}
