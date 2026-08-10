package companies

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateCompRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(nil)

	request := httptest.NewRequest(
		http.MethodPost,
		"/companies/",
		strings.NewReader(`"{name":`),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.CreateComp(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf(
			"status code = %d want %d",
			response.StatusCode,
			http.StatusBadRequest,
		)
	}
}

func TestCreateCompRequiresName(t *testing.T) {
	handler := NewHandler(nil)

	request := httptest.NewRequest(
		http.MethodPost,
		"/companies/",
		strings.NewReader(`{
			"website": "https://example.com",
			"industry": "Tech"
		}`),
	)

	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.CreateComp(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf(
			"status code = %d want %d",
			response.StatusCode,
			http.StatusBadRequest,
		)
	}

}
