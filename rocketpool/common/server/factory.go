package server

import "github.com/gorilla/mux"

// Context factories can implement this generally so they can register themselves with an HTTP router.
type IContextFactory interface {
	RegisterRoute(router *mux.Router)
}
