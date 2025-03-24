package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	handler "heydarsh.in/gobill/api"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Println("Warning: Error loading .env file")

	}

	port := "8080"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RequestURI = r.URL.String()
		handler.Handler(w, r)
	})

	log.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}
