package minipool

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/validator"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type minipoolChangeCredsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolChangeCredsContextFactory) Create(args url.Values) (*minipoolChangeCredsContext, error) {
	c := &minipoolChangeCredsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.minipoolAddress),
		server.ValidateArg("mnemonic", args, input.ValidateWalletMnemonic, &c.mnemonic),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolChangeCredsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*minipoolChangeCredsContext, api.SuccessData](
		router, "change-withdrawal-creds", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolChangeCredsContext struct {
	handler     *MinipoolHandler
	rp          *rocketpool.RocketPool
	bc          beacon.Client
	nodeAddress common.Address

	mnemonic        string
	minipoolAddress common.Address
	mp              minipool.IMinipool
}

func (c *minipoolChangeCredsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireBeaconClientSynced(),
	)
	if err != nil {
		return err
	}

	// Bindings
	mpMgr, err := minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	c.mp, err = mpMgr.CreateMinipoolFromAddress(c.minipoolAddress, false, nil)
	if err != nil {
		return fmt.Errorf("error creating minipool binding from address: %w", err)
	}
	return nil
}

func (c *minipoolChangeCredsContext) GetState(mc *batch.MultiCaller) {
	c.mp.Common().Pubkey.AddToQuery(mc)
}

func (c *minipoolChangeCredsContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	// Get minipool validator pubkey
	pubkey := c.mp.Common().Pubkey.Get()

	// Get the index for this validator based on the mnemonic
	index := uint(0)
	validatorKeyPath := validator.ValidatorKeyPath
	var validatorKey *eth2types.BLSPrivateKey
	for index < validatorKeyRetrievalLimit {
		key, err := validator.GetPrivateKey(c.mnemonic, index, validatorKeyPath)
		if err != nil {
			return fmt.Errorf("error deriving key for index %d: %w", index, err)
		}
		candidatePubkey := key.PublicKey().Marshal()
		if bytes.Equal(pubkey[:], candidatePubkey) {
			validatorKey = key
			break
		}
		index++
	}
	if validatorKey == nil {
		return fmt.Errorf("couldn't find the validator key for this mnemonic after %d tries", validatorKeyRetrievalLimit)
	}

	// Get the withdrawal creds from this index
	withdrawalKey, err := validator.GetWithdrawalKey(c.mnemonic, index, validatorKeyPath)
	if err != nil {
		return fmt.Errorf("error getting withdrawal key for validator: %w", err)
	}

	// Get beacon head
	head, err := c.bc.GetBeaconHead()
	if err != nil {
		return fmt.Errorf("error getting Beacon head: %w", err)
	}

	// Get voluntary exit signature domain
	signatureDomain, err := c.bc.GetDomainData(eth2types.DomainBlsToExecutionChange[:], head.Epoch, true)
	if err != nil {
		return fmt.Errorf("error getting signature domain: %w", err)
	}

	// Get validator index
	validatorIndex, err := c.bc.GetValidatorIndex(pubkey)
	if err != nil {
		return fmt.Errorf("error getting validator index: %w", err)
	}

	// Get signed withdrawal creds change message
	signature, err := validator.GetSignedWithdrawalCredsChangeMessage(withdrawalKey, validatorIndex, c.minipoolAddress, signatureDomain)
	if err != nil {
		return fmt.Errorf("error getting signed withdrawal credentials change message: %w", err)
	}

	// Broadcast withdrawal creds change message
	withdrawalPubkey := types.BytesToValidatorPubkey(withdrawalKey.PublicKey().Marshal())
	if err := c.bc.ChangeWithdrawalCredentials(validatorIndex, withdrawalPubkey, c.minipoolAddress, signature); err != nil {
		return fmt.Errorf("error submitting withdrawal credentials change message: %w", err)
	}
	data.Success = true
	return nil
}
