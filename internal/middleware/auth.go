package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const sessionName = "techstore-session"

func RequireAuthMiddleware(store *sessions.FilesystemStore) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, sessionName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, ok := session.Values["user_id"].(int); !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdminMiddleware(store *sessions.FilesystemStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, sessionName)
			if err != nil {
				http.Error(w, "Session error", http.StatusInternalServerError)
				return
			}

			if _, ok := session.Values["user_id"].(int); !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			role, ok := session.Values["role"].(string)
			if !ok || role != "admin" {
				http.Error(w, "Access denied: admin role required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
