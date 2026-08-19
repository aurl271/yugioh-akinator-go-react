package game

import (
	"reflect"
	"testing"
	"yugioh-akinator-backend/internal/model"
)

// TestConfirmAnswerReturnsFeedbackResponse は正解・不正解の入力がフィードバック用レスポンスへ保持されることを確認する。
func TestConfirmAnswerReturnsFeedbackResponse(t *testing.T) {
	tests := []struct {
		name      string
		isCorrect bool
	}{
		{name: "correct", isCorrect: true},
		{name: "incorrect", isCorrect: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := serviceTestMeta()
			history := []model.AnsweredQuestion{
				serviceAnsweredQuestion(serviceQuestionAID, serviceQuestionAText, model.AnswerYes),
			}
			answerCard := model.CardData{CardID: serviceCardAID, CardName: serviceCardAName}
			request := model.ConfirmAnswerRequest{
				Meta:              meta,
				AnsweredQuestions: history,
				AnswerCard:        answerCard,
				IsCorrect:         tt.isCorrect,
			}
			service := NewService(serviceTestGameData())

			response, err := service.ConfirmAnswer(request)
			if err != nil {
				t.Fatalf("ConfirmAnswer returned error: %v", err)
			}
			if response.Meta != meta {
				t.Errorf("ConfirmAnswer Meta = %+v; want %+v", response.Meta, meta)
			}
			if !reflect.DeepEqual(response.History, history) {
				t.Errorf("ConfirmAnswer History = %+v; want %+v", response.History, history)
			}
			if !response.IsAnswer {
				t.Error("ConfirmAnswer IsAnswer = false; want true")
			}
			if response.IsCorrect == nil {
				t.Fatal("ConfirmAnswer IsCorrect = nil; want bool pointer")
			}
			if *response.IsCorrect != tt.isCorrect {
				t.Errorf("ConfirmAnswer IsCorrect = %t; want %t", *response.IsCorrect, tt.isCorrect)
			}
			if response.AnswerCard == nil {
				t.Fatal("ConfirmAnswer AnswerCard = nil; want answer card")
			}
			if *response.AnswerCard != answerCard {
				t.Errorf("ConfirmAnswer AnswerCard = %+v; want %+v", *response.AnswerCard, answerCard)
			}
			if response.Question != nil {
				t.Errorf("ConfirmAnswer Question = %+v; want nil", response.Question)
			}
			if response.Candidates == nil || len(response.Candidates) != 0 {
				t.Errorf("ConfirmAnswer Candidates = %#v; want non-nil empty slice", response.Candidates)
			}
		})
	}
}
