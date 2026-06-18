package game

const (
	AnswerScoreNo         = -1.0
	AnswerScoreProbablyNo = -0.5
	AnswerScoreUnknown    = 0.0
	AnswerScoreProbably   = 0.5
	AnswerScoreYes        = 1.0
)

var AnswerScores = []float64{
	AnswerScoreNo,
	AnswerScoreProbablyNo,
	AnswerScoreUnknown,
	AnswerScoreProbably,
	AnswerScoreYes,
}

func isValidAnswerScore(answer float64) bool {
	switch answer {
	case AnswerScoreYes,
		AnswerScoreProbably,
		AnswerScoreUnknown,
		AnswerScoreProbablyNo,
		AnswerScoreNo:
		return true
	default:
		return false
	}
}

type GameData struct {
	Cards        []Card
	Questions    []Question
	QuestionByID map[int]Question
	Answers      map[int]map[int64]int
}

type Card struct {
	ID        int64
	CardID    int64
	Name      string
	Reading   string
	Desc      string
	Setcode   int64
	Type      int64
	Atk       int
	Def       int
	Level     int
	Race      int64
	Attribute int64
}

type Question struct {
	ID        int
	Text      string
	Category  int
	Query     string
	Condition *Condition
	UnsetBit  int
	NewState  int
}

type Condition struct {
	Logic      string          `json:"logic"`
	Conditions []ConditionItem `json:"conditions"`
}

type ConditionItem struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value int64  `json:"value"`
	Min   int64  `json:"min"`
	Max   int64  `json:"max"`
	Mask  int64  `json:"mask"`
	Shift int64  `json:"shift"`
}

type Candidate struct {
	Rank        int
	CardID      int64
	CardName    string
	Probability float64
}

type AnsweredQuestion struct {
	QuestionID int
	Answer     float64
}
