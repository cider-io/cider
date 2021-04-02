
package api

import (
	"cider/handle"
	"net/http"
	"encoding/json"
	"fmt"
)

// writeMessage: Write headers + formatted message to the response
func writeMessage(response *http.ResponseWriter, status int, format string, a...interface{}) {
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
