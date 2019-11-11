package proxy

import (
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"
)


// Config
const INFURA_URL = "https://%s.infura.io/v3/%s";


// Proxy server
type ProxyServer struct {
    Port string
    Network string
    ProjectId string
}


/**
 * Create proxy server
 */
func NewProxyServer(port string, network string, projectId string) *ProxyServer {
    return &ProxyServer{
        Port: port,
        Network: network,
        ProjectId: projectId,
    }
}


/**
 * Start proxy server
 */
func (p *ProxyServer) Start() error {

    // Log
    log.Println(fmt.Sprintf("Proxy server listening on port %s", p.Port))

    // Listen on RPC port
    return http.ListenAndServe(":" + p.Port, p)

}


/**
 * Handle request / serve response
 */
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

    // Forward request to infura
    response, err := http.Post(fmt.Sprintf(INFURA_URL, p.Network, p.ProjectId), contentTypes[0], r.Body)
    if err != nil {
        log.Println(errors.New("Error forwarding request to remote server: " + err.Error()))
        fmt.Fprintln(w, errors.New("Error forwarding request to remote server: " + err.Error()))
        return
    }
    defer response.Body.Close()

    // Copy infura response body to response writer
    _, err = io.Copy(w, response.Body)
    if err != nil {
        log.Println(errors.New("Error reading response from remote server: " + err.Error()))
        fmt.Fprintln(w, errors.New("Error reading response from remote server: " + err.Error()))
        return
    }

    // Log success
    log.Println(fmt.Sprintf("Response sent to %s successfully", r.RemoteAddr))

}

