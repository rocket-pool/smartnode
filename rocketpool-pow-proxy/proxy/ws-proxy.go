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
}


// Handle request / serve response
func (p *WsProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    var upgrader = websocket.Upgrader{} // use default options

    // Establish a websocket with the requester
    eth2Connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
        log.Println(fmt.Errorf("Error upgrading websocket: %w", err))
        fmt.Fprintln(w, fmt.Errorf("Error upgrading websocket: %w", err))
		return
	}
	defer c.Close()

    // Connect to Infura
    infuraConnection, _, err := websocket.DefaultDialer.Dial(p.ProviderUrl, nil)
    if err != nil {
        log.Println(fmt.Errorf("Error connecting to Infura: %w", err))
        fmt.Fprintln(w, fmt.Errorf("Error connecting to Infura: %w", err))
	}
	defer c.Close()

    // Wait groups for the proxy loops
    wg := new(sync.WaitGroup)
    wg.Add(2)

    // Run the eth2-to-Infura loop
	go func() {
        for {
            // Read from eth2
            mt, message, err := eth2Connection.ReadMessage()
		    if err != nil {
                log.Println(fmt.Errorf("Error reading from eth2: %w", err))
                fmt.Fprintln(w, fmt.Errorf("Error reading from eth2: %w", err))
			    break
		    }

            // Log request
            log.Print("New websocket message request received from Infura\n")

            // Send it to Infura
            if err = infuraConnection.WriteMessage(mt, message); err != nil {
                log.Println(fmt.Errorf("Error writing to Infura: %w", err))
                fmt.Fprintln(w, fmt.Errorf("Error writing to Infura: %w", err))
			    break
		    }
        }

        wg.Done()
	}()
	
    // Run the Infura-to-eth2 loop
    go func() {
        for {
            // Read from Infura
            mt, message, err := infuraConnection.ReadMessage()
		    if err != nil {
                log.Println(fmt.Errorf("Error reading from Infura: %w", err))
                fmt.Fprintln(w, fmt.Errorf("Error reading from Infura: %w", err))
			    break
		    }

            // Log request
            log.Print("New websocket message request received from Infura\n")

            // Send it to eth2
            if err = eth2Connection.WriteMessage(mt, message); err != nil {
                log.Println(fmt.Errorf("Error writing to eth2: %w", err))
                fmt.Fprintln(w, fmt.Errorf("Error writing to eth2: %w", err))
			    break
		    }
        }

        wg.Done()
    }()

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

    // Wait for both loops to stop
    return wg.Wait()
}
