package snroute

import (
	"net/http"
	"strings"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services/apitoken"
)

// Route is an HTTP handler bound to a path with a compile-time auth class.
type Route interface {
	http.Handler
	Path() string
	RequiresToken(sensitiveOnly bool) bool
}

// OpenRoute never requires a bearer token (health checks).
type OpenRoute struct {
	path    string
	handler http.Handler
}

// ReadRoute requires a token only when Token Scope is "all endpoints".
type ReadRoute struct {
	path    string
	handler http.Handler
}

// WriteRoute always requires a bearer token: transaction submission and
// high-impact local operations (wallet secrets, exits, signing).
type WriteRoute struct {
	path    string
	handler http.Handler
}

var (
	_ Route = OpenRoute{}
	_ Route = ReadRoute{}
	_ Route = WriteRoute{}
)

// Open returns a route that never requires a bearer token.
func Open(path string, h http.HandlerFunc) OpenRoute {
	return OpenRoute{path: path, handler: h}
}

// Read returns a route that requires a token only when Token Scope is all endpoints.
func Read(path string, h http.HandlerFunc) ReadRoute {
	return ReadRoute{path: path, handler: h}
}

// Write returns a route that always requires a bearer token.
func Write(path string, h http.HandlerFunc) WriteRoute {
	return WriteRoute{path: path, handler: h}
}

func (r OpenRoute) Path() string  { return r.path }
func (r ReadRoute) Path() string  { return r.path }
func (r WriteRoute) Path() string { return r.path }

func (r OpenRoute) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}
func (r ReadRoute) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}
func (r WriteRoute) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (OpenRoute) RequiresToken(bool) bool               { return false }
func (ReadRoute) RequiresToken(sensitiveOnly bool) bool { return !sensitiveOnly }
func (WriteRoute) RequiresToken(bool) bool              { return true }

// Router registers typed routes and applies per-route bearer auth.
type Router struct {
	mux           *http.ServeMux
	expectedToken string
	sensitiveOnly bool
	routes        []Route
}

// NewRouter returns a router that requires a bearer token according to each
// route's class and the API Token Scope setting.
func NewRouter(expectedToken string, sensitiveOnly bool) *Router {
	return &Router{
		mux:           http.NewServeMux(),
		expectedToken: expectedToken,
		sensitiveOnly: sensitiveOnly,
	}
}

// Handle registers rt. Auth is applied here so a Write route cannot be
// registered without a token check.
func (r *Router) Handle(rt Route) {
	r.routes = append(r.routes, rt)
	var h http.Handler = rt
	if rt.RequiresToken(r.sensitiveOnly) {
		h = requireBearer(r.expectedToken, h)
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

func requireBearer(expectedToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expectedToken == "" || !apitoken.Equal(expectedToken, bearerToken(r)) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			response.WriteErrorResponse(w, &response.UnauthorizedError{})
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
