package game

import (
	"reflect"
	"strings"
	"testing"
	"yugioh-akinator-backend/internal/model"
)

// TestAnswerValueToScore はAPIの5種類の回答文字列が対応するEngineのscoreへ変換されることを確認する。
func TestAnswerValueToScore(t *testing.T) {
	tests := []struct {
		name   string
		answer model.AnswerValue
		want   float64
	}{
		{name: "yes", answer: model.AnswerYes, want: AnswerScoreYes},
		{name: "probably", answer: model.AnswerProbably, want: AnswerScoreProbably},
		{name: "unknown", answer: model.AnswerUnknown, want: AnswerScoreUnknown},
		{name: "probably no", answer: model.AnswerProbablyNo, want: AnswerScoreProbablyNo},
		{name: "no", answer: model.AnswerNo, want: AnswerScoreNo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := answerValueToScore(tt.answer)
			if err != nil {
				t.Fatalf("answerValueToScore returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("answerValueToScore = %f; want %f", got, tt.want)
			}
		})
	}
}

// TestAnswerValueToScoreRejectsInvalidValue は定義されていない回答文字列がエラーになることを確認する。
func TestAnswerValueToScoreRejectsInvalidValue(t *testing.T) {
	const invalidAnswer model.AnswerValue = "invalid"

	got, err := answerValueToScore(invalidAnswer)
	if err == nil {
		t.Fatal("answerValueToScore returned nil error for invalid answer")
	}
	if got != 0 {
		t.Errorf("answerValueToScore = %f; want 0", got)
	}
}

// TestAnswerQuestionReturnsNextQuestionBelowThreshold は候補確率がしきい値未満なら次の質問を返すことを確認する。
func TestAnswerQuestionReturnsNextQuestionBelowThreshold(t *testing.T) {
	meta := serviceTestMeta()
	history := []model.AnsweredQuestion{
		serviceAnsweredQuestion(serviceQuestionAID, serviceQuestionAText, model.AnswerUnknown),
	}
	request := model.GameRequest{Meta: meta, AnsweredQuestions: history}
	service := NewService(serviceTestGameData())

	response, err := service.AnswerQuestion(request)
	if err != nil {
		t.Fatalf("AnswerQuestion returned error: %v", err)
	}
	if response.IsAnswer {
		t.Error("AnswerQuestion IsAnswer = true; want false")
	}
	if response.Question == nil {
		t.Fatal("AnswerQuestion Question = nil; want next question")
	}
	if response.Question.ID != serviceQuestionBID {
		t.Errorf("AnswerQuestion questionID = %d; want %d", response.Question.ID, serviceQuestionBID)
	}
	if response.AnswerCard != nil {
		t.Errorf("AnswerQuestion AnswerCard = %+v; want nil", response.AnswerCard)
	}
	if response.IsCorrect != nil {
		t.Errorf("AnswerQuestion IsCorrect = %v; want nil", response.IsCorrect)
	}
	if !reflect.DeepEqual(response.History, history) {
		t.Errorf("AnswerQuestion History = %+v; want %+v", response.History, history)
	}
	if len(response.Candidates) != meta.TopCandidatesCount {
		t.Errorf(
			"AnswerQuestion Candidates length = %d; want %d",
			len(response.Candidates),
			meta.TopCandidatesCount,
		)
	}
}

// TestAnswerQuestionReturnsAnswerCardAboveThreshold はYES/NOで確率がしきい値を超えたカードを最終回答として返すことを確認する。
func TestAnswerQuestionReturnsAnswerCardAboveThreshold(t *testing.T) {
	tests := []struct {
		name       string
		answer     model.AnswerValue
		wantCardID int64
		wantName   string
	}{
		{
			name:       "yes selects card A",
			answer:     model.AnswerYes,
			wantCardID: serviceCardAID,
			wantName:   serviceCardAName,
		},
		{
			name:       "no selects card B",
			answer:     model.AnswerNo,
			wantCardID: serviceCardBID,
			wantName:   serviceCardBName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := []model.AnsweredQuestion{
				serviceAnsweredQuestion(serviceQuestionAID, serviceQuestionAText, tt.answer),
			}
			request := model.GameRequest{
				Meta:              serviceTestMeta(),
				AnsweredQuestions: history,
			}
			service := NewService(serviceTestGameData())

			response, err := service.AnswerQuestion(request)
			if err != nil {
				t.Fatalf("AnswerQuestion returned error: %v", err)
			}
			if !response.IsAnswer {
				t.Error("AnswerQuestion IsAnswer = false; want true")
			}
			if response.Question != nil {
				t.Errorf("AnswerQuestion Question = %+v; want nil", response.Question)
			}
			if response.AnswerCard == nil {
				t.Fatal("AnswerQuestion AnswerCard = nil; want top card")
			}
			if response.AnswerCard.CardID != tt.wantCardID {
				t.Errorf("AnswerQuestion cardID = %d; want %d", response.AnswerCard.CardID, tt.wantCardID)
			}
			if response.AnswerCard.CardName != tt.wantName {
				t.Errorf("AnswerQuestion card name = %q; want %q", response.AnswerCard.CardName, tt.wantName)
			}
			if len(response.Candidates) == 0 || response.Candidates[0].CardID != tt.wantCardID {
				t.Errorf("AnswerQuestion first candidate = %+v; want cardID %d", response.Candidates, tt.wantCardID)
			}
		})
	}
}

// TestAnswerQuestionReturnsAnswerCardAtThreshold は候補確率がしきい値と等しい場合も回答カードを返すことを確認する。
func TestAnswerQuestionReturnsAnswerCardAtThreshold(t *testing.T) {
	meta := serviceTestMeta()
	meta.AnswerThreshold = 0.5
	history := []model.AnsweredQuestion{
		serviceAnsweredQuestion(serviceQuestionAID, serviceQuestionAText, model.AnswerUnknown),
	}
	service := NewService(serviceTestGameData())

	response, err := service.AnswerQuestion(model.GameRequest{
		Meta:              meta,
		AnsweredQuestions: history,
	})
	if err != nil {
		t.Fatalf("AnswerQuestion returned error: %v", err)
	}
	if !response.IsAnswer {
		t.Error("AnswerQuestion IsAnswer = false; want true at threshold")
	}
	if response.AnswerCard == nil {
		t.Fatal("AnswerQuestion AnswerCard = nil; want card at threshold")
	}
}

// TestAnswerQuestionRejectsInvalidAnswerValue は回答履歴に不正な回答文字列がある場合にエラーになることを確認する。
func TestAnswerQuestionRejectsInvalidAnswerValue(t *testing.T) {
	history := []model.AnsweredQuestion{
		serviceAnsweredQuestion(serviceQuestionAID, serviceQuestionAText, model.AnswerValue("invalid")),
	}
	service := NewService(serviceTestGameData())

	_, err := service.AnswerQuestion(model.GameRequest{
		Meta:              serviceTestMeta(),
		AnsweredQuestions: history,
	})
	if err == nil {
		t.Fatal("AnswerQuestion returned nil error for invalid answer value")
	}
	if !strings.Contains(err.Error(), "invalid answer value") {
		t.Errorf("AnswerQuestion error = %q; want invalid answer value error", err.Error())
	}
}

// TestAnswerQuestionRejectsUnknownQuestion は回答履歴に未登録の質問IDがある場合にエラーになることを確認する。
func TestAnswerQuestionRejectsUnknownQuestion(t *testing.T) {
	const unknownQuestionID = 999
	history := []model.AnsweredQuestion{
		serviceAnsweredQuestion(unknownQuestionID, "存在しない質問", model.AnswerYes),
	}
	service := NewService(serviceTestGameData())

	_, err := service.AnswerQuestion(model.GameRequest{
		Meta:              serviceTestMeta(),
		AnsweredQuestions: history,
	})
	if err == nil {
		t.Fatal("AnswerQuestion returned nil error for unknown question")
	}
	if !strings.Contains(err.Error(), "question not found") {
		t.Errorf("AnswerQuestion error = %q; want question not found error", err.Error())
	}
}

// TestAnswerQuestionRejectsDuplicateHistory は回答履歴に同じ質問が2回含まれる場合にエラーになることを確認する。
func TestAnswerQuestionRejectsDuplicateHistory(t *testing.T) {
	history := []model.AnsweredQuestion{
		serviceAnsweredQuestion(serviceQuestionAID, serviceQuestionAText, model.AnswerYes),
		serviceAnsweredQuestion(serviceQuestionAID, serviceQuestionAText, model.AnswerNo),
	}
	service := NewService(serviceTestGameData())

	_, err := service.AnswerQuestion(model.GameRequest{
		Meta:              serviceTestMeta(),
		AnsweredQuestions: history,
	})
	if err == nil {
		t.Fatal("AnswerQuestion returned nil error for duplicate history")
	}
	if !strings.Contains(err.Error(), "question already answered") {
		t.Errorf("AnswerQuestion error = %q; want duplicate question error", err.Error())
	}
}

// TestAnswerQuestionRejectsInvalidCandidateCount は候補件数が0以下の場合にエラーになることを確認する。
func TestAnswerQuestionRejectsInvalidCandidateCount(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{name: "zero", count: 0},
		{name: "negative", count: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := serviceTestMeta()
			meta.TopCandidatesCount = tt.count
			service := NewService(serviceTestGameData())

			_, err := service.AnswerQuestion(model.GameRequest{Meta: meta})
			if err == nil {
				t.Fatalf("AnswerQuestion returned nil error for candidate count: %d", tt.count)
			}
			if !strings.Contains(err.Error(), "topCandidatesCount must be greater than 0") {
				t.Errorf("AnswerQuestion error = %q; want candidate count error", err.Error())
			}
		})
	}
}

// TestAnswerQuestionReturnsErrorWithoutCandidates は推理対象カードがない場合にエラーになることを確認する。
func TestAnswerQuestionReturnsErrorWithoutCandidates(t *testing.T) {
	data := serviceTestGameData()
	data.Cards = nil
	service := NewService(data)

	_, err := service.AnswerQuestion(model.GameRequest{Meta: serviceTestMeta()})
	if err == nil {
		t.Fatal("AnswerQuestion returned nil error without candidates")
	}
	if err.Error() != "no candidates found" {
		t.Errorf("AnswerQuestion error = %q; want %q", err.Error(), "no candidates found")
	}
}

// TestAnswerQuestionReturnsErrorWithoutNextQuestion はしきい値未満でも未回答質問が残っていない場合にエラーになることを確認する。
func TestAnswerQuestionReturnsErrorWithoutNextQuestion(t *testing.T) {
	data := serviceTestGameData()
	data.Questions = data.Questions[:1]
	delete(data.QuestionByID, serviceQuestionBID)
	delete(data.Answers, serviceQuestionBID)
	service := NewService(data)
	history := []model.AnsweredQuestion{
		serviceAnsweredQuestion(serviceQuestionAID, serviceQuestionAText, model.AnswerUnknown),
	}

	_, err := service.AnswerQuestion(model.GameRequest{
		Meta:              serviceTestMeta(),
		AnsweredQuestions: history,
	})
	if err == nil {
		t.Fatal("AnswerQuestion returned nil error without next question")
	}
	if err.Error() != "next question not found" {
		t.Errorf("AnswerQuestion error = %q; want %q", err.Error(), "next question not found")
	}
}
