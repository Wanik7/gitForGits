package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func (uh *UserHandler) RenderRegisterForm(w http.ResponseWriter, r *http.Request) {
	err := uh.Tmpl.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		http.Error(w, "Ошибка рендеринга страницы", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка чтения формы", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if name == "" || email == "" || password == "" {
		http.Error(w, "All the fields are necessary", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	query := `INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3)`
	_, err = uh.DB.Exec(query, name, email, string(hashedPassword))
	if err != nil {
		log.Println("Error registration in Database:", err)
		http.Error(w, "This email already in use", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
