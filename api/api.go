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

var Tasks map[string]Task

var updateLoad chan int

func handleLoadUpdates() {
	for loadChange := range updateLoad {
		exportapi.Load += loadChange
		if loadChange < 0 {
			// Reputation updates with the completion of tasks
			exportapi.Reputation -= loadChange
		}
	}
}

func Start() {
	Tasks = make(map[string]Task)

	// allow API handlers to update the node's load w/o a locking mechanism
	updateLoad = make(chan int)
	go handleLoadUpdates() // we don't need a WaitGroup here since ListenAndServe blocks

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

	// TODO serve API over HTTPS
	// ListenAndServe is a blocking call-- shouldn't exit!
	handle.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.ApiPort), router))
}
