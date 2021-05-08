package functions

import (
	"errors"
	"math"
	"strconv"
	"time"
)

var Map = map[string](func([]float64, chan bool) (float64, error)){
	"sum":  sum,
	"max":  max,
	"min":  min,
	"test": test,
}

func sum(input []float64, abort chan bool) (float64, error) {
	result := 0.0
	for index, val := range input {
		select {
		case <-abort: // return a partial result if the function is async aborted
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
		case <-abort:
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
		case <-abort:
			return result, errors.New("Aborted after " + strconv.Itoa(index) + " iterations.")
		default:
			if result > val {
				result = val
			}
		}
	}
	return result, nil
}

// test: function only for test scripts. Takes in a delay and
// 		 returns 1010.1010 after the specified delay.
func test(input []float64, abort chan bool) (float64, error) {
	result := 1010.1010
	sleepTime := 0
	if len(input) > 0 {
		sleepTime = int(input[0])
	}
	for i := 0; i < sleepTime; i++ {
		select {
		case <-abort:
			return result, errors.New("Aborted after " + strconv.Itoa(i) + " iterations.")
		default:
			time.Sleep(time.Second)
		}
	}
	return result, nil
}
