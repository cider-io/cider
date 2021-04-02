package functions

import (
	"math"
)

var Map = map[string](func([]float64) float64){
	"sum": sum,
	"max": max,
	"min": min,
}

func sum(input []float64) float64 {
	result := 0.0
	for _, val := range input {
		result += val
	}
	return result
}

func max(input []float64) float64 {
	result := math.Inf(-1)
	for _, val := range input {
		if result < val {
			result = val
		}
	}
	return result
}

func min(input []float64) float64 {
	result := math.Inf(1)
	for _, val := range input {
		if result > val {
			result = val
		}
	}
	return result
}
