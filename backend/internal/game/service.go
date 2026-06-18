package game

import (
	"fmt"
	"yugioh-akinator-backend/internal/model"
)

type Service struct {
	data GameData
}

func NewService(data GameData) *Service {
	return &Service{
		data: data,
	}
}

func (s *Service) StartGame(meta model.Hyperparameters) (model.GameResponse, error) {
	engine := NewEngine(s.data, meta.Beta)

	question, ok := engine.NextQuestion()
	if !ok {
		return model.GameResponse{}, fmt.Errorf("next question not found")
	}

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

func (s *Service) AnswerQuestion(req model.GameRequest) (model.GameResponse, error) {
	engine := NewEngine(s.data, req.Meta.Beta)

	for _, answeredQuestion := range req.AnsweredQuestions {
		answerValue, err := answerValueToScore(answeredQuestion.Answer)
		if err != nil {
			return model.GameResponse{}, fmt.Errorf("invalid answer value: %w", err)
		}
		if err := engine.ApplyAnswer(answeredQuestion.Question.ID, answerValue); err != nil {
			return model.GameResponse{}, fmt.Errorf("apply answer: %w", err)
		}
	}

	if req.Meta.TopCandidatesCount <= 0 {
		return model.GameResponse{}, fmt.Errorf("topCandidatesCount must be greater than 0: got %d", req.Meta.TopCandidatesCount)
	}
	gameCandidates := engine.TopCandidates(req.Meta.TopCandidatesCount)
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

	question, ok := engine.NextQuestion()
	if !ok {
		return model.GameResponse{}, fmt.Errorf("next question not found")
	}

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
