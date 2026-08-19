package game

const (
	// AnswerScore* はユーザー回答を数値化したもの。
	// YESを1、NOを-1に置くことで、カード側の期待回答との差分を距離として扱える。
	AnswerScoreNo         = -1.0
	AnswerScoreProbablyNo = -0.5
	AnswerScoreUnknown    = 0.0
	AnswerScoreProbably   = 0.5
	AnswerScoreYes        = 1.0
)

const (
	// ExpectedAnswerYes は、GameData.AnswersのValue。
	ExpectedAnswerYes int = 1
)

// AnswerScores は NextQuestion が「各回答を選んだら候補がどう分かれるか」を試すための一覧。
var AnswerScores = []float64{
	AnswerScoreNo,
	AnswerScoreProbablyNo,
	AnswerScoreUnknown,
	AnswerScoreProbably,
	AnswerScoreYes,
}

// isValidAnswerScore はフロントエンドから来た回答値が、エンジンの想定範囲内かを確認する。
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

// GameData は推理エンジンが使う全データをメモリ上にまとめたもの。
// サーバー起動時にDBから読み込み、リクエストごとにこのデータを使ってEngineを作る。
type GameData struct {
	// Cards は推理対象カード。scoresのindexと同じ順番で扱う。
	Cards []Card
	// Questions は出題候補。NextQuestionはこの順に全質問を評価する。
	Questions []Question
	// QuestionByID は回答履歴のquestion idから質問メタ情報を引くためのmap。
	QuestionByID map[int]Question
	// Answers は questionID -> cardID -> YES/NO の対応。
	// mapに存在しないカードは、その質問に対してNOとして扱う。
	Answers map[int]map[int64]int
}

// Card はcardsテーブル1行に対応する、推理対象カードの情報。
type Card struct {
	// ID はDB内部の連番ID。
	ID int64
	// CardID は遊戯王カード固有のID。answersとの対応にはこちらを使う。
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

// Question はquestionsテーブル1行に対応する出題候補。
type Question struct {
	ID   int
	Text string
	// Category はscript由来かcards条件由来かを表す。
	Category int
	// Query は旧SQLite実装との互換用。Goの判定では主にConditionを使う。
	Query string
	// Condition はcards由来質問をGoで判定するための構造化条件。
	Condition *Condition
	// UnsetBit は現在stateにこのビットが立っていたら質問を出さない、という除外条件。
	UnsetBit int
	// NewState はYES/Probablyで答えたときにstateへ追加するビット。
	NewState int
}

// Condition は複数の条件をand/orでまとめる構造。
type Condition struct {
	Logic      string          `json:"logic"`
	Conditions []ConditionItem `json:"conditions"`
}

// ConditionItem は1つのフィールド判定を表す。
// 数値条件ではValue/Min/Max/Mask/Shiftを使い、reading条件ではTextを使う。
type ConditionItem struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value int64  `json:"value"`
	Text  string `json:"text"`
	Min   int64  `json:"min"`
	Max   int64  `json:"max"`
	Mask  int64  `json:"mask"`
	Shift int64  `json:"shift"`
}

// Candidate は候補カードを確率付きでフロントエンドへ返すための内部表現。
type Candidate struct {
	Rank        int
	CardID      int64
	CardName    string
	Probability float64
}

// AnsweredQuestion はEngine内部で保持する回答履歴。
type AnsweredQuestion struct {
	QuestionID int
	Answer     float64
}
