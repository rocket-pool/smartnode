package api

import (
    "encoding/json"
    "fmt"
    "os"
)


// Error response type
type ErrorResponse struct {
    Success bool `json:"success"`
    Error string `json:"error"`
}


// Print a response
func PrintResponse(output *os.File, response interface{}) {
    if output == nil { output = os.Stdout }
    responseBytes, err := json.Marshal(response)
    if err == nil { fmt.Fprintln(output, string(responseBytes)) }
}


// Print an error response
func PrintErrorResponse(output *os.File, err error) {
    PrintResponse(output, ErrorResponse{
        Success: false,
        Error: err.Error(),
    })
}

