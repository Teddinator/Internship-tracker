package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	app := &applicationServer{}

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	app.healthHandler(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d want %d",
			response.StatusCode,
			http.StatusOK,
		)
	}

	contentType := response.Header.Get("Content-Type")

	if contentType != "application/json" {
		t.Errorf(
			"Content type = %q want %q",
			contentType,
			"application/json",
		)
	}

	var body map[string]string

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf(
			"status body = %q want %q",
			body["status"],
			"ok",
		)
	}
}
