package exportapi

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

type TaskMetrics struct {
	Id        string  `json:"id"`
	Function  string  `json:"function"`
	DataSize  float64 `json:"datasize"`
	StartTime string  `json:"start"`
	EndTime   string  `json:"end"`
}
type Task struct {
	Id       string      `json:"id"`
	Status   TaskStatus  `json:"status"`
	Function string      `json:"function"`
	Data     []float64   `json:"-"` // ignore all fields other than Id/Status in the JSON representation
	Result   float64     `json:"-"`
	Error    string      `json:"-"`
	Abort    chan bool   `json:"-"`
	Metrics  TaskMetrics `Json:"-"`
}

type TaskRequest struct {
	Function string    `json:"function"`
	Data     []float64 `json:"data"`
}

type TaskResult struct {
	Result float64 `json:"result"`
	Error  string  `json:"error"`
}
