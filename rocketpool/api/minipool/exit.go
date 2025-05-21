package minipool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

func canExitMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanExitMinipoolResponse, error) {

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
	response := api.CanExitMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
		return nil, err
	}

	// Check minipool status
	status, err := mp.GetStatus(nil)
	if err != nil {
		return nil, err
	}
	response.InvalidStatus = (status != types.Staking)

	// Update & return response
	response.CanExit = !response.InvalidStatus
	return &response, nil

}

func exitMinipool(c *cli.Context, minipoolAddress common.Address) (*api.ExitMinipoolResponse, error) {

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
	response := api.ExitMinipoolResponse{}

	// Get minipool validator pubkey
	validatorPubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get validator private key
	validatorKey, err := w.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return nil, err
	}

	// Get beacon head
	head, err := bc.GetBeaconHead()
	if err != nil {
		return nil, err
	}

	// Get voluntary exit signature domain
	signatureDomain, err := bc.GetDomainData(eth2types.DomainVoluntaryExit[:], head.Epoch, false)
	if err != nil {
		return nil, err
	}

	// Get validator index
	validatorIndex, err := bc.GetValidatorIndex(validatorPubkey)
	if err != nil {
		return nil, err
	}

	// Get signed voluntary exit message
	signature, err := validator.GetSignedExitMessage(validatorKey, validatorIndex, head.Epoch, signatureDomain)
	if err != nil {
		return nil, err
	}

	// Broadcast voluntary exit message
	if err := bc.ExitValidator(validatorIndex, head.Epoch, signature); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
