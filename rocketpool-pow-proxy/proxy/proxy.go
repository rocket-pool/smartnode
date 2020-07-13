package proxy

import (
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"
)


// Config
const INFURA_URL = "https://%s.infura.io/v3/%s"


// Proxy server
type ProxyServer struct {
    Port string
    ProviderUrl string
}


// Create new proxy server
func NewProxyServer(port string, providerUrl string, network string, projectId string) *ProxyServer {

    // Default provider to Infura
    if providerUrl == "" {
        providerUrl = fmt.Sprintf(INFURA_URL, network, projectId)
    }

    // Create and return proxy server
    return &ProxyServer{
        Port: port,
        ProviderUrl: providerUrl,
    }

}


// Start proxy server
func (p *ProxyServer) Start() error {

    // Log
    log.Println(fmt.Sprintf("Proxy server listening on port %s", p.Port))

    // Listen on RPC port
    return http.ListenAndServe(":" + p.Port, p)

}


// Handle request / serve response
func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    // Log request
    log.Println(fmt.Sprintf("New %s request received from %s", r.Method, r.RemoteAddr))

    // Get request content type
    contentTypes, ok := r.Header["Content-Type"]
    if !ok || len(contentTypes) == 0 {
        log.Println(errors.New("Request Content-Type header not specified"))
        fmt.Fprintln(w, errors.New("Request Content-Type header not specified"))
        return
    }

    // Forward request to provider
    response, err := http.Post(p.ProviderUrl, contentTypes[0], r.Body)
    if err != nil {
        log.Println(fmt.Errorf("Error forwarding request to remote server: %w", err))
        fmt.Fprintln(w, fmt.Errorf("Error forwarding request to remote server: %w", err))
        return
    }
    defer response.Body.Close()

    // Set response writer header
    w.Header().Set("Content-Type", "application/json")

    // Copy provider response body to response writer
    _, err = io.Copy(w, response.Body)
    if err != nil {
        log.Println(fmt.Errorf("Error reading response from remote server: %w", err))
        fmt.Fprintln(w, fmt.Errorf("Error reading response from remote server: %w", err))
        return
    }

    // Log success
    log.Println(fmt.Sprintf("Response sent to %s successfully", r.RemoteAddr))

}

