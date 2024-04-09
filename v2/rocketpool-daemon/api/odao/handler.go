package odao

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type OracleDaoHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewOracleDaoHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *OracleDaoHandler {
	h := &OracleDaoHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&oracleDaoStatusContextFactory{h},
		&oracleDaoCancelProposalContextFactory{h},
		&oracleDaoExecuteProposalsContextFactory{h},
		&oracleDaoSettingsContextFactory{h},
		&oracleDaoJoinContextFactory{h},
		&oracleDaoLeaveContextFactory{h},
		&oracleDaoMembersContextFactory{h},
		&oracleDaoProposalsContextFactory{h},
		&oracleDaoVoteContextFactory{h},
		&oracleDaoProposeInviteContextFactory{h},
		&oracleDaoProposeKickContextFactory{h},
		&oracleDaoProposeLeaveContextFactory{h},
		&oracleDaoProposeSettingContextFactory{h},
	}
	return h
}

func (h *OracleDaoHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/odao").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
