package views

import (
	"log"
	"net/http"
	"time"

	"lenslocked.com/models"
)

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"

	// AlertMsgGeneric is displayed when any random error is encountered by the backend.
	AlertMsgGeneric = "Something went wrong. Please try again, or contact us if the problem persists."
)

// Alert is used to render alert bootstrap messages in the bootstrap.html
type Alert struct {
	Level   string
	Message string
}

// Data is the top level structure that views expect data to come in
type Data struct {
	Alert *Alert
	Yield interface{}
	User  *models.User
}

func (d *Data) SetAlert(err error) {

	// Errors passed from the validators contain 2 methods
	// Error() and Public() - see models/errors.go
	// err.(PublicError) conducts Type Assertion to determine
	// - using PublicError interface - see below as reference
	// if there's a match (where ok is returned), then execute
	// accordingly

	if pErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: pErr.Public(),
		}
	} else {
		log.Print(err) // if this is a private error, print it out to debug
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: AlertMsgGeneric,
		}
	}
}

func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}

func persistAlert(w http.ResponseWriter, alert Alert) {

	expiresAt := time.Now().Add(5 * time.Minute)

	level := http.Cookie{
		Name:     "alert_level",
		Value:    alert.Level,
		Expires:  expiresAt,
		HttpOnly: true,
	}

	message := http.Cookie{
		Name:     "alert_message",
		Value:    alert.Message,
		Expires:  expiresAt,
		HttpOnly: true,
	}

	http.SetCookie(w, &level)
	http.SetCookie(w, &message)
}

func clearAlert(w http.ResponseWriter) {

	level := http.Cookie{
		Name:     "alert_level",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	message := http.Cookie{
		Name:     "alert_message",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &level)
	http.SetCookie(w, &message)
}

func getAlert(r *http.Request) *Alert {
	level, err := r.Cookie("alert_level")
	if err != nil {
		return nil
	}
	message, err := r.Cookie("alert_message")

	if err != nil {
		return nil
	}

	alert := Alert{
		Level:   level.Value,
		Message: message.Value,
	}

	return &alert
}

// RedirectAlert Accepts all normal parameters for http.Redirect requests
// and performs a redirect, but only afer persisting the provided alert in a cookie
// so that it can be displayed when the new pages is loaded
func RedirectAlert(w http.ResponseWriter, r *http.Request, urlString string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, urlString, code)
}
