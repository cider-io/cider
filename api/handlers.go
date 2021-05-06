package api

import (
	"cider/exportapi"
	"cider/log"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// getTasks: GET /tasks handler
func getTasks(response http.ResponseWriter, request *http.Request) {
	writeStruct(&response, exportapi.Tasks)
}

// deployTasks: PUT /tasks handler
func deployTask(response http.ResponseWriter, request *http.Request) {
	ip, _, _ := net.SplitHostPort(request.RemoteAddr)

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	var taskRequest exportapi.TaskRequest
	err = json.Unmarshal(body, &taskRequest)
	if err != nil {
		log.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Debug(request.Method, request.URL.Path, taskRequest)

	if isLocalIp(ip) {
		suitableIp := findSuitableComputeNode(taskRequest)
		if !isLocalIp(suitableIp) {
			// TODO: We need to figure out how to send
			//  the request to the suitable remote node.
			//  We should return whatever response received
			//  from the remote node.
		}
	} else if !isValidRemote(ip) {
		log.Warning("request from invalid remote ip:", ip)
		response.WriteHeader(http.StatusNotFound)
		return
	}

	// generate a globally unique id for a task if it doesn't exist
	taskId, err := generateUUID()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	exportapi.Tasks[taskId] = exportapi.Task{Id: taskId, Status: exportapi.Deploying, Data: taskRequest.Data, Function: taskRequest.Function, Result: 0, Abort: make(chan bool), Metrics: exportapi.TaskMetrics{Id: taskId, Function: taskRequest.Function, StartTime: time.Now().Format("15:04:05.000000")}}

	go completeTask(taskId) // async launch the function

	// FIXME: this needs to be sent confidentially (via HTTPS)
	writeStruct(&response, exportapi.Tasks[taskId])
}

// getTask: GET /tasks/{id} handler
func getTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := exportapi.Tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else {
		writeStruct(&response, exportapi.Tasks[taskId])
	}
}

// abortTask: PUT /tasks/{id} handler
// TODO abortTask -> updateTask: add URL parameter for action=abort/pause/resume etc.
func abortTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := exportapi.Tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else if exportapi.Tasks[taskId].Status == exportapi.Stopped {
		writeMessage(&response, http.StatusConflict, "This task has already stopped.")
	} else {
		exportapi.Tasks[taskId].Abort <- true
		writeMessage(&response, http.StatusOK, "Aborted task %s", taskId)
	}
}

// deleteTask: DELETE /tasks/{id} handler
func deleteTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := exportapi.Tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else if exportapi.Tasks[taskId].Status != exportapi.Stopped {
		writeMessage(&response, http.StatusConflict, "Cannot delete a running task; please abort it first.")
	} else {
		delete(exportapi.Tasks, taskId)
		writeMessage(&response, http.StatusOK, "Deleted task %s", taskId)
	}
}

// getTaskResult: GET /tasks/{id}/result handler
func getTaskResult(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	log.Debug(taskId)
	if _, ok := exportapi.Tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else if exportapi.Tasks[taskId].Status != exportapi.Stopped {
		writeMessage(&response, http.StatusNotFound, "Result is not available yet.")
	} else {
		writeStruct(&response, exportapi.TaskResult{Result: exportapi.Tasks[taskId].Result, Error: exportapi.Tasks[taskId].Error})
	}
}
