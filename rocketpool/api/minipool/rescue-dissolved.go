package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

func getMinipoolRescueDissolvedDetailsForNode(c *cli.Context) (*api.GetMinipoolRescueDissolvedDetailsForNodeResponse, error) {
	return runMinipoolQuery(c, MinipoolQuerier[api.GetMinipoolRescueDissolvedDetailsForNodeResponse]{
		CreateBindings: nil,
		GetState:       nil,
		CheckState:     nil,
		GetMinipoolDetails: func(mc *batch.MultiCaller, mp minipool.Minipool) {
			mpCommon := mp.GetMinipoolCommon()
			mpCommon.GetFinalised(mc)
			mpCommon.GetStatus(mc)
			mpCommon.GetPubkey(mc)
		},
		PrepareResponse: func(rp *rocketpool.RocketPool, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolRescueDissolvedDetailsForNodeResponse) error {
			// Get the Beacon Node client
			bc, err := services.GetBeaconClient(c)
			if err != nil {
				return fmt.Errorf("error getting Beacon Node binding: %w", err)
			}

			// Get the rescue details
			pubkeys := []types.ValidatorPubkey{}
			detailsMap := map[types.ValidatorPubkey]int{}
			details := make([]api.MinipoolRescueDissolvedDetails, len(addresses))
			for i, mp := range mps {
				mpCommon := mp.GetMinipoolCommon()
				mpDetails := api.MinipoolRescueDissolvedDetails{
					Address:       mpCommon.Details.Address,
					MinipoolState: mpCommon.Details.Status.Formatted(),
					IsFinalized:   mpCommon.Details.IsFinalised,
				}

				if mpDetails.MinipoolState != types.Dissolved || mpDetails.IsFinalized {
					mpDetails.InvalidElState = true
				} else {
					pubkeys = append(pubkeys, mpCommon.Details.Pubkey)
					detailsMap[mpCommon.Details.Pubkey] = i
				}

				details[i] = mpDetails
			}

			// Get the statuses on Beacon
			beaconStatuses, err := bc.GetValidatorStatuses(pubkeys, nil)
			if err != nil {
				return fmt.Errorf("error getting validator statuses on Beacon: %w", err)
			}

			// Do a complete viability check
			for pubkey, beaconStatus := range beaconStatuses {
				i := detailsMap[pubkey]
				mpDetails := &details[i]
				mpDetails.BeaconState = beaconStatus.Status
				mpDetails.InvalidBeaconState = beaconStatus.Status != beacon.ValidatorState_PendingInitialized

				if !mpDetails.InvalidBeaconState {
					beaconBalanceGwei := big.NewInt(0).SetUint64(beaconStatus.Balance)
					mpDetails.BeaconBalance = big.NewInt(0).Mul(beaconBalanceGwei, big.NewInt(1e9))

					// Make sure it doesn't already have 32 ETH in it
					requiredBalance := eth.EthToWei(32)
					if mpDetails.BeaconBalance.Cmp(requiredBalance) >= 0 {
						mpDetails.HasFullBalance = true
					}
				}

				mpDetails.CanRescue = !(mpDetails.IsFinalized || mpDetails.InvalidElState || mpDetails.InvalidBeaconState || mpDetails.HasFullBalance)
			}

			response.Details = details
			return nil
		},
	})
}

// Create a transaction for submitting a rescue deposit, optionally simulating it only for gas estimation
func getDepositTx(rp *rocketpool.RocketPool, w *wallet.Wallet, bc beacon.Client, minipoolAddress common.Address, amount *big.Int, opts *bind.TransactOpts) (*types.Transaction, error) {

	blankAddress := common.Address{}
	casperAddress, err := rp.GetAddress("casperDeposit", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon deposit contract address: %w", err)
	}
	if casperAddress == nil || *casperAddress == blankAddress {
		return nil, fmt.Errorf("Beacon deposit contract address was empty (0x0).")
	}

	depositContract, err := contracts.NewBeaconDeposit(*casperAddress, rp.Client)
	if err != nil {
		return nil, fmt.Errorf("error creating Beacon deposit contract binding: %w", err)
	}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Get minipool withdrawal credentials
	withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, mp.GetAddress(), nil)
	if err != nil {
		return nil, err
	}

	// Get the validator key for the minipool
	validatorPubkey, err := minipool.GetMinipoolPubkey(rp, mp.GetAddress(), nil)
	if err != nil {
		return nil, err
	}
	validatorKey, err := w.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return nil, err
	}

	// Get the deposit amount in gwei
	amountGwei := big.NewInt(0).Div(amount, big.NewInt(1e9)).Uint64()

	// Get validator deposit data
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, amountGwei)
	if err != nil {
		return nil, err
	}
	signature := types.BytesToValidatorSignature(depositData.Signature)

	// Get the tx
	tx, err := depositContract.Deposit(opts, validatorPubkey[:], withdrawalCredentials[:], signature[:], depositDataRoot)
	if err != nil {
		return nil, fmt.Errorf("error performing rescue deposit: %s", err.Error())
	}

	// Return
	return tx, nil

}

func rescueDissolvedMinipool(c *cli.Context, minipoolAddress common.Address, amount *big.Int) (*api.RescueDissolvedMinipoolResponse, error) {

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

	// Response
	response := api.RescueDissolvedMinipoolResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	opts.Value = amount

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit the rescue deposit
	tx, err := getDepositTx(rp, w, bc, minipoolAddress, amount, opts)
	if err != nil {
		return nil, fmt.Errorf("error submitting rescue deposit: %w", err)
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
