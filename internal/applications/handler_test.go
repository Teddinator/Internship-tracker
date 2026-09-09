package applications

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/teddinator/Internship-tracker/internal/testutil"
)

type fakeApplicationStore struct {
	app Application
	err error
}

func (f fakeApplicationStore) GetByID(
	ctx context.Context,
	id int64,
) (Application, error) {
	return f.app, f.err
}

func TestGetAppByID(t *testing.T) {
	tests := []struct {
		name       string
		store      fakeApplicationStore
		wantStatus int
	}{
		{
			name: "application found",
			store: fakeApplicationStore{
				app: Application{
					ID:      2,
					Company: "Volvo Cars",
					Role:    "Software Developer Intern",
					Status:  "interview",
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "application not found",
			store: fakeApplicationStore{
				err: sql.ErrNoRows,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "store error",
			store: fakeApplicationStore{
				err: errors.New("database failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := &Handler{
				store: test.store,
			}

			request := httptest.NewRequest(
				http.MethodGet,
				"/applications/2",
				nil,
			)

			routeContext := chi.NewRouteContext()
			routeContext.URLParams.Add("id", "2")

			request = request.WithContext(
				context.WithValue(
					request.Context(),
					chi.RouteCtxKey,
					routeContext,
				),
			)

			recorder := httptest.NewRecorder()

			handler.GetAppByID(recorder, request)

			testutil.AssertStatus(
				t,
				recorder,
				test.wantStatus,
			)
		})
	}
}

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
		{
			name: "missing role",
			body: `{
					"company_id":"22222222-2222-2222-2222-222222222222",
					"role":"",
					"status":"applied"
					}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid applied at",
			body: `{
					"company_id":"22222222-2222-2222-2222-222222222222",
					"role":"Backend Intern",
					"status":"applied",
					"applied_at":"07-10-2026"
					}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := testutil.NewJSONRequest(
				http.MethodPost,
				"/applications",
				test.body,
			)

			recorder := httptest.NewRecorder()

			handler.CreateApp(recorder, request)

			testutil.AssertStatus(t, recorder, test.wantStatus)
		})
	}
}
