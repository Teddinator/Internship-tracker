package testutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func NewJSONRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(
		method,
		target,
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	return req
}

func AssertStatus(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	want int,
) {
	t.Helper()

	if recorder.Code != want {
		t.Errorf(
			"got status %d, want %d",
			recorder.Code,
			want,
		)
	}
}
