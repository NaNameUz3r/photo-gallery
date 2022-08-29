package controllers

import (
	"fmt"
	"net/http"
	"photo-gallery/views"
)

type Users struct {
	NewView *views.View
}

type SignupForm struct {
	Email    string
	Password string
}

func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, nil)
}

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	fmt.Fprintln(w, r.PostFormValue("email"))
	fmt.Fprintln(w, r.PostFormValue("password"))
}
