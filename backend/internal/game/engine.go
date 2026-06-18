package game

import (
	"fmt"
	"math"
	"slices"
)

type Engine struct {
	data                GameData
	beta                float64
	scores              []float64
	state               int
	answeredQuestions   []AnsweredQuestion
	answeredQuestionIDs map[int]bool
}

func NewEngine(data GameData, beta float64) *Engine {
	return &Engine{
		data:                data,
		beta:                beta,
		scores:              make([]float64, len(data.Cards)),
		state:               0,
		answeredQuestions:   []AnsweredQuestion{},
		answeredQuestionIDs: map[int]bool{},
	}
}

func (e *Engine) ApplyAnswer(questionID int, answer float64) error {
	question, ok := e.data.QuestionByID[questionID]
	if !ok {
		return fmt.Errorf("question not found: %d", questionID)
	}

	if e.answeredQuestionIDs[questionID] {
		return fmt.Errorf("question already answered: %d", questionID)
	}

	if !isValidAnswerScore(answer) {
		return fmt.Errorf("invalid answer: %f", answer)
	}

	if _, ok := e.data.Answers[questionID]; !ok {
		return fmt.Errorf("answers not found for question: %d", questionID)
	}

	e.answeredQuestions = append(e.answeredQuestions, AnsweredQuestion{
		QuestionID: questionID,
		Answer:     answer,
	})
	e.answeredQuestionIDs[questionID] = true

	e.updateScores(questionID, answer)
	e.updateState(question, answer)

	return nil
}

func (e *Engine) updateScores(questionID int, answer float64) {
	questionAnswers := e.data.Answers[questionID]

	for i, card := range e.data.Cards {
		expected := -1.0

		if value, ok := questionAnswers[card.CardID]; ok {
			expected = float64(value)
		}

		diff := answer - expected
		e.scores[i] += diff * diff
	}
}

func (e *Engine) updateState(question Question, answer float64) {
	if answer != AnswerScoreYes && answer != AnswerScoreProbably {
		return
	}

	e.state |= question.NewState
}

func (e *Engine) Probabilities() []float64 {
	if len(e.scores) == 0 {
		return []float64{}
	}

	percent := make([]float64, len(e.scores))
	sum := 0.0
	for _, v := range e.scores {
		sum += math.Exp(-e.beta * v)
	}

	if sum == 0 {
		return []float64{}
	}

	for i, v := range e.scores {
		percent[i] = math.Exp(-e.beta*v) / sum
	}

	return percent
}

func (e *Engine) TopCandidates(limit int) []Candidate {

	if limit <= 0 {
		return []Candidate{}
	}

	probabilities := e.Probabilities()

	candidates := make([]Candidate, 0, len(probabilities))
	for i, probability := range probabilities {
		card := e.data.Cards[i]
		candidates = append(candidates, Candidate{
			Rank:        0,
			CardID:      card.CardID,
			CardName:    card.Name,
			Probability: probability,
		})
	}

	slices.SortFunc(candidates, func(a, b Candidate) int {
		switch {
		case a.Probability > b.Probability:
			return -1
		case a.Probability < b.Probability:
			return 1
		default:
			return 0
		}
	})

	if limit > len(candidates) {
		limit = len(candidates)
	}

	candidates = candidates[:limit]

	for i := range candidates {
		candidates[i].Rank = i + 1
	}

	return candidates
}

func (e *Engine) NextQuestion() (Question, bool) {

	bestScore := math.Inf(-1)
	bestQuestion := Question{}
	found := false

	for _, question := range e.data.Questions {

		if e.answeredQuestionIDs[question.ID] {
			continue
		}

		if e.state&question.UnsetBit != 0 {
			continue
		}

		probabilities := make([]float64, len(AnswerScores))
		for j, answerScore := range AnswerScores {
			for cardIndex, card := range e.data.Cards {
				expected := -1.0

				if value, ok := e.data.Answers[question.ID][card.CardID]; ok {
					expected = float64(value)
				}

				diff := answerScore - expected
				probabilities[j] += math.Exp(-e.beta*diff*diff - e.beta*e.scores[cardIndex])
			}
		}

		sum := 0.0
		for _, p := range probabilities {
			sum += p
		}
		if sum == 0 {
			continue
		}
		for i := range probabilities {
			probabilities[i] /= sum
		}

		entropy := shannon_entropy(probabilities)
		if entropy > bestScore {
			bestScore = entropy
			bestQuestion = question
			found = true
		}
	}
	return bestQuestion, found
}
