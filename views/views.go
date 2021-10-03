package views

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"
	"lenslocked.com/context"
)

var (
	LayoutDir   string = "views/layout/"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

// this function takes in variatic parameters
func NewView(layout string, files ...string) *View {

	addTemplatePath(files)
	addTemplateExt(files)

	// append the layout file(s) to be used along with the passed-in files,
	// e.g. bootstrap.gohtml, footer.gohtml and navbar.gohtml
	// using Globbing, ie. find all gohtml files (refer to function layoutFiles, below)
	// the 3 dots layoutFiles()... unpacks the slice []string
	files = append(files, layoutFiles()...)

	// template.Parsefiles will unpack all the individual strings
	// therefore, there is no need to use a slice type
	// t, err := template.ParseFiles(files...)

	// UPDATED version on parsing the template files
	// For the HTML template, its Func takes in a FuncMap function that has
	// has a key and returns a value
	// We are writing our own function named "csrfField" and attaching it to our template
	// so that it can be used
	// We want the csrfField to include a hidden field to indicate that this is a valid form
	t, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrfField not implemented")
		},
	}).ParseFiles(files...)

	if err != nil {
		log.Print(err)
	}

	//return the template
	return &View{
		Template: t,
		Layout:   layout,
	}
}

type View struct {
	Template *template.Template
	Layout   string
}

// layoutFiles returns as slice of strings
// reprsenting the layout files used in our application
func layoutFiles() []string {

	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)

	if err != nil {
		panic(err)
	}
	return files
}

// addTemplatePath takes in a slice of strings
// representing file paths for templates, and it prepends
// the TemplateDir directory to each stirng in the slice
//
// e.g. the input  {"home"} would result in the output
// {"views/home"} if TemplateDir == "views"
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// addTemplateExt takes in a slice of string
// representing file paths for templates and it appends
// the TemplateExt extension to each string in the slice
//
// e.g. the input {"views/home"} would result in the output
// {"views/home.gohtml"} if TemplateExt == ".gohtml"
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}

//Render the view automatically using Go's Duck Typing
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

// Create a method to render the views
// The method is specfic to the View Type, using (v *View)
// Alternatively, consider passing the context instead of http.Request
// Refer to chapter 14 for the alternative
func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {

	vd := Data{} // create an instance of Data{}

	switch d := data.(type) {
	case Data:
		vd = d // If the data type received is Data, store d into vd
	default:
		vd = Data{
			Yield: data, // By default, pass the value of the data to vd
		}
	}

	if alert := getAlert(r); alert != nil && vd.Alert == nil {
		vd.Alert = alert
		clearAlert(w)
	}

	// get the context of the logged in user
	vd.User = context.User(r.Context())

	// by using a method by reference, it is implicit that
	// the Layout "bootstrap" is based on the object itself
	//currently no data is passed to the layout yet

	w.Header().Set("Content-Type", "text/html")

	//bytes.buttfer has both a read and a write method that allows us to write to it and read from it
	var buf bytes.Buffer

	// Instead of passing the response writer, pass in the buffer
	// If there's an error, handle the error as written in the statement below

	// UPDATE: create a new template based on the existing one
	// and attached a new function to it, return us a new template and assign it tpl
	csrfField := csrf.TemplateField(r)

	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})

	if err := tpl.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
		http.Error(w, "Something went wrong. If the problem persisits, please email support@lenslocked.com", http.StatusInternalServerError)
		return
	}

	// Otherwise, use the io.Copy to read from to the buffer to the destination, i.e. "w" or the ResponseWriter
	io.Copy(w, &buf) //io.Copy allows you to copy from one reader to a writer

	// if err := v.Template.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
	// 	http.Error(w, "Something went wrong. If the problem persisits, please email support@lenslocked.com", http.StatusInternalServerError)
	// 	return
	// }
}
