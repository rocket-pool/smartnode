package node

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/rocketpool/node/routes"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
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
		log.Printf("%s %s %s %d %s", requestIP(r), r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// tokenBucket is a simple per-process limiter: refill at `rate` tokens per
// second, capacity equal to the rate (burst of one second).
type tokenBucket struct {
	mu     sync.Mutex
	rate   float64
	tokens float64
	last   time.Time
}

func newTokenBucket(perSecond float64) *tokenBucket {
	if perSecond <= 0 {
		return nil
	}
	return &tokenBucket{
		rate:   perSecond,
		tokens: perSecond,
		last:   time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	if b == nil {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	b.tokens += now.Sub(b.last).Seconds() * b.rate
	if b.tokens > b.rate {
		b.tokens = b.rate
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func rateLimitMiddleware(limiter *tokenBucket, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.allow() {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Retry-After", "1")
		response.WriteErrorResponse(w, &response.TooManyRequestsError{})
	})
}

func requestIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func apiListenHost(cfg *config.RocketPoolConfig, mode cfgtypes.RPCMode) (string, bool) {
	if cfg.IsNativeMode {
		switch mode {
		case cfgtypes.RPC_OpenExternal:
			return "0.0.0.0", true
		case cfgtypes.RPC_OpenLocalhost:
			return "127.0.0.1", true
		default:
			return "", false
		}
	}
	// Docker: always bind on all interfaces inside the container so published
	// host ports (and other compose services) can reach the server.
	return "0.0.0.0", true
}

// startHTTP starts the node's HTTP API server and returns immediately.
// The server runs in the background for the lifetime of the process.
func startHTTP(ctx context.Context, c *cli.Command, cfg *config.RocketPoolConfig) {
	port, ok := cfg.Api.ApiPort.Value.(uint16)
	if !ok || port == 0 {
		log.Println("Warning: APIPort not configured, HTTP API server will not start.")
		return
	}

	mode, _ := cfg.Api.OpenApiPort.Value.(cfgtypes.RPCMode)
	host, listen := apiListenHost(cfg, mode)
	if !listen {
		log.Println("Node HTTP API server is closed; not listening.")
		return
	}

	tokenPath := cfg.Api.GetAPITokenPath()
	if err := cfg.SyncAPITokenFromDisk(false); err != nil {
		log.Printf("Warning: could not load API token from %s: %v", tokenPath, err)
	}
	expectedToken, _ := cfg.Api.APIToken.Value.(string)
	if expectedToken == "" {
		log.Printf("Warning: API token is empty (file %s); authenticated API routes will reject all requests.", tokenPath)
	}

	scope, _ := cfg.Api.TokenScope.Value.(cfgtypes.APITokenScope)
	sensitiveOnly := scope == cfgtypes.APITokenScope_Sensitive

	var perSecond float64
	v := cfg.Api.RateLimit.Value.(uint16)
	perSecond = float64(v)
	limiter := newTokenBucket(perSecond)

	router := snroute.NewRouter(expectedToken, sensitiveOnly)
	routes.RegisterRoutes(router, c)

	handler := loggingMiddleware(rateLimitMiddleware(limiter, router))

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Node HTTP API server listening on %s:%d (token file %s)\n", host, port, tokenPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Node HTTP API server error: %v\n", err)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
}
