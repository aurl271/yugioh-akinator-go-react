package game

import (
	"math"
	"testing"
)

// TestProbabilitiesSumToOne は各カードの確率を合計すると1になることを確認する。
func TestProbabilitiesSumToOne(t *testing.T) {
	const tolerance = 1e-9
	engine := NewEngine(testGameData(), 1.0)
	engine.scores = []float64{0.0, 1.0}

	probabilities := engine.Probabilities()
	if len(probabilities) != 2 {
		t.Fatalf("Probabilities length = %d; want 2", len(probabilities))
	}

	sum := 0.0
	for _, probability := range probabilities {
		sum += probability
	}
	if math.Abs(sum-1.0) > tolerance {
		t.Errorf("probability sum = %f; want 1.0", sum)
	}
}

// TestProbabilitiesGiveHigherProbabilityToLowerScore はscoreが低いカードほど高い確率になることを確認する。
func TestProbabilitiesGiveHigherProbabilityToLowerScore(t *testing.T) {
	engine := NewEngine(testGameData(), 1.0)
	engine.scores = []float64{0.0, 4.0}

	probabilities := engine.Probabilities()
	if len(probabilities) != 2 {
		t.Fatalf("Probabilities length = %d; want 2", len(probabilities))
	}
	if probabilities[0] <= probabilities[1] {
		t.Errorf(
			"lower score probability = %f; want greater than %f",
			probabilities[0],
			probabilities[1],
		)
	}
}

// TestProbabilitiesReturnsEmptyWithoutCards はカードがない場合に空の確率一覧を返すことを確認する。
func TestProbabilitiesReturnsEmptyWithoutCards(t *testing.T) {
	engine := NewEngine(GameData{}, 1.0)

	probabilities := engine.Probabilities()
	if len(probabilities) != 0 {
		t.Errorf("Probabilities length = %d; want 0", len(probabilities))
	}
}

// TestTopCandidatesSortsByProbability は候補カードが確率の高い順に並ぶことを確認する。
func TestTopCandidatesSortsByProbability(t *testing.T) {
	engine := NewEngine(testGameData(), 1.0)
	engine.scores = []float64{4.0, 0.0}

	candidates := engine.TopCandidates(2)
	if len(candidates) != 2 {
		t.Fatalf("TopCandidates length = %d; want 2", len(candidates))
	}
	if candidates[0].CardID != testCardBID {
		t.Errorf("first candidate cardID = %d; want %d", candidates[0].CardID, testCardBID)
	}
	if candidates[1].CardID != testCardAID {
		t.Errorf("second candidate cardID = %d; want %d", candidates[1].CardID, testCardAID)
	}
	if candidates[0].Probability <= candidates[1].Probability {
		t.Errorf(
			"first probability = %f; want greater than second probability %f",
			candidates[0].Probability,
			candidates[1].Probability,
		)
	}
}

// TestTopCandidatesAssignsRanks は並び順に対応して1から始まる順位が付くことを確認する。
func TestTopCandidatesAssignsRanks(t *testing.T) {
	engine := NewEngine(testGameData(), 1.0)

	candidates := engine.TopCandidates(2)
	if len(candidates) != 2 {
		t.Fatalf("TopCandidates length = %d; want 2", len(candidates))
	}
	for i, candidate := range candidates {
		wantRank := i + 1
		if candidate.Rank != wantRank {
			t.Errorf("candidate[%d].Rank = %d; want %d", i, candidate.Rank, wantRank)
		}
	}
}

// TestTopCandidatesRespectsLimit は指定件数を守り、カード枚数を超えるlimitでは全カードを返すことを確認する。
func TestTopCandidatesRespectsLimit(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{name: "smaller than card count", limit: 1, wantCount: 1},
		{name: "greater than card count", limit: 10, wantCount: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(testGameData(), 1.0)

			candidates := engine.TopCandidates(tt.limit)
			if len(candidates) != tt.wantCount {
				t.Errorf("TopCandidates length = %d; want %d", len(candidates), tt.wantCount)
			}
		})
	}
}

// TestTopCandidatesReturnsEmptyForInvalidLimit は0以下のlimitに対して空の候補一覧を返すことを確認する。
func TestTopCandidatesReturnsEmptyForInvalidLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{name: "zero", limit: 0},
		{name: "negative", limit: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(testGameData(), 1.0)

			candidates := engine.TopCandidates(tt.limit)
			if len(candidates) != 0 {
				t.Errorf("TopCandidates length = %d; want 0", len(candidates))
			}
		})
	}
}
