package game

import (
	"fmt"
	"math"
	"testing"
)

// TestApplyAnswerAcceptsValidAnswers は5種類の有効な回答をApplyAnswerが受け付けることを確認する。
func TestApplyAnswerAcceptsValidAnswers(t *testing.T) {
	tests := []struct {
		name   string
		answer float64
	}{
		{name: "accepts yes answer", answer: AnswerScoreYes},
		{name: "accepts probably answer", answer: AnswerScoreProbably},
		{name: "accepts unknown answer", answer: AnswerScoreUnknown},
		{name: "accepts probably no answer", answer: AnswerScoreProbablyNo},
		{name: "accepts no answer", answer: AnswerScoreNo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(testGameData(), 1.0)

			err := engine.ApplyAnswer(testQuestionAID, tt.answer)
			if err != nil {
				t.Fatalf("ApplyAnswer returned error: %v", err)
			}
		})
	}
}

// TestApplyAnswerRejectsUnknownQuestion は登録されていない質問IDへの回答がエラーになることを確認する。
func TestApplyAnswerRejectsUnknownQuestion(t *testing.T) {
	const unknownQuestionID = 999
	engine := NewEngine(testGameData(), 1.0)

	err := engine.ApplyAnswer(unknownQuestionID, AnswerScoreYes)
	if err == nil {
		t.Fatalf("ApplyAnswer returned nil error for unknown questionID: %d", unknownQuestionID)
	}

	want := fmt.Sprintf("question not found: %d", unknownQuestionID)
	if err.Error() != want {
		t.Errorf("ApplyAnswer error = %q; want %q", err.Error(), want)
	}
}

// TestApplyAnswerRejectsDuplicateQuestion は同じ質問への2回目の回答がエラーになることを確認する。
func TestApplyAnswerRejectsDuplicateQuestion(t *testing.T) {
	engine := NewEngine(testGameData(), 1.0)

	firstErr := engine.ApplyAnswer(testQuestionAID, AnswerScoreYes)
	if firstErr != nil {
		t.Fatalf("first ApplyAnswer returned error: %v", firstErr)
	}

	secondErr := engine.ApplyAnswer(testQuestionAID, AnswerScoreYes)
	if secondErr == nil {
		t.Fatalf(
			"second ApplyAnswer returned nil error for duplicate questionID: %d",
			testQuestionAID,
		)
	}

	want := fmt.Sprintf("question already answered: %d", testQuestionAID)
	if secondErr.Error() != want {
		t.Errorf("second ApplyAnswer error = %q; want %q", secondErr.Error(), want)
	}
}

// TestApplyAnswerRejectsInvalidAnswer は定義されていない回答値がエラーになることを確認する。
func TestApplyAnswerRejectsInvalidAnswer(t *testing.T) {
	const invalidAnswer = 2.0
	engine := NewEngine(testGameData(), 1.0)

	err := engine.ApplyAnswer(testQuestionAID, invalidAnswer)
	if err == nil {
		t.Fatalf("ApplyAnswer returned nil error for invalid answer: %f", invalidAnswer)
	}

	want := fmt.Sprintf("invalid answer: %f", invalidAnswer)
	if err.Error() != want {
		t.Errorf("ApplyAnswer error = %q; want %q", err.Error(), want)
	}
}

// TestApplyAnswerRejectsQuestionWithoutAnswers は質問に対応する期待回答データがない場合にエラーになることを確認する。
func TestApplyAnswerRejectsQuestionWithoutAnswers(t *testing.T) {
	data := testGameData()
	delete(data.Answers, testQuestionAID)
	engine := NewEngine(data, 1.0)

	err := engine.ApplyAnswer(testQuestionAID, AnswerScoreYes)
	if err == nil {
		t.Fatalf("ApplyAnswer returned nil error for questionID without answers: %d", testQuestionAID)
	}

	want := fmt.Sprintf("answers not found for question: %d", testQuestionAID)
	if err.Error() != want {
		t.Errorf("ApplyAnswer error = %q; want %q", err.Error(), want)
	}
}

// TestApplyAnswerUpdatesScores は5種類の回答に応じてYESカードとNOカードのscoreが正しく更新されることを確認する。
func TestApplyAnswerUpdatesScores(t *testing.T) {
	const tolerance = 1e-9
	tests := []struct {
		name       string
		answer     float64
		wantScores []float64
	}{
		{name: "yes", answer: AnswerScoreYes, wantScores: []float64{0.0, 4.0}},
		{name: "probably", answer: AnswerScoreProbably, wantScores: []float64{0.25, 2.25}},
		{name: "unknown", answer: AnswerScoreUnknown, wantScores: []float64{1.0, 1.0}},
		{name: "probably no", answer: AnswerScoreProbablyNo, wantScores: []float64{2.25, 0.25}},
		{name: "no", answer: AnswerScoreNo, wantScores: []float64{4.0, 0.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(testGameData(), 1.0)

			if err := engine.ApplyAnswer(testQuestionAID, tt.answer); err != nil {
				t.Fatalf("ApplyAnswer returned error: %v", err)
			}

			if len(engine.scores) != len(tt.wantScores) {
				t.Fatalf("scores length = %d; want %d", len(engine.scores), len(tt.wantScores))
			}
			for i, want := range tt.wantScores {
				if got := engine.scores[i]; math.Abs(got-want) > tolerance {
					t.Errorf("scores[%d] = %f; want %f", i, got, want)
				}
			}
		})
	}
}

// TestApplyAnswerRecordsAnsweredQuestion は適用した質問IDと回答値が履歴と回答済みIDに記録されることを確認する。
func TestApplyAnswerRecordsAnsweredQuestion(t *testing.T) {
	engine := NewEngine(testGameData(), 1.0)

	if err := engine.ApplyAnswer(testQuestionAID, AnswerScoreProbably); err != nil {
		t.Fatalf("ApplyAnswer returned error: %v", err)
	}

	if len(engine.answeredQuestions) != 1 {
		t.Fatalf("answeredQuestions length = %d; want 1", len(engine.answeredQuestions))
	}

	got := engine.answeredQuestions[0]
	if got.QuestionID != testQuestionAID {
		t.Errorf("answered questionID = %d; want %d", got.QuestionID, testQuestionAID)
	}
	if got.Answer != AnswerScoreProbably {
		t.Errorf("answered value = %f; want %f", got.Answer, AnswerScoreProbably)
	}
	if !engine.answeredQuestionIDs[testQuestionAID] {
		t.Errorf("answeredQuestionIDs[%d] = false; want true", testQuestionAID)
	}
}

// TestApplyAnswerUpdatesStateForPositiveAnswers はYESとProbablyで質問のNewStateがstateへ追加されることを確認する。
func TestApplyAnswerUpdatesStateForPositiveAnswers(t *testing.T) {
	tests := []struct {
		name   string
		answer float64
	}{
		{name: "yes", answer: AnswerScoreYes},
		{name: "probably", answer: AnswerScoreProbably},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(testGameData(), 1.0)

			if err := engine.ApplyAnswer(testQuestionAID, tt.answer); err != nil {
				t.Fatalf("ApplyAnswer returned error: %v", err)
			}

			if engine.state != testStateBit {
				t.Errorf("state = %d; want %d", engine.state, testStateBit)
			}
		})
	}
}

// TestApplyAnswerDoesNotUpdateStateForOtherAnswers はUnknown、ProbablyNo、NOではstateが更新されないことを確認する。
func TestApplyAnswerDoesNotUpdateStateForOtherAnswers(t *testing.T) {
	tests := []struct {
		name   string
		answer float64
	}{
		{name: "unknown", answer: AnswerScoreUnknown},
		{name: "probably no", answer: AnswerScoreProbablyNo},
		{name: "no", answer: AnswerScoreNo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(testGameData(), 1.0)

			if err := engine.ApplyAnswer(testQuestionAID, tt.answer); err != nil {
				t.Fatalf("ApplyAnswer returned error: %v", err)
			}

			if engine.state != 0 {
				t.Errorf("state = %d; want 0", engine.state)
			}
		})
	}
}
