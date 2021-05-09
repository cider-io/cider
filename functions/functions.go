package functions

import (
	"errors"
	"math"
	"strconv"
	"time"
)

/* see: https://golang.org/pkg/encoding/json/#Unmarshal for how JSON fields are unmarshalled into interfaces
 * - bool, for JSON booleans
 * - float64, for JSON numbers
 * - string, for JSON strings
 * - []interface{}, for JSON arrays
 * - map[string]interface{}, for JSON objects
 * - nil for JSON null */
type Data interface {}

var Map = map[string](func(Data, chan bool) (Data, error)){
	"sum":  sum,
	"max":  max,
	"min":  min,
	"sleep": sleep,
}

func sum(data Data, abort chan bool) (Data, error) {
	result := 0.0

	// see: https://golang.org/pkg/encoding/json/#Unmarshal
	input, ok := data.([]interface{})
	if !ok {
		return result, errors.New("Expected input type []float64.")
	}
	
	for index, val := range input {
		select {
		case <-abort: // return a partial result if the function is async aborted
			return result, errors.New("Aborted after " + strconv.Itoa(index) + " iterations.")
		default:
			value, ok := val.(float64)
			if !ok {
				return result, errors.New("Expected input type []float64.")
			}
			result += value
		}
	}
	return result, nil
}

func max(data Data, abort chan bool) (Data, error) {
	result := math.Inf(-1)

	input, ok := data.([]interface{})
	if !ok {
		return result, errors.New("Expected input type []float64.")
	}

	for index, val := range input {
		select {
		case <-abort:
			return result, errors.New("Aborted after " + strconv.Itoa(index) + " iterations.")
		default:
			value, ok := val.(float64)
			if !ok {
				return result, errors.New("Expected input type []float64.")
			}
			if result < value {
				result = value
			}
		}
	}
	return result, nil
}

func min(data Data, abort chan bool) (Data, error) {
	result := math.Inf(1)

	input, ok := data.([]interface{})
	if !ok {
		return result, errors.New("Expected input type []float64.")
	}

	for index, val := range input {
		select {
		case <-abort:
			return result, errors.New("Aborted after " + strconv.Itoa(index) + " iterations.")
		default:
			value, ok := val.(float64)
			if !ok {
				return result, errors.New("Expected input type []float64.")
			}
			if result > value {
				result = value
			}
		}
	}
	return result, nil
}

// sleep: Sleeps for a specified delay and returns 1010.1010 when it wakes up
func sleep(data Data, abort chan bool) (Data, error) {
	result := 0

	// see: https://golang.org/pkg/encoding/json/#Unmarshal
	sleepTime, ok := data.(float64)
	if !ok {
		return result, errors.New("Expected input type float64.")
	}

	for i := 0; i < int(sleepTime); i++ {
		select {
		case <-abort:
			return result, errors.New("Aborted after " + strconv.Itoa(i) + " iterations.")
		default:
			time.Sleep(time.Second)
		}
	}
	return 1010.101, nil
}
