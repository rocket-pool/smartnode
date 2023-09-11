package minipool

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

func getMinipoolStakeDetailsForNode(c *cli.Context) (*api.GetMinipoolStakeDetailsForNodeResponse, error) {
	var oSettings *settings.OracleDaoSettings

	return runMinipoolQuery(c, MinipoolQuerier[api.GetMinipoolStakeDetailsForNodeResponse]{
		CreateBindings: func(rp *rocketpool.RocketPool) error {
			var err error
			oSettings, err = settings.NewOracleDaoSettings(rp)
			if err != nil {
				return fmt.Errorf("error creating oDAO settings binding: %w", err)
			}
			return nil
		},
		GetState: func(node *node.Node, mc *batch.MultiCaller) {
			oSettings.GetScrubPeriod(mc)
		},
		CheckState: nil,
		GetMinipoolDetails: func(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
			mpCommon := mp.GetMinipoolCommon()
			mpCommon.GetStatus(mc)
			mpCommon.GetStatusTime(mc)
		},
		PrepareResponse: func(rp *rocketpool.RocketPool, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolStakeDetailsForNodeResponse) error {
			scrubPeriod := oSettings.Details.Minipools.ScrubPeriod.Formatted()

			// Get the time of the latest block
			latestEth1Block, err := rp.Client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				return fmt.Errorf("error getting the latest block header: %w", err)
			}
			latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

			// Get the stake details
			details := make([]api.MinipoolStakeDetails, len(addresses))
			for i, mp := range mps {
				mpCommonDetails := mp.GetMinipoolCommon().Details
				mpDetails := api.MinipoolStakeDetails{
					Address: mpCommonDetails.Address,
				}

				mpDetails.State = mpCommonDetails.Status.Formatted()
				if mpDetails.State != types.Prelaunch {
					mpDetails.InvalidState = true
				} else {
					creationTime := mpCommonDetails.StatusTime.Formatted()
					mpDetails.RemainingTime = creationTime.Add(scrubPeriod).Sub(latestBlockTime)
					if mpDetails.RemainingTime > 0 {
						mpDetails.StillInScrubPeriod = true
					}
				}

				mpDetails.CanStake = !(mpDetails.InvalidState || mpDetails.StillInScrubPeriod)
				details[i] = mpDetails
			}

			// Update & return response
			response.Details = details
			return nil
		},
	})
}

func stakeMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxResponse{}

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Get the relevant details
	err = rp.BatchQuery(len(minipoolAddresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpCommon := mps[i].GetMinipoolCommon()
		mpCommon.GetWithdrawalCredentials(mc)
		mpCommon.GetPubkey(mc)
		mpCommon.GetDepositType(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()

		withdrawalCredentials := mpCommon.Details.WithdrawalCredentials
		validatorKey, err := w.GetValidatorKeyByPubkey(mpCommon.Details.Pubkey)
		if err != nil {
			return nil, err
		}
		depositType := mpCommon.Details.DepositType.Formatted()

		var depositAmount uint64
		switch depositType {
		case rptypes.Full, rptypes.Half, rptypes.Empty:
			depositAmount = uint64(16e9) // 16 ETH in gwei
		case rptypes.Variable:
			depositAmount = uint64(31e9) // 31 ETH in gwei
		default:
			return nil, fmt.Errorf("error staking minipool %s: unknown deposit type %d", mpCommon.Details.Address.Hex(), depositType)
		}

		// Get validator deposit data
		depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
		if err != nil {
			return nil, err
		}
		signature := rptypes.BytesToValidatorSignature(depositData.Signature)

		txInfo, err := mpCommon.Stake(signature, depositDataRoot, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating stake transaction for minipool %s: %w", mpCommon.Details.Address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}
