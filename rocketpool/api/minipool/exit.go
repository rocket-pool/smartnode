package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

func getMinipoolExitDetailsForNode(c *cli.Context) (*api.GetMinipoolExitDetailsForNodeResponse, error) {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.GetMinipoolExitDetailsForNodeResponse{}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetMinipoolCount(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
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
		mpCommon.GetNodeAddress(mc)
		mpCommon.GetStatus(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the exit details
	details := make([]api.MinipoolExitDetails, len(addresses))
	for i, mp := range mps {
		mpCommonDetails := mp.GetMinipoolCommon().Details
		status := mpCommonDetails.Status.Formatted()

		// Validate minipool owner
		if mpCommonDetails.NodeAddress != nodeAccount.Address {
			return nil, fmt.Errorf("minipool %s does not belong to the node", mpCommonDetails.Address.Hex())
		}

		mpDetails := api.MinipoolExitDetails{
			Address:       mpCommonDetails.Address,
			InvalidStatus: (status != types.Staking),
		}
		mpDetails.CanExit = !mpDetails.InvalidStatus
		details[i] = mpDetails
	}

	response.Details = details
	return &response, nil
}

func exitMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.ApiResponse, error) {

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

	// Response
	response := api.ApiResponse{}

	// Sync
	var wg errgroup.Group
	var head beacon.BeaconHead
	var signatureDomain []byte
	var mps []minipool.Minipool

	wg.Go(func() error {
		// Create minipools
		var err error
		mps, err = minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
		if err != nil {
			return fmt.Errorf("error creating minipool bindings: %w", err)
		}

		// Run the details getter
		err = rp.BatchQuery(len(minipoolAddresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
			mps[i].GetMinipoolCommon().GetPubkey(mc)
			return nil
		}, nil)
		if err != nil {
			return fmt.Errorf("error getting minipool details: %w", err)
		}
		return nil
	})

	// Get Beacon info
	wg.Go(func() error {
		// Get beacon head
		var err error
		head, err = bc.GetBeaconHead()
		if err != nil {
			return fmt.Errorf("error getting beacon head: %w", err)
		}

		// Get voluntary exit signature domain
		signatureDomain, err = bc.GetDomainData(eth2types.DomainVoluntaryExit[:], head.Epoch, false)
		if err != nil {
			return fmt.Errorf("error getting beacon domain data: %w", err)
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	for _, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		minipoolAddress := mpCommon.Details.Address
		validatorPubkey := mpCommon.Details.Pubkey

		// Get validator private key
		validatorKey, err := w.GetValidatorKeyByPubkey(validatorPubkey)
		if err != nil {
			return nil, fmt.Errorf("error getting private key for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Get validator index
		validatorIndex, err := bc.GetValidatorIndex(validatorPubkey)
		if err != nil {
			return nil, fmt.Errorf("error getting index of minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Get signed voluntary exit message
		signature, err := validator.GetSignedExitMessage(validatorKey, validatorIndex, head.Epoch, signatureDomain)
		if err != nil {
			return nil, fmt.Errorf("error getting exit message signature for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Broadcast voluntary exit message
		if err := bc.ExitValidator(validatorIndex, head.Epoch, signature); err != nil {
			return nil, fmt.Errorf("error submitting exit message for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}
	}

	return &response, nil
}
