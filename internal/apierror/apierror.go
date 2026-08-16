package apierror

import (
	"encoding/json"
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
