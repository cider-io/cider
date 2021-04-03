package api

import (
	"bytes"
)

type TaskStatus int
const (
	Deploying TaskStatus = iota
	Running
	Stopped
)

func (status TaskStatus) String() string {
	return [...]string{"Deploying", "Running", "Stopped"}[status]
}

func (status TaskStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(status.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

type Task struct {
	Id string `json:"id"`
	Status TaskStatus `json:"status"`
	Function string `json:"-"` // ignore all fields other than Id/Status in the JSON representation
	Data []float64 `json:"-"`
	Result float64 `json:"-"`
	Error string `json:"-"`
	Abort chan bool `json:"-"`
}

type TaskRequest struct {
	Function string `json:"function"`
	Data []float64 `json:"data"`
}

type TaskResult struct {
	Result float64 `json:"result"`
	Error string `json:"error"`
}
