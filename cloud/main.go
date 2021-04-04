package main

import (
	"math"
	"fmt"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
)

type ComputeRequest struct {
	Function string
	Data []float64
}

var functionMap = map[string](func([]float64) float64){
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

func compute(request ComputeRequest) (string, error) {
	if _, ok := functionMap[request.Function]; ok {
		return fmt.Sprintf("%f", functionMap[request.Function](request.Data)), nil
	} else {
		return fmt.Sprintf("Invalid function."), errors.New("Invalid function.")
	}
}

func main() {
	lambda.Start(compute)
}
