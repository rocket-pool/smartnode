package minipool

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStatus(c *cli.Context) (*api.MinipoolStatusResponse, error) {
	var delegate *core.Contract
	var pSettings *settings.ProtocolDaoSettings
	var oSettings *settings.OracleDaoSettings
	var reth *tokens.TokenReth
	var rpl *tokens.TokenRpl
	var fsrpl *tokens.TokenRplFixedSupply
	var rethBalances []*big.Int
	var rplBalances []*big.Int
	var fsrplBalances []*big.Int

	return runMinipoolQuery(c, MinipoolQuerier[api.MinipoolStatusResponse]{
		CreateBindings: func(rp *rocketpool.RocketPool) error {
			var err error
			delegate, err = rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
			if err != nil {
				return fmt.Errorf("error getting minipool delegate binding: %w", err)
			}
			pSettings, err = settings.NewProtocolDaoSettings(rp)
			if err != nil {
				return fmt.Errorf("error creating pDAO settings binding: %w", err)
			}
			oSettings, err = settings.NewOracleDaoSettings(rp)
			if err != nil {
				return fmt.Errorf("error creating oDAO settings binding: %w", err)
			}
			reth, err = tokens.NewTokenReth(rp)
			if err != nil {
				return fmt.Errorf("error creating rETH token binding: %w", err)
			}
			rpl, err = tokens.NewTokenRpl(rp)
			if err != nil {
				return fmt.Errorf("error creating RPL token binding: %w", err)
			}
			fsrpl, err = tokens.NewTokenRplFixedSupply(rp)
			if err != nil {
				return fmt.Errorf("error creating legacy RPL token binding: %w", err)
			}
			return nil
		},
		GetState: func(node *node.Node, mc *batch.MultiCaller) {
			pSettings.GetMinipoolLaunchTimeout(mc)
			oSettings.GetScrubPeriod(mc)
			oSettings.GetPromotionScrubPeriod(mc)
		},
		CheckState: func(node *node.Node, response *api.MinipoolStatusResponse) bool {
			// Provision the token balance counts
			minipoolCount := node.Details.MinipoolCount.Formatted()
			rethBalances = make([]*big.Int, minipoolCount)
			rplBalances = make([]*big.Int, minipoolCount)
			fsrplBalances = make([]*big.Int, minipoolCount)
			return true
		},
		GetMinipoolDetails: func(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
			address := mp.GetMinipoolCommon().Details.Address
			mp.QueryAllDetails(mc)
			reth.GetBalance(mc, &rethBalances[index], address)
			rpl.GetBalance(mc, &rplBalances[index], address)
			fsrpl.GetBalance(mc, &fsrplBalances[index], address)
		},
		PrepareResponse: func(rp *rocketpool.RocketPool, addresses []common.Address, mps []minipool.Minipool, response *api.MinipoolStatusResponse) error {
			// Get the Beacon Node client
			bc, err := services.GetBeaconClient(c)
			if err != nil {
				return fmt.Errorf("error getting Beacon Node binding: %w", err)
			}

			// Data
			var wg1 errgroup.Group
			var eth2Config beacon.Eth2Config
			var currentHeader *types.Header
			var balances []*big.Int

			// Get the current ETH balances of each minipool
			wg1.Go(func() error {
				var err error
				balances, err = rp.BalanceBatcher.GetEthBalances(addresses, nil)
				if err != nil {
					return fmt.Errorf("error getting minipool balances: %w", err)
				}
				return nil
			})

			// Get eth2 config
			wg1.Go(func() error {
				var err error
				eth2Config, err = bc.GetEth2Config()
				if err != nil {
					return fmt.Errorf("error getting Beacon config: %w", err)
				}
				return nil
			})

			// Get current block header
			wg1.Go(func() error {
				var err error
				currentHeader, err = rp.Client.HeaderByNumber(context.Background(), nil)
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

			// Get the statuses on Beacon
			pubkeys := make([]rptypes.ValidatorPubkey, len(addresses))
			for i, mp := range mps {
				pubkey := mp.GetMinipoolCommon().Details.Pubkey
				pubkeys[i] = pubkey
			}
			beaconStatuses, err := bc.GetValidatorStatuses(pubkeys, nil)
			if err != nil {
				return fmt.Errorf("error getting validator statuses on Beacon: %w", err)
			}

			// Assign the details
			details := make([]api.MinipoolDetails, len(mps))
			for i, mp := range mps {
				mpCommonDetails := mp.GetMinipoolCommon().Details
				pubkey := mpCommonDetails.Pubkey
				beaconStatus, existsOnBeacon := beaconStatuses[pubkey]
				mpv3, isv3 := minipool.GetMinipoolAsV3(mp)

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
				mpDetails.Balances.Reth = rethBalances[i]
				mpDetails.Balances.Rpl = rplBalances[i]
				mpDetails.Balances.FixedSupplyRpl = fsrplBalances[i]
				mpDetails.UseLatestDelegate = mpCommonDetails.IsUseLatestDelegateEnabled
				mpDetails.Delegate = mpCommonDetails.DelegateAddress
				mpDetails.PreviousDelegate = mpCommonDetails.PreviousDelegateAddress
				mpDetails.EffectiveDelegate = mpCommonDetails.EffectiveDelegateAddress
				mpDetails.Finalised = mpCommonDetails.IsFinalised
				mpDetails.Penalties = mpCommonDetails.PenaltyCount.Formatted()
				mpDetails.Queue.Position = mpCommonDetails.QueuePosition.Formatted() + 1 // Queue pos is -1 indexed so make it 0

				if isv3 {
					mpDetails.Status.IsVacant = mpv3.Details.IsVacant
					mpDetails.ReduceBondTime = mpv3.Details.ReduceBondTime.Formatted()
				}

				details[i] = mpDetails
			}

			// Calculate the node share of each minipool balance
			err = rp.BatchQuery(len(addresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
				mpCommon := mps[i].GetMinipoolCommon()
				mpDetails := &details[i]
				if mpDetails.Balances.Eth.Cmp(mpDetails.Node.RefundBalance) == -1 {
					mpDetails.NodeShareOfEthBalance = big.NewInt(0)
				} else {
					effectiveBalance := big.NewInt(0).Sub(mpDetails.Balances.Eth, mpDetails.Node.RefundBalance)
					mpCommon.CalculateNodeShare(mc, &mpDetails.NodeShareOfEthBalance, effectiveBalance)
				}
				return nil
			}, nil)

			response.LatestDelegate = *delegate.Address
			return nil
		},
	})

	///
	///
	///

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
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
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MinipoolStatusResponse{}

	// Get the legacy MinipoolQueue contract address
	legacyMinipoolQueueAddress := cfg.Smartnode.GetV110MinipoolQueueAddress()

	// Get minipool details
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	details, err := getNodeMinipoolDetails(rp, bc, nodeAccount.Address, &legacyMinipoolQueueAddress)
	if err != nil {
		return nil, err
	}
	response.Minipools = details

	delegate, err := rp.GetContract("rocketMinipoolDelegate", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting latest minipool delegate contract: %w", err)
	}

	response.LatestDelegate = *delegate.Address

	// Return response
	return &response, nil
}
