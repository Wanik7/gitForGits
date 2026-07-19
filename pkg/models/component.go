package models

type Component struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Manufacturer string  `json:"manufacturer"`
	Category     string  `json:"category"`
	Price        float64 `json:"price"`
	Description  string  `json:"description"`
	Rating       float64 `json:"rating"`
	Stock        int     `json:"stock_quantity"`
}
