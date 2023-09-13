package node

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

type nodeCreateVacantHandler struct {
	bc         beacon.Client
	amountWei  *big.Int
	minNodeFee float64
	salt       *big.Int
	pubkey     rptypes.ValidatorPubkey
	pSettings  *settings.ProtocolDaoSettings
	oSettings  *settings.OracleDaoSettings
	mpMgr      *minipool.MinipoolManager
}

func (h *nodeCreateVacantHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	var err error
	h.pSettings, err = settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error getting pDAO settings binding: %w", err)
	}
	h.oSettings, err = settings.NewOracleDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error getting oDAO settings binding: %w", err)
	}
	h.mpMgr, err = minipool.NewMinipoolManager(rp)
	if err != nil {
		return fmt.Errorf("error getting minipool manager binding: %w", err)
	}
	return nil
}

func (h *nodeCreateVacantHandler) GetState(node *node.Node, mc *batch.MultiCaller) {
	node.GetEthMatched(mc)
	node.GetEthMatchedLimit(mc)
	h.pSettings.GetVacantMinipoolsEnabled(mc)
	h.oSettings.GetPromotionScrubPeriod(mc)
}

func (h *nodeCreateVacantHandler) PrepareResponse(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, node *node.Node, opts *bind.TransactOpts, response *api.CreateVacantMinipoolResponse) error {
	// Initial population
	response.DepositDisabled = !h.pSettings.Details.Node.AreVacantMinipoolsEnabled
	response.ScrubPeriod = h.oSettings.Details.Minipools.ScrubPeriod.Formatted()

	// Adjust the salt
	if h.salt.Cmp(big.NewInt(0)) == 0 {
		nonce, err := rp.Client.NonceAt(context.Background(), node.Details.Address, nil)
		if err != nil {
			return fmt.Errorf("error getting node's latest nonce: %w", err)
		}
		h.salt.SetUint64(nonce)
	}

	// Get the next minipool address
	err := rp.Query(func(mc *batch.MultiCaller) error {
		node.GetExpectedMinipoolAddress(mc, &response.MinipoolAddress, h.salt)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting expected minipool address: %w", err)
	}

	// Get the withdrawal credentials
	err = rp.Query(func(mc *batch.MultiCaller) error {
		h.mpMgr.GetMinipoolWithdrawalCredentials(mc, &response.WithdrawalCredentials, response.MinipoolAddress)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting minipool withdrawal credentials: %w", err)
	}

	// Check data
	validatorEthWei := eth.EthToWei(ValidatorEth)
	matchRequest := big.NewInt(0).Sub(validatorEthWei, h.amountWei)
	availableToMatch := big.NewInt(0).Sub(node.Details.EthMatchedLimit, node.Details.EthMatched)
	response.InsufficientRplStake = (availableToMatch.Cmp(matchRequest) == -1)

	// Update response
	response.CanDeposit = !(response.InsufficientRplStake || response.InvalidAmount || response.DepositDisabled)
	if response.CanDeposit {
		// Make sure ETH2 is on the correct chain
		depositContractInfo, err := rputils.GetDepositContractInfoImpl(rp, cfg, h.bc)
		if err != nil {
			return fmt.Errorf("error verifying the EL and BC are on the same chain: %w", err)
		}
		if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
			depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
			return fmt.Errorf("FATAL: Beacon network mismatch! Expected %s on chain %d, but beacon is using %s on chain %d.",
				depositContractInfo.RPDepositContract.Hex(),
				depositContractInfo.RPNetwork,
				depositContractInfo.BeaconDepositContract.Hex(),
				depositContractInfo.BeaconNetwork)
		}

		// Check if the pubkey is for an existing active_ongoing validator
		validatorStatus, err := h.bc.GetValidatorStatus(h.pubkey, nil)
		if err != nil {
			return fmt.Errorf("error checking status of existing validator: %w", err)
		}
		if !validatorStatus.Exists {
			return fmt.Errorf("validator %s does not exist on the Beacon chain. If you recently created it, please wait until the Consensus layer has processed your deposits.", h.pubkey.Hex())
		}
		if validatorStatus.Status != beacon.ValidatorState_ActiveOngoing {
			return fmt.Errorf("validator %s must be in the active_ongoing state to be migrated, but it is currently in %s.", h.pubkey.Hex(), string(validatorStatus.Status))
		}
		if cfg.Smartnode.Network.Value.(cfgtypes.Network) != cfgtypes.Network_Devnet && validatorStatus.WithdrawalCredentials[0] != 0x00 {
			return fmt.Errorf("validator %s already has withdrawal credentials [%s], which are not BLS credentials.", h.pubkey.Hex(), validatorStatus.WithdrawalCredentials.Hex())
		}

		// Convert the existing balance from gwei to wei
		balanceWei := big.NewInt(0).SetUint64(validatorStatus.Balance)
		balanceWei.Mul(balanceWei, big.NewInt(1e9))

		// Run the deposit gas estimator
		txInfo, err := node.CreateVacantMinipool(h.amountWei, h.minNodeFee, h.pubkey, h.salt, response.MinipoolAddress, balanceWei, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for CreateVacantMinipool: %w", err)
		}
		response.TxInfo = txInfo
	}
	return nil
}
