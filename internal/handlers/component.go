package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"techstore/pkg/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const sessionName = "techstore-session"

// parseIDParam extracts and parses the "id" path parameter from the request URL.
func parseIDParam(r *http.Request) (int, error) {
	return strconv.Atoi(mux.Vars(r)["id"])
}

// respondJSON writes a JSON response with the given status code and data.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

type ComponentHandler struct {
	DB    *sql.DB
	Tmpl  *template.Template
	Store *sessions.FilesystemStore
}

type PageData struct {
	User       *models.User
	Components []models.Component
}

func (ch *ComponentHandler) RenderHomeHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := ch.DB.Query("SELECT id, name, manufacturer, category, price, rating, stock FROM components ORDER BY id DESC")
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var components []models.Component
	for rows.Next() {
		var comp models.Component
		if err := rows.Scan(&comp.ID, &comp.Name, &comp.Manufacturer, &comp.Category, &comp.Price, &comp.Rating, &comp.Stock); err != nil {
			http.Error(w, "Error reading data", http.StatusInternalServerError)
			return
		}
		components = append(components, comp)
	}

	data := PageData{
		Components: components,
	}

	session, err := ch.Store.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Error processing session", http.StatusInternalServerError)
		return
	}

	if UserID, ok := session.Values["user_id"].(int); ok && UserID != 0 {
		var localUser models.User
		err := ch.DB.QueryRow("SELECT name, email FROM users WHERE id = $1", UserID).Scan(&localUser.Username, &localUser.Email)
		if err == nil {
			data.User = &localUser
		}
	}

	err = ch.Tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (ch *ComponentHandler) GetComponentsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := ch.DB.Query("SELECT id, name, manufacturer, category, price FROM components ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var components []models.Component = make([]models.Component, 0)
	for rows.Next() {
		var comp models.Component
		if err := rows.Scan(&comp.ID, &comp.Name, &comp.Manufacturer, &comp.Category, &comp.Price); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		components = append(components, comp)
	}
	respondJSON(w, http.StatusOK, components)
}

func (ch *ComponentHandler) GetComponentByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, "Invalid component ID", http.StatusBadRequest)
		return
	}

	var comp models.Component
	query := "SELECT id, name, manufacturer, category, price FROM components WHERE id = $1"

	err = ch.DB.QueryRow(query, id).Scan(&comp.ID, &comp.Name, &comp.Manufacturer, &comp.Category, &comp.Price)
	if err == sql.ErrNoRows {
		http.Error(w, "Component not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, comp)
}

func (ch *ComponentHandler) CreateComponentFormHandler(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")
	manufacturer := r.FormValue("manufacturer")
	category := r.FormValue("category")
	priceStr := r.FormValue("price")
	ratingStr := r.FormValue("rating")
	stockStr := r.FormValue("stock")

	if len(name) < 3 || manufacturer == "" || category == "" || priceStr == "" || ratingStr == "" || stockStr == "" {
		http.Error(w, "Missing mandatory fields", http.StatusBadRequest)
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price format", http.StatusBadRequest)
		return
	}

	rating, err := strconv.ParseFloat(ratingStr, 64)
	if err != nil {
		http.Error(w, "Invalid rating format", http.StatusBadRequest)
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		http.Error(w, "Invalid stock format", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO components (name, manufacturer, category, price, rating, stock)
			VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = ch.DB.Exec(query, name, manufacturer, category, price, rating, stock)
	if err != nil {
		http.Error(w, "Error creating component", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (ch *ComponentHandler) CreateComponentHandler(w http.ResponseWriter, r *http.Request) {
	var comp models.Component
	if err := json.NewDecoder(r.Body).Decode(&comp); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO components (name, manufacturer, category, price, rating, stock)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := ch.DB.QueryRow(query, comp.Name, comp.Manufacturer, comp.Category, comp.Price, comp.Rating, comp.Stock).Scan(&comp.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, comp)
}

func (ch *ComponentHandler) UpdateComponentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, "Invalid component ID", http.StatusBadRequest)
		return
	}

	var comp models.Component
	if err := json.NewDecoder(r.Body).Decode(&comp); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `UPDATE components SET name = $1, manufacturer = $2, category = $3, price = $4 WHERE id = $5`
	result, err := ch.DB.Exec(query, comp.Name, comp.Manufacturer, comp.Category, comp.Price, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Component not found", http.StatusNotFound)
		return
	}

	comp.ID = id
	respondJSON(w, http.StatusOK, comp)
}

func (ch *ComponentHandler) DeleteComponentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, "Invalid component ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM components WHERE id = $1`
	result, err := ch.DB.Exec(query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Component not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ch *ComponentHandler) RenderAdminHandler(w http.ResponseWriter, r *http.Request) {
	err := ch.Tmpl.ExecuteTemplate(w, "admin.html", nil)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
