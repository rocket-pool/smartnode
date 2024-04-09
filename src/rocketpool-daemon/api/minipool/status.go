package minipool

import (
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	rpbeacon "github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/core"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolStatusContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolStatusContextFactory) Create(args url.Values) (*minipoolStatusContext, error) {
	c := &minipoolStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolStatusContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolStatusContext, api.MinipoolStatusData](
		router, "status", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolStatusContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	bc      beacon.IBeaconClient

	delegate      *core.Contract
	pSettings     *protocol.ProtocolDaoSettings
	oSettings     *oracle.OracleDaoSettings
	reth          *tokens.TokenReth
	rpl           *tokens.TokenRpl
	fsrpl         *tokens.TokenRplFixedSupply
	rethBalances  []*big.Int
	rplBalances   []*big.Int
	fsrplBalances []*big.Int
}

func (c *minipoolStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()

	// Requirements
	err := sp.RequireBeaconClientSynced(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}

	// Bindings
	c.delegate, err = c.rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool delegate binding: %w", err)
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
	c.reth, err = tokens.NewTokenReth(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating rETH token binding: %w", err)
	}
	c.rpl, err = tokens.NewTokenRpl(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating RPL token binding: %w", err)
	}
	c.fsrpl, err = tokens.NewTokenRplFixedSupply(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating legacy RPL token binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *minipoolStatusContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pSettings.Minipool.LaunchTimeout,
		c.oSettings.Minipool.ScrubPeriod,
		c.oSettings.Minipool.PromotionScrubPeriod,
	)
}

func (c *minipoolStatusContext) CheckState(node *node.Node, response *api.MinipoolStatusData) bool {
	// Provision the token balance counts
	minipoolCount := node.MinipoolCount.Formatted()
	c.rethBalances = make([]*big.Int, minipoolCount)
	c.rplBalances = make([]*big.Int, minipoolCount)
	c.fsrplBalances = make([]*big.Int, minipoolCount)
	return true
}

func (c *minipoolStatusContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	address := mp.Common().Address
	eth.QueryAllFields(mp, mc)
	c.reth.BalanceOf(mc, &c.rethBalances[index], address)
	c.rpl.BalanceOf(mc, &c.rplBalances[index], address)
	c.fsrpl.BalanceOf(mc, &c.fsrplBalances[index], address)
}

func (c *minipoolStatusContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolStatusData) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Data
	var wg1 errgroup.Group
	var eth2Config beacon.Eth2Config
	var currentHeader *ethtypes.Header
	var balances []*big.Int

	// Get the current ETH balances of each minipool
	wg1.Go(func() error {
		var err error
		balances, err = c.rp.BalanceBatcher.GetEthBalances(addresses, nil)
		if err != nil {
			return fmt.Errorf("error getting minipool balances: %w", err)
		}
		return nil
	})

	// Get eth2 config
	wg1.Go(func() error {
		var err error
		eth2Config, err = c.bc.GetEth2Config(ctx)
		if err != nil {
			return fmt.Errorf("error getting Beacon config: %w", err)
		}
		return nil
	})

	// Get current block header
	wg1.Go(func() error {
		var err error
		currentHeader, err = c.rp.Client.HeaderByNumber(ctx, nil)
		if err != nil {
			return fmt.Errorf("error getting latest block header: %w", err)
		}
		return nil
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return types.ResponseStatus_Error, err
	}

	// Calculate the current epoch from the header and Beacon config
	genesis := time.Unix(int64(eth2Config.GenesisTime), 0)
	currentTime := time.Unix(int64(currentHeader.Time), 0)
	timeSinceGenesis := currentTime.Sub(genesis)
	currentEpoch := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerEpoch

	// Get some protocol settings
	launchTimeout := c.pSettings.Minipool.LaunchTimeout.Formatted()
	scrubPeriod := c.oSettings.Minipool.ScrubPeriod.Formatted()
	promotionScrubPeriod := c.oSettings.Minipool.PromotionScrubPeriod.Formatted()

	// Get the statuses on Beacon
	pubkeys := make([]rpbeacon.ValidatorPubkey, 0, len(addresses))
	for _, mp := range mps {
		mpCommon := mp.Common()
		status := mpCommon.Status.Formatted()
		if status == rptypes.MinipoolStatus_Staking || (status == rptypes.MinipoolStatus_Dissolved && !mpCommon.IsFinalised.Get()) {
			pubkeys = append(pubkeys, mpCommon.Pubkey.Get())
		}
	}
	beaconStatuses, err := c.bc.GetValidatorStatuses(ctx, pubkeys, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting validator statuses on Beacon: %w", err)
	}

	// Assign the details
	details := make([]api.MinipoolDetails, len(mps))
	for i, mp := range mps {
		mpCommonDetails := mp.Common()
		pubkey := mpCommonDetails.Pubkey.Get()
		mpv3, isv3 := minipool.GetMinipoolAsV3(mp)

		// Basic info
		mpDetails := api.MinipoolDetails{
			Address: mpCommonDetails.Address,
		}
		mpDetails.ValidatorPubkey = pubkey
		mpDetails.Version = mpCommonDetails.Version
		mpDetails.Status.Status = mpCommonDetails.Status.Formatted()
		mpDetails.Status.StatusBlock = mpCommonDetails.StatusBlock.Formatted()
		mpDetails.Status.StatusTime = mpCommonDetails.StatusTime.Formatted()
		mpDetails.DepositType = mpCommonDetails.DepositType.Formatted()
		mpDetails.Node.Address = mpCommonDetails.NodeAddress.Get()
		mpDetails.Node.DepositAssigned = mpCommonDetails.NodeDepositAssigned.Get()
		mpDetails.Node.DepositBalance = mpCommonDetails.NodeDepositBalance.Get()
		mpDetails.Node.Fee = mpCommonDetails.NodeFee.Formatted()
		mpDetails.Node.RefundBalance = mpCommonDetails.NodeRefundBalance.Get()
		mpDetails.User.DepositAssigned = mpCommonDetails.UserDepositAssigned.Get()
		mpDetails.User.DepositAssignedTime = mpCommonDetails.UserDepositAssignedTime.Formatted()
		mpDetails.User.DepositBalance = mpCommonDetails.UserDepositBalance.Get()
		mpDetails.Balances.Eth = balances[i]
		mpDetails.Balances.Reth = c.rethBalances[i]
		mpDetails.Balances.Rpl = c.rplBalances[i]
		mpDetails.Balances.FixedSupplyRpl = c.fsrplBalances[i]
		mpDetails.UseLatestDelegate = mpCommonDetails.IsUseLatestDelegateEnabled.Get()
		mpDetails.Delegate = mpCommonDetails.DelegateAddress.Get()
		mpDetails.PreviousDelegate = mpCommonDetails.PreviousDelegateAddress.Get()
		mpDetails.EffectiveDelegate = mpCommonDetails.EffectiveDelegateAddress.Get()
		mpDetails.Finalised = mpCommonDetails.IsFinalised.Get()
		mpDetails.Penalties = mpCommonDetails.PenaltyCount.Formatted()
		mpDetails.Queue.Position = mpCommonDetails.QueuePosition.Formatted() + 1 // Queue pos is -1 indexed so make it 0
		mpDetails.RefundAvailable = (mpDetails.Node.RefundBalance.Cmp(zero()) > 0) && (mpDetails.Balances.Eth.Cmp(mpDetails.Node.RefundBalance) >= 0)
		mpDetails.CloseAvailable = (mpDetails.Status.Status == rptypes.MinipoolStatus_Dissolved)
		mpDetails.WithdrawalAvailable = (mpDetails.Status.Status == rptypes.MinipoolStatus_Withdrawable)

		// Check the stake status of each minipool
		if mpDetails.Status.Status == rptypes.MinipoolStatus_Prelaunch {
			creationTime := mpDetails.Status.StatusTime
			dissolveTime := creationTime.Add(launchTimeout)
			remainingTime := creationTime.Add(scrubPeriod).Sub(currentTime)
			if remainingTime < 0 {
				mpDetails.CanStake = true
				mpDetails.TimeUntilDissolve = time.Until(dissolveTime)
			}
		}

		// Atlas info
		if isv3 {
			mpDetails.Status.IsVacant = mpv3.IsVacant.Get()
			mpDetails.ReduceBondTime = mpv3.ReduceBondTime.Formatted()

			// Check the promotion status of each minipool
			if mpDetails.Status.IsVacant {
				creationTime := mpDetails.Status.StatusTime
				dissolveTime := creationTime.Add(launchTimeout)
				remainingTime := creationTime.Add(promotionScrubPeriod).Sub(currentTime)
				if remainingTime < 0 {
					mpDetails.CanPromote = true
					mpDetails.TimeUntilDissolve = time.Until(dissolveTime)
				}
			}
		}

		// Beacon info
		beaconStatus, existsOnBeacon := beaconStatuses[pubkey]
		validatorActivated := false
		mpDetails.Validator.Exists = existsOnBeacon
		if existsOnBeacon {
			mpDetails.Validator.Active = (beaconStatus.ActivationEpoch < currentEpoch && beaconStatus.ExitEpoch > currentEpoch)
			mpDetails.Validator.Index = beaconStatus.Index
			validatorActivated = (beaconStatus.ActivationEpoch < currentEpoch)
		}
		if !validatorActivated {
			// Use deposit balances if the validator isn't activated yet
			mpDetails.Validator.Balance = big.NewInt(0).Add(mpDetails.Node.DepositBalance, mpDetails.User.DepositBalance)
			mpDetails.Validator.NodeBalance = big.NewInt(0).Set(mpDetails.Node.DepositBalance)
		} else {
			mpDetails.Validator.Balance = eth.GweiToWei(float64(beaconStatus.Balance))
		}

		details[i] = mpDetails
	}

	// Calculate the node share of each minipool balance
	err = c.rp.BatchQuery(len(addresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mp := mps[i]
		mpCommon := mp.Common()
		mpDetails := &details[i]

		// Get the node share of the ETH balance
		if mpDetails.Balances.Eth.Cmp(mpDetails.Node.RefundBalance) == -1 {
			mpDetails.NodeShareOfEthBalance = big.NewInt(0)
		} else {
			effectiveBalance := big.NewInt(0).Sub(mpDetails.Balances.Eth, mpDetails.Node.RefundBalance)
			mpCommon.CalculateNodeShare(mc, &mpDetails.NodeShareOfEthBalance, effectiveBalance)
		}

		// Get the node share of the Beacon balance
		pubkey := mpCommon.Pubkey.Get()
		beaconStatus, existsOnBeacon := beaconStatuses[pubkey]
		validatorActivated := (beaconStatus.ActivationEpoch < currentEpoch)
		if validatorActivated && existsOnBeacon {
			mpCommon.CalculateNodeShare(mc, &mpDetails.Validator.NodeBalance, mpDetails.Validator.Balance)
		}

		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting node share of minipools: %w", err)
	}

	data.Minipools = details
	data.LatestDelegate = c.delegate.Address
	return types.ResponseStatus_Success, nil
}
