package proxy

import (
    "fmt"
    "io"
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

    // Listen on RPC port
    return http.ListenAndServe(":" + p.Port, p)

}


/**
 * Handle request / serve response
 */
func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    // Get request content type
    contentTypes, ok := r.Header["Content-Type"]
    if !ok || len(contentTypes) == 0 {
        fmt.Fprintln(w, "Content-Type header not specified")
        return
    }

    // Forward request to infura
    response, err := http.Post(fmt.Sprintf(INFURA_URL, p.Network, p.ProjectId), contentTypes[0], r.Body)
    if err != nil {
        fmt.Fprintln(w, "Error forwarding request to remote server")
        return
    }
    defer response.Body.Close()

    // Copy infura response body to response writer
    _, err = io.Copy(w, response.Body)
    if err != nil {
        fmt.Fprintln(w, "Error reading response from remote server")
        return
    }

}

