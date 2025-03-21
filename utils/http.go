package utils

import (
	"context"
	"encoding/json"
	"net/http"
)

// WriteJSON writes JSON response
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError writes error response
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// GetUserIDFromContext gets user ID from context
func GetUserIDFromContext(ctx context.Context) uint {
	if id, ok := ctx.Value("user_id").(uint); ok {
		return id
	}
	return 0
}

// MustMarshalJSON marshals value to JSON string
func MustMarshalJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
