package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/beacon"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	rpbeacon "github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

type minipoolRescueDissolvedManager struct {
}

func (m *minipoolRescueDissolvedManager) CreateBindings(rp *rocketpool.RocketPool) error {
	return nil
}

func (m *minipoolRescueDissolvedManager) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (m *minipoolRescueDissolvedManager) CheckState(node *node.Node, response *api.MinipoolRescueDissolvedDetailsResponse) bool {
	return true
}

func (m *minipoolRescueDissolvedManager) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpCommon := mp.GetMinipoolCommon()
	mpCommon.GetFinalised(mc)
	mpCommon.GetStatus(mc)
	mpCommon.GetPubkey(mc)
}

func (m *minipoolRescueDissolvedManager) PrepareResponse(rp *rocketpool.RocketPool, bc rpbeacon.Client, addresses []common.Address, mps []minipool.Minipool, response *api.MinipoolRescueDissolvedDetailsResponse) error {
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
		mpDetails.InvalidBeaconState = beaconStatus.Status != rpbeacon.ValidatorState_PendingInitialized

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
}

func rescueDissolvedMinipools(c *cli.Context, minipoolAddresses []common.Address, amounts []*big.Int) (*api.BatchTxInfoResponse, error) {
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
		return nil, fmt.Errorf("error getting Beacon Node binding: %w", err)
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxInfoResponse{}

	// Sanity check
	if len(minipoolAddresses) != len(amounts) {
		return nil, fmt.Errorf("addresses and amounts must have the same length (%d vs. %d)", len(minipoolAddresses), len(amounts))
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, address := range minipoolAddresses {
		amount := amounts[i]
		opts.Value = amount
		txInfo, err := getDepositTx(rp, w, bc, address, amount, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating deposit transaction for minipool %s: %w", address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}

// Create a transaction for submitting a rescue deposit, optionally simulating it only for gas estimation
func getDepositTx(rp *rocketpool.RocketPool, w *wallet.Wallet, bc rpbeacon.Client, minipoolAddress common.Address, amount *big.Int, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	beaconDeposit, err := beacon.NewBeaconDeposit(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating Beacon deposit contract binding: %w", err)
	}

	// Create minipool
	mp, err := minipool.CreateMinipoolFromAddress(rp, minipoolAddress, false, nil)
	if err != nil {
		return nil, err
	}
	mpCommon := mp.GetMinipoolCommon()

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Get the contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		mpCommon.GetWithdrawalCredentials(mc)
		mpCommon.GetPubkey(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get minipool withdrawal credentials and keys
	withdrawalCredentials := mpCommon.Details.WithdrawalCredentials
	validatorPubkey := mpCommon.Details.Pubkey
	validatorKey, err := w.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return nil, fmt.Errorf("error getting validator private key for pubkey %s: %w", validatorPubkey.Hex(), err)
	}

	// Get validator deposit data
	amountGwei := big.NewInt(0).Div(amount, big.NewInt(1e9)).Uint64()
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, amountGwei)
	if err != nil {
		return nil, err
	}
	signature := types.BytesToValidatorSignature(depositData.Signature)

	// Get the tx info
	txInfo, err := beaconDeposit.Deposit(opts, validatorPubkey, withdrawalCredentials, signature, depositDataRoot)
	if err != nil {
		return nil, fmt.Errorf("error performing rescue deposit: %s", err.Error())
	}
	return txInfo, nil
}
