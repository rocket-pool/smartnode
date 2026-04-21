package node

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/node/routes"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// statusRecorder wraps http.ResponseWriter to capture the written status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Write implements the ResponseWriter interface so we don't lose the original
// Write behavior when wrapping.
func (r *statusRecorder) Write(b []byte) (int, error) {
	return r.ResponseWriter.Write(b)
}

// loggingMiddleware logs method, path, status code, and elapsed time for every request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// startHTTP starts the node's HTTP API server and returns immediately.
// The server runs in the background for the lifetime of the process.
func startHTTP(ctx context.Context, c *cli.Command, cfg *config.RocketPoolConfig) {
	port, ok := cfg.Smartnode.APIPort.Value.(uint16)
	if !ok || port == 0 {
		log.Println("Warning: APIPort not configured, HTTP API server will not start.")
		return
	}

	var host string
	if !cfg.IsNativeMode {
		// In Docker mode the server must bind to 0.0.0.0, so other containers can reach it.
		host = "0.0.0.0"
	} else {
		host = "127.0.0.1"
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, c)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: loggingMiddleware(mux),
	}

	go func() {
		log.Printf("Node HTTP API server listening on %s:%d\n", host, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Node HTTP API server error: %v\n", err)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
}
