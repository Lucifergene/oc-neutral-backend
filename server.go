package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func initializeRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	r.HandleFunc("/upload-config", UploadHandler).Methods("POST")

	log.Println("Starting server on port 9000")

	log.Fatal(http.ListenAndServe(":9000", r))
}

func main() {
	initializeRouter()

}
