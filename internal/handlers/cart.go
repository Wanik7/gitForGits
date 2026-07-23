package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"techstore/pkg/models"

	"github.com/gorilla/sessions"
)

type CartHandler struct {
	DB    *sql.DB
	Tmpl  *template.Template
	Store *sessions.FilesystemStore
}

// GetCartCountHelper returns the total number of items in the cart
func GetCartCountHelper(r *http.Request, store *sessions.FilesystemStore) int {
	session, err := store.Get(r, sessionName)
	if err != nil {
		return 0
	}

	cart, ok := session.Values["cart"].(map[string]int)
	if !ok {
		return 0
	}

	count := 0
	for _, q := range cart {
		count += q
	}
	return count
}

func (ch *CartHandler) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	sku := r.FormValue("sku")
	qtyStr := r.FormValue("quantity")
	qty, _ := strconv.Atoi(qtyStr)
	if qty <= 0 {
		qty = 1
	}

	session, _ := ch.Store.Get(r, sessionName)
	cart, ok := session.Values["cart"].(map[string]int)
	if !ok {
		cart = make(map[string]int)
	}

	cart[sku] += qty
	session.Values["cart"] = cart
	session.Save(r, w)

	http.Redirect(w, r, "/component/"+sku, http.StatusSeeOther)
}

type CartItem struct {
	Component models.Component
	Quantity  int
	Subtotal  float64
}

type CartPageData struct {
	User      *models.User
	CartCount int
	Items     []CartItem
	Total     float64
	Success   bool
}

func (ch *CartHandler) ViewCartHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := ch.Store.Get(r, sessionName)
	cart, ok := session.Values["cart"].(map[string]int)
	if !ok {
		cart = make(map[string]int)
	}

	success := r.URL.Query().Get("success") == "1"

	data := CartPageData{
		CartCount: GetCartCountHelper(r, ch.Store),
		Items:     []CartItem{},
		Total:     0,
		Success:   success,
	}

	if userID, ok := session.Values["user_id"].(int); ok && userID != 0 {
		var localUser models.User
		err := ch.DB.QueryRow("SELECT name, email FROM users WHERE id = $1", userID).Scan(&localUser.Username, &localUser.Email)
		if err == nil {
			data.User = &localUser
		}
	}

	for sku, qty := range cart {
		if qty <= 0 {
			continue
		}
		var comp models.Component
		err := ch.DB.QueryRow("SELECT sku, name, price, image_path FROM components WHERE sku = $1", sku).Scan(&comp.SKU, &comp.Name, &comp.Price, &comp.ImagePath)
		if err == nil {
			subtotal := comp.Price * float64(qty)
			data.Items = append(data.Items, CartItem{
				Component: comp,
				Quantity:  qty,
				Subtotal:  subtotal,
			})
			data.Total += subtotal
		}
	}

	ch.Tmpl.ExecuteTemplate(w, "cart", data)
}

func (ch *CartHandler) UpdateCartHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	sku := r.FormValue("sku")
	action := r.FormValue("action")

	session, _ := ch.Store.Get(r, sessionName)
	cart, ok := session.Values["cart"].(map[string]int)
	if !ok {
		cart = make(map[string]int)
	}

	switch action {
	case "remove":
		delete(cart, sku)
	case "update":
		qty, _ := strconv.Atoi(r.FormValue("quantity"))
		if qty > 0 {
			cart[sku] = qty
		} else {
			delete(cart, sku)
		}
	}

	session.Values["cart"] = cart
	session.Save(r, w)
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (ch *CartHandler) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := ch.Store.Get(r, sessionName)
	session.Values["cart"] = make(map[string]int)
	session.Save(r, w)
	http.Redirect(w, r, "/cart?success=1", http.StatusSeeOther)
}
