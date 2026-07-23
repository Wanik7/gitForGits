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
	err := uh.Tmpl.ExecuteTemplate(w, "register", nil)
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

	query := `INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`
	var userID int
	err = uh.DB.QueryRow(query, name, email, string(hashedPassword)).Scan(&userID)
	if err != nil {
		log.Println("Error registration in Database:", err)
		http.Error(w, "This email already in use", http.StatusInternalServerError)
		return
	}

	session, err := uh.Store.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = userID
	session.Values["role"] = "customer"
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LOGIN

func (uh *UserHandler) RenderLoginForm(w http.ResponseWriter, r *http.Request) {
	err := uh.Tmpl.ExecuteTemplate(w, "login", nil)
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
		id           int
		passwordHash string
		role         string
	}

	err = uh.DB.QueryRow("SELECT id, password_hash, role FROM users WHERE email = $1", email).Scan(&localUser.id, &localUser.passwordHash, &localUser.role)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(localUser.passwordHash), []byte(password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	session, err := uh.Store.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = localUser.id
	session.Values["role"] = localUser.role

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Error serving session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (uh *UserHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := uh.Store.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	delete(session.Values, "user_id")

	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Error finishing session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ADMIN USER MANAGEMENT

func (uh *UserHandler) RenderAdminUserHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	if errCookie, err := r.Cookie("admin_err"); err == nil && errCookie != nil {
		data["Error"] = errCookie.Value
		http.SetCookie(w, &http.Cookie{Name: "admin_err", MaxAge: -1, Path: "/"})
	}

	err := uh.Tmpl.ExecuteTemplate(w, "admin_user", data)
	if err != nil {
		log.Println("Error rendering admin_user:", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) CreateUserByAdminHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		setFlashCookie(w, "admin_err", "Ошибка чтения формы")
		http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role")

	if name == "" || email == "" || password == "" || role == "" {
		setFlashCookie(w, "admin_err", "Все поля обязательны")
		http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		setFlashCookie(w, "admin_err", "Ошибка хеширования пароля")
		http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
		return
	}

	query := `INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4)`
	_, err = uh.DB.Exec(query, name, email, string(hashedPassword), role)
	if err != nil {
		log.Println("Error creating user:", err)
		setFlashCookie(w, "admin_err", "Ошибка сохранения в БД (возможно email занят)")
		http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
}

func setFlashCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	})
}
