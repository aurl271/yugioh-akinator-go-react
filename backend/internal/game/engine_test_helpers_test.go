package game

const (
	testQuestionAID = 1
	testQuestionBID = 2
	testCardAID     = 1001
	testCardBID     = 1002
	testStateBit    = 1
)

// testGameData はEngineの単体テストで使う、2枚のカードと1つの質問を持つデータを毎回新しく作る。
func testGameData() GameData {
	questionA := Question{
		ID:       testQuestionAID,
		Text:     "質問A",
		NewState: testStateBit,
	}

	return GameData{
		Cards: []Card{
			{CardID: testCardAID, Name: "カードA"},
			{CardID: testCardBID, Name: "カードB"},
		},
		Questions: []Question{questionA},
		QuestionByID: map[int]Question{
			testQuestionAID: questionA,
		},
		Answers: map[int]map[int64]int{
			testQuestionAID: {
				testCardAID: ExpectedAnswerYes,
			},
		},
	}
}

// testCards はカード件数による候補制限をテストするため、指定された枚数のカードを作る。
func testCards(count int) []Card {
	cards := make([]Card, count)
	for i := range cards {
		cards[i] = Card{CardID: int64(i + 1)}
	}
	return cards
}
