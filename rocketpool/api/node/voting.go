package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func estimateSetSnapshotDelegateGas(c *cli.Context, address common.Address) (*api.EstimateSetSnapshotDelegateGasResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.EstimateSetSnapshotDelegateGasResponse{}

	// Get the snapshot address
	addressString := cfg.Smartnode.GetSnapshotDelegationAddress()
	if addressString == "" {
		return nil, fmt.Errorf("Network [%v] does not have a snapshot delegation contract.", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}
	snapshotDelegationAddress := common.HexToAddress(addressString)

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Create the snapshot delegation contract binding
	snapshotDelegationAbi, err := abi.JSON(strings.NewReader(contracts.SnapshotDelegationABI))
	if err != nil {
		return nil, err
	}
	contract := &rocketpool.Contract{
		Contract: bind.NewBoundContract(snapshotDelegationAddress, snapshotDelegationAbi, ec, ec, ec),
		Address:  &snapshotDelegationAddress,
		ABI:      &snapshotDelegationAbi,
		Client:   ec,
	}

	// Create the ID hash
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	// Get the gas info
	gasInfo, err := contract.GetTransactionGasInfo(opts, "setDelegate", idHash, address)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil

}

func setSnapshotDelegate(c *cli.Context, address common.Address) (*api.SetSnapshotDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	s, err := services.GetSnapshotDelegation(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetSnapshotDelegateResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Create the ID hash
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	// Set the delegate
	tx, err := s.SetDelegate(opts, idHash, address)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}

func estimateClearSnapshotDelegateGas(c *cli.Context) (*api.EstimateClearSnapshotDelegateGasResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.EstimateClearSnapshotDelegateGasResponse{}

	// Get the snapshot address
	addressString := cfg.Smartnode.GetSnapshotDelegationAddress()
	if addressString == "" {
		return nil, fmt.Errorf("Network [%v] does not have a snapshot delegation contract.", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}
	snapshotDelegationAddress := common.HexToAddress(addressString)

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Create the snapshot delegation contract binding
	snapshotDelegationAbi, err := abi.JSON(strings.NewReader(contracts.SnapshotDelegationABI))
	if err != nil {
		return nil, err
	}
	contract := &rocketpool.Contract{
		Contract: bind.NewBoundContract(snapshotDelegationAddress, snapshotDelegationAbi, ec, ec, ec),
		Address:  &snapshotDelegationAddress,
		ABI:      &snapshotDelegationAbi,
		Client:   ec,
	}

	// Create the ID hash
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	// Get the gas info
	gasInfo, err := contract.GetTransactionGasInfo(opts, "clearDelegate", idHash)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil

}

func clearSnapshotDelegate(c *cli.Context) (*api.ClearSnapshotDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	s, err := services.GetSnapshotDelegation(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ClearSnapshotDelegateResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Create the ID hash
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	// Set the delegate
	tx, err := s.ClearDelegate(opts, idHash)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
