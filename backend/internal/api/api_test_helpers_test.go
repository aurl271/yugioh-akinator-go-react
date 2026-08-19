package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"yugioh-akinator-backend/internal/game"
	"yugioh-akinator-backend/internal/model"
)

const (
	apiQuestionAID   = 101
	apiQuestionBID   = 102
	apiQuestionAText = "質問A"
	apiQuestionBText = "質問B"
	apiCardAID       = 2001
	apiCardBID       = 2002
	apiCardAName     = "カードA"
	apiCardBName     = "カードB"
)

// apiTestGameData はAPIテスト用の2枚のカードと2つの質問を毎回新しく作る。
func apiTestGameData() game.GameData {
	questionA := game.Question{ID: apiQuestionAID, Text: apiQuestionAText}
	questionB := game.Question{ID: apiQuestionBID, Text: apiQuestionBText}

	return game.GameData{
		Cards: []game.Card{
			{CardID: apiCardAID, Name: apiCardAName},
			{CardID: apiCardBID, Name: apiCardBName},
		},
		Questions: []game.Question{questionA, questionB},
		QuestionByID: map[int]game.Question{
			apiQuestionAID: questionA,
			apiQuestionBID: questionB,
		},
		Answers: map[int]map[int64]int{
			apiQuestionAID: {apiCardAID: game.ExpectedAnswerYes},
			apiQuestionBID: {
				apiCardAID: game.ExpectedAnswerYes,
				apiCardBID: game.ExpectedAnswerYes,
			},
		},
	}
}

// apiTestMeta はAPIテストで共通して送信する推理設定を返す。
func apiTestMeta() model.Hyperparameters {
	return model.Hyperparameters{
		Beta:               1.0,
		AnswerThreshold:    0.9,
		TopCandidatesCount: 2,
	}
}

// apiTestHandler は実DBを使わないServiceを持つGameHandlerを作る。
func apiTestHandler(data game.GameData) *GameHandler {
	return NewGameHandler(game.NewService(data), nil)
}

// apiAnsweredQuestion は質問IDと回答からAPIへ送信する回答履歴を作る。
func apiAnsweredQuestion(questionID int, questionText string, answer model.AnswerValue) model.AnsweredQuestion {
	return model.AnsweredQuestion{
		Question: model.QuestionContent{ID: questionID, Text: questionText},
		Answer:   answer,
	}
}

// apiJSONBody は値をJSONリクエストbodyへ変換し、変換失敗時はテストを終了する。
func apiJSONBody(t *testing.T, value any) *bytes.Buffer {
	t.Helper()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(value); err != nil {
		t.Fatalf("encode request body: %v", err)
	}
	return &body
}

// decodeJSONResponse はRecorderのJSONレスポンスを指定された型へ変換する。
func decodeJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.NewDecoder(recorder.Body).Decode(target); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

// assertJSONResponse はHTTPステータスとContent-Typeが期待どおりかを確認する。
func assertJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()

	if recorder.Code != wantStatus {
		t.Errorf("status code = %d; want %d", recorder.Code, wantStatus)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q; want %q", got, "application/json")
	}
}

// assertErrorResponse はエラーレスポンスのstatus、Content-Type、error本文を確認する。
func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantError string) {
	t.Helper()
	assertJSONResponse(t, recorder, wantStatus)

	var response map[string]string
	decodeJSONResponse(t, recorder, &response)
	if response["error"] != wantError {
		t.Errorf("error response = %q; want %q", response["error"], wantError)
	}
}

// newAPIRequest はテスト対象のURLとbodyからHTTPリクエストを作る。
func newAPIRequest(method string, target string, body *bytes.Buffer) *http.Request {
	return httptest.NewRequest(method, target, body)
}
