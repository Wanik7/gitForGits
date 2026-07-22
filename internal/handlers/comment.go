package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type CommentHandler struct {
	DB    *sql.DB
	Tmpl  *template.Template
	Store *sessions.FilesystemStore
}

// CreateCommentHandler обрабатывает POST-форму добавления комментария.
// Маршрут: POST /component/{sku}/comment
func (ch *CommentHandler) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	sku := mux.Vars(r)["sku"]

	// Проверяем авторизацию
	session, err := ch.Store.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	body := r.FormValue("body")
	if body == "" {
		http.Redirect(w, r, "/component/"+sku, http.StatusSeeOther)
		return
	}

	// Получаем component_id по SKU
	var componentID int
	err = ch.DB.QueryRow("SELECT id FROM components WHERE sku = $1", sku).Scan(&componentID)
	if err != nil {
		http.Error(w, "Component not found", http.StatusNotFound)
		return
	}

	// Вставляем комментарий
	query := `INSERT INTO comments (user_id, component_id, body) VALUES ($1, $2, $3)`
	_, err = ch.DB.Exec(query, userID, componentID, body)
	if err != nil {
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}

	// Редирект обратно на страницу компонента
	http.Redirect(w, r, "/component/"+sku+"#comments", http.StatusSeeOther)
}
