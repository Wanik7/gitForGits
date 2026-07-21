package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"techstore/internal/handlers"
	"techstore/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type App struct {
	Router        *mux.Router
	DB            *sql.DB
	TemplateCache *template.Template
	Store         *sessions.FilesystemStore
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
	// Global middleware
	a.Router.Use(middleware.InternalServerErrorHandler)
	a.Router.Use(middleware.RateLimit)

	// Custom 404 error
	a.Router.NotFoundHandler = http.HandlerFunc(CustomNotFoundHandler)

	// ==========================================
	// 1. USER INTERFACE (HTML UI)
	// ==========================================

	compHandler := &handlers.ComponentHandler{
		DB:    a.DB,
		Tmpl:  a.TemplateCache,
		Store: a.Store,
	}
	a.Router.HandleFunc("/", compHandler.RenderHomeHandler).Methods("GET")
	a.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	userHandler := &handlers.UserHandler{DB: a.DB, Tmpl: a.TemplateCache, Store: a.Store}

	a.Router.HandleFunc("/register", userHandler.RenderRegisterForm).Methods("GET")
	a.Router.HandleFunc("/register", userHandler.RegisterUser).Methods("POST")

	a.Router.HandleFunc("/login", userHandler.RenderLoginForm).Methods("GET")
	a.Router.HandleFunc("/login", userHandler.LoginUser).Methods("POST")

	a.Router.HandleFunc("/logout", userHandler.LogoutUser).Methods("GET") /* using GET instead of POST
	   because I'm using simple a tag in html that supports only GET method */

	// ==========================================
	// 2.1 DEV INTERFACE (ADMIN)
	// ==========================================

	adminRouter := a.Router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.RequireAdminMiddleware(a.Store))

	adminRouter.HandleFunc("/component", compHandler.RenderAdminHandler).Methods("GET")
	adminRouter.HandleFunc("/component", compHandler.CreateComponentFormHandler).Methods("POST")

	// ==========================================
	// 2.2 DEV INTERFACE (JSON API)
	// ==========================================
	apiRouter := a.Router.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("/components", compHandler.GetComponentsHandler).Methods("GET")
	apiRouter.HandleFunc("/components/{id:[0-9]+}", compHandler.GetComponentByIDHandler).Methods("GET")

	adminApiRouter := apiRouter.PathPrefix("/components").Subrouter()
	adminApiRouter.Use(middleware.RequireAdminMiddleware(a.Store))

	adminApiRouter.HandleFunc("", compHandler.CreateComponentHandler).Methods("POST")
	adminApiRouter.HandleFunc("/{id:[0-9]+}", compHandler.UpdateComponentHandler).Methods("PUT")
	adminApiRouter.HandleFunc("/{id:[0-9]+}", compHandler.DeleteComponentHandler).Methods("DELETE")
}

func (a *App) createTables() {
	query := `
	--- USERS TABLE ---
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'customer',
		created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	--- COMPONENTS TABLE ---
	CREATE TABLE IF NOT EXISTS components (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		manufacturer VARCHAR(100) NOT NULL,
		category VARCHAR(50) NOT NULL,
		price NUMERIC(10, 2) NOT NULL,
	    description TEXT,
	    rating NUMERIC(3, 2) NOT NULL CHECK ( rating >= 1 AND rating <= 5 ),
	    stock INTEGER DEFAULT 0
	);

	--- REVIEWS TABLE ---
	CREATE TABLE IF NOT EXISTS reviews (
	    id SERIAL PRIMARY KEY,
	    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	    component_id INTEGER NOT NULL REFERENCES components(id) ON DELETE CASCADE,
	    rating INTEGER NOT NULL,
	    body TEXT,
	    likes INTEGER NOT NULL DEFAULT 0,
	    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	--- COMMENTS TABLE ---
	CREATE TABLE IF NOT EXISTS comments (
	    id SERIAL PRIMARY KEY,
	    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	    component_id INTEGER NOT NULL REFERENCES components(id) ON DELETE CASCADE,
	    parent_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
	    body TEXT NOT NULL,
	    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := a.DB.Exec(query)
	if err != nil {
		log.Fatal("Table creation failed: ", err)
	}
}

func (a *App) LoadTemplates() {
	templates, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatal("Template load failed: ", err)
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

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPass == "" || dbName == "" {
		log.Fatal("Fatal error: System defaults not found in environment variables.")
	}

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("Fatal error: Session key not found in environment variables.")
	}

	store := sessions.NewFilesystemStore("", []byte(sessionKey))

	app := &App{
		Store: store,
	}

	app.Initialize(dbHost, dbPort, dbUser, dbPass, dbName)
	app.Run(":8080")
}
