package node

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type nodeSendContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSendContextFactory) Create(vars map[string]string) (*nodeSendContext, error) {
	c := &nodeSendContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", vars, input.ValidateBigInt, &c.amount),
		server.GetStringFromVars("token", vars, &c.token),
		server.ValidateArg("recipient", vars, input.ValidateAddress, &c.recipient),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSendContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessRoute[]()[*nodeSendContext, api.NodeDepositData](
		router, "send", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSendContext struct {
	handler     *NodeHandler

	amount        *big.Int
	token         string
	recipient     common.Address
}

func (c *nodeSendContext) Initialize() error {
}

func (c *nodeSendContext) GetState(mc *batch.MultiCaller) {
	if c.tokenContract != nil {
		c.tokenContract.BalanceOf(mc, &c.balance, c.nodeAddress)
	} else {
		switch c.token {
		case "rpl":
			
		}
		core.AddQueryablesToMulticall(mc,
			c.node.Credit,
			c.depositPool.Balance,
			c.pSettings.Node.IsDepositingEnabled,
			c.oSettings.Minipool.ScrubPeriod,
		)
	}
}

func (c *nodeSendContext) PrepareData(data *api.NodeSendData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	ec := sp.GetEthClient()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return err
	}

	// Bindings
	if strings.HasPrefix(c.token, "0x") {
		// Make sure the contract address is legal
		if !common.IsHexAddress(c.token) {
			return fmt.Errorf("[%s] is not a valid token address", c.token)
		}
		tokenAddress := common.HexToAddress(c.token)

		// Make a binding for it
		tokenContract, err := utils.NewErc20Contract(c.rp, tokenAddress, c.ec, nil)
		if err != nil {
			return fmt.Errorf("error creating ERC20 contract binding: %w", err)
		}
		data.TokenSymbol = tokenContract.Details.Symbol
		data.TokenName = tokenContract.Details.Name

		// Get the token balance
		err = rp.Query(func(mc *batch.MultiCaller) error {
			tokenContract.BalanceOf(mc, &data.Balance, nodeAddress)
		}, nil)
		if err != nil {
			return fmt.Errorf("error querying ERC20 balance: %w", err)
		}

		// Check the balance
		if data.Balance.Cmp(common.Big0) == 0 {
			data.InsufficientBalance = true
			data.CanSend = false
		} else {
			txInfo, err := tokenContract.Transfer(c.recipient, c.amount, opts)
			if err != nil {
				return fmt.Errorf("error getting TX info for Transfer: %w", err)
			}
			data.TxInfo = txInfo
		}
	} else {
		switch c.token {
		case "eth":
			// Check the balance
			var err error
			data.Balance, err = c.ec.BalanceAt(context.Background(), c.nodeAddress, nil)
			if err != nil {
				return fmt.Errorf("error getting ETH balance: %w", err)
			}

			// txInfo, err := c.ec.Transfer
			// TODO: need to do SendTransaction but capture the TX before sending it


		case "eth", "rpl", "fsrpl", "reth":
			break
		default:
			return fmt.Errorf("[%s] is not a valid token name or address", c.token)
		}
	}

	return nil

	// Handle ETH balance
	switch c.token {
	case "eth":
	}
}

func canNodeSend(c *cli.Context, amountWei *big.Int, token string, to common.Address) (*api.CanNodeSendResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
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

	// Response
	response := api.CanNodeSendResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the sending opts
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the well-known contracts
	rplContract, err := rp.GetContract("rocketTokenRPL", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting RPL contract address: %w", err)
	}
	fsrplContract, err := rp.GetContract("rocketTokenRPLFixedSupply", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting legacy RPL contract address: %w", err)
	}
	rethContract, err := rp.GetContract("rocketTokenRETH", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting rETH contract address: %w", err)
	}

	// Error out if the recipient is one of the contract addresses
	if to == *rplContract.Address {
		return nil, fmt.Errorf("sending tokens to the RPL contract address is prohibited for safety")
	}
	if to == *fsrplContract.Address {
		return nil, fmt.Errorf("sending tokens to the legacy RPL contract address is prohibited for safety")
	}
	if to == *rethContract.Address {
		return nil, fmt.Errorf("sending tokens to the rETH contract address is prohibited for safety")
	}

	// Handle explicit token addresses
	if strings.HasPrefix(token, "0x") {
		tokenAddress := common.HexToAddress(token)

		// Error out if using one of the well-known ones
		if tokenAddress == *rplContract.Address {
			return nil, fmt.Errorf("sending RPL via the token address is prohibited for safety; please use 'rpl' as the token to send instead of its address")
		}
		if tokenAddress == *fsrplContract.Address {
			return nil, fmt.Errorf("sending legacy RPL via the token address is prohibited for safety; please use 'fsrpl' as the token to send instead of its address")
		}
		if tokenAddress == *rethContract.Address {
			return nil, fmt.Errorf("sending rETH via the token address is prohibited for safety; please use 'reth' as the token to send instead of its address")
		}

		// Create the ERC20 binding
		contract, err := eth.NewErc20Contract(tokenAddress, ec, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating ERC20 contract binding: %w", err)
		}
		response.TokenName = contract.Name
		response.TokenSymbol = contract.Symbol

		// Get the balance
		balance, err := contract.BalanceOf(nodeAccount.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting ERC20 balance: %w", err)
		}
		response.Balance = balance
		response.InsufficientBalance = (amountWei.Cmp(balance) > 0)

		// Get the gas info
		gasInfo, err := contract.EstimateTransferGas(to, amountWei, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	} else {
		// Handle well-known token types
		switch token {
		case "eth":

			// Check node ETH balance
			ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			response.Balance = ethBalanceWei
			response.InsufficientBalance = (amountWei.Cmp(ethBalanceWei) > 0)
			gasInfo, err := eth.EstimateSendTransactionGas(ec, to, nil, false, opts)
			if err != nil {
				return nil, err
			}
			response.GasInfo = gasInfo

		case "rpl":

			// Get RocketStorage
			if err := services.RequireRocketStorage(c); err != nil {
				return nil, err
			}
			// Check node RPL balance
			rplBalanceWei, err := tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			response.Balance = rplBalanceWei
			response.InsufficientBalance = (amountWei.Cmp(rplBalanceWei) > 0)
			gasInfo, err := tokens.EstimateTransferRPLGas(rp, to, amountWei, opts)
			if err != nil {
				return nil, err
			}
			response.GasInfo = gasInfo

		case "fsrpl":

			// Get RocketStorage
			if err := services.RequireRocketStorage(c); err != nil {
				return nil, err
			}
			// Check node fixed-supply RPL balance
			fixedSupplyRplBalanceWei, err := tokens.GetFixedSupplyRPLBalance(rp, nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			response.Balance = fixedSupplyRplBalanceWei
			response.InsufficientBalance = (amountWei.Cmp(fixedSupplyRplBalanceWei) > 0)
			gasInfo, err := tokens.EstimateTransferFixedSupplyRPLGas(rp, to, amountWei, opts)
			if err != nil {
				return nil, err
			}
			response.GasInfo = gasInfo

		case "reth":

			// Get RocketStorage
			if err := services.RequireRocketStorage(c); err != nil {
				return nil, err
			}
			// Check node rETH balance
			rethBalanceWei, err := tokens.GetRETHBalance(rp, nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			response.Balance = rethBalanceWei
			response.InsufficientBalance = (amountWei.Cmp(rethBalanceWei) > 0)
			gasInfo, err := tokens.EstimateTransferRETHGas(rp, to, amountWei, opts)
			if err != nil {
				return nil, err
			}
			response.GasInfo = gasInfo

		}
	}

	// Update & return response
	response.CanSend = !response.InsufficientBalance
	return &response, nil

}

func nodeSend(c *cli.Context, amountWei *big.Int, token string, to common.Address) (*api.NodeSendResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
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

	// Response
	response := api.NodeSendResponse{}

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

	// Handle explicit token addresses
	if strings.HasPrefix(token, "0x") {
		tokenAddress := common.HexToAddress(token)
		contract, err := eth.NewErc20Contract(tokenAddress, ec, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating ERC20 contract binding: %w", err)
		}

		tx, err := contract.Transfer(to, amountWei, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = tx.Hash()
	} else {
		// Handle token type
		switch token {
		case "eth":

			// Transfer ETH
			opts.Value = amountWei
			hash, err := eth.SendTransaction(ec, to, w.GetChainID(), nil, false, opts)
			if err != nil {
				return nil, err
			}
			response.TxHash = hash

		case "rpl":

			// Get RocketStorage
			if err := services.RequireRocketStorage(c); err != nil {
				return nil, err
			}
			// Transfer RPL
			hash, err := tokens.TransferRPL(rp, to, amountWei, opts)
			if err != nil {
				return nil, err
			}
			response.TxHash = hash

		case "fsrpl":

			// Get RocketStorage
			if err := services.RequireRocketStorage(c); err != nil {
				return nil, err
			}
			// Transfer fixed-supply RPL
			hash, err := tokens.TransferFixedSupplyRPL(rp, to, amountWei, opts)
			if err != nil {
				return nil, err
			}
			response.TxHash = hash

		case "reth":

			// Get RocketStorage
			if err := services.RequireRocketStorage(c); err != nil {
				return nil, err
			}
			// Transfer rETH
			hash, err := tokens.TransferRETH(rp, to, amountWei, opts)
			if err != nil {
				return nil, err
			}
			response.TxHash = hash

		}
	}

	// Return response
	return &response, nil

}
