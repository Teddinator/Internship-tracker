package apierror

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type Response struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func Write(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(Response{
		Error:   code,
		Message: message,
	})
}

func BadRequest(w http.ResponseWriter, code, message string) {
	Write(w, http.StatusBadRequest, code, message)
}

func NotFound(w http.ResponseWriter, code, message string) {
	Write(w, http.StatusNotFound, code, message)
}

func Internal(w http.ResponseWriter) {
	Write(
		w,
		http.StatusInternalServerError,
		"internal_error",
		"an unexpected error occured",
	)
}

func ContextError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, context.Canceled):
		// Client/request already gone
		// No response needed.
		return true

	case errors.Is(err, context.DeadlineExceeded):
		Write(
			w,
			http.StatusGatewayTimeout,
			"query_timeout",
			"the request timed out",
		)
		return true

	default:
		return false
	}
}
