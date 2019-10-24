package api

import (
    "encoding/json"
    "fmt"
)


// Error response type
type ErrorResponse struct {
    Success bool `json:"success"`
    Error string `json:"error"`
}


// Print a response
func PrintResponse(response interface{}) {
    responseBytes, err := json.Marshal(response)
    if err == nil { fmt.Println(string(responseBytes)) }
}


// Print an error response
func PrintErrorResponse(err error) {
    PrintResponse(ErrorResponse{
        Success: false,
        Error: err.Error(),
    })
}

