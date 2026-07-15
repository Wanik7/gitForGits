package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"techstore/pkg/models"

	"github.com/gorilla/mux"
)

var ComponentDB = []models.Component{
	{ID: 1, Name: "Ryzen 5 5600", Manufacturer: "AMD", Category: "CPU", Price: 135.50},
	{ID: 2, Name: "GeForce RTX 4060", Manufacturer: "NVIDIA", Category: "GPU", Price: 299.99},
}

func GetComponentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(ComponentDB)
}

func GetComponentByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid component ID", http.StatusBadRequest)
	}

	for _, component := range ComponentDB {
		if component.ID == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(component)
			return
		}
	}

	http.Error(w, "Component not found", http.StatusNotFound)
}

func CreateComponentHandler(w http.ResponseWriter, r *http.Request) {
	var newComponent models.Component

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newComponent); err != nil {
		http.Error(w, "Invalid component data", http.StatusBadRequest)
		return
	}

	newComponent.ID = len(ComponentDB) + 1
	ComponentDB = append(ComponentDB, newComponent)

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

	for i, comp := range ComponentDB {
		if comp.ID == id {
			updatedComponent.ID = id
			ComponentDB[i] = updatedComponent

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

	for i, comp := range ComponentDB {
		if comp.ID == id {
			ComponentDB = append(ComponentDB[:i], ComponentDB[i+1:]...)

			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Component not found", http.StatusNotFound)
}

func RenderHomeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, "base", ComponentDB)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
