package httpx

import (
	"encoding/json"
	"net/http"
)

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e HTTPError) Error() string { return e.Message }

func BadRequest(msg string) error      { return HTTPError{Code: http.StatusBadRequest, Message: msg} }
func Unauthorized(msg string) error    { return HTTPError{Code: http.StatusUnauthorized, Message: msg} }
func Forbidden(msg string) error       { return HTTPError{Code: http.StatusForbidden, Message: msg} }
func NotFound(msg string) error        { return HTTPError{Code: http.StatusNotFound, Message: msg} }
func Conflict(msg string) error        { return HTTPError{Code: http.StatusConflict, Message: msg} }
func TooManyRequests(msg string) error { return HTTPError{Code: http.StatusTooManyRequests, Message: msg} }
func Internal(msg string) error        { return HTTPError{Code: http.StatusInternalServerError, Message: msg} }

func JSONError(w http.ResponseWriter, err error) {
	he := HTTPError{Code: http.StatusInternalServerError, Message: "internal error"}
	if e, ok := err.(HTTPError); ok {
		he = e
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(he.Code)
	_ = json.NewEncoder(w).Encode(he)
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
