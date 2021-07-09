package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	port := "8080"

	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Printf("Starting up on http://localhost%s", port)

	r := chi.NewRouter()

	r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		(JSON(rw, map[string]string{"welcome to": "the universe"}))
	})
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func JSON(rw http.ResponseWriter, data interface{}) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	b, err := json.Marshal(data)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		JSON(rw, map[string]string{"error": err.Error()})
		return
	}
	rw.Write(b)
}
