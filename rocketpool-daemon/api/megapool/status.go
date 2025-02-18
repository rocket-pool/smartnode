package megapool

import (
	"net/url"

	"github.com/gorilla/mux"
)

// ===============
// === Factory ===
// ===============

type megapoolStatusContextFactory struct {
	handler *MegapoolHandler
}

func (f *megapoolStatusContextFactory) Create(args url.Values) (*megapoolStatusContext, error) {
	c := &megapoolStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *megapoolStatusContextFactory) RegisterRoute(router *mux.Router) {
	// RegisterMegapoolRoute[*megapoolStatusContext, api.MegapoolStatusData](
	// 	router, "status", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	// )
}

// ===============
// === Context ===
// ===============

type megapoolStatusContext struct {
	handler *MegapoolHandler
}
