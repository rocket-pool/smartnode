package pdao

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"

	"github.com/urfave/cli"
)

func canSetSnapshotAddress(c *cli.Context, snapshotAddress common.Address, signature string) (*api.PDAOCanSetSnapshotAddressResponse, error) {

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
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOCanSetSnapshotAddressResponse{}

	response.VotingInitialized, err = network.GetVotingInitialized(rp, nodeAccount.Address, nil)
	if !response.VotingInitialized {
		return nil, fmt.Errorf("Voting must be initialized to set a snapshot address. Use 'rocketpool pdao initialize-voting' to initialize voting first.")
	}

	// Get signer registry contract address
	addressString := cfg.Smartnode.GetRocketSignerRegistryAddress()
	if addressString == "" {
		return nil, fmt.Errorf("Network [%v] does not have a signer registry contract.", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}
	rocketSignerRegistryAddress := common.HexToAddress(addressString)

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Create the signer registry contract binding
	rocketSignerRegistryAbi, err := abi.JSON(strings.NewReader(contracts.RocketSignerRegistryABI))
	if err != nil {
		return nil, err
	}
	contract := &rocketpool.Contract{
		Contract: bind.NewBoundContract(rocketSignerRegistryAddress, rocketSignerRegistryAbi, ec, ec, ec),
		Address:  &rocketSignerRegistryAddress,
		ABI:      &rocketSignerRegistryAbi,
		Client:   ec,
	}

	// Parse signature into vrs components, v to uint8 and v,s to [32]byte
	sig, err := apiutils.ParseEIP712(signature)
	if err != nil {
		fmt.Println("Error parsing signature", err)
	}

	// Get the gas info
	gasInfo, err := contract.GetTransactionGasInfo(opts, "setSigningDelegate", snapshotAddress, sig.V, sig.R, sig.S)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil
}

func setSnapshotAddress(c *cli.Context, snapshotAddress common.Address, signature string) (*api.PDAOSetSnapshotAddressResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	reg, err := services.GetRocketSignerRegistry(c)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("Error getting the signer registry on network [%v].", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}

	// Response
	response := api.PDAOSetSnapshotAddressResponse{}

	// Parse signature into vrs components, v to uint8 and v,s to [32]byte
	sig, err := apiutils.ParseEIP712(signature)
	if err != nil {
		fmt.Println("Error parsing signature", err)
	}

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

	// Call SetSigningDelegate on RocketSignerRegistry
	tx, err := reg.SetSigningDelegate(opts, snapshotAddress, sig.V, sig.R, sig.S)
	if err != nil {
		return nil, fmt.Errorf("Error setting snapshot address: %w", err)
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil
}

func canClearSnapshotAddress(c *cli.Context) (*api.PDAOCanClearSnapshotAddressResponse, error) {
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

	response := api.PDAOCanClearSnapshotAddressResponse{}

	// Get signer registry contract address
	addressString := cfg.Smartnode.GetRocketSignerRegistryAddress()
	if addressString == "" {
		return nil, fmt.Errorf("Network [%v] does not have a signer registry contract.", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}
	rocketSignerRegistryAddress := common.HexToAddress(addressString)

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Create the signer registry contract binding
	rocketSignerRegistryAbi, err := abi.JSON(strings.NewReader(contracts.RocketSignerRegistryABI))
	if err != nil {
		return nil, err
	}
	contract := &rocketpool.Contract{
		Contract: bind.NewBoundContract(rocketSignerRegistryAddress, rocketSignerRegistryAbi, ec, ec, ec),
		Address:  &rocketSignerRegistryAddress,
		ABI:      &rocketSignerRegistryAbi,
		Client:   ec,
	}

	// Get the gas info
	gasInfo, err := contract.GetTransactionGasInfo(opts, "clearSigningDelegate")
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	return &response, nil
}

func clearSnapshotAddress(c *cli.Context) (*api.PDAOClearSnapshotAddressResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	reg, err := services.GetRocketSignerRegistry(c)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("Error getting the signer registry on network [%v].", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}

	response := api.PDAOClearSnapshotAddressResponse{}

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

	// Clear the snapshot address
	tx, err := reg.ClearSigningDelegate(opts)
	if err != nil {
		return nil, fmt.Errorf("Error clearing the snapshot address: %w", err)
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
