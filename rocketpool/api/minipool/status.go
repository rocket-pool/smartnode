package minipool

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolStatusContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolStatusContextFactory) Create(vars map[string]string) (*minipoolStatusContext, error) {
	c := &minipoolStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterMinipoolRoute[*minipoolStatusContext, api.MinipoolStatusData](
		router, "status", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolStatusContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	bc      beacon.Client

	delegate      *core.Contract
	pSettings     *settings.ProtocolDaoSettings
	oSettings     *settings.OracleDaoSettings
	reth          *tokens.TokenReth
	rpl           *tokens.TokenRpl
	fsrpl         *tokens.TokenRplFixedSupply
	rethBalances  []*big.Int
	rplBalances   []*big.Int
	fsrplBalances []*big.Int
}

func (c *minipoolStatusContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireBeaconClientSynced(),
	)
	if err != nil {
		return err
	}

	// Bindings
	c.delegate, err = c.rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return fmt.Errorf("error getting minipool delegate binding: %w", err)
	}
	c.pSettings, err = settings.NewProtocolDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	c.oSettings, err = settings.NewOracleDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating oDAO settings binding: %w", err)
	}
	c.reth, err = tokens.NewTokenReth(c.rp)
	if err != nil {
		return fmt.Errorf("error creating rETH token binding: %w", err)
	}
	c.rpl, err = tokens.NewTokenRpl(c.rp)
	if err != nil {
		return fmt.Errorf("error creating RPL token binding: %w", err)
	}
	c.fsrpl, err = tokens.NewTokenRplFixedSupply(c.rp)
	if err != nil {
		return fmt.Errorf("error creating legacy RPL token binding: %w", err)
	}
	return nil
}

func (c *minipoolStatusContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	c.pSettings.GetMinipoolLaunchTimeout(mc)
	c.oSettings.GetScrubPeriod(mc)
	c.oSettings.GetPromotionScrubPeriod(mc)
}

func (c *minipoolStatusContext) CheckState(node *node.Node, response *api.MinipoolStatusData) bool {
	// Provision the token balance counts
	minipoolCount := node.Details.MinipoolCount.Formatted()
	c.rethBalances = make([]*big.Int, minipoolCount)
	c.rplBalances = make([]*big.Int, minipoolCount)
	c.fsrplBalances = make([]*big.Int, minipoolCount)
	return true
}

func (c *minipoolStatusContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	address := mp.GetMinipoolCommon().Details.Address
	mp.QueryAllDetails(mc)
	c.reth.GetBalance(mc, &c.rethBalances[index], address)
	c.rpl.GetBalance(mc, &c.rplBalances[index], address)
	c.fsrpl.GetBalance(mc, &c.fsrplBalances[index], address)
}

func (c *minipoolStatusContext) PrepareData(addresses []common.Address, mps []minipool.Minipool, data *api.MinipoolStatusData) error {
	// Data
	var wg1 errgroup.Group
	var eth2Config beacon.Eth2Config
	var currentHeader *types.Header
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
		eth2Config, err = c.bc.GetEth2Config()
		if err != nil {
			return fmt.Errorf("error getting Beacon config: %w", err)
		}
		return nil
	})

	// Get current block header
	wg1.Go(func() error {
		var err error
		currentHeader, err = c.rp.Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("error getting latest block header: %w", err)
		}
		return nil
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return err
	}

	// Calculate the current epoch from the header and Beacon config
	genesis := time.Unix(int64(eth2Config.GenesisTime), 0)
	currentTime := time.Unix(int64(currentHeader.Time), 0)
	timeSinceGenesis := currentTime.Sub(genesis)
	currentEpoch := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerEpoch

	// Get some protocol settings
	launchTimeout := c.pSettings.Details.Minipool.LaunchTimeout.Formatted()
	scrubPeriod := c.oSettings.Details.Minipools.ScrubPeriod.Formatted()
	promotionScrubPeriod := c.oSettings.Details.Minipools.PromotionScrubPeriod.Formatted()

	// Get the statuses on Beacon
	pubkeys := make([]rptypes.ValidatorPubkey, 0, len(addresses))
	for _, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		status := mpCommon.Details.Status.Formatted()
		if status == rptypes.Staking || (status == rptypes.Dissolved && !mpCommon.Details.IsFinalised) {
			pubkeys = append(pubkeys, mpCommon.Details.Pubkey)
		}
	}
	beaconStatuses, err := c.bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return fmt.Errorf("error getting validator statuses on Beacon: %w", err)
	}

	// Assign the details
	details := make([]api.MinipoolDetails, len(mps))
	for i, mp := range mps {
		mpCommonDetails := mp.GetMinipoolCommon().Details
		pubkey := mpCommonDetails.Pubkey
		mpv3, isv3 := minipool.GetMinipoolAsV3(mp)

		// Basic info
		mpDetails := api.MinipoolDetails{
			Address: mpCommonDetails.Address,
		}
		mpDetails.ValidatorPubkey = pubkey
		mpDetails.Status.Status = mpCommonDetails.Status.Formatted()
		mpDetails.Status.StatusBlock = mpCommonDetails.StatusBlock.Formatted()
		mpDetails.Status.StatusTime = mpCommonDetails.StatusTime.Formatted()
		mpDetails.DepositType = mpCommonDetails.DepositType.Formatted()
		mpDetails.Node.Address = mpCommonDetails.NodeAddress
		mpDetails.Node.DepositAssigned = mpCommonDetails.NodeDepositAssigned
		mpDetails.Node.DepositBalance = mpCommonDetails.NodeDepositBalance
		mpDetails.Node.Fee = mpCommonDetails.NodeFee.Formatted()
		mpDetails.Node.RefundBalance = mpCommonDetails.NodeRefundBalance
		mpDetails.User.DepositAssigned = mpCommonDetails.UserDepositAssigned
		mpDetails.User.DepositAssignedTime = mpCommonDetails.UserDepositAssignedTime.Formatted()
		mpDetails.User.DepositBalance = mpCommonDetails.UserDepositBalance
		mpDetails.Balances.Eth = balances[i]
		mpDetails.Balances.Reth = c.rethBalances[i]
		mpDetails.Balances.Rpl = c.rplBalances[i]
		mpDetails.Balances.FixedSupplyRpl = c.fsrplBalances[i]
		mpDetails.UseLatestDelegate = mpCommonDetails.IsUseLatestDelegateEnabled
		mpDetails.Delegate = mpCommonDetails.DelegateAddress
		mpDetails.PreviousDelegate = mpCommonDetails.PreviousDelegateAddress
		mpDetails.EffectiveDelegate = mpCommonDetails.EffectiveDelegateAddress
		mpDetails.Finalised = mpCommonDetails.IsFinalised
		mpDetails.Penalties = mpCommonDetails.PenaltyCount.Formatted()
		mpDetails.Queue.Position = mpCommonDetails.QueuePosition.Formatted() + 1 // Queue pos is -1 indexed so make it 0
		mpDetails.RefundAvailable = (mpDetails.Node.RefundBalance.Cmp(zero()) > 0) && (mpDetails.Balances.Eth.Cmp(mpDetails.Node.RefundBalance) >= 0)
		mpDetails.CloseAvailable = (mpDetails.Status.Status == rptypes.Dissolved)
		mpDetails.WithdrawalAvailable = (mpDetails.Status.Status == rptypes.Withdrawable)

		// Check the stake status of each minipool
		if mpDetails.Status.Status == rptypes.Prelaunch {
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
			mpDetails.Status.IsVacant = mpv3.Details.IsVacant
			mpDetails.ReduceBondTime = mpv3.Details.ReduceBondTime.Formatted()

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
		mpCommon := mps[i].GetMinipoolCommon()
		mpDetails := &details[i]

		// Get the node share of the ETH balance
		if mpDetails.Balances.Eth.Cmp(mpDetails.Node.RefundBalance) == -1 {
			mpDetails.NodeShareOfEthBalance = big.NewInt(0)
		} else {
			effectiveBalance := big.NewInt(0).Sub(mpDetails.Balances.Eth, mpDetails.Node.RefundBalance)
			mpCommon.CalculateNodeShare(mc, &mpDetails.NodeShareOfEthBalance, effectiveBalance)
		}

		// Get the node share of the Beacon balance
		pubkey := mpCommon.Details.Pubkey
		beaconStatus, existsOnBeacon := beaconStatuses[pubkey]
		validatorActivated := (beaconStatus.ActivationEpoch < currentEpoch)
		if validatorActivated && existsOnBeacon {
			mpCommon.CalculateNodeShare(mc, &mpDetails.Validator.NodeBalance, mpDetails.Validator.Balance)
		}

		return nil
	}, nil)

	data.LatestDelegate = *c.delegate.Address
	return nil
}
