package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yugioh-akinator-backend/internal/model"
)

// TestAnswerQuestionHandlerRejectsNonPostMethod はPOST以外のメソッドに405を返すことを確認する。
func TestAnswerQuestionHandlerRejectsNonPostMethod(t *testing.T) {
	handler := apiTestHandler(apiTestGameData())
	request := httptest.NewRequest(http.MethodGet, "/api/game/answer", nil)
	recorder := httptest.NewRecorder()

	handler.AnswerQuestionHandler(recorder, request)

	assertErrorResponse(t, recorder, http.StatusMethodNotAllowed, "method not allowed")
}

// TestAnswerQuestionHandlerRejectsInvalidJSON は壊れたJSONリクエストに400を返すことを確認する。
func TestAnswerQuestionHandlerRejectsInvalidJSON(t *testing.T) {
	handler := apiTestHandler(apiTestGameData())
	request := httptest.NewRequest(http.MethodPost, "/api/game/answer", strings.NewReader("{"))
	recorder := httptest.NewRecorder()

	handler.AnswerQuestionHandler(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid request body")
}

// TestAnswerQuestionHandlerReturnsNextQuestion はしきい値未満の回答履歴から次の質問をJSONで返すことを確認する。
func TestAnswerQuestionHandlerReturnsNextQuestion(t *testing.T) {
	handler := apiTestHandler(apiTestGameData())
	history := []model.AnsweredQuestion{
		apiAnsweredQuestion(apiQuestionAID, apiQuestionAText, model.AnswerUnknown),
	}
	requestBody := model.GameRequest{Meta: apiTestMeta(), AnsweredQuestions: history}
	request := newAPIRequest(http.MethodPost, "/api/game/answer", apiJSONBody(t, requestBody))
	recorder := httptest.NewRecorder()

	handler.AnswerQuestionHandler(recorder, request)

	assertJSONResponse(t, recorder, http.StatusOK)
	var response model.GameResponse
	decodeJSONResponse(t, recorder, &response)
	if response.IsAnswer {
		t.Error("response IsAnswer = true; want false")
	}
	if response.Question == nil {
		t.Fatal("response Question = nil; want next question")
	}
	if response.Question.ID != apiQuestionBID {
		t.Errorf("response questionID = %d; want %d", response.Question.ID, apiQuestionBID)
	}
	if len(response.History) != len(history) {
		t.Errorf("response History length = %d; want %d", len(response.History), len(history))
	}
	if len(response.Candidates) != apiTestMeta().TopCandidatesCount {
		t.Errorf(
			"response Candidates length = %d; want %d",
			len(response.Candidates),
			apiTestMeta().TopCandidatesCount,
		)
	}
}

// TestAnswerQuestionHandlerReturnsAnswerCard は十分に確率が高い候補を最終回答カードとしてJSONで返すことを確認する。
func TestAnswerQuestionHandlerReturnsAnswerCard(t *testing.T) {
	handler := apiTestHandler(apiTestGameData())
	history := []model.AnsweredQuestion{
		apiAnsweredQuestion(apiQuestionAID, apiQuestionAText, model.AnswerYes),
	}
	requestBody := model.GameRequest{Meta: apiTestMeta(), AnsweredQuestions: history}
	request := newAPIRequest(http.MethodPost, "/api/game/answer", apiJSONBody(t, requestBody))
	recorder := httptest.NewRecorder()

	handler.AnswerQuestionHandler(recorder, request)

	assertJSONResponse(t, recorder, http.StatusOK)
	var response model.GameResponse
	decodeJSONResponse(t, recorder, &response)
	if !response.IsAnswer {
		t.Error("response IsAnswer = false; want true")
	}
	if response.Question != nil {
		t.Errorf("response Question = %+v; want nil", response.Question)
	}
	if response.AnswerCard == nil {
		t.Fatal("response AnswerCard = nil; want top card")
	}
	if response.AnswerCard.CardID != apiCardAID {
		t.Errorf("response cardID = %d; want %d", response.AnswerCard.CardID, apiCardAID)
	}
	if response.AnswerCard.CardName != apiCardAName {
		t.Errorf("response card name = %q; want %q", response.AnswerCard.CardName, apiCardAName)
	}
}

// TestAnswerQuestionHandlerReturnsServiceError はServiceがリクエストを処理できない場合に400とエラー本文を返すことを確認する。
func TestAnswerQuestionHandlerReturnsServiceError(t *testing.T) {
	meta := apiTestMeta()
	meta.TopCandidatesCount = 0
	handler := apiTestHandler(apiTestGameData())
	request := newAPIRequest(
		http.MethodPost,
		"/api/game/answer",
		apiJSONBody(t, model.GameRequest{Meta: meta}),
	)
	recorder := httptest.NewRecorder()

	handler.AnswerQuestionHandler(recorder, request)

	assertErrorResponse(
		t,
		recorder,
		http.StatusBadRequest,
		"topCandidatesCount must be greater than 0: got 0",
	)
}
