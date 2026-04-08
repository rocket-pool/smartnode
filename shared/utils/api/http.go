package api

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// BadRequestError signals that the caller supplied invalid input.
// WriteResponse maps it to HTTP 400 rather than 500.
type BadRequestError struct{ Err error }

func (e *BadRequestError) Error() string { return e.Err.Error() }
func (e *BadRequestError) Unwrap() error { return e.Err }

// NotFoundError signals that the requested resource or route does not exist.
// WriteResponse maps it to HTTP 404.
type NotFoundError struct{ Path string }

func (e *NotFoundError) Error() string { return fmt.Sprintf("not found: %s", e.Path) }

// WriteResponse serialises response as JSON and writes it to w.
// response must be a pointer to a struct with string fields named Status and Error.
// On error it writes 400 for BadRequestError and 500 for everything else.
func WriteResponse(w http.ResponseWriter, response interface{}, responseError error) {
	r := reflect.ValueOf(response)
	if !(r.Kind() == reflect.Ptr && r.Type().Elem().Kind() == reflect.Struct) {
		WriteErrorResponse(w, errors.New("invalid API response"))
		return
	}

	if r.IsNil() {
		response = reflect.New(r.Type().Elem()).Interface()
		r = reflect.ValueOf(response)
	}

	sf := r.Elem().FieldByName("Status")
	ef := r.Elem().FieldByName("Error")
	if !(sf.IsValid() && sf.CanSet() && sf.Kind() == reflect.String &&
		ef.IsValid() && ef.CanSet() && ef.Kind() == reflect.String) {
		WriteErrorResponse(w, errors.New("invalid API response"))
		return
	}

	if responseError != nil {
		ef.SetString(responseError.Error())
	}
	if ef.String() == "" {
		sf.SetString("success")
	} else {
		sf.SetString("error")
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		WriteErrorResponse(w, fmt.Errorf("could not encode API response: %w", err))
		return
	}

	statusCode := http.StatusOK
	if ef.String() != "" {
		var br *BadRequestError
		var nf *NotFoundError
		switch {
		case errors.As(responseError, &br):
			statusCode = http.StatusBadRequest
		case errors.As(responseError, &nf):
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(append(responseBytes, '\n'))
}

// WriteErrorResponse writes a generic error response to w.
func WriteErrorResponse(w http.ResponseWriter, err error) {
	WriteResponse(w, &api.APIResponse{}, err)
}
