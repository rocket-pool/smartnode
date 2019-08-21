package beaconchain

import (
    "errors"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/gorilla/websocket"

    "github.com/rocket-pool/smartnode/shared/utils/messaging"
)


// Config
const RECONNECT_INTERVAL string = "10s"
var reconnectInterval, _ = time.ParseDuration(RECONNECT_INTERVAL)


// Client
type Client struct {
    providerUrl string
    publisher *messaging.Publisher
    log *log.Logger
    connection *websocket.Conn
    connectionTimer *time.Timer
    readLock sync.Mutex
    writeLock sync.Mutex
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
    Balance uint    `json:"balance"`
    Number uint     `json:"number"`
    Action string   `json:"action"`
    Error string    `json:"error"`
}


/**
 * Create client
 */
func NewClient(providerUrl string, publisher *messaging.Publisher, logger *log.Logger) *Client {
    return &Client{
        providerUrl: providerUrl,
        publisher: publisher,
        log: logger,
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

    // Lock for write
    c.writeLock.Lock()
    defer c.writeLock.Unlock()

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
        c.log.Println(errors.New("Error connecting to beacon chain server: " + err.Error()))
        c.log.Println(fmt.Sprintf("Retrying in %s...", reconnectInterval.String()))
        c.connectionTimer.Reset(reconnectInterval)
        return

    } else {

        // Log success & notify
        defer connection.Close()
        c.connection = connection
        c.log.Println("Connected to beacon chain server at", c.providerUrl)
        c.publisher.Notify("beacon.client.connected", struct{Client *Client}{c})

    }

    // Handle beacon messages
    closed := make(chan struct{})
    go (func() {
        defer close(closed)
        for {
            if message, err, didClose := c.readMessage(); err != nil {
                c.log.Println(err)
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
            c.log.Println(fmt.Sprintf("Connection closed, reconnecting in %s...", reconnectInterval.String()))
            c.connectionTimer.Reset(reconnectInterval)
    }

}


/**
 * Read message from beacon chain server
 */
func (c *Client) readMessage() ([]byte, error, bool) {

    // Lock for read
    c.readLock.Lock()
    defer c.readLock.Unlock()

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

