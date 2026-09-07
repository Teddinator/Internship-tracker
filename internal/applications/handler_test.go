package applications

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateApplication(t *testing.T) {
	handler := NewHandler(nil)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid json",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid status",
			body: `{
			"company_id":"22222222-2222-2222-2222-222222222222",
			"role":"Software Developer Intern",
			"status":"banana"
			}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing company id",
			body: `{
					"company_id":"",
					"role":"Software Developer Intern",
					"status":"applied"
					}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"/applications",
				strings.NewReader(test.body),
			)

			recorder := httptest.NewRecorder()

			handler.CreateApp(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Errorf(
					"got status %d, want %d",
					recorder.Code,
					test.wantStatus,
				)
			}
		})
	}
}
