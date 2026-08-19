//go:build integration

package repository

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"yugioh-akinator-backend/internal/game"
)

// testLoadCardsFromPostgres はcardsテーブルの各列とNULL文字列がGameData用Cardへ変換されることを確認する。
func testLoadCardsFromPostgres(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	insertIntegrationGameData(t, ctx, db)

	cards, err := LoadCards(ctx, db)
	if err != nil {
		t.Fatalf("LoadCards returned error: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("LoadCards length = %d; want 2", len(cards))
	}

	cardA := cards[0]
	if cardA.ID != 1 || cardA.CardID != integrationCardAID || cardA.Name != "カードA" {
		t.Errorf("LoadCards first card = %+v; want card A", cardA)
	}
	if cardA.Reading != "カードエー" || cardA.Desc != "このカードは特殊召喚できる。" {
		t.Errorf("LoadCards first card strings = reading %q desc %q", cardA.Reading, cardA.Desc)
	}
	if cardA.Setcode != 52 || cardA.Type != 10 || cardA.Atk != 2500 || cardA.Def != 2100 || cardA.Level != 8 {
		t.Errorf("LoadCards first card numeric fields = %+v; want fixture values", cardA)
	}

	cardB := cards[1]
	if cardB.CardID != integrationCardBID {
		t.Errorf("LoadCards second cardID = %d; want %d", cardB.CardID, integrationCardBID)
	}
	if cardB.Reading != "" || cardB.Desc != "" {
		t.Errorf("LoadCards NULL strings = reading %q desc %q; want empty strings", cardB.Reading, cardB.Desc)
	}
}

// testLoadQuestionsFromPostgres はquestionsテーブルのNULL値とcondition_jsonがQuestionへ変換されることを確認する。
func testLoadQuestionsFromPostgres(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	insertIntegrationGameData(t, ctx, db)

	questions, err := LoadQuestions(ctx, db)
	if err != nil {
		t.Fatalf("LoadQuestions returned error: %v", err)
	}
	if len(questions) != 2 {
		t.Fatalf("LoadQuestions length = %d; want 2", len(questions))
	}

	questionA := questions[0]
	if questionA.ID != integrationQuestionAID || questionA.Category != QuestionCategoryScript {
		t.Errorf("LoadQuestions first question = %+v; want script question", questionA)
	}
	if questionA.Query != "legacy query" || questionA.Condition != nil {
		t.Errorf("LoadQuestions first question query/condition = %+v", questionA)
	}

	questionB := questions[1]
	if questionB.ID != integrationQuestionBID || questionB.Category != QuestionCategoryCards {
		t.Errorf("LoadQuestions second question = %+v; want cards question", questionB)
	}
	if questionB.Query != "" || questionB.Condition == nil {
		t.Fatalf("LoadQuestions second question query/condition = %+v", questionB)
	}
	if questionB.Condition.Logic != "and" || len(questionB.Condition.Conditions) != 1 {
		t.Errorf("LoadQuestions parsed condition = %+v; want one and condition", questionB.Condition)
	}
	condition := questionB.Condition.Conditions[0]
	if condition.Field != CardFieldAtk || condition.Op != ConditionOpBetween || condition.Min != 2000 || condition.Max != 3000 {
		t.Errorf("LoadQuestions condition item = %+v; want atk between 2000 and 3000", condition)
	}
	if questionB.UnsetBit != 4 || questionB.NewState != 8 {
		t.Errorf("LoadQuestions state fields = unset %d new %d; want 4 and 8", questionB.UnsetBit, questionB.NewState)
	}
}

// testLoadQuestionsRejectsInvalidConditionJSON は壊れたcondition_jsonを含む行でLoadQuestionsがエラーになることを確認する。
func testLoadQuestionsRejectsInvalidConditionJSON(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO questions (id, question_text, category, condition_json)
		VALUES (1, '不正な条件の質問', 1, '{')
	`)
	if err != nil {
		t.Fatalf("insert invalid condition question: %v", err)
	}

	questions, err := LoadQuestions(ctx, db)
	if err == nil {
		t.Fatal("LoadQuestions returned nil error for invalid condition_json")
	}
	if questions != nil && len(questions) != 0 {
		t.Errorf("LoadQuestions questions = %+v; want empty result", questions)
	}
	if !strings.Contains(err.Error(), "parse condition_json question_id=1") {
		t.Errorf("LoadQuestions error = %q; want condition_json parse error", err.Error())
	}
}

// testLoadAnswersFromPostgres はanswersテーブルの質問ID、カードID、回答値をRawAnswerとして読み込めることを確認する。
func testLoadAnswersFromPostgres(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	insertIntegrationGameData(t, ctx, db)

	answers, err := LoadAnswers(ctx, db)
	if err != nil {
		t.Fatalf("LoadAnswers returned error: %v", err)
	}
	if len(answers) != 1 {
		t.Fatalf("LoadAnswers length = %d; want 1", len(answers))
	}
	want := RawAnswer{
		QuestionID: integrationQuestionAID,
		CardID:     integrationCardAID,
		Answer:     integrationExpectedYes,
	}
	if answers[0] != want {
		t.Errorf("LoadAnswers first answer = %+v; want %+v", answers[0], want)
	}
}

// testLoadGameDataFromPostgres はDBの全テーブルからEngine用GameDataとcondition由来回答を構築できることを確認する。
func testLoadGameDataFromPostgres(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	insertIntegrationGameData(t, ctx, db)

	data, err := LoadGameData(ctx, db)
	if err != nil {
		t.Fatalf("LoadGameData returned error: %v", err)
	}
	if len(data.Cards) != 2 || len(data.Questions) != 2 || len(data.QuestionByID) != 2 {
		t.Fatalf(
			"LoadGameData sizes = cards %d questions %d questionByID %d; want 2 each",
			len(data.Cards),
			len(data.Questions),
			len(data.QuestionByID),
		)
	}
	if data.QuestionByID[integrationQuestionBID].Text != "攻撃力が2000以上3000以下ですか？" {
		t.Errorf("LoadGameData questionByID = %+v", data.QuestionByID[integrationQuestionBID])
	}
	if got := data.Answers[integrationQuestionAID][integrationCardAID]; got != game.ExpectedAnswerYes {
		t.Errorf("LoadGameData raw answer = %d; want %d", got, game.ExpectedAnswerYes)
	}
	if got := data.Answers[integrationQuestionBID][integrationCardAID]; got != game.ExpectedAnswerYes {
		t.Errorf("LoadGameData generated answer = %d; want %d", got, game.ExpectedAnswerYes)
	}
	if _, ok := data.Answers[integrationQuestionBID][integrationCardBID]; ok {
		t.Errorf("LoadGameData generated answers contain non-matching cardID: %d", integrationCardBID)
	}
}
