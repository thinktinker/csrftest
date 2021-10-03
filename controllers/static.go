package controllers

import "lenslocked.com/views"

type Static struct {
	Home    *views.View
	Contact *views.View
	About   *views.View
}

func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "statics/home"),
		Contact: views.NewView("bootstrap", "statics/contact"),
		About:   views.NewView("bootstrap", "statics/about"),
	}
}
