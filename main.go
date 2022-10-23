package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/", getBlockchain).Methods("GET")
	route.HandleFunc("/", writeBlock).Methods("POST")
	route.HandleFunc("/new", newBook).Methods("POST")

	log.Println("listening on port:3000")
	log.Fatal(http.ListenAndServe(":3000", route))

}
