package controllers

import (
	"net/http"
	"time"

	"lenslocked.com/context"
	"lenslocked.com/models"
	"lenslocked.com/rand"
	"lenslocked.com/views"
)

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        models.UserService
}

// NewUsers is used to create a Users controller
// This function will panic if the templates are not parsed correctly
// and should only be used during initial setup
func NewUsers(us models.UserService) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
	}
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, r, nil)
}

// we are going to use struct tag which are meta data to handle incoming requests
// this is easier to manage if the form has a lot of data
type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
	Age      uint   `schema:"age"`
}

// Create is used to process the signup form when a user submits it.
// This is to creat a new user account.

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {

	var vd views.Data
	var form SignupForm

	if err := parseForm(r, &form); err != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: views.AlertMsgGeneric,
		}
		u.NewView.Render(w, r, vd)
		return
	}

	var user = models.User{
		Name:     form.Name,
		Email:    form.Email,
		Age:      form.Age,
		Password: form.Password,
	}

	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	err := u.signIn(w, &user)

	if err != nil {
		http.Redirect(w, r, "login", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)

}

// Post/Login verifies the the provided email and password and
// the logs in if the user is correct, with stuct tags (meta data) to handle incoming requests

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {

	var vd views.Data
	var form LoginForm

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	user, err := u.us.Authenticate(form.Email, form.Password)

	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("Invalid Email Address .")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return
	}

	err = u.signIn(w, user)

	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Welcome to Lenslocked.com",
	}
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, alert)

}

// Logout deletes a user's session cookie (remember_token)
// and then will update the user resource with a new member token
// POST/logout
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	user := context.User(r.Context())
	token, _ := rand.RememberToken()

	user.Remember = token
	u.us.Update(user)
	http.Redirect(w, r, "/", http.StatusFound)

}

// SignIn is used to sign the given user in via cookies

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {

	if user.Remember == "" {
		remember, err := rand.RememberToken()

		if err != nil {
			return err
		}

		user.Remember = remember
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	return nil
}

// CookieTest is used to display the cookies set on the current user
/*
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("remember_token")

	if err != nil {

		// If a remember token is not found in the cookie,
		// redirect the user to the login page

		http.Redirect(w, r, "/login", http.StatusFound)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := u.us.ByRemember(cookie.Value)

	if err != nil {

		// Using the remember token from the cookie, it is hashed
		// and checked if it such a hashed token belongs to a user
		// When no user is found, redirect the user to the login page

		http.Redirect(w, r, "/login", http.StatusFound)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, user)
}
*/
