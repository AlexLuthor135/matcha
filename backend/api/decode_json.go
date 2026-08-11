package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxJSONRequestSize int64 = 64 << 10

func DecodeJSONRequest(w http.ResponseWriter, r *http.Request, destination any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONRequestSize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeJSONDecodeError(w, err)
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			http.Error(w, "Request body must contain only one JSON object", http.StatusBadRequest)
		} else {
			writeJSONDecodeError(w, err)
		}
		return false
	}
	return true
}

func writeJSONDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	switch {
	case errors.As(err, &maxBytesError):
		http.Error(w, "Request payload is too large", http.StatusRequestEntityTooLarge)
	case errors.Is(err, io.EOF):
		http.Error(w, "Request body is required", http.StatusBadRequest)
	default:
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
	}
}
