package api

import (
	"cider/functions"
	"cider/handle"
	"cider/log"
	"encoding/json"
	"errors"
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
		handle.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	var taskRequest TaskRequest
	err = json.Unmarshal(body, &taskRequest)
	log.Debug(taskRequest)
	if err != nil {
		handle.Warning(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	task := Task{Id: taskRequest.Id, Status: Waiting, Data: taskRequest.Data, Function: taskRequest.Function, Result: 0}

	var msg string
	if _, present := tasks[task.Id]; !present {
		tasks[task.Id] = task
		writeMessage(&response, http.StatusOK, "Task with ID %s deployed.", task.Id)
		task.Status = Computing
		tasks[task.Id] = task
		task.Result = functions.Map[task.Function](task.Data)
		task.Status = Succeeded
		tasks[task.Id] = task
	} else {
		writeMessage(&response, http.StatusConflict, "Task with ID %s already deployed.", task.Id)
		handle.Warning(errors.New(msg))
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
		if tasks[taskId].Status != Computing {
			delete(tasks, taskId)
			writeMessage(&response, http.StatusOK, "Removed task with id: %v", taskId)
		} else {
			writeMessage(&response, http.StatusConflict, "Cannot remove a task in Computing state.")
		}
	}
}
