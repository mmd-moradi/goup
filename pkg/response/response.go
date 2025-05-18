package response

import (
	"encoding/json"
	"net/http"

	"github.com/mmd-moradi/goup/pkg/apperrors"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
	}

	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func Error(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusInternalServerError
	errorType := "INTERNAL_SERVER"
	message := "An unexpected error occurred"

	if appErr, ok := err.(apperrors.Error); ok {
		statusCode = appErr.StatusCode()
		errorType = string(appErr.Type)
		message = appErr.Message
	}
	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Type:    errorType,
			Message: message,
		},
	}

	w.WriteHeader(statusCode)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
