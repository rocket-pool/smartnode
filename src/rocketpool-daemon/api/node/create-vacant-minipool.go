package node

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rputils "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeCreateVacantMinipoolContextFactory struct {
	handler *NodeHandler
}

func (f *nodeCreateVacantMinipoolContextFactory) Create(args url.Values) (*nodeCreateVacantMinipoolContext, error) {
	c := &nodeCreateVacantMinipoolContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", args, input.ValidateBigInt, &c.amount),
		server.ValidateArg("min-node-fee", args, input.ValidateFraction, &c.minNodeFee),
		server.ValidateArg("salt", args, input.ValidateBigInt, &c.salt),
		server.ValidateArg("pubkey", args, input.ValidatePubkey, &c.pubkey),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeCreateVacantMinipoolContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeCreateVacantMinipoolContext, api.NodeCreateVacantMinipoolData](
		router, "create-vacant-minipool", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeCreateVacantMinipoolContext struct {
	handler *NodeHandler
	cfg     *config.SmartNodeConfig
	rp      *rocketpool.RocketPool
	bc      beacon.IBeaconClient

	amount     *big.Int
	minNodeFee float64
	salt       *big.Int
	pubkey     beacon.ValidatorPubkey
	node       *node.Node
	pSettings  *protocol.ProtocolDaoSettings
	oSettings  *oracle.OracleDaoSettings
	mpMgr      *minipool.MinipoolManager
}

func (c *nodeCreateVacantMinipoolContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.cfg = sp.GetConfig()
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	oMgr, err := oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting oDAO manager binding: %w", err)
	}
	c.oSettings = oMgr.Settings
	c.mpMgr, err = minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *nodeCreateVacantMinipoolContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.EthMatched,
		c.node.EthMatchedLimit,
		c.pSettings.Node.AreVacantMinipoolsEnabled,
		c.oSettings.Minipool.PromotionScrubPeriod,
	)
}

func (c *nodeCreateVacantMinipoolContext) PrepareData(data *api.NodeCreateVacantMinipoolData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Initial population
	data.DepositDisabled = !c.pSettings.Node.AreVacantMinipoolsEnabled.Get()
	data.ScrubPeriod = c.oSettings.Minipool.PromotionScrubPeriod.Formatted()

	// Adjust the salt
	if c.salt.Cmp(common.Big0) == 0 {
		nonce, err := c.rp.Client.NonceAt(context.Background(), c.node.Address, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting node's latest nonce: %w", err)
		}
		c.salt.SetUint64(nonce)
	}

	// Get the next minipool address
	err := c.rp.Query(func(mc *batch.MultiCaller) error {
		c.node.GetExpectedMinipoolAddress(mc, &data.MinipoolAddress, c.salt)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting expected minipool address: %w", err)
	}

	// Get the withdrawal credentials
	err = c.rp.Query(func(mc *batch.MultiCaller) error {
		c.mpMgr.GetMinipoolWithdrawalCredentials(mc, &data.WithdrawalCredentials, data.MinipoolAddress)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool withdrawal credentials: %w", err)
	}

	// Check data
	validatorEthWei := eth.EthToWei(ValidatorEth)
	matchRequest := big.NewInt(0).Sub(validatorEthWei, c.amount)
	availableToMatch := big.NewInt(0).Sub(c.node.EthMatchedLimit.Get(), c.node.EthMatched.Get())
	data.InsufficientRplStake = (availableToMatch.Cmp(matchRequest) == -1)

	// Update response
	data.CanDeposit = !(data.InsufficientRplStake || data.InvalidAmount || data.DepositDisabled)
	if !data.CanDeposit {
		return types.ResponseStatus_Success, nil
	}
	// Make sure the BN is on the correct chain
	depositContractInfo, err := rputils.GetDepositContractInfo(ctx, c.rp, c.cfg, c.bc)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error verifying the EL and BC are on the same chain: %w", err)
	}
	if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
		depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("FATAL: Beacon network mismatch! Expected %s on chain %d, but beacon is using %s on chain %d.",
			depositContractInfo.RPDepositContract.Hex(),
			depositContractInfo.RPNetwork,
			depositContractInfo.BeaconDepositContract.Hex(),
			depositContractInfo.BeaconNetwork)
	}

	// Check if the pubkey is for an existing active_ongoing validator
	validatorStatus, err := c.bc.GetValidatorStatus(ctx, c.pubkey, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error checking status of existing validator: %w", err)
	}
	if !validatorStatus.Exists {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("validator %s does not exist on the Beacon chain. If you recently created it, please wait until the Consensus layer has processed your deposits.", c.pubkey.Hex())
	}
	if validatorStatus.Status != beacon.ValidatorState_ActiveOngoing {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("validator %s must be in the active_ongoing state to be migrated, but it is currently in %s.", c.pubkey.Hex(), string(validatorStatus.Status))
	}
	if c.cfg.Network.Value != config.Network_Devnet && validatorStatus.WithdrawalCredentials[0] != 0x00 {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("validator %s already has withdrawal credentials [%s], which are not BLS credentials.", c.pubkey.Hex(), validatorStatus.WithdrawalCredentials.Hex())
	}

	// Convert the existing balance from gwei to wei
	balanceWei := big.NewInt(0).SetUint64(validatorStatus.Balance)
	balanceWei.Mul(balanceWei, big.NewInt(1e9))

	// Get tx info
	txInfo, err := c.node.CreateVacantMinipool(c.amount, c.minNodeFee, c.pubkey, c.salt, data.MinipoolAddress, balanceWei, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for CreateVacantMinipool: %w", err)
	}
	data.TxInfo = txInfo
	return types.ResponseStatus_Success, nil
}
