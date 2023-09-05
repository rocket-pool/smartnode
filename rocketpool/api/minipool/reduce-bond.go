package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func getMinipoolBeginReduceBondDetailsForNode(c *cli.Context, newBondAmountWei *big.Int) (*api.GetMinipoolBeginReduceBondDetailsForNodeResponse, error) {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.GetMinipoolBeginReduceBondDetailsForNodeResponse{}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}
	pSettings, err := settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating pDAO settings binding: %w", err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetMinipoolCount(mc)
		pSettings.GetBondReductionEnabled(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Return if bond reductions are disabled
	if !pSettings.Details.Minipool.IsBondReductionEnabled {
		response.BondReductionDisabled = true
		return &response, nil
	}

	// Get the minipool addresses for this node
	addresses, err := node.GetMinipoolAddresses(node.Details.MinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Create each minipool binding
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, addresses, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the relevant details
	err = rp.BatchQuery(len(addresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpCommon := mps[i].GetMinipoolCommon()
		mpCommon.GetNodeDepositBalance(mc)
		mpCommon.GetStatus(mc)
		mpCommon.GetPubkey(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the bond reduction details
	pubkeys := []types.ValidatorPubkey{}
	detailsMap := map[types.ValidatorPubkey]int{}
	details := make([]api.MinipoolBeginReduceBondDetails, len(addresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		mpDetails := api.MinipoolBeginReduceBondDetails{
			Address: mpCommon.Details.Address,
		}

		eligibleForBeacon := true
		if mpCommon.Details.Version < 3 {
			mpDetails.MinipoolVersionTooLow = true
			eligibleForBeacon = false
		}

		if mpCommon.Details.Status.Formatted() != types.Staking {
			mpDetails.InvalidElState = true
			eligibleForBeacon = false
		}

		details[i] = mpDetails

		if eligibleForBeacon {
			pubkeys = append(pubkeys, mpCommon.Details.Pubkey)
			detailsMap[mpCommon.Details.Pubkey] = i
		} else {
			mpDetails.CanReduce = false
		}
	}

	// Get the statuses on Beacon
	beaconStatuses, err := bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting validator statuses on Beacon: %w", err)
	}

	// Do a complete viability check
	for pubkey, beaconStatus := range beaconStatuses {
		i := detailsMap[pubkey]
		mpDetails := &details[i]
		mpDetails.Balance = beaconStatus.Balance
		mpDetails.BeaconState = beaconStatus.Status

		// How much more ETH they're requesting from the staking pool
		mpCommon := mps[i].GetMinipoolCommon()
		mpDetails.MatchRequest = big.NewInt(0).Sub(mpCommon.Details.NodeDepositBalance, newBondAmountWei)

		// Check the beacon state
		mpDetails.InvalidBeaconState = !(mpDetails.BeaconState == beacon.ValidatorState_PendingInitialized ||
			mpDetails.BeaconState == beacon.ValidatorState_PendingQueued ||
			mpDetails.BeaconState == beacon.ValidatorState_ActiveOngoing)

		// Make sure the balance is high enough
		threshold := uint64(32000000000)
		mpDetails.BalanceTooLow = mpDetails.Balance < threshold

		mpDetails.CanReduce = !(response.BondReductionDisabled || mpDetails.MinipoolVersionTooLow || mpDetails.BalanceTooLow || mpDetails.InvalidBeaconState)
	}

	response.Details = details
	return &response, nil
}

func beginReduceBondAmounts(c *cli.Context, minipoolAddresses []common.Address, newBondAmountWei *big.Int) (*api.BatchTxResponse, error) {
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
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxResponse{}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		minipoolAddress := mpCommon.Details.Address

		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if !success {
			return nil, fmt.Errorf("cannot begin bond reduction for minipool %s because its delegate version is too low (v%d); please update the delegate first", minipoolAddress.Hex(), mpCommon.Details.Version)
		}

		txInfo, err := mpv3.BeginReduceBondAmount(newBondAmountWei, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating begin-reduce-bond transaction for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}

func canReduceBondAmount(c *cli.Context, minipoolAddress common.Address) (*api.CanReduceBondAmountResponse, error) {
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

	// Response
	response := api.CanReduceBondAmountResponse{}

	// Make the minipool binding
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool binding for %s: %w", minipoolAddress.Hex(), err)
	}
	response.MinipoolVersion = mp.GetVersion()
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		// Get gas estimate
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, err
		}
		gasInfo, err := mpv3.EstimateReduceBondAmountGas(opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
	}

	response.CanReduce = success

	// Update & return response
	return &response, nil
}

func reduceBondAmount(c *cli.Context, minipoolAddress common.Address) (*api.ReduceBondAmountResponse, error) {
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

	// Response
	response := api.ReduceBondAmountResponse{}

	// Make the minipool binding
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool binding for %s: %w", minipoolAddress.Hex(), err)
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return nil, fmt.Errorf("bond reduction is not supported for minipool version %d; please upgrade the delegate for minipool %s to reduce its bond", mp.GetVersion(), minipoolAddress.Hex())
	}

	// Get the node transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Start bond reduction
	hash, err := mpv3.ReduceBondAmount(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}
