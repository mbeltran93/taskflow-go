// Package handler is the HTTP transport layer: it decodes requests, calls
// the service layer, and encodes responses. It knows about chi and JSON; it
// never talks to the database or issues JWTs directly.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/mbeltran93/taskflow-go/internal/apperror"
	"github.com/mbeltran93/taskflow-go/internal/auth"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

// handleServiceError maps a service-layer error to the right HTTP status by
// walking the apperror sentinels with errors.Is - the Go analogue of a
// @ExceptionHandler, implemented with plain error values instead of exceptions.
func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, apperror.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, apperror.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, apperror.ErrUnauthorized), errors.Is(err, auth.ErrInvalidToken):
		writeError(w, http.StatusUnauthorized, err.Error())
	default:
		slog.Error("unhandled service error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
