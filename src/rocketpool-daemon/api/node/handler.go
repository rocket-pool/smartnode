package node

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type NodeHandler struct {
	serviceProvider *services.ServiceProvider
	context         context.Context
	factories       []server.IContextFactory
}

func NewNodeHandler(context context.Context, serviceProvider *services.ServiceProvider) *NodeHandler {
	h := &NodeHandler{
		serviceProvider: serviceProvider,
		context:         context,
	}
	h.factories = []server.IContextFactory{
		&nodeBalanceContextFactory{h},
		&nodeBurnContextFactory{h},
		&nodeCheckCollateralContextFactory{h},
		&nodeClaimAndStakeContextFactory{h},
		&nodeClearSnapshotDelegateContextFactory{h},
		&nodeConfirmPrimaryWithdrawalAddressContextFactory{h},
		&nodeConfirmRplWithdrawalAddressContextFactory{h},
		&nodeCreateVacantMinipoolContextFactory{h},
		&nodeDepositContextFactory{h},
		&nodeDistributeContextFactory{h},
		&nodeRewardsContextFactory{h},
		&nodeGetRewardsInfoContextFactory{h},
		&nodeGetSnapshotProposalsContextFactory{h},
		&nodeGetSnapshotVotingPowerContextFactory{h},
		&nodeInitializeFeeDistributorContextFactory{h},
		&nodeRegisterContextFactory{h},
		&nodeResolveEnsContextFactory{h},
		&nodeSendMessageContextFactory{h},
		&nodeSendContextFactory{h},
		&nodeSetPrimaryWithdrawalAddressContextFactory{h},
		&nodeSetRplLockingAllowedContextFactory{h},
		&nodeSetRplWithdrawalAddressContextFactory{h},
		&nodeSetSnapshotDelegateContextFactory{h},
		&nodeSetSmoothingPoolRegistrationStatusContextFactory{h},
		&nodeSetStakeRplForAllowedContextFactory{h},
		&nodeSetTimezoneContextFactory{h},
		&nodeStakeRplContextFactory{h},
		&nodeStatusContextFactory{h},
		&nodeSwapRplContextFactory{h},
		&nodeWithdrawEthContextFactory{h},
		&nodeWithdrawRplContextFactory{h},
	}
	return h
}

func (h *NodeHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/node").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
