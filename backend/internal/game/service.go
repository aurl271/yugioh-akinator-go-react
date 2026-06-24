package game

import (
	"fmt"
	"log"
	"time"
	"yugioh-akinator-backend/internal/model"
)

// Service はAPI層とEngineの間に置くアプリケーション層。
// HTTPやJSONの詳細を知らず、ゲーム開始・回答・確認のユースケースだけを扱う。
type Service struct {
	// data はサーバー起動時にDBから読み込んだ固定データ。
	data GameData
}

// NewService はGameDataを受け取り、ゲーム操作用Serviceを作る。
func NewService(data GameData) *Service {
	return &Service{
		data: data,
	}
}

// StartGame は新しいゲームを開始し、最初の質問を返す。
func (s *Service) StartGame(meta model.Hyperparameters) (model.GameResponse, error) {
	start := time.Now()
	engine := NewEngine(s.data, meta.Beta)

	nextQuestionStart := time.Now()
	question, ok := engine.NextQuestion()
	nextQuestionElapsed := time.Since(nextQuestionStart)
	if !ok {
		return model.GameResponse{}, fmt.Errorf("next question not found")
	}
	log.Printf(
		"service start game: nextQuestion=%s total=%s questionID=%d",
		nextQuestionElapsed,
		time.Since(start),
		question.ID,
	)

	return model.GameResponse{
		Meta: meta,
		Question: &model.QuestionContent{
			ID:   question.ID,
			Text: question.Text,
		},
		History:    []model.AnsweredQuestion{},
		Candidates: []model.CandidateResponse{},
		IsAnswer:   false,
		IsCorrect:  nil,
		AnswerCard: nil,
	}, nil
}

// AnswerQuestion は回答履歴を受け取り、次の質問または最終回答候補を返す。
// backendは状態を保存しないため、履歴をEngineへ毎回再適用する。
func (s *Service) AnswerQuestion(req model.GameRequest) (model.GameResponse, error) {
	start := time.Now()
	engine := NewEngine(s.data, req.Meta.Beta)

	applyAnswersStart := time.Now()
	for _, answeredQuestion := range req.AnsweredQuestions {
		answerValue, err := answerValueToScore(answeredQuestion.Answer)
		if err != nil {
			return model.GameResponse{}, fmt.Errorf("invalid answer value: %w", err)
		}
		if err := engine.ApplyAnswer(answeredQuestion.Question.ID, answerValue); err != nil {
			return model.GameResponse{}, fmt.Errorf("apply answer: %w", err)
		}
	}
	applyAnswersElapsed := time.Since(applyAnswersStart)

	if req.Meta.TopCandidatesCount <= 0 {
		return model.GameResponse{}, fmt.Errorf("topCandidatesCount must be greater than 0: got %d", req.Meta.TopCandidatesCount)
	}
	topCandidatesStart := time.Now()
	gameCandidates := engine.TopCandidates(req.Meta.TopCandidatesCount)
	topCandidatesElapsed := time.Since(topCandidatesStart)
	if len(gameCandidates) <= 0 {
		return model.GameResponse{}, fmt.Errorf("no candidates found")
	}

	cardCandidates := make([]model.CandidateResponse, len(gameCandidates))
	for i, candidate := range gameCandidates {
		cardCandidates[i] = model.CandidateResponse{
			Rank:        candidate.Rank,
			CardID:      candidate.CardID,
			CardName:    candidate.CardName,
			Probability: candidate.Probability,
		}
	}

	topCard := cardCandidates[0]
	if topCard.Probability >= req.Meta.AnswerThreshold {
		log.Printf(
			"service answer question: history=%d applyAnswers=%s topCandidates=%s nextQuestion=skipped total=%s isAnswer=true topProbability=%.6f",
			len(req.AnsweredQuestions),
			applyAnswersElapsed,
			topCandidatesElapsed,
			time.Since(start),
			topCard.Probability,
		)
		return model.GameResponse{
			Meta:       req.Meta,
			Question:   nil,
			History:    req.AnsweredQuestions,
			Candidates: cardCandidates,
			IsAnswer:   true,
			IsCorrect:  nil,
			AnswerCard: &model.CardData{
				CardID:   topCard.CardID,
				CardName: topCard.CardName,
			},
		}, nil
	}

	nextQuestionStart := time.Now()
	question, ok := engine.NextQuestion()
	nextQuestionElapsed := time.Since(nextQuestionStart)
	if !ok {
		return model.GameResponse{}, fmt.Errorf("next question not found")
	}
	log.Printf(
		"service answer question: history=%d applyAnswers=%s topCandidates=%s nextQuestion=%s total=%s isAnswer=false topProbability=%.6f questionID=%d",
		len(req.AnsweredQuestions),
		applyAnswersElapsed,
		topCandidatesElapsed,
		nextQuestionElapsed,
		time.Since(start),
		topCard.Probability,
		question.ID,
	)

	return model.GameResponse{
		Meta: req.Meta,
		Question: &model.QuestionContent{
			ID:   question.ID,
			Text: question.Text,
		},
		History:    req.AnsweredQuestions,
		Candidates: cardCandidates,
		IsAnswer:   false,
		IsCorrect:  nil,
		AnswerCard: nil,
	}, nil
}

// ConfirmAnswer は提示した回答が正しかったかを受け取り、フィードバック画面用の結果を返す。
// 現時点では保存処理を持たないが、将来のログ保存や学習データ収集の入口にできる。
func (s *Service) ConfirmAnswer(req model.ConfirmAnswerRequest) (model.GameResponse, error) {
	correct := req.IsCorrect

	return model.GameResponse{
		Meta:       req.Meta,
		Question:   nil,
		History:    req.AnsweredQuestions,
		Candidates: []model.CandidateResponse{},
		IsAnswer:   true,
		IsCorrect:  &correct,
		AnswerCard: &req.AnswerCard,
	}, nil
}

// answerValueToScore はAPIで受け取った文字列回答をEngine用の数値回答へ変換する。
func answerValueToScore(answer model.AnswerValue) (float64, error) {
	switch answer {
	case model.AnswerYes:
		return AnswerScoreYes, nil
	case model.AnswerProbably:
		return AnswerScoreProbably, nil
	case model.AnswerUnknown:
		return AnswerScoreUnknown, nil
	case model.AnswerProbablyNo:
		return AnswerScoreProbablyNo, nil
	case model.AnswerNo:
		return AnswerScoreNo, nil
	default:
		return 0, fmt.Errorf("invalid answer: %s", answer)
	}
}
