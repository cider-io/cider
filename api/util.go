package api

import (
	"cider/config"
	"cider/functions"
	"cider/gossip"
	"cider/handle"
	"cider/log"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// writeMessage: Write headers + formatted message to the response
func writeMessage(response *http.ResponseWriter, status int, format string, args ...interface{}) {
	message := []byte(fmt.Sprintf(format, args...))

	// changing the Header after calling WriteHeader(statusCode) has no effect,
	// so these lines must remain in this order
	(*response).Header().Add("Content-Type", "text/plain")
	(*response).WriteHeader(status)

	// HTTP response body is terminated with CRLF
	(*response).Write(append(message, '\r', '\n'))
}

// writeStruct: Write headers + struct (as json) to the response
func writeStruct(response *http.ResponseWriter, body interface{}) {
	jsonBody, err := json.Marshal(body)
	handle.Fatal(err)

	// changing the Header after calling WriteHeader(statusCode) has no effect,
	// so these lines must remain in this order
	(*response).Header().Add("Content-Type", "text/json")
	(*response).WriteHeader(http.StatusOK)

	// HTTP response body is terminated with CRLF
	(*response).Write(append(jsonBody, '\r', '\n'))
}

// generateUUID: Generate UUID using time, local ip
func generateUUID() (string, error) {
	input, err := time.Now().MarshalBinary()
	if err != nil {
		return "", err
	}

	rand.Seed(time.Now().UnixNano())
	nonce := make([]byte, config.NonceLength)
	rand.Read(nonce)
	input = append(input, nonce...)

	input = append(input, ([]byte(gossip.Self.IpAddress))...)

	return fmt.Sprintf("%x", sha256.Sum256(input)), nil
}

// runTask: Async run a task
func runTask(taskId string) {
	task := Tasks[taskId]
	task.Status = Running
	Tasks[taskId] = task

	updateLoad <- 1
	taskResult, taskErr := functions.Map[task.Function](task.Data, task.Abort)
	updateLoad <- -1

	task.Status = Stopped
	task.Result = taskResult

	task.Metrics.EndTime = time.Now().Format("15:04:05.000000")
	LogMetrics(task.Metrics)

	if taskErr != nil {
		task.Error = taskErr.Error()
	} else {
		task.Error = ""
	}

	Tasks[taskId] = task
}

// LogMetrics: Log metrics
func LogMetrics(metricsIn TaskMetrics) {
	metrics, _ := json.Marshal(metricsIn)
	log.Output("METRIC ", 3, string(metrics))
}
