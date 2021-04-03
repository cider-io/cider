package api

import (
	"cider/functions"
	"cider/log"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// getTasks: GET /tasks handler
func getTasks(response http.ResponseWriter, request *http.Request) {
	writeStruct(&response, tasks)
}

// deployTasks: PUT /tasks handler
func deployTask(response http.ResponseWriter, request *http.Request) {
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

	// generate a globally unique id for a task if it doesn't exist
	taskId, err := generateUUID()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	tasks[taskId] = Task{Id: taskId, Status: Deploying, Data: taskRequest.Data, Function: taskRequest.Function, Result: 0, Abort: make(chan bool)}

	go func(taskId string) { // async launch the function
		task := tasks[taskId]
		task.Status = Running
		tasks[taskId] = task

		taskResult, taskErr := functions.Map[task.Function](task.Data, task.Abort)
		task.Status = Stopped
		task.Result = taskResult
		if taskErr != nil {
			task.Error = taskErr.Error()
		} else {
			task.Error = ""
		}
		
		tasks[taskId] = task
	}(taskId)

	// FIXME: this needs to be sent confidentially (via HTTPS)
	writeStruct(&response, tasks[taskId])
}

// getTask: GET /tasks/{id} handler
func getTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else {
		writeStruct(&response, tasks[taskId])
	}
}

// abortTask: PUT /tasks/{id} handler
// TODO abortTask -> updateTask: add URL parameter for action=abort/pause/resume etc.
func abortTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else {
		if tasks[taskId].Status != Stopped {
			tasks[taskId].Abort <- true	
		} 
		writeMessage(&response, http.StatusOK, "Aborted task %s", taskId)
	}
}

// deleteTask: DELETE /tasks/{id} handler
func deleteTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else {
		if tasks[taskId].Status != Stopped {
			writeMessage(&response, http.StatusConflict, "Cannot delete a running task.")
		} 
		writeMessage(&response, http.StatusOK, "Deleted task %s", taskId)
	}
}

// getTaskResult: GET /tasks/{id}/result handler
func getTaskResult(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := chi.URLParam(request, "id")
	if _, ok := tasks[taskId]; !ok {
		if tasks[taskId].Status != Stopped {
			writeMessage(&response, http.StatusNotFound, "Result is not available yet.")
		} else {
			writeStruct(&response, TaskResult{Result: tasks[taskId].Result, Error: tasks[taskId].Error})
		}
	}
}
