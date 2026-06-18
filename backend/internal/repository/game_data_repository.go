package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"yugioh-akinator-backend/internal/game"
)

const (
	QuestionCategoryScript = 0
	QuestionCategoryCards  = 1
)

type RawAnswer struct {
	QuestionID int
	CardID     int64
	Answer     int
}

const (
	CardFieldType    = "type"
	CardFieldAtk     = "atk"
	CardFieldDef     = "def"
	CardFieldLevel   = "level"
	CardFieldSetcode = "setcode"
)

const (
	ConditionOpEq          = "eq"
	ConditionOpBetween     = "between"
	ConditionOpBitOn       = "bit_on"
	ConditionOpBitOff      = "bit_off"
	ConditionOpBitMaskEq   = "bit_mask_eq"
	ConditionOpShiftMaskEq = "shift_mask_eq"
)

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

func matchConditionItem(card game.Card, item game.ConditionItem) (bool, error) {
	value, err := cardFieldValue(card, item.Field)
	if err != nil {
		return false, err
	}

	switch item.Op {
	case ConditionOpEq:
		if value == item.Value {
			return true, nil
		}
		return false, nil
	case ConditionOpBetween:
		if item.Min <= value && value <= item.Max {
			return true, nil
		}
		return false, nil
	case ConditionOpBitOn:
		if value&item.Value != 0 {
			return true, nil
		}
		return false, nil
	case ConditionOpBitOff:
		if value&item.Value == 0 {
			return true, nil
		}
		return false, nil
	case ConditionOpBitMaskEq:
		if value&item.Mask == item.Value {
			return true, nil
		}
		return false, nil
	case ConditionOpShiftMaskEq:
		if (value>>item.Shift)&item.Mask == item.Value {
			return true, nil
		}
		return false, nil
	}
	return false, fmt.Errorf("unknown condition op: %s", item.Op)
}

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

func buildQuestionByID(questions []game.Question) map[int]game.Question {
	questionByID := make(map[int]game.Question, len(questions))
	for _, question := range questions {
		questionByID[question.ID] = question
	}
	return questionByID
}
