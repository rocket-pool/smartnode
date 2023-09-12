package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

type Server struct {
	log        *log.ColorLogger
	socketPath string
	socket     net.Listener
	server     http.Server
}

// Creates a new Server instance
func NewServer(cfg *config.RocketPoolConfig, log *log.ColorLogger) *Server {
	// Get the socket file path
	return &Server{
		log:        log,
		socketPath: cfg.Smartnode.GetSocketPath(),
	}
}

// Starts the server and begins listening for incoming HTTP requests
func (s *Server) Start() error {
	// Create the socket
	socket, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	s.socket = socket

	// Register the routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Default route"))
	})

	// Create the HTTP server
	s.server = http.Server{
		Handler: mux,
	}

	// Start listening
	go func() {
		err := s.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			s.log.Printlnf("error while listening for HTTP requests: %s", err.Error())
		}
	}()

	return nil
}

// Stops the server
func (s *Server) Stop() error {
	// Shutdown the listener
	err := s.server.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("error stopping listener: %w", err)
	}

	// Remove the socket file
	err = os.Remove(s.socketPath)
	if err != nil {
		return fmt.Errorf("error removing socket file: %w", err)
	}

	return nil
}
