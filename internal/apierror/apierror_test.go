package apierror

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		writeError func(http.ResponseWriter)
		wantStatus int
		wantCode   string
	}{
		{
			name: "bad request",
			writeError: func(w http.ResponseWriter) {
				BadRequest(w, "bad_request", "bad request")
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "bad_request",
		},
		{
			name: "not found",
			writeError: func(w http.ResponseWriter) {
				NotFound(w, "not_found", "not found")
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
		},
		{
			name: "conflict",
			writeError: func(w http.ResponseWriter) {
				Conflict(w, "conflict", "conflict")
			},
			wantStatus: http.StatusConflict,
			wantCode:   "conflict",
		},
		{
			name: "internal",
			writeError: func(w http.ResponseWriter) {
				Internal(w)
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			tt.writeError(rr)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rr.Code)
			}

			var resp Response
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Error != tt.wantCode {
				t.Errorf("expected error %q, got %q", tt.wantCode, resp.Error)
			}
		})
	}
}

func TestHandleContextError(t *testing.T) {
	t.Run("canceled", func(t *testing.T) {
		rr := httptest.NewRecorder()

		handled := HandleContextError(rr, context.Canceled)

		if !handled {
			t.Fatal("expected context.Canceled to be handled")
		}

		if rr.Code != http.StatusOK {
			t.Fatalf("excpeted no response to be written, got %d", rr.Code)
		}
	})

	t.Run("deadline exceeded", func(t *testing.T) {
		rr := httptest.NewRecorder()

		handled := HandleContextError(rr, context.DeadlineExceeded)

		if !handled {
			t.Fatal("expected deadline to be handled")
		}

		if rr.Code != http.StatusGatewayTimeout {
			t.Fatalf("expected 504, got %d", rr.Code)
		}
	})

	t.Run("other error", func(t *testing.T) {
		rr := httptest.NewRecorder()

		handled := HandleContextError(rr, errors.New("database exploded, poof"))

		if handled {
			t.Fatal("expected unrelated error not to be handled")
		}
	})
}
