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
	// custom 404 error
	a.Router.NotFoundHandler = http.HandlerFunc(CustomNotFoundHandler)

	// === global middleware that using through the whole app

	// custom 500 error
	a.Router.Use(middleware.InternalServerErrorHandler)

	// rate and limit inhibit middleware
	a.Router.Use(middleware.RateLimit)
	a.Router.Use(middleware.RequestThrottle)

	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// Глобальные middleware
	a.Router.Use(middleware.InternalServerErrorHandler)
	a.Router.Use(middleware.RateLimit)
	a.Router.Use(middleware.RequestThrottle)

	// Кастомная 404 ошибка
	a.Router.NotFoundHandler = http.HandlerFunc(CustomNotFoundHandler)

	// ==========================================
	// 1. ПОЛЬЗОВАТЕЛЬСКИЙ ИНТЕРФЕЙС (HTML UI)
	// ==========================================
	// Публичная главная страница
	a.Router.HandleFunc("/", handlers.RenderHomeHandler(templateCache)).Methods("GET")

	// Создаем изолированный саб-роутер для Админки
	adminRouter := a.Router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminAuthMiddleware) // Защищаем ВСЕ маршруты внутри /admin

	// Обработка формы добавления товара (теперь путь относительно префикса "/admin")
	adminRouter.HandleFunc("/components/add", handlers.CreateComponentFormHandler).Methods("POST")

	// ==========================================
	// 2. ИНТЕРФЕЙС РАЗРАБОТЧИКА (JSON API)
	// ==========================================
	apiRouter := a.Router.PathPrefix("/api").Subrouter()

	// А. Публичные API-маршруты
	apiRouter.HandleFunc("/components", handlers.GetComponentsHandler).Methods("GET")
	apiRouter.HandleFunc("/components/{id:[0-9]+}", handlers.GetComponentByIDHandler).Methods("GET")

	// Б. Админские API-маршруты
	adminApiRouter := apiRouter.PathPrefix("/components").Subrouter()
	adminApiRouter.Use(middleware.AdminAuthMiddleware)

	adminApiRouter.HandleFunc("", handlers.CreateComponentHandler).Methods("POST")
	adminApiRouter.HandleFunc("/{id:[0-9]+}", handlers.UpdateComponentHandler).Methods("PUT")
	adminApiRouter.HandleFunc("/{id:[0-9]+}", handlers.DeleteComponentHandler).Methods("DELETE")
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
