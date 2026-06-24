package game

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"slices"
	"time"
)

// nextQuestionCardLimit は次質問を評価するときに見るカード数。
// 全カードで評価すると重いため、精度と速度のバランスとして500枚に絞っている。
const nextQuestionCardLimit = 500

// Engine は1ゲーム分の推理状態を持つ。
// backendは状態を保存しないため、リクエストごとに回答履歴からEngineを作り直す。
type Engine struct {
	// data はDBから読み込んだ固定データ。
	data                GameData
	// beta はscoreを確率へ変換するときの鋭さ。大きいほど上位候補に確率が寄る。
	beta                float64
	// scores は各カードの回答とのズレの累積。小さいほど正解候補として強い。
	scores              []float64
	// state は質問の出し分けに使うビット集合。
	state               int
	// answeredQuestions はこのEngineに適用済みの回答。
	answeredQuestions   []AnsweredQuestion
	// answeredQuestionIDs は同じ質問を二重に適用しないためのset。
	answeredQuestionIDs map[int]bool
}

// NewEngine はまだ回答を適用していない初期状態のEngineを作る。
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

// ApplyAnswer は1つの回答をEngineへ反映する。
// score更新とstate更新を同時に行うため、回答履歴の再適用でも使う。
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

// updateScores は回答と各カードの期待回答の差をscoreへ加算する。
func (e *Engine) updateScores(questionID int, answer float64) {
	questionAnswers := e.data.Answers[questionID]

	for i, card := range e.data.Cards {
		// answersに無いカードは、その質問に対してNOとみなす。
		expected := -1.0

		if value, ok := questionAnswers[card.CardID]; ok {
			expected = float64(value)
		}

		diff := answer - expected
		e.scores[i] += diff * diff
	}
}

// updateState はYES寄りの回答を、以降の質問除外条件に反映する。
func (e *Engine) updateState(question Question, answer float64) {
	if answer != AnswerScoreYes && answer != AnswerScoreProbably {
		return
	}

	// stateは質問カテゴリの絞り込み用ビット。回答履歴から毎回再構築される。
	e.state |= question.NewState
}

// Probabilities は現在のscoresをカード確率へ変換する。
// 戻り値のindexはdata.Cardsと同じ順番。
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

// TopCandidates は現在もっとも正解らしいカードを確率順に返す。
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

// topScoreCardIndexes はNextQuestionで評価するカードindexを選ぶ。
// 初回はランダム、回答後はscoreの低い有力候補に絞る。
func (e *Engine) topScoreCardIndexes(limit int) []int {
	cardCount := len(e.data.Cards)
	if cardCount == 0 {
		return []int{}
	}

	indexes := make([]int, cardCount)
	for i := range indexes {
		indexes[i] = i
	}

	if limit <= 0 || cardCount <= limit {
		return indexes
	}

	if len(e.answeredQuestions) == 0 {
		// 初回は全カードのscoreが同じなので、固定の先頭500枚ではなくランダムに見る。
		rand.Shuffle(len(indexes), func(i, j int) {
			indexes[i], indexes[j] = indexes[j], indexes[i]
		})
		return indexes[:limit]
	}

	// scoreは小さいほど正解候補として強い。次質問計算では有力候補だけを見る。
	slices.SortFunc(indexes, func(a, b int) int {
		switch {
		case e.scores[a] < e.scores[b]:
			return -1
		case e.scores[a] > e.scores[b]:
			return 1
		default:
			return a - b
		}
	})

	return indexes[:limit]
}

// NextQuestion は回答が分かれやすい質問をエントロピーで選ぶ。
// boolは出題可能な質問が見つかったかどうかを表す。
func (e *Engine) NextQuestion() (Question, bool) {
	start := time.Now()

	bestScore := math.Inf(-1)
	bestQuestion := Question{}
	found := false
	skippedAnswered := 0
	skippedState := 0
	checkedQuestions := 0
	cardIndexes := e.topScoreCardIndexes(nextQuestionCardLimit)

	for _, question := range e.data.Questions {

		if e.answeredQuestionIDs[question.ID] {
			skippedAnswered++
			continue
		}

		if e.state&question.UnsetBit != 0 {
			skippedState++
			continue
		}

		checkedQuestions++
		// 各回答を選んだ場合に候補カードがどう分かれるかを見て、情報量の大きい質問を選ぶ。
		probabilities := make([]float64, len(AnswerScores))
		for j, answerScore := range AnswerScores {
			for _, cardIndex := range cardIndexes {
				card := e.data.Cards[cardIndex]
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
	log.Printf(
		"engine next question: checkedQuestions=%d skippedAnswered=%d skippedState=%d evaluatedCards=%d totalCards=%d found=%t bestQuestionID=%d elapsed=%s",
		checkedQuestions,
		skippedAnswered,
		skippedState,
		len(cardIndexes),
		len(e.data.Cards),
		found,
		bestQuestion.ID,
		time.Since(start),
	)
	return bestQuestion, found
}
