package node

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeSend(c *cli.Context, amountWei *big.Int, token string) (*api.CanNodeSendResponse, error) {

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

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Handle explicit token addresses
	if strings.HasPrefix(token, "0x") {
		tokenAddress := common.HexToAddress(token)
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

		response.InsufficientBalance = (amountWei.Cmp(balance) > 0)
		gasInfo, err := contract.EstimateTransferGas(nodeAccount.Address, amountWei, opts)
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
			response.InsufficientBalance = (amountWei.Cmp(ethBalanceWei) > 0)
			gasInfo, err := eth.EstimateSendTransactionGas(ec, nodeAccount.Address, nil, false, opts)
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
			response.InsufficientBalance = (amountWei.Cmp(rplBalanceWei) > 0)
			gasInfo, err := tokens.EstimateTransferRPLGas(rp, nodeAccount.Address, amountWei, opts)
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
			response.InsufficientBalance = (amountWei.Cmp(fixedSupplyRplBalanceWei) > 0)
			gasInfo, err := tokens.EstimateTransferFixedSupplyRPLGas(rp, nodeAccount.Address, amountWei, opts)
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
			response.InsufficientBalance = (amountWei.Cmp(rethBalanceWei) > 0)
			gasInfo, err := tokens.EstimateTransferRETHGas(rp, nodeAccount.Address, amountWei, opts)
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
