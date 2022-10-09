package main

import (
	"fmt"
	"net/http"
	"photo-gallery/controllers"
	"photo-gallery/middleware"
	"photo-gallery/models"

	"github.com/gorilla/mux"
)

const (
	host     = "localhost"
	port     = "5432"
	user     = "admin"
	password = "qwerty"
	dbname   = "photogallery_dev"
	sslmode  = "disable"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Contetn-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1> 404 not found !!!!!!</h1>")
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	services, err := models.NewServices(psqlInfo)
	must(err)
	// services.DestructiveReset()
	defer services.CloseConnection()
	services.AutoMigrate()

	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery)

	requireUserMw := middleware.RequireUser{UserService: services.User}

	r := mux.NewRouter()

	// Main routes
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	// Gallery routes
	r.Handle("/galleries/new", requireUserMw.Apply(galleriesC.New)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesC.Create)).Methods("POST")

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
