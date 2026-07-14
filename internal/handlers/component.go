package handlers

import (
	"encoding/json"
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
