package functions

import (
	"math"
	"errors"
	"strconv"
)

var Map = map[string](func([]float64, chan bool) (float64, error)){
	"sum": sum,
	"max": max,
	"min": min,
}

func sum(input []float64, abort chan bool) (float64, error) {
	result := 0.0
	for index, val := range input {
		select {
		case <- abort: // return a partial result if the function is async aborted
			return result, errors.New("Aborted after " + strconv.Itoa(index) + " iterations.")
		default:
			result += val
		}	
	}
	return result, nil
}

func max(input []float64, abort chan bool) (float64, error) {
	result := math.Inf(-1)
	for index, val := range input {
		select {
		case <- abort:
			return result, errors.New("Aborted after " + strconv.Itoa(index) + " iterations.")
		default:
			if result < val {
				result = val
			}
		}	
	}
	return result, nil
}

func min(input []float64, abort chan bool) (float64, error) {
	result := math.Inf(1)
	for index, val := range input {
		select {
		case <- abort:
			return result, errors.New("Aborted after " + strconv.Itoa(index) + " iterations.")
		default:
			if result > val {
				result = val
			}
		}
	}
	return result, nil
}
