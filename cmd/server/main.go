package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"techstore/internal/middleware"

	"techstore/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type App struct {
	Router        *mux.Router
	DB            *sql.DB
	TemplateCache *template.Template
}

func (a *App) Initialize(dbHost, dbPort, dbUser, dbPassword, dbName string) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	a.DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = a.DB.Ping()
	if err != nil {
		log.Fatal("DB Connection failed: ", err)
	}
	log.Println("Successful connected to PostgreSQL!")

	a.createTables()

	a.LoadTemplates()

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// Глобальные middleware
	a.Router.Use(middleware.InternalServerErrorHandler)
	a.Router.Use(middleware.RateLimit)
	a.Router.Use(middleware.RequestThrottle)

	// Кастомная 404 ошибка
	a.Router.NotFoundHandler = http.HandlerFunc(CustomNotFoundHandler)

	compHandler := &handlers.ComponentHandler{DB: a.DB, Tmpl: a.TemplateCache}

	// ==========================================
	// 1. ПОЛЬЗОВАТЕЛЬСКИЙ ИНТЕРФЕЙС (HTML UI)
	// ==========================================

	adminRouter := a.Router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminAuthMiddleware)

	adminRouter.HandleFunc("/components/add", compHandler.CreateComponentFormHandler).Methods("POST")

	// ==========================================
	// 2. ИНТЕРФЕЙС РАЗРАБОТЧИКА (JSON API)
	// ==========================================
	apiRouter := a.Router.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("/components", compHandler.GetComponentsHandler).Methods("GET")
	apiRouter.HandleFunc("/components/{id:[0-9]+}", compHandler.GetComponentByIDHandler).Methods("GET")

	adminApiRouter := apiRouter.PathPrefix("/components").Subrouter()
	adminApiRouter.Use(middleware.AdminAuthMiddleware)

	adminApiRouter.HandleFunc("", compHandler.CreateComponentHandler).Methods("POST")
	adminApiRouter.HandleFunc("/{id:[0-9]+}", compHandler.UpdateComponentHandler).Methods("PUT")
	adminApiRouter.HandleFunc("/{id:[0-9]+}", compHandler.DeleteComponentHandler).Methods("DELETE")
}

func (a *App) createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS components (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		manufacturer VARCHAR(100) NOT NULL,
		category VARCHAR(50) NOT NULL,
		price NUMERIC(10, 2) NOT NULL
	);`

	_, err := a.DB.Exec(query)
	if err != nil {
		log.Fatal("Ошибка создания таблицы: ", err)
	}
}

func (a *App) LoadTemplates() {
	templates, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatal("Ошибка загрузки шаблонов: ", err)
	}
	a.TemplateCache = templates
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
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env not found, using system defaults.")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbUser == "" || dbPass == "" || dbName == "" {
		log.Fatal("Fatal error: System defaults not found in environment variables.")
	}

	app := &App{}
	app.Initialize(dbHost, dbPort, dbUser, dbPass, dbName)
	app.Run(":8080")
}
