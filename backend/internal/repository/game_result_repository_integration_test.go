//go:build integration

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"yugioh-akinator-backend/internal/model"
)

// testSaveGameResultToPostgres はConfirmAnswerRequestがgame_resultsへ全項目保存されることを確認する。
func testSaveGameResultToPostgres(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	history := []model.AnsweredQuestion{
		{
			Question: model.QuestionContent{ID: integrationQuestionAID, Text: "質問A"},
			Answer:   model.AnswerYes,
		},
	}
	request := model.ConfirmAnswerRequest{
		Meta: model.Hyperparameters{
			Beta:               1.0,
			AnswerThreshold:    0.9,
			TopCandidatesCount: 10,
		},
		AnsweredQuestions: history,
		AnswerCard: model.CardData{
			CardID:   integrationCardAID,
			CardName: "カードA",
		},
		IsCorrect: true,
	}
	repository := NewGameResultRepository(db)

	if err := repository.SaveGameResult(ctx, request); err != nil {
		t.Fatalf("SaveGameResult returned error: %v", err)
	}

	var cardID int64
	var cardName string
	var isCorrect bool
	var historyJSON string
	var beta float64
	var answerThreshold float64
	var topCandidatesCount int
	err := db.QueryRowContext(ctx, `
		SELECT
			answer_card_id,
			answer_card_name,
			is_correct,
			answered_questions::text,
			beta,
			answer_threshold,
			top_candidates_count
		FROM game_results
	`).Scan(
		&cardID,
		&cardName,
		&isCorrect,
		&historyJSON,
		&beta,
		&answerThreshold,
		&topCandidatesCount,
	)
	if err != nil {
		t.Fatalf("read saved game result: %v", err)
	}

	if cardID != request.AnswerCard.CardID || cardName != request.AnswerCard.CardName {
		t.Errorf("saved answer card = %d %q; want %+v", cardID, cardName, request.AnswerCard)
	}
	if isCorrect != request.IsCorrect {
		t.Errorf("saved isCorrect = %t; want %t", isCorrect, request.IsCorrect)
	}
	if beta != request.Meta.Beta || answerThreshold != request.Meta.AnswerThreshold || topCandidatesCount != request.Meta.TopCandidatesCount {
		t.Errorf(
			"saved meta = beta %f threshold %f count %d; want %+v",
			beta,
			answerThreshold,
			topCandidatesCount,
			request.Meta,
		)
	}

	var savedHistory []model.AnsweredQuestion
	if err := json.Unmarshal([]byte(historyJSON), &savedHistory); err != nil {
		t.Fatalf("decode saved answered_questions: %v", err)
	}
	if !reflect.DeepEqual(savedHistory, history) {
		t.Errorf("saved history = %+v; want %+v", savedHistory, history)
	}
}
