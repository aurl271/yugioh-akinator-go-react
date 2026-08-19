package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yugioh-akinator-backend/internal/model"
)

// TestStartGameHandlerRejectsNonPostMethod はPOST以外のメソッドに405を返すことを確認する。
func TestStartGameHandlerRejectsNonPostMethod(t *testing.T) {
	handler := apiTestHandler(apiTestGameData())
	request := httptest.NewRequest(http.MethodGet, "/api/game/start", nil)
	recorder := httptest.NewRecorder()

	handler.StartGameHandler(recorder, request)

	assertErrorResponse(t, recorder, http.StatusMethodNotAllowed, "method not allowed")
}

// TestStartGameHandlerRejectsInvalidJSON は壊れたJSONリクエストに400を返すことを確認する。
func TestStartGameHandlerRejectsInvalidJSON(t *testing.T) {
	handler := apiTestHandler(apiTestGameData())
	request := httptest.NewRequest(http.MethodPost, "/api/game/start", strings.NewReader("{"))
	recorder := httptest.NewRecorder()

	handler.StartGameHandler(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid request body")
}

// TestStartGameHandlerReturnsGameResponse は正しいJSONからゲームを開始し、最初の質問をJSONで返すことを確認する。
func TestStartGameHandlerReturnsGameResponse(t *testing.T) {
	data := apiTestGameData()
	data.Questions = data.Questions[:1]
	handler := apiTestHandler(data)
	meta := apiTestMeta()
	request := newAPIRequest(
		http.MethodPost,
		"/api/game/start",
		apiJSONBody(t, model.SettingHyperParameter{Meta: meta}),
	)
	recorder := httptest.NewRecorder()

	handler.StartGameHandler(recorder, request)

	assertJSONResponse(t, recorder, http.StatusOK)
	var response model.GameResponse
	decodeJSONResponse(t, recorder, &response)
	if response.Meta != meta {
		t.Errorf("response Meta = %+v; want %+v", response.Meta, meta)
	}
	if response.Question == nil {
		t.Fatal("response Question = nil; want first question")
	}
	if response.Question.ID != apiQuestionAID {
		t.Errorf("response questionID = %d; want %d", response.Question.ID, apiQuestionAID)
	}
	if response.Question.Text != apiQuestionAText {
		t.Errorf("response question text = %q; want %q", response.Question.Text, apiQuestionAText)
	}
	if response.IsAnswer {
		t.Error("response IsAnswer = true; want false")
	}
}

// TestStartGameHandlerReturnsServiceError はServiceがゲームを開始できない場合に400とエラー本文を返すことを確認する。
func TestStartGameHandlerReturnsServiceError(t *testing.T) {
	data := apiTestGameData()
	data.Questions = nil
	handler := apiTestHandler(data)
	request := newAPIRequest(
		http.MethodPost,
		"/api/game/start",
		apiJSONBody(t, model.SettingHyperParameter{Meta: apiTestMeta()}),
	)
	recorder := httptest.NewRecorder()

	handler.StartGameHandler(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "next question not found")
}
