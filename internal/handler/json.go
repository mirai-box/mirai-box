package handler

import (
	"encoding/json"
	"net/http"
)

// jsonResponse represents the structure of the response.
type jsonResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// respondWithJSON writes a JSON response.
func respondWithJSON(w http.ResponseWriter, status int, response jsonResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// respondWithError writes an error response.
func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, jsonResponse{Status: "error", Message: message})
}

// respondWithSuccess writes a success response.
func respondWithSuccess(w http.ResponseWriter, data interface{}) {
	respondWithJSON(w, http.StatusOK, jsonResponse{Status: "success", Data: data})
}
