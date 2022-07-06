package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	// "github.com/rs/cors"
)

func initializeRouter() {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	r.HandleFunc("/test-connect", ConnectHandler).Methods(http.MethodPost) 
	r.HandleFunc("/disconnect-cluster", DisconnectHandler).Methods(http.MethodPost)
	r.HandleFunc("/deployments", DeploymentHandler).Methods(http.MethodGet)
	r.HandleFunc("/services", ServiceHandler).Methods(http.MethodGet)

	// Then wrap in the CORS handler
	// corsHandler := cors.New(cors.Options{
	// 	AllowedOrigins: []string{"http://localhost", "http://localhost:3000"},
	// 	AllowedHeaders: []string{"*"},
	// }).Handler(r)
	// handler := cors.Default().Handler(r)

	log.Println("Starting server on port 9001")

	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}

	log.Fatal(http.ListenAndServe(":" + port, r))
}

func main() {
	initializeRouter()

}
