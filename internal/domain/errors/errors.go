package errors
import (
	"encoding/json"
	"net/http"
)
type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
func (e APIError) Error() string {
	return e.Message
}
func NewNotFoundError(message string) APIError {
	return APIError{
		Status:  http.StatusNotFound,
		Code:    "NOT_FOUND",
		Message: message,
	}
}
func NewBadRequestError(message string) APIError {
	return APIError{
		Status:  http.StatusBadRequest,
		Code:    "BAD_REQUEST",
		Message: message,
	}
}
func RespondWithError(w http.ResponseWriter, err error) {
	apiErr, ok := err.(APIError)
	if !ok {
		apiErr = APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Erro interno do servidor",
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(apiErr)
}
