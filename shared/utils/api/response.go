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
func PrintResponse(response interface{}, responseError error) {

    // Check response type
    r := reflect.ValueOf(response)
    if !(r.Kind() == reflect.Ptr && r.Type().Elem().Kind() == reflect.Struct) {
        PrintErrorResponse(errors.New("Invalid API response"))
        return
    }

    // Create zero response value if nil
    if r.IsNil() {
        response = reflect.New(r.Type().Elem()).Interface()
        r = reflect.ValueOf(response)
    }

    // Get and check response fields
    sf := r.Elem().FieldByName("Status")
    ef := r.Elem().FieldByName("Error")
    if !(sf.IsValid() && sf.CanSet() && sf.Kind() == reflect.String && ef.IsValid() && ef.CanSet() && ef.Kind() == reflect.String) {
        PrintErrorResponse(errors.New("Invalid API response"))
        return
    }

    // Populate error
    if responseError != nil {
        ef.SetString(responseError.Error())
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
    PrintResponse(&api.APIResponse{}, err)
}

