package proxy

import (
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)


// Config
const InfuraURL = "wss://%s.infura.io/ws/v3/%s"


// Proxy server
type WsProxyServer struct {
    Port string
    ProviderUrl string
}


// Create new proxy server
func NewWsProxyServer(port string, providerUrl string, network string, projectId string) *WsProxyServer {

    // Default provider to Infura
    if providerUrl == "" {
        providerUrl = fmt.Sprintf(InfuraURL, network, projectId)
    }

    // Create and return proxy server
    return &WsProxyServer{
        Port: port,
        ProviderUrl: providerUrl,
    }

}


// Start proxy server
func (p *WsProxyServer) Start() error {

    // Log
    log.Printf("Proxy server listening on port %s\n", p.Port)

    // Listen on RPC port
    return http.ListenAndServe(":" + p.Port, p)

    // TODO: CONNECT TO INFURA

}


// Handle request / serve response
func (p *WsProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    var upgrader = websocket.Upgrader{} // use default options

    c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
        log.Println(fmt.Errorf("Error upgrading websocket: %w", err))
        fmt.Fprintln(w, fmt.Errorf("Error upgrading websocket: %w", err))
		return
	}

	defer c.Close()
	
    for {
		mt, message, err := c.ReadMessage()
		if err != nil {
            log.Println(fmt.Errorf("Error reading from websocket: %w", err))
            fmt.Fprintln(w, fmt.Errorf("Error reading from websocket: %w", err))
			break
		}

        // Log request
        log.Printf("New websocket message request received from %s\n", r.RemoteAddr)

        // TODO: FORWARD IT


		err = c.WriteMessage(mt, message)
		if err != nil {
            log.Println(fmt.Errorf("Error writing to websocket: %w", err))
            fmt.Fprintln(w, fmt.Errorf("Error writing to websocket: %w", err))
			break
		}
	}

    /*

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

    */

}

