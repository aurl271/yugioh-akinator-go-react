package game

import (
	"math"
)

func shannon_entropy(p []float64) float64{
	entropy := 0.0
	for _,v := range p {
		if v > 0 {
			entropy -=  v * math.Log2(v)
		}
	}
	return entropy
}