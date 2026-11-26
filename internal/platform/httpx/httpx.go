package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(
		r.Context(), "request failed",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err,
	)
	Error(w, http.StatusInternalServerError, "internal error")
}

func Decode(w http.ResponseWriter, r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v)
}
