package api

import (
    "encoding/json"
    "fmt"
    "reflect"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Print an API response
// response MUST be a pointer to a struct type with Error and Status string fields
func PrintResponse(response interface{}) error {

    // Set status
    responseVal := reflect.ValueOf(response).Elem()
    if responseVal.FieldByName("Error").String() == "" {
        responseVal.FieldByName("Status").SetString("success")
    } else {
        responseVal.FieldByName("Status").SetString("error")
    }

    // Encode, print and return
    responseBytes, err := json.Marshal(response)
    if err != nil {
        PrintErrorResponse(fmt.Errorf("Could not encode API response: %w", err))
    } else {
        fmt.Println(string(responseBytes))
    }
    return nil

}


// Print an API error response
func PrintErrorResponse(err error) {
    _ = PrintResponse(&api.APIResponse{
        Error: err.Error(),
    })
}

