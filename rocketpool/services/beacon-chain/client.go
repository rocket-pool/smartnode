package beaconchain

import (
    "errors"
    "fmt"
    "log"
    "time"

    "github.com/gorilla/websocket"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Config
const RECONNECT_INTERVAL string = "15s"
var reconnectInterval, _ = time.ParseDuration(RECONNECT_INTERVAL)


// Client
type Client struct {
    providerUrl string
    publisher *messaging.Publisher
    connection *websocket.Conn
    connectionTimer *time.Timer
}


// Client message to server
type ClientMessage struct {
    Message string  `json:"message"`
    Pubkey string   `json:"pubkey"`
}


// Server message to client
type ServerMessage struct {
    Message string  `json:"message"`
    Pubkey string   `json:"pubkey"`
    Status struct {
        Code string     `json:"code"`
        Initiated struct {
            Exit uint       `json:"exit"`
        }               `json:"initiated"`
    }               `json:"status"`
    Action string   `json:"action"`
    Error string    `json:"error"`
}


/**
 * Create client
 */
func NewClient(providerUrl string, publisher *messaging.Publisher) *Client {
    return &Client{
        providerUrl: providerUrl,
        publisher: publisher,
    }
}


/**
 * Get persistent connection to beacon chain server
 */
func (c *Client) Connect() {

    // Cancel if already maintaining connection
    if c.connectionTimer != nil {
        return
    }

    // Initialise beacon connection timer
    go (func() {
        now, _ := time.ParseDuration("0s")
        c.connectionTimer = time.NewTimer(now)
        for _ = range c.connectionTimer.C {
            c.connect()
        }
    })()

}


/**
 * Send message to beacon chain server
 */
func (c *Client) Send(payload []byte) error {

    // Check connection is open
    if c.connection == nil {
        return errors.New("Cannot send to beacon chain server while connection is closed")
    }

    // Send
    return c.connection.WriteMessage(websocket.TextMessage, payload)

}


/**
 * Connect to beacon chain server
 */
func (c *Client) connect() {

    // Open websocket connection to beacon chain server
    if connection, _, err := websocket.DefaultDialer.Dial(c.providerUrl, nil); err != nil {

        // Log connection errors and retry
        log.Println(errors.New("Error connecting to beacon chain server: " + err.Error()))
        log.Println(fmt.Sprintf("Retrying in %s...", reconnectInterval.String()))
        c.connectionTimer.Reset(reconnectInterval)
        return

    } else {

        // Log success & notify
        defer connection.Close()
        c.connection = connection
        log.Println("Connected to beacon chain server at", c.providerUrl)
        c.publisher.Notify("beacon.client.connected", struct{Client *Client}{c})

    }

    // Handle beacon messages
    closed := make(chan struct{})
    go (func() {
        defer close(closed)
        for {
            if message, err, didClose := c.readMessage(); err != nil {
                log.Println(err)
                if didClose { return }
            } else {
                c.publisher.Notify("beacon.client.message", struct{Client *Client; Message []byte}{c, message})
            }
        }
    })()

    // Block thread until closed, notify & reconnect
    select {
        case <-closed:
            c.connection = nil
            c.publisher.Notify("beacon.client.disconnected", struct{Client *Client}{c})
            log.Println(fmt.Sprintf("Connection closed, reconnecting in %s...", reconnectInterval.String()))
            c.connectionTimer.Reset(reconnectInterval)
    }

}


/**
 * Read message from beacon chain server
 */
func (c *Client) readMessage() ([]byte, error, bool) {

    // Read message
    messageType, message, err := c.connection.ReadMessage()
    if err != nil {
        return nil, errors.New("Error reading beacon message: " + err.Error()), websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err)
    }

    // Check message type
    if messageType != websocket.TextMessage {
        return nil, errors.New("Unrecognised beacon message type"), false
    }

    // Return
    return message, nil, false

}

