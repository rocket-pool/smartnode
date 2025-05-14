package node

import (
	"bytes"
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

func canNodeSend(c *cli.Context, amountRaw float64, token string, to common.Address) (*api.CanNodeSendResponse, error) {

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

		if bytes.Equal(to.Bytes(), tokenAddress.Bytes()) {
			return nil, fmt.Errorf("sending tokens to the same address as the token is prohibited for safety")
		}

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

		amountWei := eth.EthToWeiWithDecimals(amountRaw, contract.Decimals)
		response.TokenName = contract.Name
		response.TokenSymbol = contract.Symbol

		// Get the balance
		balance, err := contract.BalanceOf(nodeAccount.Address, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting ERC20 balance: %w", err)
		}

		response.Balance = eth.WeiToEthWithDecimals(balance, contract.Decimals)
		response.InsufficientBalance = (amountWei.Cmp(balance) > 0)

		// Get the gas info
		gasInfo, err := contract.EstimateTransferGas(to, amountWei, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	} else {
		// Handle well-known token types
		amountWei := eth.EthToWei(amountRaw)
		var balanceWei *big.Int
		switch token {
		case "eth":

			// Check node ETH balance
			balanceWei, err = ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			response.InsufficientBalance = (amountWei.Cmp(balanceWei) > 0)
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
			balanceWei, err = tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			response.InsufficientBalance = (amountWei.Cmp(balanceWei) > 0)
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
			balanceWei, err = tokens.GetFixedSupplyRPLBalance(rp, nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			response.InsufficientBalance = (amountWei.Cmp(balanceWei) > 0)
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
			balanceWei, err = tokens.GetRETHBalance(rp, nodeAccount.Address, nil)
			if err != nil {
				return nil, err
			}
			response.InsufficientBalance = (amountWei.Cmp(balanceWei) > 0)
			gasInfo, err := tokens.EstimateTransferRETHGas(rp, to, amountWei, opts)
			if err != nil {
				return nil, err
			}
			response.GasInfo = gasInfo

		}
		response.Balance = eth.WeiToEth(balanceWei)
	}

	// Update & return response
	response.CanSend = !response.InsufficientBalance
	return &response, nil

}

func nodeSend(c *cli.Context, amountRaw float64, token string, to common.Address) (*api.NodeSendResponse, error) {

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

		amountWei := eth.EthToWeiWithDecimals(amountRaw, contract.Decimals)

		tx, err := contract.Transfer(to, amountWei, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = tx.Hash()
	} else {
		amountWei := eth.EthToWei(amountRaw)
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
