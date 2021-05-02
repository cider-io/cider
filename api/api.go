package api

import (
	"cider/config"
	"cider/exportapi"
	"cider/handle"
	"cider/log"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func Start() {
	exportapi.Tasks = make(map[string]exportapi.Task)
	router := chi.NewRouter()

	router.Route("/tasks", func(router chi.Router) {
		router.Get("/", getTasks) // TODO filtering, authorized endpoint
		router.Put("/", deployTask)
		router.Route("/{id}", func(router chi.Router) {
			router.Get("/", getTask)
			router.Put("/", abortTask)
			router.Delete("/", deleteTask)
			router.Get("/result", getTaskResult)
		})
	})

	log.Info("Serving CIDER API at port " + strconv.Itoa(config.ApiPort))
	handle.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.ApiPort), router))
}
