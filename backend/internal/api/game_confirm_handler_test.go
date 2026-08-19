package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yugioh-akinator-backend/internal/model"
)

// TestConfirmAnswerHandlerRejectsNonPostMethod はPOST以外のメソッドに405を返すことを確認する。
func TestConfirmAnswerHandlerRejectsNonPostMethod(t *testing.T) {
	handler := apiTestHandler(apiTestGameData())
	request := httptest.NewRequest(http.MethodGet, "/api/game/confirm", nil)
	recorder := httptest.NewRecorder()

	handler.ConfirmAnswerHandler(recorder, request)

	assertErrorResponse(t, recorder, http.StatusMethodNotAllowed, "method not allowed")
}

// TestConfirmAnswerHandlerRejectsInvalidJSON は壊れたJSONリクエストに400を返すことを確認する。
func TestConfirmAnswerHandlerRejectsInvalidJSON(t *testing.T) {
	handler := apiTestHandler(apiTestGameData())
	request := httptest.NewRequest(http.MethodPost, "/api/game/confirm", strings.NewReader("{"))
	recorder := httptest.NewRecorder()

	handler.ConfirmAnswerHandler(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid request body")
}

// TestConfirmAnswerHandlerReturnsFeedback は正解・不正解の確認結果をフィードバック用JSONとして返すことを確認する。
func TestConfirmAnswerHandlerReturnsFeedback(t *testing.T) {
	tests := []struct {
		name      string
		isCorrect bool
	}{
		{name: "correct", isCorrect: true},
		{name: "incorrect", isCorrect: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := []model.AnsweredQuestion{
				apiAnsweredQuestion(apiQuestionAID, apiQuestionAText, model.AnswerYes),
			}
			answerCard := model.CardData{CardID: apiCardAID, CardName: apiCardAName}
			requestBody := model.ConfirmAnswerRequest{
				Meta:              apiTestMeta(),
				AnsweredQuestions: history,
				AnswerCard:        answerCard,
				IsCorrect:         tt.isCorrect,
			}
			handler := apiTestHandler(apiTestGameData())
			request := newAPIRequest(http.MethodPost, "/api/game/confirm", apiJSONBody(t, requestBody))
			recorder := httptest.NewRecorder()

			handler.ConfirmAnswerHandler(recorder, request)

			assertJSONResponse(t, recorder, http.StatusOK)
			var response model.GameResponse
			decodeJSONResponse(t, recorder, &response)
			if !response.IsAnswer {
				t.Error("response IsAnswer = false; want true")
			}
			if response.IsCorrect == nil {
				t.Fatal("response IsCorrect = nil; want bool pointer")
			}
			if *response.IsCorrect != tt.isCorrect {
				t.Errorf("response IsCorrect = %t; want %t", *response.IsCorrect, tt.isCorrect)
			}
			if response.AnswerCard == nil {
				t.Fatal("response AnswerCard = nil; want answer card")
			}
			if *response.AnswerCard != answerCard {
				t.Errorf("response AnswerCard = %+v; want %+v", *response.AnswerCard, answerCard)
			}
		})
	}
}
