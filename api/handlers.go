package api

import (
	"cider/log"
	"cider/gossip"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// getTasks: GET /tasks handler
func getTasks(response http.ResponseWriter, request *http.Request) {
	writeStruct(&response, Tasks)
}

// deployTasks: PUT /tasks handler
func deployTask(response http.ResponseWriter, request *http.Request) {
	requestSourceIp, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		log.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	var taskRequest TaskRequest
	err = json.Unmarshal(body, &taskRequest)
	if err != nil {
		log.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Debug(request.Method, request.URL.Path, taskRequest)

	if isLocalIp(requestSourceIp) {
		// TODO deploy the task locally if possible
		// otherwise, send the task to a remote node
	} else if _, ok := gossip.Self.MembershipList[requestSourceIp]; !ok {
		log.Warning("Denied request from unknown node", requestSourceIp)
		response.WriteHeader(http.StatusNotFound)
		return
	}

	// generate a globally unique id for a task if it doesn't exist
	taskId, err := generateUUID()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	Tasks[taskId] = Task{Id: taskId, Status: Deploying, Data: taskRequest.Data, Function: taskRequest.Function, Result: 0, Abort: make(chan bool), Metrics: TaskMetrics{Id: taskId, Function: taskRequest.Function, StartTime: time.Now().Format("15:04:05.000000")}}

	go runTask(taskId) // async launch the function

	// FIXME this needs to be sent confidentially (via HTTPS)
	writeStruct(&response, Tasks[taskId])
}

// getTask: GET /tasks/{id} handler
func getTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := Tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else {
		writeStruct(&response, Tasks[taskId])
	}
}

// abortTask: PUT /tasks/{id} handler
// TODO abortTask -> updateTask: add URL parameter for action=abort/pause/resume etc.
func abortTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := Tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else if Tasks[taskId].Status == Stopped {
		writeMessage(&response, http.StatusConflict, "This task has already stopped.")
	} else {
		Tasks[taskId].Abort <- true
		writeMessage(&response, http.StatusOK, "Aborted task %s", taskId)
	}
}

// deleteTask: DELETE /tasks/{id} handler
func deleteTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := Tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else if Tasks[taskId].Status != Stopped {
		writeMessage(&response, http.StatusConflict, "Cannot delete a running task; please abort it first.")
	} else {
		delete(Tasks, taskId)
		writeMessage(&response, http.StatusOK, "Deleted task %s", taskId)
	}
}

// getTaskResult: GET /tasks/{id}/result handler
func getTaskResult(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	log.Debug(taskId)
	if _, ok := Tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else if Tasks[taskId].Status != Stopped {
		writeMessage(&response, http.StatusNotFound, "Result is not available yet.")
	} else {
		writeStruct(&response, TaskResult{Result: Tasks[taskId].Result, Error: Tasks[taskId].Error})
	}
}
