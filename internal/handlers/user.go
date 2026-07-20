package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	DB    *sql.DB
	Tmpl  *template.Template
	Store *sessions.FilesystemStore
}

// REGISTRATION

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

// LOGIN

func (uh *UserHandler) RenderLoginForm(w http.ResponseWriter, r *http.Request) {
	err := uh.Tmpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

func (uh *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error processing form", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var localUser struct {
		ID           int
		PasswordHash string
	}

	err = uh.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&localUser.ID, &localUser.PasswordHash)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(localUser.PasswordHash), []byte(password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	session, _ := uh.Store.Get(r, "techstore-session")
	session.Values["user_id"] = localUser.ID

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error serving session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
