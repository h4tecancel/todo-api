package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func WriteJSONError(w http.ResponseWriter, log *slog.Logger, op string, code int, msg string, err error) {
	errDTO := ErrorDTO{
		Message: msg,
		Time:    time.Now(),
	}

	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errDTO)


	args := []any{
		slog.String("op", op),
		slog.Int("status", code),
		slog.String("message", msg),
	}
	if err != nil {
		args = append(args, slog.Any("error", err))
	}

	log.Error("request failed", args...)
}


func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
