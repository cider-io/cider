package api

import (
	"cider/log"
	"net/http"
	"strconv"
	"cider/handle"
	"cider/config"
	"encoding/json"
)

type Results struct { // conventional JSON response for a list of things
	Results interface{} `json:"results"`
}

type Task struct {
	id string
}

var tasks []Task

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

func Start() {
	tasks := make([]Task, 0)
	mux := http.NewServeMux() // routes requests to registered handlers

	mux.HandleFunc("/tasks/", func(response http.ResponseWriter, request *http.Request){
		log.Debug(request.Method, request.URL.Path)
		if request.URL.Path != "/tasks/" {
			response.WriteHeader(http.StatusNotFound)
		} else if request.Method == http.MethodGet {
			writeStruct(&response, Results{Results: tasks})
		} else if request.Method == http.MethodPut {
			// TODO deployTask()
			response.WriteHeader(http.StatusNotImplemented)
		} else {
			response.WriteHeader(http.StatusNotImplemented)
		}
	})

	log.Info("Serving CIDER API at port", config.ApiPort)
	handle.Fatal(http.ListenAndServe(":" + strconv.Itoa(config.ApiPort), mux))
}
