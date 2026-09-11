package applications

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/teddinator/Internship-tracker/internal/config"
	"github.com/teddinator/Internship-tracker/internal/testutil"
)

type fakeApplicationStore struct {
	app  Application
	apps []Application
	err  error

	gotID  int64
	called bool

	gotFilter applicationFilter

	deleted    bool
	deletedErr error
}

func (f *fakeApplicationStore) GetAll(
	ctx context.Context,
	filter applicationFilter,
) ([]Application, error) {
	f.called = true
	f.gotFilter = filter

	return f.apps, f.err
}

func (f *fakeApplicationStore) GetByID(
	ctx context.Context,
	id int64,
) (Application, error) {
	f.gotID = id
	f.called = true
	return f.app, f.err
}

func (f *fakeApplicationStore) Delete(
	ctx context.Context,
	id int64,
) (bool, error) {
	f.gotID = id
	f.called = true
	return f.deleted, f.deletedErr
}

func TestGetAllApp(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		store      fakeApplicationStore
		wantStatus int
	}{
		{
			name: "applications returned",
			url:  "/applications?status=interview&location=Gothenburg",
			store: fakeApplicationStore{
				apps: []Application{
					{
						ID:      1,
						Company: "Volvo Cars",
						Role:    "Backend Intern",
						Status:  "interview",
					},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid status",
			url:        "/applications?status=banana",
			store:      fakeApplicationStore{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid company id",
			url:        "/applications?company_id=abc",
			store:      fakeApplicationStore{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "store error",
			url:  "/applications",
			store: fakeApplicationStore{
				err: errors.New("database failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &test.store

			handler := NewHandlerWithStore(
				store,
				config.New(),
			)

			request := httptest.NewRequest(
				http.MethodGet,
				test.url,
				nil,
			)

			recorder := httptest.NewRecorder()

			handler.GetAll(recorder, request)

			testutil.AssertStatus(
				t,
				recorder,
				test.wantStatus,
			)

			if test.name == "invalid status" && store.called {
				t.Error("store should not be called for invalid status")
			}

			if test.name == "invalid company id" && store.called {
				t.Error("store should not be called for invalid company id")
			}

			if test.name == "applications returned" {
				if !store.called {
					t.Fatal("store should have been called")
				}

				if store.gotFilter.Status != "interview" {
					t.Errorf(
						"got status filter %q, want %q",
						store.gotFilter.Status,
						"interview",
					)
				}

				if store.gotFilter.Location != "Gothenburg" {
					t.Errorf(
						"got location filter %q, want %q",
						store.gotFilter.Location,
						"Gothenburg",
					)
				}

				body := recorder.Body.String()

				if !strings.Contains(body, `"company":"Volvo Cars"`) {
					t.Errorf(
						"unexpected response body: %s",
						body,
					)
				}

				if !strings.Contains(body, `"status":"interview"`) {
					t.Errorf(
						"unexpected response body: %s",
						body,
					)
				}
			}
		})
	}
}

func TestGetAppByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		store      fakeApplicationStore
		wantStatus int
	}{
		{
			name: "application found",
			id:   "2",
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
			id:   "2",
			store: fakeApplicationStore{
				err: sql.ErrNoRows,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "store error",
			id:   "2",
			store: fakeApplicationStore{
				err: errors.New("database failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid id",
			id:         "abc",
			store:      fakeApplicationStore{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &test.store

			handler := &Handler{
				store:  store,
				config: config.New(),
			}

			request := httptest.NewRequest(
				http.MethodGet,
				"/applications/"+test.id,
				nil,
			)

			routeContext := chi.NewRouteContext()
			routeContext.URLParams.Add("id", test.id)

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

			if test.id == "abc" && store.called {
				t.Error("store should not be called for invalid id")
			}

			if test.id == "2" && test.wantStatus == http.StatusOK {
				if test.store.gotID != 2 {
					t.Errorf(
						"store got id %d, want 2",
						test.store.gotID,
					)
				}
				body := recorder.Body.String()

				if !strings.Contains(body, `"company":"Volvo Cars"`) {
					t.Errorf("unexpected response body: %s", body)
				}

				if !strings.Contains(body, `"status":"interview"`) {
					t.Errorf("unexpected response body: %s", body)
				}
			}
		})
	}
}

func TestDeleteApplication(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		store      fakeApplicationStore
		wantStatus int
	}{
		{
			name: "application deleted",
			id:   "2",
			store: fakeApplicationStore{
				deleted: true,
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "application not found",
			id:   "2",
			store: fakeApplicationStore{
				deleted: false,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "store error",
			id:   "2",
			store: fakeApplicationStore{
				deletedErr: errors.New("database failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid id",
			id:         "abc",
			store:      fakeApplicationStore{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &test.store

			handler := NewHandlerWithStore(
				store,
				config.New(),
			)

			request := httptest.NewRequest(
				http.MethodDelete,
				"/applications/"+test.id,
				nil,
			)

			routeContext := chi.NewRouteContext()
			routeContext.URLParams.Add("id", test.id)

			request = request.WithContext(
				context.WithValue(
					request.Context(),
					chi.RouteCtxKey,
					routeContext,
				),
			)

			recorder := httptest.NewRecorder()

			handler.DeleteApp(recorder, request)

			testutil.AssertStatus(
				t,
				recorder,
				test.wantStatus,
			)

			if test.id == "abc" && store.called {
				t.Error("store should not be called for invalid id")
			}

			if test.id == "2" && !store.called {
				t.Error("store should have been called")
			}

			if test.id == "2" && store.gotID != 2 {
				t.Errorf(
					"store got id %d, want 2",
					store.gotID,
				)
			}

		})
	}
}

func TestCreateApplication(t *testing.T) {
	handler := NewHandlerWithStore(
		&fakeApplicationStore{},
		config.New(),
	)

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
