package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type nodeClaimAndStakeHandler struct {
	indices     []*big.Int
	stakeAmount *big.Int
	distMainnet *rewards.MerkleDistributorMainnet
}

func (h *nodeClaimAndStakeHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	var err error
	h.distMainnet, err = rewards.NewMerkleDistributorMainnet(rp)
	if err != nil {
		return fmt.Errorf("error getting merkle distributor mainnet binding: %w", err)
	}
	return nil
}

func (h *nodeClaimAndStakeHandler) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (h *nodeClaimAndStakeHandler) PrepareResponse(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, node *node.Node, opts *bind.TransactOpts, response *api.TxResponse) error {
	// Read the tree files to get the details
	rplAmount := []*big.Int{}
	ethAmount := []*big.Int{}
	merkleProofs := [][]common.Hash{}

	// Populate the interval info for each one
	for _, index := range h.indices {
		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, node.Details.Address, index.Uint64(), nil)
		if err != nil {
			return fmt.Errorf("error getting interval info for interval %d: %w", index, err)
		}

		// Validate
		if !intervalInfo.TreeFileExists {
			return fmt.Errorf("rewards tree file '%s' doesn't exist", intervalInfo.TreeFilePath)
		}
		if !intervalInfo.MerkleRootValid {
			return fmt.Errorf("merkle root for rewards tree file '%s' doesn't match the canonical merkle root for interval %d", intervalInfo.TreeFilePath, index.Uint64())
		}

		// Get the rewards from it
		if intervalInfo.NodeExists {
			rplForInterval := big.NewInt(0)
			rplForInterval.Add(rplForInterval, &intervalInfo.CollateralRplAmount.Int)
			rplForInterval.Add(rplForInterval, &intervalInfo.ODaoRplAmount.Int)

			ethForInterval := big.NewInt(0)
			ethForInterval.Add(ethForInterval, &intervalInfo.SmoothingPoolEthAmount.Int)

			rplAmount = append(rplAmount, rplForInterval)
			ethAmount = append(ethAmount, ethForInterval)
			merkleProofs = append(merkleProofs, intervalInfo.MerkleProof)
		}
	}

	// Get tx info
	var txInfo *core.TransactionInfo
	var funcName string
	var err error
	if h.stakeAmount == nil {
		txInfo, err = h.distMainnet.Claim(node.Details.Address, h.indices, rplAmount, ethAmount, merkleProofs, opts)
		funcName = "Claim"
	} else {
		txInfo, err = h.distMainnet.ClaimAndStake(node.Details.Address, h.indices, rplAmount, ethAmount, merkleProofs, h.stakeAmount, opts)
		funcName = "ClaimAndStake"
	}
	if err != nil {
		return fmt.Errorf("error getting TX info for %s: %w", funcName, err)
	}
	response.TxInfo = txInfo
	return nil
}
