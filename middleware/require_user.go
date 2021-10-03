package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"lenslocked.com/context"
	"lenslocked.com/models"
)

type User struct {
	UserService models.UserService
}

// Apply function accepts and returns a http Handler method
// It calls the ApplyFn function by passing to it the Handler function's method: ServeHTTP
// This is where ServeHTTP passes to ResponseWriter and Request to ApplyFn

// Apply Method for User struct
func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn is the MIDDLEWARE that allows us to check if there's a valid remember token
// ApplyFn then checks if the hased remember token has a valid user
// It the user cookie is valid, it sets the context and calls the next handler to further process
// the routing request for the requested page

// ApplyFn Method for User struct
func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip user lookup for static assets
		path := r.URL.Path
		if strings.HasPrefix(path, "/assets/") || strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("remember_token")

		if err != nil {
			next(w, r)
			return
		}

		user, err := mw.UserService.ByRemember(cookie.Value)

		if err != nil {
			next(w, r)
			return
		}

		cxt := r.Context()                // get the context that is part of the request
		cxt = context.WithUser(cxt, user) // provide the current context of the remember token's user
		r = r.WithContext(cxt)            // this will update request with the new context that was just created

		fmt.Println("Found User: ", user)
		next(w, r)

	})
}

// RequireUser embeds the User object
// RequireUser assumes that User middleware has already been run
// Otherwise, it will not work correctly
type RequireUser struct {
	User
}

// Apply Method for RequireUser struct
// Apply assumes that the User middleware has already been run
// Otherwise, it will not work correctly.
func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn Method for RequireUser struct
// Apply assumes that the User middleware has already been run
// Otherwise, it will not work correctly.
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
