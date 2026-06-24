package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"yugioh-akinator-backend/internal/game"
)

const (
	// QuestionCategoryScript はanswersテーブルにYESカードを持つ質問。
	QuestionCategoryScript = 0
	// QuestionCategoryCards はcardsテーブルの条件からYES/NOを生成する質問。
	QuestionCategoryCards  = 1
)

// RawAnswer はanswersテーブルから読み取った1行分の生データ。
type RawAnswer struct {
	QuestionID int
	CardID     int64
	Answer     int
}

const (
	// CardField* はcondition_jsonのfieldで使えるカード列名。
	CardFieldType    = "type"
	CardFieldAtk     = "atk"
	CardFieldDef     = "def"
	CardFieldLevel   = "level"
	CardFieldSetcode = "setcode"
	CardFieldReading = "reading"
)

const (
	// ConditionOp* はcondition_jsonのopで使える判定演算子。
	ConditionOpEq          = "eq"
	ConditionOpBetween     = "between"
	ConditionOpBitOn       = "bit_on"
	ConditionOpBitOff      = "bit_off"
	ConditionOpBitMaskEq   = "bit_mask_eq"
	ConditionOpShiftMaskEq = "shift_mask_eq"
	ConditionOpStartsWith  = "starts_with"
	ConditionOpEndsWith    = "ends_with"
)

// LoadGameData は推理に必要なDBデータをまとめて読み込み、Engineが使う形へ変換する。
func LoadGameData(ctx context.Context, db *sql.DB) (game.GameData, error) {
	cards, err := LoadCards(ctx, db)
	if err != nil {
		return game.GameData{}, fmt.Errorf("load cards: %w", err)
	}
	questions, err := LoadQuestions(ctx, db)
	if err != nil {
		return game.GameData{}, fmt.Errorf("load questions: %w", err)
	}
	rawAnswers, err := LoadAnswers(ctx, db)
	if err != nil {
		return game.GameData{}, fmt.Errorf("load answers: %w", err)
	}
	answers, err := buildAnswers(questions, cards, rawAnswers)
	if err != nil {
		return game.GameData{}, fmt.Errorf("build answers: %w", err)
	}
	questionByID := buildQuestionByID(questions)

	return game.GameData{
		Cards:        cards,
		Questions:    questions,
		QuestionByID: questionByID,
		Answers:      answers,
	}, nil
}

// LoadCards はcardsテーブルをGameData用のCard一覧へ変換する。
func LoadCards(ctx context.Context, db *sql.DB) ([]game.Card, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			card_id,
			name,
			reading,
			"desc",
			setcode,
			type,
			atk,
			def,
			level,
			race,
			attribute
		FROM cards
		ORDER BY id
	`)
	if err != nil {
		return []game.Card{}, err
	}
	defer rows.Close()

	cards := make([]game.Card, 0)
	for rows.Next() {
		var id int64
		var cardID int64
		var name string
		var reading sql.NullString
		var desc sql.NullString
		var setcode int64
		var cardType int64
		var atk int
		var def int
		var level int
		var race int64
		var attribute int64

		if err := rows.Scan(&id, &cardID, &name, &reading, &desc, &setcode, &cardType, &atk, &def, &level, &race, &attribute); err != nil {
			return []game.Card{}, err
		}
		cards = append(cards, game.Card{
			ID:        id,
			CardID:    cardID,
			Name:      name,
			Reading:   reading.String,
			Desc:      desc.String,
			Setcode:   setcode,
			Type:      cardType,
			Atk:       atk,
			Def:       def,
			Level:     level,
			Race:      race,
			Attribute: attribute,
		})
	}

	if err := rows.Err(); err != nil {
		return []game.Card{}, err
	}

	return cards, nil
}

// LoadQuestions はquestionsテーブルを読み込み、condition_jsonを構造体へ戻す。
func LoadQuestions(ctx context.Context, db *sql.DB) ([]game.Question, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			question_text,
			category,
			query,
			condition_json,
			unset_bit,
			new_state
		FROM questions
		ORDER BY id
	`)
	if err != nil {
		return []game.Question{}, err
	}
	defer rows.Close()

	questions := make([]game.Question, 0)
	for rows.Next() {
		var id int
		var text string
		var category int
		var query sql.NullString
		var conditionText sql.NullString
		var unsetBit int
		var newState int
		if err := rows.Scan(&id, &text, &category, &query, &conditionText, &unsetBit, &newState); err != nil {
			return []game.Question{}, err
		}

		var condition *game.Condition
		if conditionText.Valid && conditionText.String != "" {
			var parsed game.Condition
			if err := json.Unmarshal([]byte(conditionText.String), &parsed); err != nil {
				return []game.Question{}, fmt.Errorf("parse condition_json question_id=%d: %w", id, err)
			}
			condition = &parsed
		}
		questions = append(questions, game.Question{
			ID:        id,
			Text:      text,
			Category:  category,
			Query:     query.String,
			Condition: condition,
			UnsetBit:  unsetBit,
			NewState:  newState,
		})
	}

	if err := rows.Err(); err != nil {
		return []game.Question{}, err
	}

	return questions, nil
}

// LoadAnswers はscript由来質問のYESカードをanswersテーブルから読み込む。
func LoadAnswers(ctx context.Context, db *sql.DB) ([]RawAnswer, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			card_id,
			question_id,
			answer
		FROM answers
		ORDER BY id
	`)
	if err != nil {
		return []RawAnswer{}, err
	}
	defer rows.Close()

	answers := make([]RawAnswer, 0)
	for rows.Next() {
		var cardID int64
		var questionID int
		var answer int

		if err := rows.Scan(&cardID, &questionID, &answer); err != nil {
			return []RawAnswer{}, err
		}

		answers = append(answers, RawAnswer{
			QuestionID: questionID,
			CardID:     cardID,
			Answer:     answer,
		})
	}

	if err := rows.Err(); err != nil {
		return []RawAnswer{}, err
	}

	return answers, nil
}

// buildAnswers はscript由来answersとcards条件由来answersを1つのmapへまとめる。
// cards条件由来の質問はDBに全回答を保存せず、Go側でconditionからYESカードを生成する。
func buildAnswers(
	questions []game.Question,
	cards []game.Card,
	rawAnswers []RawAnswer,
) (map[int]map[int64]int, error) {
	answers := make(map[int]map[int64]int, len(questions))

	for _, question := range questions {
		answers[question.ID] = make(map[int64]int)
	}

	for _, rawAnswer := range rawAnswers {
		answers[rawAnswer.QuestionID][rawAnswer.CardID] = rawAnswer.Answer
	}

	for _, question := range questions {
		if question.Category != QuestionCategoryCards {
			continue
		}

		if question.Condition == nil {
			return nil, fmt.Errorf("cards question has no condition: question_id=%d text=%q", question.ID, question.Text)
		}

		for _, card := range cards {
			matched, err := matchCondition(card, *question.Condition)
			if err != nil {
				return nil, err
			}
			if matched {
				answers[question.ID][card.CardID] = 1
			}
		}
	}

	return answers, nil
}

// matchCondition はCondition全体のand/orを評価する。
func matchCondition(card game.Card, condition game.Condition) (bool, error) {
	switch condition.Logic {
	case "and":
		for _, item := range condition.Conditions {
			matched, err := matchConditionItem(card, item)
			if err != nil {
				return false, err
			}
			if !matched {
				return false, nil
			}
		}
		return true, nil
	case "or":
		for _, item := range condition.Conditions {
			matched, err := matchConditionItem(card, item)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown condition logic: %s", condition.Logic)
	}
}

// matchConditionItem はConditionItemを1つ評価する入口。
func matchConditionItem(card game.Card, item game.ConditionItem) (bool, error) {
	return matchNumberCondition(card, item)
}

// matchNumberCondition は数値条件とreading文字列条件を評価する。
// 名前は歴史的にNumberだが、starts_with/ends_withもここで扱っている。
func matchNumberCondition(card game.Card, item game.ConditionItem) (bool, error) {
	numberValue := func() (int64, error) {
		return cardFieldValue(card, item.Field)
	}

	switch item.Op {
	case ConditionOpEq:
		value, err := numberValue()
		if err != nil {
			return false, err
		}
		if value == item.Value {
			return true, nil
		}
		return false, nil
	case ConditionOpBetween:
		value, err := numberValue()
		if err != nil {
			return false, err
		}
		if item.Min <= value && value <= item.Max {
			return true, nil
		}
		return false, nil
	case ConditionOpBitOn:
		value, err := numberValue()
		if err != nil {
			return false, err
		}
		if value&item.Value != 0 {
			return true, nil
		}
		return false, nil
	case ConditionOpBitOff:
		value, err := numberValue()
		if err != nil {
			return false, err
		}
		if value&item.Value == 0 {
			return true, nil
		}
		return false, nil
	case ConditionOpBitMaskEq:
		value, err := numberValue()
		if err != nil {
			return false, err
		}
		if value&item.Mask == item.Value {
			return true, nil
		}
		return false, nil
	case ConditionOpShiftMaskEq:
		value, err := numberValue()
		if err != nil {
			return false, err
		}
		if (value>>item.Shift)&item.Mask == item.Value {
			return true, nil
		}
		return false, nil
	case ConditionOpStartsWith:
		if item.Field != CardFieldReading {
			return false, fmt.Errorf("condition op %s requires field %s", item.Op, CardFieldReading)
		}
		return strings.HasPrefix(card.Reading, item.Text), nil
	case ConditionOpEndsWith:
		if item.Field != CardFieldReading {
			return false, fmt.Errorf("condition op %s requires field %s", item.Op, CardFieldReading)
		}
		return strings.HasSuffix(card.Reading, item.Text), nil
	}
	return false, fmt.Errorf("unknown condition op: %s", item.Op)
}

// cardFieldValue はConditionItem.FieldをCardの数値フィールドへ対応させる。
func cardFieldValue(card game.Card, field string) (int64, error) {
	switch field {
	case CardFieldType:
		return card.Type, nil
	case CardFieldAtk:
		return int64(card.Atk), nil
	case CardFieldDef:
		return int64(card.Def), nil
	case CardFieldLevel:
		return int64(card.Level), nil
	case CardFieldSetcode:
		return card.Setcode, nil
	default:
		return 0, fmt.Errorf("unknown field: %s", field)
	}
}

// buildQuestionByID は回答履歴のquestion idからQuestionを高速に引くためのmapを作る。
func buildQuestionByID(questions []game.Question) map[int]game.Question {
	questionByID := make(map[int]game.Question, len(questions))
	for _, question := range questions {
		questionByID[question.ID] = question
	}
	return questionByID
}
