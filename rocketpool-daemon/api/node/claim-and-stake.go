package node

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/rewards"
	rprewards "github.com/rocket-pool/smartnode/rocketpool-daemon/common/rewards"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

const (
	claimAndStakeBatchLimit int = 100
)

// ===============
// === Factory ===
// ===============

type nodeClaimAndStakeContextFactory struct {
	handler *NodeHandler
}

func (f *nodeClaimAndStakeContextFactory) Create(args url.Values) (*nodeClaimAndStakeContext, error) {
	c := &nodeClaimAndStakeContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("indices", args, claimAndStakeBatchLimit, input.ValidateBigInt, &c.indices),
		server.ValidateArg("stake-amount", args, input.ValidateBigInt, &c.stakeAmount),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeClaimAndStakeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeClaimAndStakeContext, api.TxInfoData](
		router, "claim-and-stake", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeClaimAndStakeContext struct {
	handler *NodeHandler

	indices     []*big.Int
	stakeAmount *big.Int
}

func (c *nodeClaimAndStakeContext) PrepareData(data *api.TxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	cfg := sp.GetConfig()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	distMainnet, err := rewards.NewMerkleDistributorMainnet(rp)
	if err != nil {
		return fmt.Errorf("error getting merkle distributor mainnet binding: %w", err)
	}

	// Read the tree files to get the details
	rplAmount := []*big.Int{}
	ethAmount := []*big.Int{}
	merkleProofs := [][]common.Hash{}

	// Populate the interval info for each one
	for _, index := range c.indices {
		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAddress, index.Uint64(), nil)
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
	if c.stakeAmount.Cmp(common.Big0) == 0 {
		txInfo, err = distMainnet.Claim(nodeAddress, c.indices, rplAmount, ethAmount, merkleProofs, opts)
		funcName = "Claim"
	} else {
		txInfo, err = distMainnet.ClaimAndStake(nodeAddress, c.indices, rplAmount, ethAmount, merkleProofs, c.stakeAmount, opts)
		funcName = "ClaimAndStake"
	}
	if err != nil {
		return fmt.Errorf("error getting TX info for %s: %w", funcName, err)
	}
	data.TxInfo = txInfo
	return nil
}
