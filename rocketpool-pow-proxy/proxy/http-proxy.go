package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Config
const InfuraURL = "https://%s.infura.io/v3/%s"
const PocketURL = "https://%s.gateway.pokt.network/v1/%s"
const HandleRequestRecursionLimit = 3


// Proxy server
type HttpProxyServer struct {
    Port string
    ProviderUrl string
    Verbose bool
    ProviderType string
    idLock sync.Mutex
    id uint64
}

// Infura rate limit error
type InfuraRateLimitError struct {
    Error struct {
        Code int                    `json:"code"`
        Message string              `json:"message"`
        Data struct {
            Rate struct {
                BackoffSeconds float64  `json:"backoff_seconds"`
            }   `json:"rate"`
        }   `json:"data"`
    }   `json:"error"`
}

// Create new proxy server
func NewHttpProxyServer(port string, providerUrl string, network string, projectId string, providerType string, verbose bool) *HttpProxyServer {

    // Default provider to Infura
    if providerType == "infura" {
        providerUrl = fmt.Sprintf(InfuraURL, network, projectId)
    } else if providerType == "pocket" {
        providerUrl = fmt.Sprintf(PocketURL, network, projectId)
    } else if providerUrl == "" {
        fmt.Printf("Unknown provider [%s] and no providerUrl was provided, exiting.\n", providerType)
        os.Exit(1)
    }

    // Create and return proxy server
    return &HttpProxyServer{
        Port: port,
        ProviderUrl: providerUrl,
        Verbose: verbose,
        ProviderType: providerType,
    }

}


// Start proxy server
func (p *HttpProxyServer) Start() error {

    // Log
    log.Printf("Proxy server listening on port %s\n", p.Port)

    // Listen on RPC port
    return http.ListenAndServe(":" + p.Port, p)

}


// Handle request / serve response
func (p *HttpProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    p.idLock.Lock()
    messageId := p.id
    p.id++
    p.idLock.Unlock()

    // Log request
    log.Printf("New %s request received from %s\n", r.Method, r.RemoteAddr)

    // Get request content type
    contentTypes, ok := r.Header["Content-Type"]
    if !ok || len(contentTypes) == 0 {
        log.Println(errors.New("Request Content-Type header not specified"))
        _,_ = fmt.Fprintln(w, errors.New("Request Content-Type header not specified"))
        return
    }

    // Get the request body as a string
    requestBuffer := new(bytes.Buffer)
    _, err := requestBuffer.ReadFrom(r.Body)
    if err != nil {
        log.Println(fmt.Errorf("Error getting request body string: %w", err))
        _, _ = fmt.Fprintln(w, fmt.Errorf("Error getting request body string: %w", err))
        return
    }
    requestBody := requestBuffer.String()

    // Log request if in verbose mode
    if p.Verbose {
        fmt.Printf("(< %d) %s\n", requestBody)
    }

    // Handle the request
    responseReader, err := p.handleRequest(contentTypes[0], requestBody, messageId, 0)
    if err != nil {
        log.Println(err.Error())
        _, _ =fmt.Fprintln(w, err.Error())
        return
    }

    // Set response writer header
    w.Header().Set("Content-Type", "application/json")

    // Copy provider response body to response writer
    _, err = io.Copy(w, responseReader)
    if err != nil {
        log.Println(fmt.Errorf("Error reading response from remote server: %w", err))
        _, _ =fmt.Fprintln(w, fmt.Errorf("Error reading response from remote server: %w", err))
        return
    }

    // Log success
    log.Printf("Response sent to %s successfully\n", r.RemoteAddr)
}


// Handle request / serve response
func (p *HttpProxyServer) handleRequest(contentType, requestBody string, messageId uint64, recursionCount int) (io.Reader, error) {

    // Error out if we've tried too many times
    if recursionCount >= HandleRequestRecursionLimit {
        return nil, fmt.Errorf("Request hit the rate limit too many times.")
    }

    // Forward request to provider
    var reader io.Reader
    reader = strings.NewReader(requestBody)
    response, err := http.Post(p.ProviderUrl, contentType, reader)
    if err != nil {
        return nil, fmt.Errorf("Error forwarding request to remote server: %w", err)
    }
    defer func() {
        _ =response.Body.Close()
    }()

    // Get the response body as a string
    responseBuffer := new(bytes.Buffer)
    _, err = responseBuffer.ReadFrom(response.Body)
    if err != nil {
        return nil, fmt.Errorf("Error getting response body string: %w", err)
    }
    responseBody := responseBuffer.String()

    // Log response if in verbose mode
    if p.Verbose {
        fmt.Printf("(> %d) %s\n", messageId, responseBody)
    }
    
    // If using Infura, check for a rate limit error
    if p.ProviderType == "infura" && response.StatusCode == 429 {
        
        // Unmarshal it into an object
        var infuraError InfuraRateLimitError
        err = json.Unmarshal(responseBuffer.Bytes(), &infuraError)
        if err != nil {
            return nil, fmt.Errorf("Received a 429 from Infura but failed deserializing: %w", err)
        }
        
        // Wait for the requested number of seconds, then try again
        secondsToWait := int(math.Ceil(infuraError.Error.Data.Rate.BackoffSeconds))
        log.Printf("Infura rate limit hit, waiting %d seconds... (Attempt %d of %d)\n", secondsToWait, recursionCount + 1, HandleRequestRecursionLimit)
        time.Sleep(time.Duration(secondsToWait) * time.Second)
        return p.handleRequest(contentType, requestBody, messageId, recursionCount + 1)
    } else if p.ProviderType == "pocket" && response.StatusCode == 502 {
        log.Printf("Pocket returned a 502 gateway error, trying again... (Attempt %d of %d)\n", recursionCount + 1, HandleRequestRecursionLimit)
        return p.handleRequest(contentType, requestBody, messageId, recursionCount + 1)
    }
    
    // Success, return the body
    return strings.NewReader(responseBody), nil

}

