package api

import (
	"cider/functions"
	"cider/log"
	"encoding/json"
	"path"
	"io/ioutil"
	"net/http"
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
	task := Task{Id: taskRequest.Id, Status: Deploying, Data: taskRequest.Data, Function: taskRequest.Function, Result: 0}

	if _, present := tasks[task.Id]; !present {
		writeMessage(&response, http.StatusOK, "Deploying task %s.", task.Id)
		tasks[task.Id] = task

		task.Status = Running
		tasks[task.Id] = task

		task.Result = functions.Map[task.Function](task.Data)
		task.Status = Succeeded
		tasks[task.Id] = task
	} else {
		writeMessage(&response, http.StatusConflict, "Task %s  is already deployed.", task.Id)
	}
}

// getTask: GET /tasks/{id} handler
func getTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := path.Base(request.URL.Path)
	if _, ok := tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else {
		writeStruct(&response, tasks[taskId])
	}
}

// deleteTask: DELETE /tasks/{id} handler
func deleteTask(response http.ResponseWriter, request *http.Request) {
	log.Debug(request.Method, request.URL.Path)
	taskId := path.Base(request.URL.Path)
	if _, ok := tasks[taskId]; !ok {
		response.WriteHeader(http.StatusNotFound)
	} else {
		if tasks[taskId].Status != Running {
			delete(tasks, taskId)
			writeMessage(&response, http.StatusOK, "Removed task %s", taskId)
		} else { // TODO use a channel to abort goroutines running tasks
			writeMessage(&response, http.StatusConflict, "Cannot remove running task!")
		}
	}
}
