package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dominic/readshelf/internal/core/domain"
)

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

func HandleError(w http.ResponseWriter, err error) {
	status := domainErrorStatus(err)
	msg := err.Error()
	if status == http.StatusInternalServerError {
		msg = "internal server error"
	}
	Error(w, status, msg)
}

func domainErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrDuplicateAnnotation):
		return http.StatusConflict
	case errors.Is(err, domain.ErrConflictAnnotation):
		return http.StatusConflict
	case errors.Is(err, domain.ErrBookLimitReached):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrEmailTaken):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInvalidToken):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrInvalidFileType):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}
