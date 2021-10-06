package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"lenslocked.com/context"
	"lenslocked.com/models"
	"lenslocked.com/views"
)

const (
	ShowGallery     = "show_gallery"
	EditGallery     = "edit_gallery"
	maxMultipartMem = 1 << 20 //1MB
)

type Galleries struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        models.GalleryService
	is        models.ImageService
	r         *mux.Router
}

// NewGalleries is used to create a Galleries controller
// and should only be used during initial setup
// Update: pased in the mux router so as to create named routes for the Create method
func NewGalleries(gs models.GalleryService, is models.ImageService, r *mux.Router) *Galleries {
	return &Galleries{
		New:       views.NewView("bootstrap", "galleries/new"),
		ShowView:  views.NewView("bootstrap", "galleries/show"),
		EditView:  views.NewView("bootstrap", "galleries/edit"),
		IndexView: views.NewView("bootstrap", "galleries/index"),
		gs:        gs,
		is:        is,
		r:         r,
	}
}

// we are going to use struct tag which are meta data to handle incoming requests
// for the galleryForm this is easier to manage if the form has a lot of data
type GalleryForm struct {
	Title string `schema:"title"`
}

// GET /galleries/:id
// Update: Split the checking of the gallery id to another function called galleryByID
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {

	gallery, err := g.galleryByID(w, r)

	if err != nil {
		return
	}

	vd := views.Data{}
	vd.Yield = gallery // If the gallery exists, store it in the Yield property of views.Data

	g.ShowView.Render(w, r, vd) //render the view with the data (temporary)
}

// GET /galleries/

func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {

	user := context.User(r.Context())
	galleries, err := g.gs.ByUserID(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	vd := views.Data{}
	vd.Yield = galleries // If the gallery exists, store it in the Yield property of views.Data

	g.IndexView.Render(w, r, vd) //render the view with the data (temporary)

}

// GET /galleries/:id/edit

func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {

	gallery, err := g.galleryByID(w, r)

	if err != nil {
		return
	}

	// To ensure that the one who edits the gallery is the originator
	// get the context of the User from the request and verify against
	// the gallery's userID
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	vd := views.Data{}
	vd.Yield = gallery // If the gallery exists, store it in the Yield property of views.Data
	// vd.User = user //Just for testing, DO NOT pass in user here. It's done in require_user middleware

	g.EditView.Render(w, r, vd) //render the view with the data (temporary)
}

// POST /galleries/:id/images

func (g *Galleries) ImageUpload(w http.ResponseWriter, r *http.Request) {

	gallery, err := g.galleryByID(w, r)

	if err != nil {
		return
	}

	// 1. To ensure that the one who edits the gallery is the originator
	// get the context of the User from the request and verify against
	// the gallery's userID
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	// To parse a single file form: https://pkg.go.dev/net/http@go1.17#Request.FormFile
	// parse the multipart form: https://pkg.go.dev/net/http@go1.17#Request.ParseMultipartForm

	// 2. Include the gallery data (with or w/o errors) to be sent
	var vd views.Data
	vd.Yield = gallery
	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	// 4. Obtain the files uploaded
	//"images" is the name of the upload input in the html

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		// Open the uploaded file
		file, err := f.Open()
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}

		defer file.Close()

		trimmedFilename := strings.Replace(f.Filename, " ", "", -1)

		err = g.is.Create(gallery.ID, file, trimmedFilename)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
	}

	//After uploading the image, get the edit_gallery named route
	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))

	//If there's an error, route to the gallery as the image has been uploaded
	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}

	//Othewise, redirect to edit gallery
	http.Redirect(w, r, url.Path, http.StatusFound)

}

// POST /galleries/:id/images/:filename/delete
// data: gallery_id, filename

func (g *Galleries) ImageDelete(w http.ResponseWriter, r *http.Request) {

	gallery, err := g.galleryByID(w, r)

	if err != nil {
		return
	}

	// To ensure that the one who edits the gallery is the originator
	// get the context of the User from the request and verify against
	// the gallery's userID
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	filename := mux.Vars(r)["filename"]
	// fmt.Fprintln(w, filename) //Check if the filename is retrieved from the url

	i := models.Image{
		Filename:  filename,
		GalleryID: gallery.ID,
	}

	err = g.is.Delete(&i)

	if err != nil {
		var vd views.Data
		vd.Yield = gallery
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}

	http.Redirect(w, r, url.Path, http.StatusFound)

}

// POST /galleries/:id/update

func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {

	gallery, err := g.galleryByID(w, r)

	if err != nil {
		return
	}

	// To ensure that the one who edits the gallery is the originator
	// get the context of the User from the request and verify against
	// the gallery's userID
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	vd := views.Data{}
	vd.Yield = gallery // If the gallery exists, store it in the Yield property of views.Data

	var form GalleryForm

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	gallery.Title = form.Title
	if err := g.gs.Update(gallery); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Gallery successfully updated!",
	}
	g.EditView.Render(w, r, vd)
}

// Create is used to process the signup form when a user submits it.
// This is to creat a new user account.

func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {

	var vd views.Data
	var form GalleryForm

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}

	// set the context of the User to be passed to the request
	user := context.User(r.Context())

	var gallery = models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}

	// *mux.Router is passed in when the NewGalleryService is created.
	// A named route (controllers.ShowGallery) is sent over by the handleFunc request from main.go
	// and the url.path is constructed based on the key-value pair recognized by the route in main.go
	// e.g.: "/gallery/123"
	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))

	fmt.Println(url)

	if err != nil {
		// Redirect the users to the galleries page
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}

	// the constructed url.Path that is routed will look like: "/gallery/123"
	// which is a recognized route in main.go
	http.Redirect(w, r, url.Path, http.StatusFound)

}

// POST /galleries/:id/delete

func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {

	gallery, err := g.galleryByID(w, r)

	if err != nil {
		return
	}

	// To ensure that the one who deletes the gallery is the originator
	// get the context of the User from the request and verify against
	// the gallery's userID
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	vd := views.Data{}

	if err := g.gs.Delete(gallery); err != nil {
		vd.SetAlert(err)
		vd.Yield = gallery // If the gallery DOES NOT exists, store it in the Yield property of views.Data
		g.EditView.Render(w, r, vd)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {

	vars := mux.Vars(r)            // Use mux's Vars to obtain all variables sent via request
	idStr := vars["id"]            // Obtain the variable with the id format
	id, err := strconv.Atoi(idStr) // And convert the id to integer

	if err != nil {
		log.Print(err)
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}

	gallery, err := g.gs.ByID(uint(id)) //check that the id of a gallery exists
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found.", http.StatusNotFound)
		default:
			log.Print(err)
			http.Error(w, "Whoops. Something went wrong!", http.StatusInternalServerError)
		}
		return nil, err
	}

	images, _ := g.is.ByGalleryID(gallery.ID)
	gallery.Images = images

	return gallery, nil
}
