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

	get.HandleFunc("/", routes.IndexHandler)
	get.HandleFunc("/page/{num}", routes.PageHandler)
	get.HandleFunc("/post/{num}", routes.SingleHandler)

	log.Print("Listening on :8000")

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
