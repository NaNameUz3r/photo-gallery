package middleware

import (
	"net/http"
	"photo-gallery/context"
	"photo-gallery/models"
	"strings"
)

type User struct {
	models.UserService
}

func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}
		user, err := mw.UserService.ByRememberedToken(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)

		next(w, r)

	})
}

// RequireUser assumes that User middleware has already been run
// otherwise it will not work correctly.
type RequireUser struct {
	User
}

// Apply assumes that User middleware has already been run
// otherwise it will not work correctly.
func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn assumes that User middleware has already been run
// otherwise it will not work correctly.
func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)

	})
}
