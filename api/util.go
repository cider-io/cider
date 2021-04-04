package api

import (
	"cider/config"
	"cider/handle"
	"cider/log"
	"cider/util"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// writeMessage: Write headers + formatted message to the response
func writeMessage(response *http.ResponseWriter, status int, format string, a ...interface{}) {
	message := []byte(fmt.Sprintf(format, a...))

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

func generateUUID() (string, error) {
	input, err := time.Now().MarshalBinary()
	if err != nil {
		return "", err
	}

	rand.Seed(time.Now().UnixNano())
	nonce := make([]byte, config.NonceLength)
	rand.Read(nonce)
	input = append(input, nonce...)

	nodeIpAddress, err := util.GetIpAddress()
	if err != nil {
		return "", err
	}
	input = append(input, ([]byte(nodeIpAddress))...)

	return fmt.Sprintf("%x", sha256.Sum256(input)), nil
}
func MetricsLog(metricsIn TaskMetrics) {

	metrics, _ := json.Marshal(metricsIn)
	log.Output("METRIC ", 3, string(metrics))
}
