package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Request struct for JSON-RPC request
type Request struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// Response struct for JSON-RPC response
type Response struct {
	Jsonrpc string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Result  string    `json:"result"`
	Error   *RpcError `json:"error"`
}

// RpcError struct for JSON-RPC error
type RpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid arguments. One argument is expected - the URL of Nethermind's admin JSON-RPC API.")
		return
	}

	url := os.Args[1]
	client := &http.Client{}

	retryTime := 3 * time.Second
	retryCount := 100
	for i := 0; i < retryCount; i++ {
		// Generate the request payload
		request := Request{
			ID:     i + 1,
			Method: "admin_prune",
			Params: []interface{}{},
		}
		if i > 0 {
			time.Sleep(retryTime)
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			fmt.Printf("Error marshaling request: %s\n", err)
			continue
		}

		// Send the request
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestData))
		if err != nil {
			fmt.Printf("Error requesting prune: %s\n", err)
			continue
		}

		// Process the response
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %s\n", err)
			continue
		}

		var response Response
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Printf("Error deserializing response JSON: %s\n", err)
			continue
		}

		if response.Error != nil {
			fmt.Printf("Error starting prune: code %d, message = %s, data = %v\n", response.Error.Code, response.Error.Message, response.Error.Data)
		} else {
			fmt.Printf("Success: Pruning is now \"%s\"\n", response.Result)
			return
		}

		fmt.Printf("Trying again in %v... (%d/%d)\n", retryTime, i+1, retryCount)
	}

	fmt.Printf("Failed starting prune after %d attempts. Please try again later.\n", retryCount)
}
