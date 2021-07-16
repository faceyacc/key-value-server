package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	// "net/http/httptest"
	"testing"
)

func makeStorage(t *testing.T) {
	err := os.Mkdir("testdata", 0755)
	if err != nil && !os.IsExist(err) {
		t.Fatalf("Couldn't create directory testdata: %s", err)
	}

	StoragePath = "testdata"
}

func cleanupStorage(t *testing.T) {
	if err := os.RemoveAll(StoragePath); err != nil {
		t.Errorf("Failed to delete storage path: %s", StoragePath)
	}
	StoragePath = "/tmp/kv"
}

func TestGet(t *testing.T) {
	makeStorage(t)
	defer cleanupStorage(t)

	key := "key"
	value := "value"
	encodedKey := base64.URLEncoding.EncodeToString([]byte(key))
	encodedValue := base64.URLEncoding.EncodeToString([]byte(value))

	fileContents, _ := json.Marshal(map[string]string{encodedKey: encodedValue})
	os.WriteFile(StoragePath+"/data.json", fileContents, 0644)

	got, err := Get(context.Background(), key)
	if err != nil {
		t.Errorf("Received unexpected error: %s", err)
	}
	if got != value {
		t.Errorf("Got %s, expected %s", got, value)
	}
}

func TestJSON(t *testing.T) {

	header := http.Header{}
	headerKey := "Content-Type"
	headerValue := "application/json; charset=utf-8"
	header.Add(headerKey, headerValue)

	testCases := []struct {
		in     map[string]string
		header http.Header
		out    string
	}{
		{map[string]string{"hello": "world"}, header, `{"hello":"world"}`},
		{map[string]string{"good morning": "test"}, header, `{"good morning":"test"}`},
	}

	for _, test := range testCases {

		recorder := httptest.NewRecorder()

		JSON(recorder, test.in)

		response := recorder.Result()
		defer response.Body.Close()

		got, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("error reading response body: %s", err)
		}

		assertError(t, got, test.out)

		contentType := response.Header.Get(headerKey)
		assertHeader(t, contentType, headerValue)

	}

}

func assertHeader(t testing.TB, contentType, headerValue string) {
	if contentType != headerValue {
		t.Errorf("Got %s, expected %s", contentType, headerValue)
	}
}

func assertError(t testing.TB, got []byte, test string) {

	if string(got) != test {
		t.Errorf("got %s expected %s", string(got), test)
	}

}
