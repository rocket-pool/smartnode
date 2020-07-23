package proxy

import (
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"
)


// Config
const InfuraURL = "https://%s.infura.io/v3/%s"


// Proxy server
type ProxyServer struct {
    Port string
    ProviderUrl string
}


// Create new proxy server
func NewProxyServer(port string, providerUrl string, network string, projectId string) *ProxyServer {

    // Default provider to Infura
    if providerUrl == "" {
        providerUrl = fmt.Sprintf(InfuraURL, network, projectId)
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
    log.Printf("Proxy server listening on port %s\n", p.Port)

    // Listen on RPC port
    return http.ListenAndServe(":" + p.Port, p)

}


// Handle request / serve response
func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    // Log request
    log.Printf("New %s request received from %s\n", r.Method, r.RemoteAddr)

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
    log.Printf("Response sent to %s successfully\n", r.RemoteAddr)

}

