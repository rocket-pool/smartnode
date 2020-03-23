package api

import (
    "encoding/json"
    "fmt"
    "os"
)


// API response type
type Response struct {
    Status string       `json:"status"`
    Data interface{}    `json:"data"`
    Error string        `json:"error"`
}


// Print an API response
func PrintResponse(output *os.File, response interface{}, errorMessage string) {

    // Get status
    var status string
    if errorMessage == "" { status = "success" }
    else { status = "error" }

    // Print
    printResponse(output, Response{
        Status: status,
        Data: response,
        Error: errorMessage,
    })

}


// Print an error response
func PrintErrorResponse(output *os.File, err error) {
    printResponse(output, Response{
        Status: "error",
        Data: struct{}{},
        Error: err.Error(),
    })
}


// Print a response
func printResponse(output *os.File, response interface{}) {
    if output == nil { output = os.Stdout }
    responseBytes, err := json.Marshal(response)
    if err == nil { fmt.Fprintln(output, string(responseBytes)) }
}

