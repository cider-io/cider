package api

import (
	"bytes"
	"cider/config"
	"cider/handle"
	"cider/log"
	"net/http"
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

type Task struct {
	Id       string    `json:"id"`
	Status   Status    `json:"status"`
	Function string    `json:"function"`
	Data     []float64 `json:"data"`
	Result   float64   `json:"result"`
}

type TaskRequest struct {
	Id       string    `json:"id"`
	Function string    `json:"function"`
	Data     []float64 `json:"data"`
}

var tasks map[string]Task

func (status Status) String() string {
	return [...]string{"Waiting", "Computing", "Succeeded", "Failed", "Cancelled"}[status]
}

func (s Status) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func Start() {
	tasks = make(map[string]Task)
	router := chi.NewRouter()

	router.Route("/tasks", func(router chi.Router) {
		router.Get("/", getTasks)
		router.Put("/", deployTask)
		router.Route("/{id}", func (router chi.Router) {
			router.Get("/", getTask)
			router.Delete("/", deleteTask)
		})
	})

	log.Info("Serving CIDER API at localhost:" + strconv.Itoa(config.ApiPort))
	handle.Fatal(http.ListenAndServe("localhost:"+strconv.Itoa(config.ApiPort), router))
}
