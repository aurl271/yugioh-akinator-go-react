package repository

import (
	"testing"
	"yugioh-akinator-backend/internal/game"
)

// testConditionCard は各条件演算子のテストで共通して使うカードを作る。
func testConditionCard() game.Card {
	return game.Card{
		CardID:    1001,
		Reading:   "ブルーアイズホワイトドラゴン",
		Desc:      "このカードは特殊召喚できる。",
		Setcode:   0b110100,
		Type:      0b1010,
		Atk:       2500,
		Def:       2100,
		Level:     8,
		Race:      1,
		Attribute: 2,
	}
}

// TestMatchConditionWithAndLogic はand条件がすべて一致した場合だけtrueになることを確認する。
func TestMatchConditionWithAndLogic(t *testing.T) {
	card := testConditionCard()
	tests := []struct {
		name       string
		conditions []game.ConditionItem
		want       bool
	}{
		{
			name: "all conditions match",
			conditions: []game.ConditionItem{
				{Field: CardFieldAtk, Op: ConditionOpEq, Value: 2500},
				{Field: CardFieldLevel, Op: ConditionOpBetween, Min: 7, Max: 8},
			},
			want: true,
		},
		{
			name: "one condition does not match",
			conditions: []game.ConditionItem{
				{Field: CardFieldAtk, Op: ConditionOpEq, Value: 2500},
				{Field: CardFieldLevel, Op: ConditionOpEq, Value: 4},
			},
			want: false,
		},
		{name: "empty conditions", conditions: []game.ConditionItem{}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchCondition(card, game.Condition{
				Logic:      "and",
				Conditions: tt.conditions,
			})
			if err != nil {
				t.Fatalf("matchCondition returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchCondition matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionWithOrLogic はor条件が1つでも一致した場合にtrueになることを確認する。
func TestMatchConditionWithOrLogic(t *testing.T) {
	card := testConditionCard()
	tests := []struct {
		name       string
		conditions []game.ConditionItem
		want       bool
	}{
		{
			name: "one condition matches",
			conditions: []game.ConditionItem{
				{Field: CardFieldAtk, Op: ConditionOpEq, Value: 1000},
				{Field: CardFieldLevel, Op: ConditionOpEq, Value: 8},
			},
			want: true,
		},
		{
			name: "no condition matches",
			conditions: []game.ConditionItem{
				{Field: CardFieldAtk, Op: ConditionOpEq, Value: 1000},
				{Field: CardFieldLevel, Op: ConditionOpEq, Value: 4},
			},
			want: false,
		},
		{name: "empty conditions", conditions: []game.ConditionItem{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchCondition(card, game.Condition{
				Logic:      "or",
				Conditions: tt.conditions,
			})
			if err != nil {
				t.Fatalf("matchCondition returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchCondition matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionRejectsUnknownLogic はand/or以外のlogicがエラーになることを確認する。
func TestMatchConditionRejectsUnknownLogic(t *testing.T) {
	matched, err := matchCondition(testConditionCard(), game.Condition{Logic: "unknown"})
	if err == nil {
		t.Fatal("matchCondition returned nil error for unknown logic")
	}
	if matched {
		t.Error("matchCondition matched = true; want false")
	}
}

// TestMatchConditionReturnsItemError はand/or内の条件項目で発生したエラーが呼び出し元へ返ることを確認する。
func TestMatchConditionReturnsItemError(t *testing.T) {
	condition := game.Condition{
		Logic: "and",
		Conditions: []game.ConditionItem{
			{Field: CardFieldAtk, Op: "unknown"},
		},
	}

	matched, err := matchCondition(testConditionCard(), condition)
	if err == nil {
		t.Fatal("matchCondition returned nil error for invalid condition item")
	}
	if matched {
		t.Error("matchCondition matched = true; want false")
	}
}

// TestMatchConditionItemEq はeqが数値フィールドの一致と不一致を判定することを確認する。
func TestMatchConditionItemEq(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  bool
	}{
		{name: "equal", value: 2500, want: true},
		{name: "not equal", value: 2400, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldAtk,
				Op:    ConditionOpEq,
				Value: tt.value,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemBetween はbetweenが最小値と最大値を含む範囲を判定することを確認する。
func TestMatchConditionItemBetween(t *testing.T) {
	tests := []struct {
		name string
		min  int64
		max  int64
		want bool
	}{
		{name: "below range", min: 2501, max: 3000, want: false},
		{name: "equal to minimum", min: 2500, max: 3000, want: true},
		{name: "inside range", min: 2000, max: 3000, want: true},
		{name: "equal to maximum", min: 2000, max: 2500, want: true},
		{name: "above range", min: 1000, max: 2499, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldAtk,
				Op:    ConditionOpBetween,
				Min:   tt.min,
				Max:   tt.max,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemBitOn は指定ビットが1つでも立っているかをbit_onが判定することを確認する。
func TestMatchConditionItemBitOn(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  bool
	}{
		{name: "bit is on", value: 0b0010, want: true},
		{name: "bit is off", value: 0b0100, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldType,
				Op:    ConditionOpBitOn,
				Value: tt.value,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemBitOff は指定ビットがすべて立っていないかをbit_offが判定することを確認する。
func TestMatchConditionItemBitOff(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  bool
	}{
		{name: "bit is off", value: 0b0100, want: true},
		{name: "bit is on", value: 0b0010, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldType,
				Op:    ConditionOpBitOff,
				Value: tt.value,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemBitMaskEq はmaskで残したビットが指定値と等しいかを判定することを確認する。
func TestMatchConditionItemBitMaskEq(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  bool
	}{
		{name: "masked value is equal", value: 0b1010, want: true},
		{name: "masked value is not equal", value: 0b0010, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldType,
				Op:    ConditionOpBitMaskEq,
				Mask:  0b1110,
				Value: tt.value,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemShiftMaskEq は右シフト後にmaskした値を比較できることを確認する。
func TestMatchConditionItemShiftMaskEq(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  bool
	}{
		{name: "shifted value is equal", value: 0b101, want: true},
		{name: "shifted value is not equal", value: 0b001, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldSetcode,
				Op:    ConditionOpShiftMaskEq,
				Shift: 2,
				Mask:  0b111,
				Value: tt.value,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemStartsWith はreadingの先頭文字列をstarts_withが判定することを確認する。
func TestMatchConditionItemStartsWith(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "has prefix", text: "ブルーアイズ", want: true},
		{name: "does not have prefix", text: "ホワイト", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldReading,
				Op:    ConditionOpStartsWith,
				Text:  tt.text,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemEndsWith はreadingの末尾文字列をends_withが判定することを確認する。
func TestMatchConditionItemEndsWith(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "has suffix", text: "ドラゴン", want: true},
		{name: "does not have suffix", text: "ホワイト", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldReading,
				Op:    ConditionOpEndsWith,
				Text:  tt.text,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemContains は効果テキストに指定文字列が含まれるかをcontainsが判定することを確認する。
func TestMatchConditionItemContains(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "contains text", text: "特殊召喚", want: true},
		{name: "does not contain text", text: "破壊する", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
				Field: CardFieldDesc,
				Op:    ConditionOpContains,
				Text:  tt.text,
			})
			if err != nil {
				t.Fatalf("matchConditionItem returned error: %v", err)
			}
			if matched != tt.want {
				t.Errorf("matchConditionItem matched = %t; want %t", matched, tt.want)
			}
		})
	}
}

// TestMatchConditionItemRejectsWrongStringField は文字列演算子に対応しないfieldを指定するとエラーになることを確認する。
func TestMatchConditionItemRejectsWrongStringField(t *testing.T) {
	tests := []struct {
		name string
		item game.ConditionItem
	}{
		{
			name: "starts_with requires reading",
			item: game.ConditionItem{Field: CardFieldAtk, Op: ConditionOpStartsWith, Text: "2500"},
		},
		{
			name: "ends_with requires reading",
			item: game.ConditionItem{Field: CardFieldDesc, Op: ConditionOpEndsWith, Text: "召喚"},
		},
		{
			name: "contains requires desc",
			item: game.ConditionItem{Field: CardFieldReading, Op: ConditionOpContains, Text: "ドラゴン"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := matchConditionItem(testConditionCard(), tt.item)
			if err == nil {
				t.Fatal("matchConditionItem returned nil error for unsupported field")
			}
			if matched {
				t.Error("matchConditionItem matched = true; want false")
			}
		})
	}
}

// TestMatchConditionItemRejectsUnknownOperation は未定義の演算子がエラーになることを確認する。
func TestMatchConditionItemRejectsUnknownOperation(t *testing.T) {
	matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
		Field: CardFieldAtk,
		Op:    "unknown",
	})
	if err == nil {
		t.Fatal("matchConditionItem returned nil error for unknown operation")
	}
	if matched {
		t.Error("matchConditionItem matched = true; want false")
	}
}

// TestMatchConditionItemRejectsUnknownField は数値演算子に未定義のfieldを指定するとエラーになることを確認する。
func TestMatchConditionItemRejectsUnknownField(t *testing.T) {
	matched, err := matchConditionItem(testConditionCard(), game.ConditionItem{
		Field: "unknown",
		Op:    ConditionOpEq,
		Value: 1,
	})
	if err == nil {
		t.Fatal("matchConditionItem returned nil error for unknown field")
	}
	if matched {
		t.Error("matchConditionItem matched = true; want false")
	}
}

// TestCardFieldValue は各field名がCardの対応する数値フィールドを返すことを確認する。
func TestCardFieldValue(t *testing.T) {
	card := testConditionCard()
	tests := []struct {
		field string
		want  int64
	}{
		{field: CardFieldType, want: card.Type},
		{field: CardFieldAtk, want: int64(card.Atk)},
		{field: CardFieldDef, want: int64(card.Def)},
		{field: CardFieldLevel, want: int64(card.Level)},
		{field: CardFieldSetcode, want: card.Setcode},
		{field: CardFieldRace, want: card.Race},
		{field: CardFieldAttribute, want: card.Attribute},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got, err := cardFieldValue(card, tt.field)
			if err != nil {
				t.Fatalf("cardFieldValue returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("cardFieldValue = %d; want %d", got, tt.want)
			}
		})
	}
}

// TestCardFieldValueRejectsUnknownField は未定義のfield名がエラーになることを確認する。
func TestCardFieldValueRejectsUnknownField(t *testing.T) {
	got, err := cardFieldValue(testConditionCard(), "unknown")
	if err == nil {
		t.Fatal("cardFieldValue returned nil error for unknown field")
	}
	if got != 0 {
		t.Errorf("cardFieldValue = %d; want 0", got)
	}
}
