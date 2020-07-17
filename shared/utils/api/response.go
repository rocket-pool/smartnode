package api

import (
    "encoding/json"
    "errors"
    "fmt"
    "reflect"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Print an API response
// response must be a pointer to a struct type with Error and Status string fields
func PrintResponse(response interface{}) {

    // Get and check response fields
    r := reflect.ValueOf(response)
    if r.Kind() != reflect.Ptr || r.IsNil() {
        PrintErrorResponse(errors.New("Invalid API response"))
        return
    }
    sf := r.Elem().FieldByName("Status")
    ef := r.Elem().FieldByName("Error")
    if !(sf.IsValid() && sf.CanSet() && sf.Kind() == reflect.String && ef.IsValid() && ef.Kind() == reflect.String) {
        PrintErrorResponse(errors.New("Invalid API response"))
        return
    }

    // Set status
    if ef.String() == "" {
        sf.SetString("success")
    } else {
        sf.SetString("error")
    }

    // Encode
    responseBytes, err := json.Marshal(response)
    if err != nil {
        PrintErrorResponse(fmt.Errorf("Could not encode API response: %w", err))
        return
    }

    // Print
    fmt.Println(string(responseBytes))

}


// Print an API error response
func PrintErrorResponse(err error) {
    PrintResponse(&api.APIResponse{
        Error: err.Error(),
    })
}

