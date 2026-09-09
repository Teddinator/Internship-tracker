package companies

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teddinator/Internship-tracker/internal/testutil"
)

func TestCreateCompRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(nil)

	request := testutil.NewJSONRequest(
		http.MethodPost,
		"/companies/",
		`"{name":`,
	)

	recorder := httptest.NewRecorder()

	handler.CreateComp(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	testutil.AssertStatus(t, recorder, http.StatusBadRequest)
}

func TestCreateCompRequiresName(t *testing.T) {
	handler := NewHandler(nil)

	request := testutil.NewJSONRequest(
		http.MethodPost,
		"/companies/",
		`{
			"website": "https://example.com",
			"industry": "Tech"
		}`,
	)

	recorder := httptest.NewRecorder()

	handler.CreateComp(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	testutil.AssertStatus(t, recorder, http.StatusBadRequest)

}
