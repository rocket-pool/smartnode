package node

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/network"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	ens "github.com/wealdtech/go-ens/v3"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/alerting"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/alerting/alertmanager/models"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/collateral"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/voting"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	minipoolInfoBatchSize int = 200
)

// ===============
// === Factory ===
// ===============

type nodeStatusContextFactory struct {
	handler *NodeHandler
}

func (f *nodeStatusContextFactory) Create(args url.Values) (*nodeStatusContext, error) {
	c := &nodeStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeStatusContext, api.NodeStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeStatusContext struct {
	handler  *NodeHandler
	cfg      *config.SmartNodeConfig
	rp       *rocketpool.RocketPool
	ec       eth.IExecutionClient
	bc       beacon.IBeaconClient
	snapshot *contracts.SnapshotDelegation

	node         *node.Node
	networkMgr   *network.NetworkManager
	mpMgr        *minipool.MinipoolManager
	odaoMember   *oracle.OracleDaoMember
	pSettings    *protocol.ProtocolDaoSettings
	oSettings    *oracle.OracleDaoSettings
	rpl          *tokens.TokenRpl
	rplBalance   *big.Int
	fsrpl        *tokens.TokenRplFixedSupply
	fsrplBalance *big.Int
	reth         *tokens.TokenReth
	rethBalance  *big.Int
	delegate     common.Address
}

func (c *nodeStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.cfg = sp.GetConfig()
	c.rp = sp.GetRocketPool()
	c.ec = sp.GetEthClient()
	c.bc = sp.GetBeaconClient()
	c.snapshot = sp.GetSnapshotDelegation()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	c.networkMgr, err = network.NewNetworkManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting network manager binding: %w", err)
	}
	c.mpMgr, err = minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool manager binding: %w", err)
	}
	c.odaoMember, err = oracle.NewOracleDaoMember(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating oracle DAO member %s binding: %w", nodeAddress.Hex(), err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	oMgr, err := oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating oDAO manager binding: %w", err)
	}
	c.oSettings = oMgr.Settings
	c.rpl, err = tokens.NewTokenRpl(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating RPL token binding: %w", err)
	}
	c.fsrpl, err = tokens.NewTokenRplFixedSupply(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating legacy RPL token binding: %w", err)
	}
	c.reth, err = tokens.NewTokenReth(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating rETH token binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *nodeStatusContext) GetState(mc *batch.MultiCaller) {
	// Node properties
	eth.AddQueryablesToMulticall(mc,
		// Node
		c.node.Exists,
		c.node.PrimaryWithdrawalAddress,
		c.node.PendingPrimaryWithdrawalAddress,
		c.node.IsRplWithdrawalAddressSet,
		c.node.RplWithdrawalAddress,
		c.node.PendingRplWithdrawalAddress,
		c.node.TimezoneLocation,
		c.node.RplStake,
		c.node.EffectiveRplStake,
		c.node.MinimumRplStake,
		c.node.MaximumRplStake,
		c.node.Credit,
		c.node.IsFeeDistributorInitialized,
		c.node.MinipoolCount,
		c.node.DistributorAddress,
		c.node.SmoothingPoolRegistrationState,
		c.node.SmoothingPoolRegistrationChanged,
		c.node.TotalCreditAndDonatedBalance,
		c.node.UsableCreditAndDonatedBalance,
		c.node.DonatedEthBalance,
		c.node.IsRplLockingAllowed,
		c.node.RplLocked,
		c.node.IsVotingInitialized,
		c.node.CurrentVotingDelegate,

		// Other
		c.odaoMember.Exists,
		c.networkMgr.RplPrice,
		c.pSettings.Node.MinimumPerMinipoolStake,
		c.pSettings.Node.MaximumPerMinipoolStake,
		c.oSettings.Minipool.BondReductionWindowStart,
		c.oSettings.Minipool.BondReductionWindowLength,
	)

	// Token balances
	c.rpl.BalanceOf(mc, &c.rplBalance, c.node.Address)
	c.fsrpl.BalanceOf(mc, &c.fsrplBalance, c.node.Address)
	c.reth.BalanceOf(mc, &c.rethBalance, c.node.Address)

	// Snapshot
	if c.snapshot != nil {
		c.snapshot.Delegation(mc, &c.delegate, c.node.Address, c.cfg.GetVotingSnapshotID())
	}
}

func (c *nodeStatusContext) PrepareData(data *api.NodeStatusData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Beacon info
	beaconConfig, err := c.bc.GetEth2Config(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Beacon config: %w", err)
	}
	beaconHead, err := c.bc.GetBeaconHead(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Beacon head: %w", err)
	}
	genesisTime := time.Unix(int64(beaconConfig.GenesisTime), 0)

	// Basic properties
	data.AccountAddress = c.node.Address
	data.AccountAddressFormatted = c.getFormattedAddress(data.AccountAddress)
	data.Trusted = c.odaoMember.Exists.Get()
	data.Registered = c.node.Exists.Get()
	data.PrimaryWithdrawalAddress = c.node.PrimaryWithdrawalAddress.Get()
	data.PrimaryWithdrawalAddressFormatted = c.getFormattedAddress(data.PrimaryWithdrawalAddress)
	data.PendingPrimaryWithdrawalAddress = c.node.PendingPrimaryWithdrawalAddress.Get()
	data.PendingPrimaryWithdrawalAddressFormatted = c.getFormattedAddress(data.PendingPrimaryWithdrawalAddress)
	data.IsRplWithdrawalAddressSet = c.node.IsRplWithdrawalAddressSet.Get()
	data.RplWithdrawalAddress = c.node.RplWithdrawalAddress.Get()
	data.RplWithdrawalAddressFormatted = c.getFormattedAddress(data.RplWithdrawalAddress)
	data.PendingRplWithdrawalAddress = c.node.PendingRplWithdrawalAddress.Get()
	data.PendingRplWithdrawalAddressFormatted = c.getFormattedAddress(data.PendingRplWithdrawalAddress)
	data.TimezoneLocation = c.node.TimezoneLocation.Get()
	data.RplStake = c.node.RplStake.Get()
	data.EffectiveRplStake = c.node.EffectiveRplStake.Get()
	data.MinimumRplStake = c.node.MinimumRplStake.Get()
	data.MaximumRplStake = c.node.MaximumRplStake.Get()
	data.MaximumStakeFraction = c.pSettings.Node.MaximumPerMinipoolStake.Raw()
	data.CreditBalance = c.node.Credit.Get()
	data.CreditAndEthOnBehalfBalance = c.node.TotalCreditAndDonatedBalance.Get()
	data.UsableCreditAndEthOnBehalfBalance = c.node.UsableCreditAndDonatedBalance.Get()
	data.EthOnBehalfBalance = c.node.DonatedEthBalance.Get()
	data.IsFeeDistributorInitialized = c.node.IsFeeDistributorInitialized.Get()
	data.IsRplLockingAllowed = c.node.IsRplLockingAllowed.Get()
	data.RplLocked = c.node.RplLocked.Get()
	data.IsVotingInitialized = c.node.IsVotingInitialized.Get()
	data.OnchainVotingDelegate = c.node.CurrentVotingDelegate.Get()
	data.OnchainVotingDelegateFormatted = c.getFormattedAddress(data.OnchainVotingDelegate)

	// Minipool info
	mps, err := c.getMinipoolInfo(data)
	if err != nil {
		return types.ResponseStatus_Error, err
	}

	// Withdrawal address and balance info
	err = c.getBalanceInfo(data)
	if err != nil {
		return types.ResponseStatus_Error, err
	}

	// Collateral
	collateral, err := collateral.CheckCollateralWithMinipoolCache(c.rp, c.node.Address, mps, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting node collateral balance: %w", err)
	}
	data.EthMatched = collateral.EthMatched
	data.EthMatchedLimit = collateral.EthMatchedLimit
	data.PendingMatchAmount = collateral.PendingMatchAmount

	// Snapshot
	if c.snapshot != nil {
		emptyAddress := common.Address{}
		data.SnapshotVotingDelegate = c.delegate
		if data.SnapshotVotingDelegate != emptyAddress {
			data.SnapshotVotingDelegateFormatted = c.getFormattedAddress(data.SnapshotVotingDelegate)
		}
		props, err := voting.GetSnapshotProposals(c.cfg, c.node.Address, c.delegate, true)
		if err != nil {
			data.SnapshotResponse.Error = fmt.Sprintf("error getting snapshot proposals: %s", err.Error())
		} else {
			data.SnapshotResponse.ActiveSnapshotProposals = props
		}
	}

	// Fee recipient and smoothing pool
	sp, err := c.rp.GetContract(rocketpool.ContractName_RocketSmoothingPool)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting smoothing pool contract: %w", err)
	}
	data.FeeRecipientInfo.SmoothingPoolAddress = sp.Address
	data.FeeRecipientInfo.FeeDistributorAddress = c.node.DistributorAddress.Get()
	data.FeeRecipientInfo.IsInSmoothingPool = c.node.SmoothingPoolRegistrationState.Get()
	if !data.FeeRecipientInfo.IsInSmoothingPool {
		// Check if the user just opted out
		optOutTime := c.node.SmoothingPoolRegistrationChanged.Formatted()
		if optOutTime != time.Unix(0, 0) {
			// Get the epoch for that time
			secondsSinceGenesis := optOutTime.Sub(genesisTime)
			epoch := uint64(secondsSinceGenesis.Seconds()) / beaconConfig.SecondsPerEpoch

			// Make sure epoch + 1 is finalized - if not, they're still on cooldown
			targetEpoch := epoch + 1
			if beaconHead.FinalizedEpoch < targetEpoch {
				data.FeeRecipientInfo.IsInOptOutCooldown = true
				data.FeeRecipientInfo.OptOutEpoch = targetEpoch
			}
		}
	}

	// True effective stakes and collateral ratios
	err = c.calculateTrueStakesAndBonds(data, mps, beaconHead.Epoch)
	if err != nil {
		return types.ResponseStatus_Error, err
	}

	// Get alerts from Alertmanager
	alerts, err := alerting.FetchAlerts(c.cfg)
	if err != nil {
		// no reason to make `rocketpool node status` fail if we can't get alerts
		// (this is more likely to happen in native mode than docker where
		// alertmanager is more complex to set up)
		// Do save a warning though to print to the user
		data.Warning = fmt.Sprintf("Error fetching alerts from Alertmanager: %s", err)
		alerts = make([]*models.GettableAlert, 0)
	}
	data.Alerts = make([]api.NodeAlert, len(alerts))

	for i, a := range alerts {
		data.Alerts[i] = api.NodeAlert{
			State:       *a.Status.State,
			Labels:      a.Labels,
			Annotations: a.Annotations,
		}
	}

	return types.ResponseStatus_Success, nil
}

// Get a formatting string containing the ENS name for an address (if it exists)
func (c *nodeStatusContext) getFormattedAddress(address common.Address) string {
	name, err := ens.ReverseResolve(c.ec, address)
	if err != nil {
		return address.Hex()
	}
	return fmt.Sprintf("%s (%s)", name, address.Hex())
}

// Get info pertaining to the node's minipools
func (c *nodeStatusContext) getMinipoolInfo(data *api.NodeStatusData) ([]minipool.IMinipool, error) {
	// Minipool info
	addresses, err := c.node.GetMinipoolAddresses(c.node.MinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}
	mps, err := c.mpMgr.CreateMinipoolsFromAddresses(addresses, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool bindings: %w", err)
	}
	err = c.rp.BatchQuery(len(addresses), minipoolInfoBatchSize, func(mc *batch.MultiCaller, i int) error {
		// Basic details
		mp := mps[i]
		mpCommon := mp.Common()
		eth.AddQueryablesToMulticall(mc,
			mpCommon.PenaltyCount,
			mpCommon.IsFinalised,
			mpCommon.Status,
			mpCommon.NodeRefundBalance,
			mpCommon.NodeDepositBalance,
			mpCommon.UserDepositBalance,
			mpCommon.Pubkey,
		)

		// Details needed for collateral checking
		mpv3, isMpv3 := minipool.GetMinipoolAsV3(mp)
		if isMpv3 {
			eth.AddQueryablesToMulticall(mc,
				mpv3.ReduceBondTime,
				mpv3.IsBondReduceCancelled,
				mpv3.ReduceBondValue,
			)
		}
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Minipools
	data.PenalizedMinipools = map[common.Address]uint64{}
	data.MinipoolCounts.Total = len(mps)
	for _, mp := range mps {
		mpCommon := mp.Common()
		penaltyCount := mpCommon.PenaltyCount.Formatted()
		if penaltyCount > 0 {
			data.PenalizedMinipools[mpCommon.Address] = penaltyCount
		}
		if mpCommon.IsFinalised.Get() {
			data.MinipoolCounts.Finalised++
		} else {
			switch mpCommon.Status.Formatted() {
			case rptypes.MinipoolStatus_Initialized:
				data.MinipoolCounts.Initialized++
			case rptypes.MinipoolStatus_Prelaunch:
				data.MinipoolCounts.Prelaunch++
			case rptypes.MinipoolStatus_Staking:
				data.MinipoolCounts.Staking++
			case rptypes.MinipoolStatus_Withdrawable:
				data.MinipoolCounts.Withdrawable++
			case rptypes.MinipoolStatus_Dissolved:
				data.MinipoolCounts.Dissolved++
			}
			if mpCommon.NodeRefundBalance.Get().Cmp(common.Big0) > 0 {
				data.MinipoolCounts.RefundAvailable++
			}
		}
	}

	return mps, nil
}

// Get token and ETH balance information for the node and its primary / RPL withdrawal addresses (if set)
func (c *nodeStatusContext) getBalanceInfo(data *api.NodeStatusData) error {
	// Withdrawal address balances
	ethAddresses := []common.Address{
		c.node.Address,
		c.node.DistributorAddress.Get(),
	}
	var primaryWithdrawRplBalance *big.Int
	var primaryWithdrawFsRplBalance *big.Int
	var primaryWithdrawRethBalance *big.Int
	var rplWithdrawRplBalance *big.Int
	var rplWithdrawFsRplBalance *big.Int
	var rplWithdrawRethBalance *big.Int
	primaryWithdrawalDifferent := (data.PrimaryWithdrawalAddress != data.AccountAddress)
	rplWithdrawalDifferent := (data.RplWithdrawalAddress != data.AccountAddress && data.IsRplWithdrawalAddressSet)
	err := c.rp.Query(func(mc *batch.MultiCaller) error {
		if primaryWithdrawalDifferent {
			c.rpl.BalanceOf(mc, &primaryWithdrawRplBalance, data.PrimaryWithdrawalAddress)
			c.fsrpl.BalanceOf(mc, &primaryWithdrawFsRplBalance, data.PrimaryWithdrawalAddress)
			c.reth.BalanceOf(mc, &primaryWithdrawRethBalance, data.PrimaryWithdrawalAddress)
			ethAddresses = append(ethAddresses, data.PrimaryWithdrawalAddress)
		}
		if rplWithdrawalDifferent {
			c.rpl.BalanceOf(mc, &rplWithdrawRplBalance, data.RplWithdrawalAddress)
			c.fsrpl.BalanceOf(mc, &rplWithdrawFsRplBalance, data.RplWithdrawalAddress)
			c.reth.BalanceOf(mc, &rplWithdrawRethBalance, data.RplWithdrawalAddress)
			ethAddresses = append(ethAddresses, data.RplWithdrawalAddress)
		}
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting withdrawal address balances: %w", err)
	}

	// Balances
	ethBalances, err := c.rp.BalanceBatcher.GetEthBalances(ethAddresses, nil)
	if err != nil {
		return fmt.Errorf("error getting ETH balances: %w", err)
	}
	data.NodeBalances.Rpl = c.rplBalance
	data.NodeBalances.Fsrpl = c.fsrplBalance
	data.NodeBalances.Reth = c.rethBalance
	data.NodeBalances.Eth = ethBalances[0]
	data.FeeDistributorBalance = ethBalances[1]
	index := 2
	if primaryWithdrawalDifferent {
		data.PrimaryWithdrawalBalances.Rpl = primaryWithdrawRplBalance
		data.PrimaryWithdrawalBalances.Fsrpl = primaryWithdrawFsRplBalance
		data.PrimaryWithdrawalBalances.Reth = primaryWithdrawRethBalance
		data.PrimaryWithdrawalBalances.Eth = ethBalances[index]
		index++
	}
	if rplWithdrawalDifferent {
		data.RplWithdrawalBalances.Rpl = rplWithdrawRplBalance
		data.RplWithdrawalBalances.Fsrpl = rplWithdrawFsRplBalance
		data.RplWithdrawalBalances.Reth = rplWithdrawRethBalance
		data.RplWithdrawalBalances.Eth = ethBalances[index]
	}

	return nil
}

// Calculate the node's borrowed and bonded RPL amounts, along with the true min and max stakes
func (c *nodeStatusContext) calculateTrueStakesAndBonds(data *api.NodeStatusData, minipools []minipool.IMinipool, epoch uint64) error {
	activeMinipools := data.MinipoolCounts.Total - data.MinipoolCounts.Finalised
	if activeMinipools > 0 {
		minStakeFraction := c.pSettings.Node.MinimumPerMinipoolStake.Raw()
		maxStakeFraction := c.pSettings.Node.MaximumPerMinipoolStake.Raw()
		rplPrice := c.networkMgr.RplPrice.Raw()

		// Calculate the *real* minimum, including the pending bond reductions
		trueMinimumStake := big.NewInt(0).Add(data.EthMatched, data.PendingMatchAmount)
		trueMinimumStake.Mul(trueMinimumStake, minStakeFraction)
		trueMinimumStake.Div(trueMinimumStake, rplPrice)

		// Calculate the *real* maximum, including the pending bond reductions
		trueMaximumStake := eth.EthToWei(32)
		trueMaximumStake.Mul(trueMaximumStake, big.NewInt(int64(activeMinipools)))
		trueMaximumStake.Sub(trueMaximumStake, data.EthMatched)
		trueMaximumStake.Sub(trueMaximumStake, data.PendingMatchAmount) // (32 * activeMinipools - ethMatched - pendingMatch)
		trueMaximumStake.Mul(trueMaximumStake, maxStakeFraction)
		trueMaximumStake.Div(trueMaximumStake, rplPrice)

		data.MinimumRplStake = trueMinimumStake
		data.MaximumRplStake = trueMaximumStake

		if data.EffectiveRplStake.Cmp(trueMinimumStake) < 0 {
			data.EffectiveRplStake.SetUint64(0)
		} else if data.EffectiveRplStake.Cmp(trueMaximumStake) > 0 {
			data.EffectiveRplStake.Set(trueMaximumStake)
		}

		data.BondedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(data.RplStake) / (float64(activeMinipools)*32.0 - eth.WeiToEth(data.EthMatched) - eth.WeiToEth(data.PendingMatchAmount))
		data.BorrowedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(data.RplStake) / (eth.WeiToEth(data.EthMatched) + eth.WeiToEth(data.PendingMatchAmount))

		// Calculate the "eligible" info (ignoring pending bond reductions) based on the Beacon Chain
		_, _, pendingEligibleBorrowedEth, pendingEligibleBondedEth, err := c.getTrueBorrowAndBondAmounts(minipools, epoch)
		if err != nil {
			return fmt.Errorf("error calculating eligible borrowed and bonded amounts: %w", err)
		}

		// Calculate the "eligible real" minimum based on the Beacon Chain, including pending bond reductions
		pendingTrueMinimumStake := big.NewInt(0).Mul(pendingEligibleBorrowedEth, minStakeFraction)
		pendingTrueMinimumStake.Div(pendingTrueMinimumStake, rplPrice)

		// Calculate the "eligible real" maximum based on the Beacon Chain, including the pending bond reductions
		pendingTrueMaximumStake := big.NewInt(0).Mul(pendingEligibleBondedEth, maxStakeFraction)
		pendingTrueMaximumStake.Div(pendingTrueMaximumStake, rplPrice)

		data.PendingMinimumRplStake = pendingTrueMinimumStake
		data.PendingMaximumRplStake = pendingTrueMaximumStake

		data.PendingEffectiveRplStake = big.NewInt(0).Set(data.RplStake)
		if data.PendingEffectiveRplStake.Cmp(pendingTrueMinimumStake) < 0 {
			data.PendingEffectiveRplStake.SetUint64(0)
		} else if data.PendingEffectiveRplStake.Cmp(pendingTrueMaximumStake) > 0 {
			data.PendingEffectiveRplStake.Set(pendingTrueMaximumStake)
		}

		pendingEligibleBondedEthFloat := eth.WeiToEth(pendingEligibleBondedEth)
		if pendingEligibleBondedEthFloat == 0 {
			data.PendingBondedCollateralRatio = 0
		} else {
			data.PendingBondedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(data.RplStake) / pendingEligibleBondedEthFloat
		}

		pendingEligibleBorrowedEthFloat := eth.WeiToEth(pendingEligibleBorrowedEth)
		if pendingEligibleBorrowedEthFloat == 0 {
			data.PendingBorrowedCollateralRatio = 0
		} else {
			data.PendingBorrowedCollateralRatio = eth.WeiToEth(rplPrice) * eth.WeiToEth(data.RplStake) / pendingEligibleBorrowedEthFloat
		}
	} else {
		data.BorrowedCollateralRatio = -1
		data.BondedCollateralRatio = -1
		data.PendingEffectiveRplStake = big.NewInt(0)
		data.PendingMinimumRplStake = big.NewInt(0)
		data.PendingMaximumRplStake = big.NewInt(0)
		data.PendingBondedCollateralRatio = -1
		data.PendingBorrowedCollateralRatio = -1
	}
	return nil
}

// Calculate the true borrowed and bonded ETH amounts for a node based on the Beacon status of the minipools
func (c *nodeStatusContext) getTrueBorrowAndBondAmounts(mps []minipool.IMinipool, epoch uint64) (*big.Int, *big.Int, *big.Int, *big.Int, error) {

	pubkeys := make([]beacon.ValidatorPubkey, len(mps))
	nodeDeposits := make([]*big.Int, len(mps))
	userDeposits := make([]*big.Int, len(mps))
	pendingNodeDeposits := make([]*big.Int, len(mps))
	pendingUserDeposits := make([]*big.Int, len(mps))

	latestBlockHeader, err := c.ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error getting latest block header: %w", err)
	}
	blockTime := time.Unix(int64(latestBlockHeader.Time), 0)
	reductionWindowStart := c.oSettings.Minipool.BondReductionWindowStart.Formatted()
	reductionWindowLength := c.oSettings.Minipool.BondReductionWindowLength.Formatted()
	reductionWindowEnd := time.Duration(reductionWindowStart+reductionWindowLength) * time.Second

	zeroTime := time.Unix(0, 0)
	for i, mp := range mps {
		mpCommon := mp.Common()
		pubkeys[i] = mpCommon.Pubkey.Get()
		nodeDeposits[i] = mpCommon.NodeDepositBalance.Get()
		userDeposits[i] = mpCommon.UserDepositBalance.Get()

		mpv3, isv3 := minipool.GetMinipoolAsV3(mp)
		if isv3 {
			reduceBondTime := mpv3.ReduceBondTime.Formatted()
			reduceBondCancelled := mpv3.IsBondReduceCancelled.Get()

			// Ignore minipools that don't have a bond reduction pending
			timeSinceReductionStart := blockTime.Sub(reduceBondTime)
			if reduceBondTime == zeroTime ||
				reduceBondCancelled ||
				timeSinceReductionStart > reductionWindowEnd {
				pendingNodeDeposits[i] = nodeDeposits[i]
				pendingUserDeposits[i] = userDeposits[i]
			} else {
				newBond := mpv3.ReduceBondValue.Get()
				pendingNodeDeposits[i] = newBond

				// New user deposit = old + delta
				pendingUserDeposits[i] = big.NewInt(0).Sub(nodeDeposits[i], newBond)
				pendingUserDeposits[i].Add(pendingUserDeposits[i], userDeposits[i])
			}
		} else {
			pendingNodeDeposits[i] = nodeDeposits[i]
			pendingUserDeposits[i] = userDeposits[i]
		}
	}

	ctx := c.handler.ctx
	statuses, err := c.bc.GetValidatorStatuses(ctx, pubkeys, nil)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error loading validator statuses: %w", err)
	}

	eligibleBorrowedEth := big.NewInt(0)
	eligibleBondedEth := big.NewInt(0)
	pendingEligibleBorrowedEth := big.NewInt(0)
	pendingEligibleBondedEth := big.NewInt(0)
	for i, pubkey := range pubkeys {
		status, exists := statuses[pubkey]
		if !exists {
			// Validator doesn't exist on Beacon yet
			continue
		}
		if status.ActivationEpoch > epoch {
			// Validator hasn't activated yet
			continue
		}
		if status.ExitEpoch <= epoch {
			// Validator exited
			continue
		}
		// It's eligible, so add up the borrowed and bonded amounts
		eligibleBorrowedEth.Add(eligibleBorrowedEth, userDeposits[i])
		eligibleBondedEth.Add(eligibleBondedEth, nodeDeposits[i])
		pendingEligibleBorrowedEth.Add(pendingEligibleBorrowedEth, pendingUserDeposits[i])
		pendingEligibleBondedEth.Add(pendingEligibleBondedEth, pendingNodeDeposits[i])
	}

	return eligibleBorrowedEth, eligibleBondedEth, pendingEligibleBorrowedEth, pendingEligibleBondedEth, nil
}
