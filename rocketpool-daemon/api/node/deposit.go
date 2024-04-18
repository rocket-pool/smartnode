package node

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/deposit"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	nodewallet "github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/collateral"
	rputils "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/utils"
	snValidator "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeDepositContextFactory struct {
	handler *NodeHandler
}

func (f *nodeDepositContextFactory) Create(args url.Values) (*nodeDepositContext, error) {
	c := &nodeDepositContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", args, input.ValidateBigInt, &c.amount),
		server.ValidateArg("min-node-fee", args, input.ValidateFraction, &c.minNodeFee),
		server.ValidateArg("salt", args, input.ValidateBigInt, &c.salt),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeDepositContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeDepositContext, api.NodeDepositData](
		router, "deposit", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeDepositContext struct {
	handler *NodeHandler
	cfg     *config.SmartNodeConfig
	rp      *rocketpool.RocketPool
	bc      beacon.IBeaconClient
	w       *nodewallet.Wallet
	vMgr    *snValidator.ValidatorManager

	amount      *big.Int
	minNodeFee  float64
	salt        *big.Int
	node        *node.Node
	depositPool *deposit.DepositPoolManager
	pSettings   *protocol.ProtocolDaoSettings
	oSettings   *oracle.OracleDaoSettings
	mpMgr       *minipool.MinipoolManager
}

func (c *nodeDepositContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.cfg = sp.GetConfig()
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	c.w = sp.GetWallet()
	c.vMgr = sp.GetValidatorManager()
	nodeAddress, _ := c.w.GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}
	err = sp.RequireWalletReady()
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	c.depositPool, err = deposit.NewDepositPoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting deposit pool binding: %w", err)
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

func (c *nodeDepositContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.UsableCreditAndDonatedBalance,
		c.depositPool.Balance,
		c.pSettings.Node.IsDepositingEnabled,
		c.oSettings.Minipool.ScrubPeriod,
	)
}

func (c *nodeDepositContext) PrepareData(data *api.NodeDepositData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	rs := c.cfg.GetNetworkResources()

	// Initial population
	data.CreditBalance = c.node.UsableCreditAndDonatedBalance.Get()
	data.DepositDisabled = !c.pSettings.Node.IsDepositingEnabled.Get()
	data.DepositBalance = c.depositPool.Balance.Get()
	data.ScrubPeriod = c.oSettings.Minipool.ScrubPeriod.Formatted()

	// Adjust the salt
	if c.salt.Cmp(big.NewInt(0)) == 0 {
		nonce, err := c.rp.Client.NonceAt(ctx, c.node.Address, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting node's latest nonce: %w", err)
		}
		c.salt.SetUint64(nonce)
	}

	// Check node balance
	var err error
	data.NodeBalance, err = c.rp.Client.BalanceAt(ctx, c.node.Address, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting node's ETH balance: %w", err)
	}

	// Check the node's collateral
	collateral, err := collateral.CheckCollateral(c.rp, c.node.Address, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error checking node collateral: %w", err)
	}
	ethMatched := collateral.EthMatched
	ethMatchedLimit := collateral.EthMatchedLimit
	pendingMatchAmount := collateral.PendingMatchAmount

	// Check for insufficient balance
	totalBalance := big.NewInt(0).Add(data.NodeBalance, data.CreditBalance)
	data.InsufficientBalance = (c.amount.Cmp(totalBalance) > 0)

	// Check if the credit balance can be used
	data.CanUseCredit = (data.DepositBalance.Cmp(eth.EthToWei(1)) >= 0)

	// Check data
	validatorEthWei := eth.EthToWei(ValidatorEth)
	matchRequest := big.NewInt(0).Sub(validatorEthWei, c.amount)
	availableToMatch := big.NewInt(0).Sub(ethMatchedLimit, ethMatched)
	availableToMatch.Sub(availableToMatch, pendingMatchAmount)
	data.InsufficientRplStake = (availableToMatch.Cmp(matchRequest) == -1)

	// Update response
	data.CanDeposit = !(data.InsufficientBalance || data.InsufficientRplStake || data.InvalidAmount || data.DepositDisabled)
	if data.CanDeposit && !data.CanUseCredit && data.NodeBalance.Cmp(c.amount) < 0 {
		// Can't use credit and there's not enough ETH in the node wallet to deposit so error out
		data.InsufficientBalanceWithoutCredit = true
		data.CanDeposit = false
	}

	// Return if depositing won't work
	if !data.CanDeposit {
		return types.ResponseStatus_Success, nil
	}

	// Make sure ETH2 is on the correct chain
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

	// Get how much credit to use
	if data.CanUseCredit {
		remainingAmount := big.NewInt(0).Sub(c.amount, data.CreditBalance)
		if remainingAmount.Cmp(big.NewInt(0)) > 0 {
			// Send the remaining amount if the credit isn't enough to cover the whole deposit
			opts.Value = remainingAmount
		}
	} else {
		opts.Value = c.amount
	}

	// Get the next available validator key without saving it
	validatorKey, index, err := c.vMgr.GetNextValidatorKey(false)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting next available validator key: %w", err)
	}
	data.Index = index

	// Get the next minipool address
	var minipoolAddress common.Address
	err = c.rp.Query(func(mc *batch.MultiCaller) error {
		c.node.GetExpectedMinipoolAddress(mc, &minipoolAddress, c.salt)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting expected minipool address: %w", err)
	}
	data.MinipoolAddress = minipoolAddress

	// Get the withdrawal credentials
	var withdrawalCredentials common.Hash
	err = c.rp.Query(func(mc *batch.MultiCaller) error {
		c.mpMgr.GetMinipoolWithdrawalCredentials(mc, &withdrawalCredentials, minipoolAddress)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool withdrawal credentials: %w", err)
	}

	// Get validator deposit data and associated parameters
	// NOTE: validation is done in the NMC now
	depositAmount := uint64(1e9) // 1 ETH in gwei
	depositData, err := validator.GetDepositData(validatorKey, withdrawalCredentials, rs.GenesisForkVersion, depositAmount, rs.EthNetworkName)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting deposit data: %w", err)
	}
	pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
	signature := beacon.ValidatorSignature(depositData.Signature)
	data.ValidatorPubkey = pubkey

	// Make sure a validator with this pubkey doesn't already exist
	status, err := c.bc.GetValidatorStatus(ctx, pubkey, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("Error checking for existing validator status: %w\nYour funds have not been deposited for your own safety.", err)
	}
	if status.Exists {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("**** ALERT ****\n"+
			"Your minipool %s has the following as a validator pubkey:\n\t%s\n"+
			"This key is already in use by validator %s on the Beacon chain!\n"+
			"Rocket Pool will not allow you to deposit this validator for your own safety so you do not get slashed.\n"+
			"PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS.\n"+
			"***************\n", minipoolAddress.Hex(), pubkey.HexWithPrefix(), status.Index)
	}

	// Get tx info
	var txInfo *eth.TransactionInfo
	var funcName string
	depositDataRoot := common.BytesToHash(depositData.DepositDataRoot)
	if data.CanUseCredit {
		txInfo, err = c.node.DepositWithCredit(c.amount, c.minNodeFee, pubkey, signature, depositDataRoot, c.salt, minipoolAddress, opts)
		funcName = "DepositWithCredit"
	} else {
		txInfo, err = c.node.Deposit(c.amount, c.minNodeFee, pubkey, signature, depositDataRoot, c.salt, minipoolAddress, opts)
		funcName = "Deposit"
	}
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for %s: %w", funcName, err)
	}
	data.TxInfo = txInfo

	return types.ResponseStatus_Success, nil
}
