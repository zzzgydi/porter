package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

// HTTPError is the internal error type used by handlers.
// On the wire it is rendered as a Response with Msg set to Message.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e HTTPError) Error() string { return e.Message }

func BadRequest(msg string) error   { return HTTPError{Code: http.StatusBadRequest, Message: msg} }
func Unauthorized(msg string) error { return HTTPError{Code: http.StatusUnauthorized, Message: msg} }
func Forbidden(msg string) error    { return HTTPError{Code: http.StatusForbidden, Message: msg} }
func NotFound(msg string) error     { return HTTPError{Code: http.StatusNotFound, Message: msg} }
func Conflict(msg string) error     { return HTTPError{Code: http.StatusConflict, Message: msg} }
func TooManyRequests(msg string) error {
	return HTTPError{Code: http.StatusTooManyRequests, Message: msg}
}
func Internal(msg string) error { return HTTPError{Code: http.StatusInternalServerError, Message: msg} }

// Response is the standard API response envelope.
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func JSONError(w http.ResponseWriter, err error) {
	he := HTTPError{Code: http.StatusInternalServerError, Message: "internal error"}
	var e HTTPError
	if errors.As(err, &e) {
		he = e
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(he.Code)
	_ = json.NewEncoder(w).Encode(Response{Code: he.Code, Msg: he.Message})
}

// JSON writes a standard wrapped response: {"code": status, "msg": "", "data": v}.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Code: status, Data: v})
}

// RawJSON writes the value directly without the Response wrapper.
// It is used for endpoints that must return a specific wire format (e.g. Docker registry token).
func RawJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
