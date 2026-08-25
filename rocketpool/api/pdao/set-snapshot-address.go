package pdao

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/rocketpool/eip712"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"

	"github.com/urfave/cli/v3"
)

func canSetSignallingAddress(c *cli.Command, signallingAddress common.Address, signature string) (*api.PDAOCanSetSignallingAddressResponse, error) {

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
	nodeAccount, err := w.GetNodeAccount()
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
	response := api.PDAOCanSetSignallingAddressResponse{}

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

	// Check if the node already has a signer
	callOpts := &bind.CallOpts{}
	nodeToSigner, err := reg.NodeToSigner(callOpts, nodeAccount.Address)
	if err != nil {
		return nil, err
	}

	// Return if there is no signer
	response.NodeToSigner = nodeToSigner
	if nodeToSigner == signallingAddress {
		return &response, nil
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
	sig := eip712.Components{}
	err = sig.UnmarshalText([]byte(signature))
	if err != nil {
		fmt.Println("Error parsing signature", err)
	}

	// Get the gas info
	gasLimits, err := contract.GetTransactionGasInfo(opts, "setSigner", signallingAddress, sig.V, sig.R, sig.S)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits

	// Return response
	return &response, nil
}

func setSignallingAddress(c *cli.Command, signallingAddress common.Address, signature string, t *snroute.TransactOpts) (*api.PDAOSetSignallingAddressResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
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
	response := api.PDAOSetSignallingAddressResponse{}

	// Parse signature into vrs components, v to uint8 and v,s to [32]byte
	sig := eip712.Components{}
	err = sig.UnmarshalText([]byte(signature))
	if err != nil {
		fmt.Println("Error parsing signature", err)
	}

	// Call SetSigner on RocketSignerRegistry
	tx, err := reg.SetSigner(opts, signallingAddress, sig.V, sig.R, sig.S)
	if err != nil {
		return nil, fmt.Errorf("Error setting signalling address: %w", err)
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil
}

func canClearSignallingAddress(c *cli.Command) (*api.PDAOCanClearSignallingAddressResponse, error) {
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

	reg, err := services.GetRocketSignerRegistry(c)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("Error getting the signer registry on network [%v].", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	response := api.PDAOCanClearSignallingAddressResponse{}

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

	// Check if the node already has a signer
	callOpts := &bind.CallOpts{}
	nodeToSigner, err := reg.NodeToSigner(callOpts, nodeAccount.Address)
	if err != nil {
		return nil, err
	}

	// Return if there is no signer
	response.NodeToSigner = nodeToSigner
	if nodeToSigner == (common.Address{}) {
		return &response, nil
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
	gasLimits, err := contract.GetTransactionGasInfo(opts, "clearSigner")
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits

	return &response, nil
}

func clearSignallingAddress(c *cli.Command, t *snroute.TransactOpts) (*api.PDAOClearSignallingAddressResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
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

	response := api.PDAOClearSignallingAddressResponse{}

	// Clear the signalling address
	tx, err := reg.ClearSigner(opts)
	if err != nil {
		return nil, fmt.Errorf("Error clearing the signalling address: %w", err)
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}

func canSetSignallingAddressHandler(ctx snroute.Context) {
	addr := common.HexToAddress(paramVal(ctx.Request, "address"))
	sig := paramVal(ctx.Request, "signature")
	resp, err := canSetSignallingAddress(ctx.Command(), addr, sig)
	response.WriteResponse(ctx.Writer, resp, err)
}

func setSignallingAddressHandler(ctx snroute.WriteContext) {
	addr := common.HexToAddress(paramVal(ctx.Request, "address"))
	sig := paramVal(ctx.Request, "signature")
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := setSignallingAddress(ctx.Command(), addr, sig, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canClearSignallingAddressHandler(ctx snroute.Context) {
	resp, err := canClearSignallingAddress(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}

func clearSignallingAddressHandler(ctx snroute.WriteContext) {
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := clearSignallingAddress(ctx.Command(), opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
