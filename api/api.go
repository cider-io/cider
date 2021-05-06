package api

import (
	"cider/config"
	"cider/handle"
	"cider/log"
	"cider/exportapi"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

var Tasks map[string]Task

var updateLoad chan int

func handleLoadUpdates() {
	for loadChange := range updateLoad {
		exportapi.Load += loadChange
	}
}

func Start() {
	Tasks = make(map[string]Task)

	// allow API handlers to update the node's load w/o a locking mechanism
	updateLoad = make(chan int)
	go handleLoadUpdates()

	router := chi.NewRouter()
	router.Route("/tasks", func(router chi.Router) {
		router.Get("/", getTasks)
		router.Put("/", deployTask)
		router.Route("/{id}", func(router chi.Router) {
			router.Get("/", getTask)
			router.Put("/", abortTask)
			router.Delete("/", deleteTask)
			router.Get("/result", getTaskResult)
		})
	})

	log.Info("Serving CIDER API at port " + strconv.Itoa(config.ApiPort))

	// ListenAndServe is a blocking call-- shouldn't exit!
	handle.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.ApiPort), router))
}
