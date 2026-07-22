package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
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

/*
	generateSKU generates unique SKU for component bases on its category and manufacturer

(first 3 letter each) and adds 4 random bytes.
*/
func generateSKU(category, manufacturer string) string {
	catRunes := []rune(strings.ToUpper(category))
	manRunes := []rune(strings.ToUpper(manufacturer))

	catPrefix := string(catRunes)
	if len(catRunes) >= 3 {
		catPrefix = string(catRunes[:3])
	}

	manPrefix := string(manRunes)
	if len(manRunes) >= 3 {
		manPrefix = string(manRunes[:3])
	}

	b := make([]byte, 2)
	rand.Read(b)
	randStr := fmt.Sprintf("%X", b)

	return fmt.Sprintf("%s-%s-%s", catPrefix, manPrefix, randStr)
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
	rows, err := ch.DB.Query("SELECT id, sku, name, manufacturer, category, price, rating, stock, image_path, specs FROM components ORDER BY id DESC")
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var components []models.Component
	for rows.Next() {
		var comp models.Component
		if err := rows.Scan(&comp.ID, &comp.SKU, &comp.Name, &comp.Manufacturer, &comp.Category, &comp.Price, &comp.Rating, &comp.Stock, &comp.ImagePath, &comp.Specs); err != nil {
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

func (ch *ComponentHandler) RenderComponentDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sku := vars["sku"]

	var comp models.Component
	query := `SELECT id, sku, name, manufacturer, category, price, description, rating, stock, image_path, specs 
	          FROM components WHERE sku = $1`
	err := ch.DB.QueryRow(query, sku).Scan(&comp.ID, &comp.SKU, &comp.Name, &comp.Manufacturer, &comp.Category,
		&comp.Price, &comp.Description, &comp.Rating, &comp.Stock, &comp.ImagePath, &comp.Specs,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Component not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	var parsedSpecs map[string]interface{}
	if len(comp.Specs) > 0 {
		json.Unmarshal(comp.Specs, &parsedSpecs)
	}

	data := struct {
		User      *models.User
		Component models.Component
		Specs     map[string]interface{}
	}{
		Component: comp,
		Specs:     parsedSpecs,
	}

	session, _ := ch.Store.Get(r, sessionName)
	if userID, ok := session.Values["user_id"].(int); ok && userID != 0 {
		var localUser models.User
		err := ch.DB.QueryRow("SELECT name, email FROM users WHERE id = $1", userID).Scan(&localUser.Username, &localUser.Email)
		if err == nil {
			data.User = &localUser
		}
	}

	err = ch.Tmpl.ExecuteTemplate(w, "component_detail", data)
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
	stockStr := r.FormValue("stock")
	specsStr := r.FormValue("specs")

	if len(name) < 3 || manufacturer == "" || category == "" || priceStr == "" || stockStr == "" {
		http.Error(w, "Missing mandatory fields", http.StatusBadRequest)
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price format", http.StatusBadRequest)
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		http.Error(w, "Invalid stock format", http.StatusBadRequest)
		return
	}

	sku := generateSKU(category, manufacturer)

	if strings.TrimSpace(specsStr) == "" {
		specsStr = "{}"
	}

	if !json.Valid([]byte(specsStr)) {
		http.Error(w, "Invalid JSON format in specs", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO components (sku, name, manufacturer, category, price, stock, specs)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = ch.DB.Exec(query, sku, name, manufacturer, category, price, stock, json.RawMessage(specsStr))
	if err != nil {
		http.Error(w, "Error creating component", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/component", http.StatusSeeOther)
}

func (ch *ComponentHandler) CreateComponentHandler(w http.ResponseWriter, r *http.Request) {
	var comp models.Component
	if err := json.NewDecoder(r.Body).Decode(&comp); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	comp.SKU = generateSKU(comp.Category, comp.Manufacturer)

	if len(comp.Specs) == 0 {
		comp.Specs = json.RawMessage("{}")
	}

	query := `INSERT INTO components (sku, name, manufacturer, category, price, rating, stock, image_path, specs)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	err := ch.DB.QueryRow(query, comp.SKU, comp.Name, comp.Manufacturer, comp.Category, comp.Price, comp.Rating, comp.Stock, comp.ImagePath, comp.Specs).Scan(&comp.ID)
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
	err := ch.Tmpl.ExecuteTemplate(w, "admin", nil)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
