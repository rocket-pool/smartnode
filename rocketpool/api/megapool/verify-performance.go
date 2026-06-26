package megapool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/node"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/performance"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// verifyPerformance computes a megapool validator's RPIP-73 target-vote
// performance over the inclusive epoch range [startEpoch, endEpoch].
//
// If megapoolAddress is the zero address, the node's own megapool address is
// looked up via rocketNodeManager.getMegapoolAddress.
func verifyPerformance(
	c *cli.Command,
	megapoolAddress common.Address,
	validatorId uint32,
	startEpoch uint64,
	endEpoch uint64,
) (*api.VerifyPerformanceResponse, error) {
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	if (megapoolAddress == common.Address{}) {
		if err := services.RequireNodeRegistered(c); err != nil {
			return nil, fmt.Errorf("no megapool address supplied and node is not registered: %w", err)
		}
		w, err := services.GetWallet(c)
		if err != nil {
			return nil, err
		}
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return nil, err
		}
		megapoolAddress, err = node.GetMegapoolAddress(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("error looking up node's megapool address: %w", err)
		}
		if (megapoolAddress == common.Address{}) {
			return nil, fmt.Errorf("node has no megapool deployed; pass --megapool to specify one")
		}
	}

	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating megapool binding for %s: %w", megapoolAddress.Hex(), err)
	}
	pubkey, err := mp.GetValidatorPubkey(validatorId, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting megapool %s validator %d pubkey: %w", megapoolAddress.Hex(), validatorId, err)
	}

	return performance.VerifyPerformance(rp, bc, pubkey, startEpoch, endEpoch)
}
