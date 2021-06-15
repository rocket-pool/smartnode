package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Config
const InfuraURL = "https://%s.infura.io/v3/%s"
const PocketURL = "https://%s.gateway.pokt.network/v1/%s"


// Proxy server
type HttpProxyServer struct {
    Port string
    ProviderUrl string
}


// Create new proxy server
func NewHttpProxyServer(port string, providerUrl string, network string, projectId string, providerType string) *HttpProxyServer {

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

    // Log request
    log.Printf("New %s request received from %s\n", r.Method, r.RemoteAddr)

    // Get request content type
    contentTypes, ok := r.Header["Content-Type"]
    if !ok || len(contentTypes) == 0 {
        log.Println(errors.New("Request Content-Type header not specified"))
        _,_ = fmt.Fprintln(w, errors.New("Request Content-Type header not specified"))
        return
    }

    // Forward request to provider
    response, err := http.Post(p.ProviderUrl, contentTypes[0], r.Body)
    if err != nil {
        log.Println(fmt.Errorf("Error forwarding request to remote server: %w", err))
        _, _ = fmt.Fprintln(w, fmt.Errorf("Error forwarding request to remote server: %w", err))
        return
    }
    defer func() {
        _ =response.Body.Close()
    }()

    // Set response writer header
    w.Header().Set("Content-Type", "application/json")

    // Copy provider response body to response writer
    _, err = io.Copy(w, response.Body)
    if err != nil {
        log.Println(fmt.Errorf("Error reading response from remote server: %w", err))
        _, _ =fmt.Fprintln(w, fmt.Errorf("Error reading response from remote server: %w", err))
        return
    }

    // Log success
    log.Printf("Response sent to %s successfully\n", r.RemoteAddr)

}

