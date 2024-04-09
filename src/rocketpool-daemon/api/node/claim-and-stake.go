package node

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/eth"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/rewards"
	rprewards "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/rewards"
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
	server.RegisterQuerylessGet[*nodeClaimAndStakeContext, types.TxInfoData](
		router, "claim-and-stake", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *nodeClaimAndStakeContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	cfg := sp.GetConfig()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	distMainnet, err := rewards.NewMerkleDistributorMainnet(rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting merkle distributor mainnet binding: %w", err)
	}

	// Read the tree files to get the details
	rplAmount := []*big.Int{}
	ethAmount := []*big.Int{}
	merkleProofs := [][]common.Hash{}

	// Populate the interval info for each one
	for _, index := range c.indices {
		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAddress, index.Uint64(), nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting interval info for interval %d: %w", index, err)
		}

		// Validate
		if !intervalInfo.TreeFileExists {
			return types.ResponseStatus_ResourceNotFound, fmt.Errorf("rewards tree file '%s' doesn't exist", intervalInfo.TreeFilePath)
		}
		if !intervalInfo.MerkleRootValid {
			return types.ResponseStatus_ResourceConflict, fmt.Errorf("merkle root for rewards tree file '%s' doesn't match the canonical merkle root for interval %d", intervalInfo.TreeFilePath, index.Uint64())
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
	var txInfo *eth.TransactionInfo
	var funcName string
	if c.stakeAmount.Cmp(common.Big0) == 0 {
		txInfo, err = distMainnet.Claim(nodeAddress, c.indices, rplAmount, ethAmount, merkleProofs, opts)
		funcName = "Claim"
	} else {
		txInfo, err = distMainnet.ClaimAndStake(nodeAddress, c.indices, rplAmount, ethAmount, merkleProofs, c.stakeAmount, opts)
		funcName = "ClaimAndStake"
	}
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for %s: %w", funcName, err)
	}
	data.TxInfo = txInfo
	return types.ResponseStatus_Success, nil
}
