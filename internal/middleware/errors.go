package middleware

import "net/http"

func InternalServerErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Oops, something went wrong... We're on it."))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
