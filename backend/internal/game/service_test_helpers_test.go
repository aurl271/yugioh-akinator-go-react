package game

import "yugioh-akinator-backend/internal/model"

const (
	serviceQuestionAID   = 101
	serviceQuestionBID   = 102
	serviceQuestionAText = "質問A"
	serviceQuestionBText = "質問B"
	serviceCardAID       = 2001
	serviceCardBID       = 2002
	serviceCardAName     = "カードA"
	serviceCardBName     = "カードB"
)

// serviceTestGameData はServiceの状態遷移テストで使う2枚のカードと2つの質問を毎回新しく作る。
func serviceTestGameData() GameData {
	questionA := Question{ID: serviceQuestionAID, Text: serviceQuestionAText}
	questionB := Question{ID: serviceQuestionBID, Text: serviceQuestionBText}

	return GameData{
		Cards: []Card{
			{CardID: serviceCardAID, Name: serviceCardAName},
			{CardID: serviceCardBID, Name: serviceCardBName},
		},
		Questions: []Question{questionA, questionB},
		QuestionByID: map[int]Question{
			serviceQuestionAID: questionA,
			serviceQuestionBID: questionB,
		},
		Answers: map[int]map[int64]int{
			serviceQuestionAID: {serviceCardAID: ExpectedAnswerYes},
			serviceQuestionBID: {serviceCardBID: ExpectedAnswerYes},
		},
	}
}

// serviceTestMeta はServiceテストで共通して使う推理設定を返す。
func serviceTestMeta() model.Hyperparameters {
	return model.Hyperparameters{
		Beta:               1.0,
		AnswerThreshold:    0.9,
		TopCandidatesCount: 2,
	}
}

// serviceAnsweredQuestion はServiceへ渡す回答履歴をテスト用の質問情報から作る。
func serviceAnsweredQuestion(questionID int, questionText string, answer model.AnswerValue) model.AnsweredQuestion {
	return model.AnsweredQuestion{
		Question: model.QuestionContent{ID: questionID, Text: questionText},
		Answer:   answer,
	}
}
