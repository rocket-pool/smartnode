package pdao

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/governance"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// GetOffchainVotingStatus keeps the legacy response shape used by status and
// metrics consumers, while sourcing proposals and effective votes from RocketDash.
func GetOffchainVotingStatus(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, node common.Address, state string) (api.SnapshotResponseStruct, error) {
	response := api.SnapshotResponseStruct{}
	client, err := governance.NewClient(cfg.Smartnode.GetRocketDashURL(), cfg.Smartnode.GetChainID())
	if err != nil {
		return response, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	response.ActiveSnapshotProposals, err = client.Proposals(ctx, state)
	if err != nil {
		return response, err
	}
	response.ProposalVotes, err = client.Votes(ctx, response.ActiveSnapshotProposals, node, func(block uint32) (common.Address, error) {
		return network.GetVotingDelegate(rp, node, block, &bind.CallOpts{Context: ctx})
	})
	return response, err
}
