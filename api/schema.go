package api

import (
	"cider/functions"
	"bytes"
	"errors"
	"strings"
)

type TaskStatus int

const (
	Deploying TaskStatus = iota
	Running
	Stopped
)

var stringToStatus = map[string]TaskStatus{"Deploying": Deploying, "Running": Running, "Stopped": Stopped}

func (status TaskStatus) String() string {
	return [...]string{"Deploying", "Running", "Stopped"}[status]
}

func (status TaskStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(status.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (status *TaskStatus) UnmarshalJSON(data []byte) error {
	taskStatus, ok := stringToStatus[strings.Trim(string(data), "\"")] // strip " "
	*status = taskStatus
	if !ok {
		return errors.New("Illegal status string")
	}
	return nil
} 

type TaskMetrics struct {
	Id        string  `json:"id"`
	Function  string  `json:"function"`
	DataSize  float64 `json:"datasize"`
	StartTime string  `json:"start"`
	EndTime   string  `json:"end"`
}

type Task struct {
	Id       string      	 `json:"id"`
	Status   TaskStatus  	 `json:"status"`
	Function string      	 `json:"function"`
	Data     functions.Data  `json:"-"` // ignore all fields other than Id/Status in the JSON representation
	Result   functions.Data  `json:"-"`
	Error    string      	 `json:"-"`
	Abort    chan bool   	 `json:"-"`
	Metrics  TaskMetrics 	 `Json:"-"`
}

type TaskRequest struct {
	Function string         `json:"function"`
	Data     functions.Data `json:"data"`
}

type TaskResult struct {
	Result functions.Data `json:"result"`
	Error  string         `json:"error"`
}
