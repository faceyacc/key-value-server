package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
)

var (
	data    = map[string]string{}
	RWMutex = sync.RWMutex{}
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

	r.Get("key/{key}", func(rw http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")

		data, err := Get(r.Context(), key)

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError) // returns 500 error
			JSON(rw, map[string]string{"error": err.Error()})
			return
		}
		rw.Write([]byte(data))

	})

	r.Delete("key/{key}", func(rw http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")

		err := Delete(r.Context(), key)

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			JSON(rw, map[string]string{"error": err.Error()})
			return
		}
		JSON(rw, map[string]string{"status": "success"})
	})

	r.Post("key/{key}", func(rw http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")

		body, err := io.ReadAll(r.Body)

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			JSON(rw, map[string]string{"error": err.Error()})
			return
		}

		err = Set(r.Context(), key, string(body))

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			JSON(rw, map[string]string{"error": err.Error()})
			return
		}

		JSON(rw, map[string]string{"status": "success"})

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

func Set(ctx context.Context, key, body string) error {

	data, err := loadData(ctx)

	if err != nil {
		return err
	}

	RWMutex.RLock()
	data[key] = body
	RWMutex.RUnlock()
	return nil
}

func Get(ctx context.Context, key string) (string, error) {

	data, err := loadData(ctx)

	if err != nil {
		return "", err
	}
	RWMutex.RLock()
	value := data[key]
	RWMutex.RUnlock()

	return value, nil
}

func Delete(ctx context.Context, key string) error {

	data, err := loadData(ctx)

	if err != nil {
		return err
	}

	RWMutex.RLock()
	delete(data, key)
	RWMutex.RUnlock()
	return nil
}
