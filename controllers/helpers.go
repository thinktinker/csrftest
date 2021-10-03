package controllers

import (
	"net/http"

	"github.com/gorilla/schema"
)

func parseForm(r *http.Request, destination interface{}) error {

	// try to test the error for parsing the form when creating a user
	// return errors.New("Something went bad.")

	//this is not necessary, but good to add for error handling when the form cannot be parsed
	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	// fmt.Fprintln(w, "Using PostFormValue "+string(r.PostFormValue("email")))
	// fmt.Fprintln(w, r.PostForm["email"]) // r.Postform = map[string][]string

	//use gorilla schema to handle the form request
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(destination, r.PostForm); err != nil {
		panic(err)
	}

	return nil
}
