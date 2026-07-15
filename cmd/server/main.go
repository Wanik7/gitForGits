package main

import (
	"html/template"
	"log"
	"net/http"
	"techstore/internal/middleware"

	"techstore/internal/handlers"

	"github.com/gorilla/mux"
)

var templateCache *template.Template

func LoadTemplates() {
	templates, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatal("Error loading templates: ", err)
	}

	templateCache = templates
	log.Println("Templates loaded successfully")
}

type App struct {
	Router *mux.Router
}

func (a *App) Initialize() {
	a.Router = mux.NewRouter()

	a.Router.Use(middleware.InternalServerErrorHandler)
	a.Router.Use(middleware.RateLimit)
	a.Router.Use(middleware.RequestThrottle)

	a.Router.NotFoundHandler = http.HandlerFunc(CustomNotFoundHandler)
	a.InitializeRoutes()
}

func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("TechStore API is up and running"))
	}).Methods("GET")

	publicCompRouter := a.Router.PathPrefix("/components").Subrouter()

	publicCompRouter.HandleFunc("/ui", handlers.RenderHomeHandler(templateCache)).Methods("GET")

	compRouter := a.Router.PathPrefix("/components").Subrouter()

	compRouter.HandleFunc("", handlers.GetComponentsHandler).Methods("GET")
	compRouter.HandleFunc("/{id:[0-9]+}", handlers.GetComponentByIDHandler).Methods("GET")

	compRouter.HandleFunc("", handlers.CreateComponentHandler).Methods("POST")
	compRouter.HandleFunc("/{id:[0-9]+}", handlers.UpdateComponentHandler).Methods("PUT")
	compRouter.HandleFunc("/{id:[0-9]+}", handlers.DeleteComponentHandler).Methods("DELETE")
}

func (a *App) Run(addr string) {
	log.Printf("Listening on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func CustomNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Component not found. Visit our catalog to explore other items."))
}

func main() {
	LoadTemplates()

	app := &App{}
	app.Initialize()
	app.Run(":8080")
}
