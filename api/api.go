package api

import (
	"bytes"
	"cider/config"
	"cider/handle"
	"cider/log"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"path"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Results struct { // conventional JSON response for a list of things
	Results interface{} `json:"results"`
}

type Status int

const (
	Waiting Status = iota
	Computing
	Succeeded
	Failed
	Cancelled
)

func (status Status) String() string {
	return [...]string{"Waiting", "Computing", "Succeeded", "Failed", "Cancelled"}[status]
}

func (s Status) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

type Task struct {
	Id       string    `json:"id"`
	Status   Status    `json:"status"`
	Function string    `json:"function"`
	Data     []float64 `json:"data"`
	Result   float64   `json:"result"`
}

type IncomingTask struct {
	Id       string    `json:"id"`
	Function string    `json:"function"`
	Data     []float64 `json:"data"`
}

var tasks map[string]Task

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

var functionMap = map[string](func([]float64) float64){
	"sum": sum,
	"max": max,
	"min": min,
}

func deployTask(response http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		handle.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	var incomingTask IncomingTask
	err = json.Unmarshal(body, &incomingTask)
	log.Debug(incomingTask)
	if err != nil {
		handle.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	var task Task
	task.Id = incomingTask.Id
	task.Status = Waiting
	task.Data = incomingTask.Data
	task.Function = incomingTask.Function
	task.Result = 0

	var msg string
	if _, present := tasks[task.Id]; !present {
		tasks[task.Id] = task
		msg = fmt.Sprintf("Task with ID %s deployed.", task.Id)
		task.Status = Computing
		tasks[task.Id] = task
		task.Result = functionMap[task.Function](task.Data)
		task.Status = Succeeded
		tasks[task.Id] = task
	} else {
		msg = fmt.Sprintf("Task with ID %s already deployed.", task.Id)
		handle.Warning(errors.New(msg))
	}
	writeStruct(&response, Results{Results: msg})
}

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

func taskHandler(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	if request.URL.Path != "/tasks" {
		response.WriteHeader(http.StatusNotFound)
	}
	switch request.Method {
	case http.MethodGet:
		writeStruct(&response, Results{Results: tasks})
	case http.MethodPut:
		deployTask(response, request)
	default:
		response.WriteHeader(http.StatusNotImplemented)
	}
}

func taskIdHandler(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := path.Base(request.URL.Path)
	if _, ok := tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	switch request.Method {
	case http.MethodGet:
		writeStruct(&response, Results{Results: tasks[taskId]})
	case http.MethodDelete:
		if tasks[taskId].Status != Computing {
			delete(tasks, taskId)
			writeStruct(&response, Results{Results: fmt.Sprintf("Removed task with id: %v", taskId)})
			return
		} else {
			writeStruct(&response, Results{Results: "Cannot remove a task in Computing state."})
		}
		writeStruct(&response, Results{Results: fmt.Sprintf("Could not remove task id: %v", taskId)})
	default:
		response.WriteHeader(http.StatusNotFound)
	}
}

func taskIdResultHandler(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := path.Base(path.Dir(request.URL.Path))
	if _, ok := tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	switch request.Method {
	case http.MethodGet:
		switch tasks[taskId].Status {
		case Waiting, Computing:
			writeStruct(&response, Results{Results: tasks[taskId].Status})
		case Succeeded:
			writeStruct(&response, Results{Results: tasks[taskId].Result})
			delete(tasks, taskId)
		case Failed, Cancelled:
			writeStruct(&response, Results{Results: tasks[taskId].Status})
			delete(tasks, taskId)
		}
	default:
		response.WriteHeader(http.StatusNotFound)
	}
}

func Start() {
	tasks = make(map[string]Task)

	chiRouter := chi.NewRouter()

	chiRouter.HandleFunc("/tasks", taskHandler)
	chiRouter.HandleFunc("/tasks/{id}", taskIdHandler)
	chiRouter.HandleFunc("/tasks/{id}/result", taskIdResultHandler)

	log.Info("Serving CIDER API at port", config.ApiPort)
	handle.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.ApiPort), chiRouter))
}
