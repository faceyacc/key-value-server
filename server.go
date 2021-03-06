package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-chi/chi/v5"
)

var (
	StoragePath = "/tmp"
	RWMutex     = sync.RWMutex{}
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

		// read user's request input to POST
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

	saveData(ctx, data)

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

	err = saveData(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func dataPath() string {
	return filepath.Join(StoragePath, "data.json")
}

func loadData(ctx context.Context) (map[string]string, error) {
	empty := map[string]string{}

	emptyData, err := encode(map[string]string{})
	if err != nil {
		return empty, err
	}

	// Check if folder exist if not create one.
	if _, err = os.Stat(StoragePath); os.IsNotExist(err) {
		err = os.MkdirAll(StoragePath, 0755)
		if err != nil {
			return empty, err
		}
	}

	// Check if file exist if not create one.
	if _, err = os.Stat(dataPath()); os.IsNotExist(err) {
		err := os.WriteFile(dataPath(), emptyData, 0644)
		if err != nil {
			return empty, err
		}
	}

	content, err := os.ReadFile(dataPath())
	if err != nil {
		return empty, err
	}

	return decode(content)

}

func saveData(ctx context.Context, data map[string]string) error {
	// check if folder exist, if not create one.

	if _, err := os.Stat(StoragePath); os.IsNotExist(err) {
		err := os.MkdirAll(StoragePath, 0755)
		if err != nil {
			return err
		}
	}

	encodedData, err := encode(data)
	if err != nil {
		return err
	}

	// write encoded data to file.
	return os.WriteFile(dataPath(), encodedData, 0644)
}

// encode and decode handles marhsalling and unmarshalling of data formats.
func encode(data map[string]string) ([]byte, error) {

	encodedData := map[string]string{}
	for k, v := range data {
		ek := base64.URLEncoding.EncodeToString([]byte(k))
		ev := base64.URLEncoding.EncodeToString([]byte(v))
		encodedData[ek] = ev
	}
	return json.Marshal(encodedData)
}

func decode(data []byte) (map[string]string, error) {
	var jsonData map[string]string

	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	returnData := map[string]string{}

	for k, v := range jsonData {
		dk, err := base64.URLEncoding.DecodeString(k)
		if err != nil {
			return nil, err
		}

		dv, err := base64.URLEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
		returnData[string(dk)] = string(dv)

	}
	return returnData, nil
}
