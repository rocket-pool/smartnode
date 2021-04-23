package proxy

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Config
const InfuraWsURL = "wss://%s.infura.io/ws/v3/%s"


// Proxy server
type WsProxyServer struct {
    Port string
    ProviderUrl string
}


// Create new proxy server
func NewWsProxyServer(port string, providerUrl string, network string, projectId string) *WsProxyServer {

    // Default provider to Infura
    if providerUrl == "" {
        providerUrl = fmt.Sprintf(InfuraWsURL, network, projectId)
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
}


// Handle request / serve response
func (p *WsProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

    // Establish a websocket with the requester
    eth2Connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
        log.Println(fmt.Errorf("Error upgrading websocket: %w", err))
        fmt.Fprintln(w, fmt.Errorf("Error upgrading websocket: %w", err))
		return
	}
	defer eth2Connection.Close()

    // Connect to Infura
    infuraConnection, _, err := websocket.DefaultDialer.Dial(p.ProviderUrl, nil)
    if err != nil {
        log.Println(fmt.Errorf("Error connecting to remote websocket: %w", err))
        fmt.Fprintln(w, fmt.Errorf("Error connecting to remote websocket: %w", err))
	}
	defer infuraConnection.Close()

    // Wait groups for the proxy loops
    wg := new(sync.WaitGroup)
    wg.Add(2)

    // Run the eth2-to-remote loop
	go func() {
        for {
            // Read from eth2
            mt, message, err := eth2Connection.ReadMessage()
		    if err != nil {
                log.Println(fmt.Errorf("Error reading from eth2: %w", err))
                fmt.Fprintln(w, fmt.Errorf("Error reading from eth2: %w", err))
			    break
		    }

            // Send it to the remote server
            if err = infuraConnection.WriteMessage(mt, message); err != nil {
                log.Println(fmt.Errorf("Error writing to remote websocket: %w", err))
                fmt.Fprintln(w, fmt.Errorf("Error writing to remote websocket: %w", err))
			    break
		    }
        }

        wg.Done()
	}()
	
    // Run the remote-to-eth2 loop
    go func() {
        for {
            // Read from the remote server
            mt, message, err := infuraConnection.ReadMessage()
		    if err != nil {
                log.Println(fmt.Errorf("Error reading from remote websocket: %w", err))
                fmt.Fprintln(w, fmt.Errorf("Error reading from remote websocket: %w", err))
			    break
		    }

            // Send it to eth2
            if err = eth2Connection.WriteMessage(mt, message); err != nil {
                log.Println(fmt.Errorf("Error writing to eth2: %w", err))
                fmt.Fprintln(w, fmt.Errorf("Error writing to eth2: %w", err))
			    break
		    }
        }

        wg.Done()
    }()

    // Wait for both loops to stop
	wg.Wait()
	return
}
