package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1> Hello there!!! </h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "To get in touch, please send an email to <a href=\"mailto:support@xyz.com\">support@xyz.com</a>.")
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h4>Every frequently asked question is usually so stupid and obvious that it already contains the answer.</h4>")

}
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Contetn-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1> 404 not found !!!!!!</h1>")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/contact", contactHandler)
	r.HandleFunc("/faq", faqHandler)
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	http.ListenAndServe(":3000", r)
}
