package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ReponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1> Hello there! </h1>")
}

func main() {
	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":3000", nil)
}