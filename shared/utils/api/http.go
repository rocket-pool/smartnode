package api

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// WriteResponse serialises response as JSON and writes it to w.
// response must be a pointer to a struct with string fields named Status and Error.
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseBytes)
}

// WriteErrorResponse writes a generic error response to w.
func WriteErrorResponse(w http.ResponseWriter, err error) {
	WriteResponse(w, &api.APIResponse{}, err)
}
