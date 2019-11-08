package proxy

import (
    "fmt"
    "io"
    "log"
    "net/http"
)


// Config
const INFURA_URL = "https://%s.infura.io/v3/%s";
const INFURA_NETWORK = "mainnet";
const INFURA_PROJECT_ID = "d690a0156a994dd785c0a64423586f52";


// Proxy server
type ProxyServer struct {}


/**
 * Create proxy server
 */
func NewProxyServer() *ProxyServer {
    return &ProxyServer{}
}


/**
 * Start proxy server
 */
func (p *ProxyServer) Start() {

    // Listen on RPC port
    log.Fatal(http.ListenAndServe(":8545", p))

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
    response, err := http.Post(fmt.Sprintf(INFURA_URL, INFURA_NETWORK, INFURA_PROJECT_ID), contentTypes[0], r.Body)
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

