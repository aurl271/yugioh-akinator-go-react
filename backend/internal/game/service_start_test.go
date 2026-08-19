package game

import "testing"

// TestStartGameReturnsAvailableQuestion はゲーム開始時に出題可能な質問と初期状態のレスポンスを返すことを確認する。
func TestStartGameReturnsAvailableQuestion(t *testing.T) {
	meta := serviceTestMeta()
	data := serviceTestGameData()
	data.Questions = data.Questions[:1]
	service := NewService(data)

	response, err := service.StartGame(meta)
	if err != nil {
		t.Fatalf("StartGame returned error: %v", err)
	}
	if response.Meta != meta {
		t.Errorf("StartGame Meta = %+v; want %+v", response.Meta, meta)
	}
	if response.Question == nil {
		t.Fatal("StartGame Question = nil; want first question")
	}
	if response.Question.ID != serviceQuestionAID {
		t.Errorf("StartGame questionID = %d; want %d", response.Question.ID, serviceQuestionAID)
	}
	if response.Question.Text != serviceQuestionAText {
		t.Errorf("StartGame question text = %q; want %q", response.Question.Text, serviceQuestionAText)
	}
	if response.History == nil || len(response.History) != 0 {
		t.Errorf("StartGame History = %#v; want non-nil empty slice", response.History)
	}
	if response.Candidates == nil || len(response.Candidates) != 0 {
		t.Errorf("StartGame Candidates = %#v; want non-nil empty slice", response.Candidates)
	}
	if response.IsAnswer {
		t.Error("StartGame IsAnswer = true; want false")
	}
	if response.IsCorrect != nil {
		t.Errorf("StartGame IsCorrect = %v; want nil", response.IsCorrect)
	}
	if response.AnswerCard != nil {
		t.Errorf("StartGame AnswerCard = %+v; want nil", response.AnswerCard)
	}
}

// TestStartGameReturnsErrorWithoutQuestions は出題候補がない場合にゲーム開始がエラーになることを確認する。
func TestStartGameReturnsErrorWithoutQuestions(t *testing.T) {
	data := serviceTestGameData()
	data.Questions = nil
	service := NewService(data)

	response, err := service.StartGame(serviceTestMeta())
	if err == nil {
		t.Fatal("StartGame returned nil error without questions")
	}
	if err.Error() != "next question not found" {
		t.Errorf("StartGame error = %q; want %q", err.Error(), "next question not found")
	}
	if response.Question != nil {
		t.Errorf("StartGame Question = %+v; want nil", response.Question)
	}
}
