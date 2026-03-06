package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Envelope{
		Status:  http.StatusText(status),
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func Ok(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusOK, message, data)
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, message, nil)
}

func InternalServerError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, message, nil)
}

func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, message, nil)
}

func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, message, nil)
}
