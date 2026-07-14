package main

import (
	"log"
	"net/http"

	"techstore/internal/handlers"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

func (a *App) Initialize() {
	a.Router = mux.NewRouter()
	a.InitializeRoutes()
}

func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("TechStore API is up and running"))
	}).Methods("GET")

	compRouter := a.Router.PathPrefix("/components").Subrouter()

	compRouter.HandleFunc("", handlers.GetComponentsHandler).Methods("GET")

	compRouter.HandleFunc("/{id:[0-9]+}", handlers.GetComponentByIDHandler).Methods("GET")
}

func (a *App) Run(addr string) {
	log.Printf("Listening on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func main() {
	app := &App{}
	app.Initialize()
	app.Run(":8080")
}
