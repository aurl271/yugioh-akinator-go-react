package game

import "testing"

// TestNextQuestionSkipsAnsweredQuestions は回答済みの質問を除外して未回答の質問を選ぶことを確認する。
func TestNextQuestionSkipsAnsweredQuestions(t *testing.T) {
	questionA := Question{ID: testQuestionAID, Text: "質問A"}
	questionB := Question{ID: testQuestionBID, Text: "質問B"}
	data := GameData{
		Cards:     testGameData().Cards,
		Questions: []Question{questionA, questionB},
		QuestionByID: map[int]Question{
			testQuestionAID: questionA,
			testQuestionBID: questionB,
		},
		Answers: map[int]map[int64]int{
			testQuestionAID: {testCardAID: ExpectedAnswerYes},
			testQuestionBID: {testCardBID: ExpectedAnswerYes},
		},
	}
	engine := NewEngine(data, 1.0)
	if err := engine.ApplyAnswer(testQuestionAID, AnswerScoreUnknown); err != nil {
		t.Fatalf("ApplyAnswer returned error: %v", err)
	}

	question, ok := engine.NextQuestion()
	if !ok {
		t.Fatal("NextQuestion found no question; want question B")
	}
	if question.ID != testQuestionBID {
		t.Errorf("NextQuestion ID = %d; want %d", question.ID, testQuestionBID)
	}
}

// TestNextQuestionSkipsQuestionsExcludedByState はstateとUnsetBitが重なる質問を出題しないことを確認する。
func TestNextQuestionSkipsQuestionsExcludedByState(t *testing.T) {
	excludedQuestion := Question{ID: testQuestionAID, Text: "除外質問", UnsetBit: testStateBit}
	availableQuestion := Question{ID: testQuestionBID, Text: "出題可能な質問"}
	data := GameData{
		Cards:     testGameData().Cards,
		Questions: []Question{excludedQuestion, availableQuestion},
		Answers: map[int]map[int64]int{
			testQuestionAID: {testCardAID: ExpectedAnswerYes},
			testQuestionBID: {testCardBID: ExpectedAnswerYes},
		},
	}
	engine := NewEngine(data, 1.0)
	engine.state = testStateBit

	question, ok := engine.NextQuestion()
	if !ok {
		t.Fatal("NextQuestion found no question; want available question")
	}
	if question.ID != testQuestionBID {
		t.Errorf("NextQuestion ID = %d; want %d", question.ID, testQuestionBID)
	}
}

// TestNextQuestionReturnsFalseWhenNoQuestionIsAvailable は出題可能な質問がない場合にfalseを返すことを確認する。
func TestNextQuestionReturnsFalseWhenNoQuestionIsAvailable(t *testing.T) {
	data := GameData{Cards: testGameData().Cards}
	engine := NewEngine(data, 1.0)

	question, ok := engine.NextQuestion()
	if ok {
		t.Errorf("NextQuestion returned question ID %d; want no question", question.ID)
	}
}

// TestNextQuestionSelectsMostInformativeQuestion は候補が均等に分かれる情報量の高い質問を選ぶことを確認する。
func TestNextQuestionSelectsMostInformativeQuestion(t *testing.T) {
	const (
		cardCID = 1003
		cardDID = 1004
	)
	uninformativeQuestion := Question{ID: testQuestionAID, Text: "全カードがYESの質問"}
	informativeQuestion := Question{ID: testQuestionBID, Text: "回答が半分に分かれる質問"}
	data := GameData{
		Cards: []Card{
			{CardID: testCardAID},
			{CardID: testCardBID},
			{CardID: cardCID},
			{CardID: cardDID},
		},
		Questions: []Question{uninformativeQuestion, informativeQuestion},
		Answers: map[int]map[int64]int{
			testQuestionAID: {
				testCardAID: ExpectedAnswerYes,
				testCardBID: ExpectedAnswerYes,
				cardCID:     ExpectedAnswerYes,
				cardDID:     ExpectedAnswerYes,
			},
			testQuestionBID: {
				testCardAID: ExpectedAnswerYes,
				testCardBID: ExpectedAnswerYes,
			},
		},
	}
	engine := NewEngine(data, 1.0)

	question, ok := engine.NextQuestion()
	if !ok {
		t.Fatal("NextQuestion found no question; want informative question")
	}
	if question.ID != testQuestionBID {
		t.Errorf("NextQuestion ID = %d; want %d", question.ID, testQuestionBID)
	}
}

// TestTopScoreCardIndexesSelectsLowestScoresAfterAnswer は回答後にscoreが低いカードのindexを優先することを確認する。
func TestTopScoreCardIndexesSelectsLowestScoresAfterAnswer(t *testing.T) {
	engine := NewEngine(GameData{Cards: testCards(5)}, 1.0)
	engine.scores = []float64{3.0, 1.0, 4.0, 0.0, 2.0}
	engine.answeredQuestions = []AnsweredQuestion{{QuestionID: testQuestionAID, Answer: AnswerScoreYes}}

	indexes := engine.topScoreCardIndexes(3)
	want := []int{3, 1, 4}
	if len(indexes) != len(want) {
		t.Fatalf("topScoreCardIndexes length = %d; want %d", len(indexes), len(want))
	}
	for i := range want {
		if indexes[i] != want[i] {
			t.Errorf("topScoreCardIndexes[%d] = %d; want %d", i, indexes[i], want[i])
		}
	}
}

// TestTopScoreCardIndexesReturnsRequestedNumberOfCards は初回候補が上限件数を守り、重複や範囲外indexを含まないことを確認する。
func TestTopScoreCardIndexesReturnsRequestedNumberOfCards(t *testing.T) {
	const cardCount = 600
	engine := NewEngine(GameData{Cards: testCards(cardCount)}, 1.0)

	indexes := engine.topScoreCardIndexes(nextQuestionCardLimit)
	if len(indexes) != nextQuestionCardLimit {
		t.Fatalf(
			"topScoreCardIndexes length = %d; want %d",
			len(indexes),
			nextQuestionCardLimit,
		)
	}

	seen := make(map[int]bool, len(indexes))
	for _, index := range indexes {
		if index < 0 || index >= cardCount {
			t.Errorf("topScoreCardIndexes returned out-of-range index: %d", index)
			continue
		}
		if seen[index] {
			t.Errorf("topScoreCardIndexes returned duplicate index: %d", index)
		}
		seen[index] = true
	}
}
