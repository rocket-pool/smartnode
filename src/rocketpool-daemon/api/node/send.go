package node

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/tokens"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeSendContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSendContextFactory) Create(args url.Values) (*nodeSendContext, error) {
	c := &nodeSendContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", args, input.ValidateBigInt, &c.amount),
		server.GetStringFromVars("token", args, &c.token),
		server.ValidateArg("recipient", args, input.ValidateAddress, &c.recipient),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSendContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeSendContext, api.NodeSendData](
		router, "send", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSendContext struct {
	handler *NodeHandler

	amount    *big.Int
	token     string
	recipient common.Address
}

func (c *nodeSendContext) PrepareData(data *api.NodeSendData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	ec := sp.GetEthClient()
	txMgr := sp.GetTransactionManager()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}

	// Get the contract (nil in the case of ETH)
	var tokenContract tokens.IErc20Token
	if c.token == "eth" {
		tokenContract = nil
	} else if strings.HasPrefix(c.token, "0x") {
		// Arbitrary token - make sure the contract address is legal
		if !common.IsHexAddress(c.token) {
			return types.ResponseStatus_InvalidArguments, fmt.Errorf("[%s] is not a valid token address", c.token)
		}
		tokenAddress := common.HexToAddress(c.token)

		// Make a binding for it
		tokenContract, err := tokens.NewErc20Contract(rp, tokenAddress, ec, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating ERC20 contract binding: %w", err)
		}
		data.TokenSymbol = tokenContract.Details.Symbol
		data.TokenName = tokenContract.Details.Name
	} else {
		// Load the contracts
		status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
		if err != nil {
			return status, err
		}
		switch c.token {
		case "rpl":
			tokenContract, err = tokens.NewTokenRpl(rp)
		case "fsrpl":
			tokenContract, err = tokens.NewTokenRplFixedSupply(rp)
		case "reth":
			tokenContract, err = tokens.NewTokenReth(rp)
		default:
			return types.ResponseStatus_InvalidArguments, fmt.Errorf("[%s] is not a valid token name", c.token)
		}
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating %s token binding: %w", c.token, err)
		}
	}

	// Get the balance
	if tokenContract != nil {
		err := rp.Query(func(mc *batch.MultiCaller) error {
			tokenContract.BalanceOf(mc, &data.Balance, nodeAddress)
			return nil
		}, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting token balance: %w", err)
		}
	} else {
		// ETH balance
		var err error
		data.Balance, err = ec.BalanceAt(context.Background(), nodeAddress, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting ETH balance: %w", err)
		}
	}

	// Check the balance
	data.InsufficientBalance = (data.Balance.Cmp(common.Big0) == 0)
	data.CanSend = !(data.InsufficientBalance)

	// Get the TX Info
	if data.CanSend {
		var txInfo *eth.TransactionInfo
		var err error
		if tokenContract != nil {
			txInfo, err = tokenContract.Transfer(c.recipient, c.amount, opts)
		} else {
			// ETH transfers
			newOpts := &bind.TransactOpts{
				From:  opts.From,
				Nonce: opts.Nonce,
				Value: c.amount,
			}
			txInfo = txMgr.CreateTransactionInfoRaw(c.recipient, nil, newOpts)
		}
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Transfer: %w", err)
		}
		data.TxInfo = txInfo
	}

	return types.ResponseStatus_Success, nil
}
