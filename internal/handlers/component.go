package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"techstore/pkg/models"

	"github.com/gorilla/mux"
)

var componentDB = []models.Component{
	{ID: 1, Name: "Ryzen 5 5600", Manufacturer: "AMD", Category: "CPU", Price: 135.50},
	{ID: 2, Name: "GeForce RTX 4060", Manufacturer: "NVIDIA", Category: "GPU", Price: 299.99},
}

func GetComponentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(componentDB)
}

func GetComponentByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid component ID", http.StatusBadRequest)
	}

	for _, component := range componentDB {
		if component.ID == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(component)
			return
		}
	}

	http.Error(w, "Component not found", http.StatusNotFound)
}

func CreateComponentFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	manufacturer := r.FormValue("manufacturer")
	category := r.FormValue("category")
	priceStr := r.FormValue("price")

	if len(name) < 3 || manufacturer == "" || priceStr == "" {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price format", http.StatusBadRequest)
		return
	}

	newComp := models.Component{
		ID:           len(componentDB) + 1,
		Name:         name,
		Manufacturer: manufacturer,
		Category:     category,
		Price:        price,
	}
	componentDB = append(componentDB, newComp)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CreateComponentHandler(w http.ResponseWriter, r *http.Request) {
	var newComponent models.Component

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newComponent); err != nil {
		http.Error(w, "Invalid component data", http.StatusBadRequest)
		return
	}

	newComponent.ID = len(componentDB) + 1
	componentDB = append(componentDB, newComponent)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newComponent)
}

func UpdateComponentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid component ID", http.StatusBadRequest)
		return
	}

	var updatedComponent models.Component

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedComponent); err != nil {
		http.Error(w, "Invalid component data", http.StatusBadRequest)
		return
	}

	for i, comp := range componentDB {
		if comp.ID == id {
			updatedComponent.ID = id
			componentDB[i] = updatedComponent

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(updatedComponent)
			return
		}
	}

	http.Error(w, "Component not found", http.StatusNotFound)
}

func DeleteComponentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid component ID", http.StatusBadRequest)
	}

	for i, comp := range componentDB {
		if comp.ID == id {
			componentDB = append(componentDB[:i], componentDB[i+1:]...)

			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Component not found", http.StatusNotFound)
}

func RenderHomeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, "base", componentDB)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
