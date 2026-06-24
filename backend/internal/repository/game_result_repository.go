package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"yugioh-akinator-backend/internal/model"
)

// GameResultRepository は1プレイ分の最終結果をgame_resultsへ保存する。
type GameResultRepository struct {
	db *sql.DB
}

// NewGameResultRepository は結果保存用Repositoryを作る。
func NewGameResultRepository(db *sql.DB) *GameResultRepository {
	return &GameResultRepository{
		db: db,
	}
}

// SaveGameResult はconfirm時のリクエスト内容をプレイ結果ログとして保存する。
func (r *GameResultRepository) SaveGameResult(ctx context.Context, req model.ConfirmAnswerRequest) error {
	answeredQuestionsJSON, err := json.Marshal(req.AnsweredQuestions)
	if err != nil {
		return fmt.Errorf("marshal answered questions: %w", err)
	}

	_, err = r.db.ExecContext(
		ctx,
		`
			INSERT INTO game_results (
				answer_card_id,
				answer_card_name,
				is_correct,
				answered_questions,
				beta,
				answer_threshold,
				top_candidates_count
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		req.AnswerCard.CardID,
		req.AnswerCard.CardName,
		req.IsCorrect,
		string(answeredQuestionsJSON),
		req.Meta.Beta,
		req.Meta.AnswerThreshold,
		req.Meta.TopCandidatesCount,
	)
	if err != nil {
		return fmt.Errorf("insert game result: %w", err)
	}

	return nil
}
