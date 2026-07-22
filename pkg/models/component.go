package models

import "encoding/json"

type Component struct {
	ID           int             `json:"id"`
	SKU          string          `json:"sku"`
	Name         string          `json:"name"`
	Manufacturer string          `json:"manufacturer"`
	Category     string          `json:"category"`
	Price        float64         `json:"price"`
	Description  string          `json:"description"`
	Rating       float64         `json:"rating"`
	Stock        int             `json:"stock_quantity"`
	ImagePath    string          `json:"image_path"`
	Specs        json.RawMessage `json:"specs"`
}
