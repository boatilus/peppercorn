package main

import (
	"log"
	"net/http"

	"github.com/boatilus/peppercorn/routes"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func main() {
	router := mux.NewRouter()
	get := router.Methods("GET").Subrouter()

	get.HandleFunc("/", routes.IndexHandler)
	get.HandleFunc("/page/{num}", routes.PageHandler)
	get.HandleFunc("/post/{num}", routes.SingleHandler)

	n := negroni.Classic()
	n.UseHandler(router)

	port := viper.GetString("port")

	log.Printf("Listening on %v", port)

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(port, nil))
}
