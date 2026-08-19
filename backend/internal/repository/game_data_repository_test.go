package repository

import (
	"testing"
	"yugioh-akinator-backend/internal/game"
)

const (
	testScriptQuestionID = 1
	testCardsQuestionID  = 2
	testMatchingCardID   = 1001
	testOtherCardID      = 1002
)

// TestBuildAnswersUsesRawAnswers はDBから読み込んだ回答が質問IDとカードIDに対応して格納されることを確認する。
func TestBuildAnswersUsesRawAnswers(t *testing.T) {
	questions := []game.Question{
		{ID: testScriptQuestionID, Category: QuestionCategoryScript},
	}
	rawAnswers := []RawAnswer{
		{
			QuestionID: testScriptQuestionID,
			CardID:     testMatchingCardID,
			Answer:     game.ExpectedAnswerYes,
		},
	}

	answers, err := buildAnswers(questions, nil, rawAnswers)
	if err != nil {
		t.Fatalf("buildAnswers returned error: %v", err)
	}

	questionAnswers, ok := answers[testScriptQuestionID]
	if !ok {
		t.Fatalf("answers does not contain questionID: %d", testScriptQuestionID)
	}
	got, ok := questionAnswers[testMatchingCardID]
	if !ok {
		t.Fatalf("question answers does not contain cardID: %d", testMatchingCardID)
	}
	if got != game.ExpectedAnswerYes {
		t.Errorf("answer = %d; want %d", got, game.ExpectedAnswerYes)
	}
}

// TestBuildAnswersGeneratesAnswersFromConditions はcards由来質問の条件に一致するカードだけがYESになることを確認する。
func TestBuildAnswersGeneratesAnswersFromConditions(t *testing.T) {
	condition := game.Condition{
		Logic: "and",
		Conditions: []game.ConditionItem{
			{Field: CardFieldAtk, Op: ConditionOpBetween, Min: 2000, Max: 3000},
			{Field: CardFieldReading, Op: ConditionOpStartsWith, Text: "ブルーアイズ"},
		},
	}
	questions := []game.Question{
		{
			ID:        testCardsQuestionID,
			Text:      "条件に一致しますか？",
			Category:  QuestionCategoryCards,
			Condition: &condition,
		},
	}
	matchingCard := testConditionCard()
	matchingCard.CardID = testMatchingCardID
	otherCard := testConditionCard()
	otherCard.CardID = testOtherCardID
	otherCard.Atk = 1000

	answers, err := buildAnswers(questions, []game.Card{matchingCard, otherCard}, nil)
	if err != nil {
		t.Fatalf("buildAnswers returned error: %v", err)
	}

	questionAnswers := answers[testCardsQuestionID]
	got, ok := questionAnswers[testMatchingCardID]
	if !ok {
		t.Fatalf("question answers does not contain matching cardID: %d", testMatchingCardID)
	}
	if got != game.ExpectedAnswerYes {
		t.Errorf("matching card answer = %d; want %d", got, game.ExpectedAnswerYes)
	}
	if _, ok := questionAnswers[testOtherCardID]; ok {
		t.Errorf("question answers contains non-matching cardID: %d", testOtherCardID)
	}
}

// TestBuildAnswersRejectsCardsQuestionWithoutCondition はcards由来質問にConditionがない場合にエラーになることを確認する。
func TestBuildAnswersRejectsCardsQuestionWithoutCondition(t *testing.T) {
	questions := []game.Question{
		{
			ID:       testCardsQuestionID,
			Text:     "条件なしの質問",
			Category: QuestionCategoryCards,
		},
	}

	answers, err := buildAnswers(questions, []game.Card{testConditionCard()}, nil)
	if err == nil {
		t.Fatal("buildAnswers returned nil error for cards question without condition")
	}
	if answers != nil {
		t.Errorf("buildAnswers answers = %v; want nil", answers)
	}
}

// TestBuildAnswersReturnsConditionError はConditionのlogicや演算子が不正な場合にエラーが返ることを確認する。
func TestBuildAnswersReturnsConditionError(t *testing.T) {
	condition := game.Condition{Logic: "unknown"}
	questions := []game.Question{
		{
			ID:        testCardsQuestionID,
			Category:  QuestionCategoryCards,
			Condition: &condition,
		},
	}

	answers, err := buildAnswers(questions, []game.Card{testConditionCard()}, nil)
	if err == nil {
		t.Fatal("buildAnswers returned nil error for invalid condition")
	}
	if answers != nil {
		t.Errorf("buildAnswers answers = %v; want nil", answers)
	}
}

// TestBuildQuestionByID は質問一覧が質問IDをkeyとするmapへ変換されることを確認する。
func TestBuildQuestionByID(t *testing.T) {
	questions := []game.Question{
		{ID: testScriptQuestionID, Text: "質問A"},
		{ID: testCardsQuestionID, Text: "質問B"},
	}

	questionByID := buildQuestionByID(questions)
	if len(questionByID) != len(questions) {
		t.Fatalf("buildQuestionByID length = %d; want %d", len(questionByID), len(questions))
	}
	for _, question := range questions {
		got, ok := questionByID[question.ID]
		if !ok {
			t.Errorf("buildQuestionByID does not contain questionID: %d", question.ID)
			continue
		}
		if got != question {
			t.Errorf("buildQuestionByID[%d] = %+v; want %+v", question.ID, got, question)
		}
	}
}
