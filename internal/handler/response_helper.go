package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mirai-box/mirai-box/internal/model"
)

// ResponseWriter is an interface for writing HTTP responses
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

// SendJSONResponse sends a JSON response with the given status code and data
func SendJSONResponse(w ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Error encoding JSON response", "error", err)
	}
}

// SendErrorResponse sends a JSON error response with the given status code and message
func SendErrorResponse(w ResponseWriter, status int, message string) {
	errResp := model.ErrorResponse{
		Status:  status,
		Message: message,
	}
	SendJSONResponse(w, status, errResp)
}

// SendValidationErrorResponse sends a JSON response for validation errors
func SendValidationErrorResponse(w ResponseWriter, errors map[string]string) {
	errResp := model.ValidationErrorResponse{
		Status:  http.StatusBadRequest,
		Message: "Validation failed",
		Errors:  errors,
	}
	SendJSONResponse(w, http.StatusBadRequest, errResp)
}

// SendPaginatedResponse sends a JSON response for paginated data
func SendPaginatedResponse(w ResponseWriter, status int, data interface{}, pagination model.Pagination) {
	resp := model.PaginatedResponse{
		Data:       data,
		Pagination: pagination,
	}
	SendJSONResponse(w, status, resp)
}
