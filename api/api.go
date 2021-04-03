package api

import (
	"cider/config"
	"cider/handle"
	"cider/log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var tasks map[string]Task

func Start() {
	tasks = make(map[string]Task)
	router := chi.NewRouter()

	router.Route("/tasks", func(router chi.Router) {
		router.Get("/", getTasks) // TODO filtering, authorized endpoint
		router.Put("/", deployTask)
		router.Route("/{id}", func (router chi.Router) {
			router.Get("/", getTask)
			router.Put("/", abortTask)
			router.Delete("/", deleteTask)
			router.Get("/result", getTaskResult)
		})
	})

	log.Info("Serving CIDER API at localhost:" + strconv.Itoa(config.ApiPort))
	handle.Fatal(http.ListenAndServe("localhost:"+strconv.Itoa(config.ApiPort), router))
}
