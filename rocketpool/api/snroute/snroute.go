package snroute

import (
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/apitoken"
)

// Route is an HTTP endpoint bound to a path with a compile-time auth class.
type Route interface {
	Path() string
	RequiresToken(sensitiveOnly bool) bool
	serve(ctx Context)
}

// Context is the request view available to Open and Read handlers.
// It has no method that returns a submitting transactor.
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	cmd     *cli.Command
}

// Command is the daemon CLI command used to load services.
func (c Context) Command() *cli.Command { return c.cmd }

// WriteContext is the request view available to Write handlers.
// Transactor is the only way HTTP handlers may obtain submit opts.
type WriteContext struct{ Context }

// TransactOpts is a node-account transactor for submitting a transaction.
// The only constructor is WriteContext.Transactor; the inner opts are
// unexported so a Read handler cannot forge one.
type TransactOpts struct {
	opts *bind.TransactOpts
}

// Opts returns the go-ethereum transactor for bindings.
func (t *TransactOpts) Opts() *bind.TransactOpts { return t.opts }

// Transactor returns submit opts with gas fields from the request form.
func (c WriteContext) Transactor() (*TransactOpts, error) {
	w, err := services.GetWallet(c.cmd)
	if err != nil {
		return nil, err
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	r := c.Request
	if maxFeeStr := r.FormValue("maxFee"); maxFeeStr != "" {
		if maxFeeGwei, parseErr := strconv.ParseFloat(maxFeeStr, 64); parseErr == nil && maxFeeGwei > 0 {
			opts.GasFeeCap = math.GweiToWei(maxFeeGwei)
		}
	}
	if maxPrioFeeStr := r.FormValue("maxPrioFee"); maxPrioFeeStr != "" {
		if maxPrioFeeGwei, parseErr := strconv.ParseFloat(maxPrioFeeStr, 64); parseErr == nil && maxPrioFeeGwei > 0 {
			opts.GasTipCap = math.GweiToWei(maxPrioFeeGwei)
		}
	}
	if gasLimitStr := r.FormValue("gasLimit"); gasLimitStr != "" {
		if gasLimit, parseErr := strconv.ParseUint(gasLimitStr, 10, 64); parseErr == nil {
			opts.GasLimit = gasLimit
		}
	}
	if nonceStr := r.FormValue("nonce"); nonceStr != "" {
		if nonce, ok := new(big.Int).SetString(nonceStr, 0); ok {
			opts.Nonce = nonce
		}
	}
	return &TransactOpts{opts: opts}, nil
}

// OpenRoute never requires a bearer token (health checks).
type OpenRoute struct {
	path    string
	handler func(Context)
}

// ReadRoute requires a token only when Token Scope is "all endpoints".
type ReadRoute struct {
	path    string
	handler func(Context)
}

// WriteRoute always requires a bearer token: transaction submission and
// high-impact local operations (wallet secrets, exits, signing).
type WriteRoute struct {
	path    string
	handler func(WriteContext)
}

var (
	_ Route = OpenRoute{}
	_ Route = ReadRoute{}
	_ Route = WriteRoute{}
)

// Open returns a route that never requires a bearer token.
func Open(path string, h func(Context)) OpenRoute {
	return OpenRoute{path: path, handler: h}
}

// Read returns a route that requires a token only when Token Scope is all endpoints.
func Read(path string, h func(Context)) ReadRoute {
	return ReadRoute{path: path, handler: h}
}

// Write returns a route that always requires a bearer token.
func Write(path string, h func(WriteContext)) WriteRoute {
	return WriteRoute{path: path, handler: h}
}

func (r OpenRoute) Path() string  { return r.path }
func (r ReadRoute) Path() string  { return r.path }
func (r WriteRoute) Path() string { return r.path }

func (r OpenRoute) serve(ctx Context)  { r.handler(ctx) }
func (r ReadRoute) serve(ctx Context)  { r.handler(ctx) }
func (r WriteRoute) serve(ctx Context) { r.handler(WriteContext{ctx}) }

func (OpenRoute) RequiresToken(bool) bool { return false }
func (ReadRoute) RequiresToken(unauthenticatedReads bool) bool {
	return !unauthenticatedReads
}
func (WriteRoute) RequiresToken(bool) bool { return true }

// RegisterTo adds the route to router. Handle remains the primitive that
// applies auth; this is the one-line registration used in routes.go files.
func (r OpenRoute) RegisterTo(router *Router)  { router.Handle(r) }
func (r ReadRoute) RegisterTo(router *Router)  { router.Handle(r) }
func (r WriteRoute) RegisterTo(router *Router) { router.Handle(r) }

// Router registers typed routes and applies per-route bearer auth.
type Router struct {
	mux                  *http.ServeMux
	cmd                  *cli.Command
	tokens               []apitoken.Record
	unauthenticatedReads bool
	routes               []Route
}

// NewRouter returns a router that authenticates against tokens.
// unauthenticatedReads allows Read routes without a bearer token; Write
// routes always require a write-scoped token.
func NewRouter(cmd *cli.Command, tokens []apitoken.Record, unauthenticatedReads bool) *Router {
	return &Router{
		mux:                  http.NewServeMux(),
		cmd:                  cmd,
		tokens:               tokens,
		unauthenticatedReads: unauthenticatedReads,
	}
}

// Handle registers rt. Auth is applied here so a Write route cannot be
// registered without a token check.
func (r *Router) Handle(rt Route) {
	r.routes = append(r.routes, rt)
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rt.serve(Context{Writer: w, Request: req, cmd: r.cmd})
	})
	switch rt.(type) {
	case OpenRoute:
		// never authed
	case ReadRoute:
		if !r.unauthenticatedReads {
			h = requireToken(r.tokens, false, h)
		}
	default:
		h = requireToken(r.tokens, true, h)
	}
	r.mux.Handle(rt.Path(), h)
}

// Routes returns registered routes in registration order.
func (r *Router) Routes() []Route {
	return r.routes
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func requireToken(tokens []apitoken.Record, writeOnly bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec, ok := (apitoken.File{Tokens: tokens}).Lookup(bearerToken(r))
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			response.WriteErrorResponse(w, &response.UnauthorizedError{})
			return
		}
		if writeOnly && rec.Scope != apitoken.ScopeWrite {
			response.WriteErrorResponse(w, &response.ForbiddenError{})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
