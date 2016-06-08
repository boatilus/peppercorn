package main

import (
	"github.com/boatilus/peppercorn/routes"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	get := router.Methods("GET").Subrouter()

	get.HandleFunc("/", indexHandler)
	get.HandleFunc("/page/{num}", pageHandler)
	get.HandleFunc("/single/{num}", singleHandler)

	log.Print("Listening on :8000")

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
