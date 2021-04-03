package api

import (
	"bytes"
)

type TaskRequest struct {
	Id string `json:"id"`
	Function string `json:"function"`
	Data []float64 `json:"data"`
}

type TaskStatus int
const (
	Deploying TaskStatus = iota
	Running
	Succeeded
	Failed
)

func (status TaskStatus) String() string {
	return [...]string{"Deploying", "Running", "Succeeded", "Failed"}[status]
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
	Function string `json:"function"`
	Data []float64 `json:"data"`
	Result float64 `json:"result"`
}
