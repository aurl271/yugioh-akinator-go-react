package game

import (
	"math"
	"testing"
)

const entropyTestTolerance = 1e-9

// assertEntropyEqual はエントロピーの計算結果を浮動小数点の許容誤差付きで比較する。
func assertEntropyEqual(t *testing.T, got float64, want float64) {
	t.Helper()

	if math.Abs(got-want) > entropyTestTolerance {
		t.Errorf("shannon_entropy = %f; want %f", got, want)
	}
}

// TestShannonEntropyReturnsZeroForCertainOutcome は結果が1つに確定した分布のエントロピーが0になることを確認する。
func TestShannonEntropyReturnsZeroForCertainOutcome(t *testing.T) {
	got := shannon_entropy([]float64{1.0, 0.0})

	assertEntropyEqual(t, got, 0.0)
}

// TestShannonEntropyReturnsOneForEqualBinaryProbabilities は2つの結果が同確率ならエントロピーが1になることを確認する。
func TestShannonEntropyReturnsOneForEqualBinaryProbabilities(t *testing.T) {
	got := shannon_entropy([]float64{0.5, 0.5})

	assertEntropyEqual(t, got, 1.0)
}

// TestShannonEntropyIgnoresZeroProbabilities は確率0の要素がエントロピーの計算結果へ影響しないことを確認する。
func TestShannonEntropyIgnoresZeroProbabilities(t *testing.T) {
	got := shannon_entropy([]float64{0.5, 0.0, 0.5})

	assertEntropyEqual(t, got, 1.0)
}

// TestShannonEntropyReturnsZeroForEmptyInput は空の確率分布に対して0を返すことを確認する。
func TestShannonEntropyReturnsZeroForEmptyInput(t *testing.T) {
	got := shannon_entropy([]float64{})

	assertEntropyEqual(t, got, 0.0)
}
