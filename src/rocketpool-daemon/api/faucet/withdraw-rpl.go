package faucet

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type faucetWithdrawContextFactory struct {
	handler *FaucetHandler
}

func (f *faucetWithdrawContextFactory) Create(args url.Values) (*faucetWithdrawContext, error) {
	c := &faucetWithdrawContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *faucetWithdrawContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*faucetWithdrawContext, api.FaucetWithdrawRplData](
		router, "withdraw-rpl", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type faucetWithdrawContext struct {
	handler     *FaucetHandler
	rp          *rocketpool.RocketPool
	f           *contracts.RplFaucet
	nodeAddress common.Address

	allowance *big.Int
}

func (c *faucetWithdrawContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.f = sp.GetRplFaucet()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	return errors.Join(
		sp.RequireNodeRegistered(c.handler.context),
		sp.RequireRplFaucet(),
	)
}

func (c *faucetWithdrawContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.f.Balance,
		c.f.WithdrawalFee,
	)
	c.f.GetAllowanceFor(mc, &c.allowance, c.nodeAddress)
}

func (c *faucetWithdrawContext) PrepareData(data *api.FaucetWithdrawRplData, opts *bind.TransactOpts) error {
	// Get node account balance
	nodeAccountBalance, err := c.rp.Client.BalanceAt(context.Background(), c.nodeAddress, nil)
	if err != nil {
		return fmt.Errorf("error getting node account balance: %w", err)
	}

	// Populate the response
	data.InsufficientFaucetBalance = (c.f.Balance.Get().Cmp(big.NewInt(0)) == 0)
	data.InsufficientAllowance = (c.allowance.Cmp(big.NewInt(0)) == 0)
	data.InsufficientNodeBalance = (nodeAccountBalance.Cmp(c.f.WithdrawalFee.Get()) < 0)
	data.CanWithdraw = !(data.InsufficientFaucetBalance || data.InsufficientAllowance || data.InsufficientNodeBalance)

	// Get withdrawal amount
	var amount *big.Int
	balance := c.f.Balance.Get()
	if balance.Cmp(c.allowance) > 0 {
		amount = c.allowance
	} else {
		amount = balance
	}
	data.Amount = amount

	// Get the TX
	if data.CanWithdraw && opts != nil {
		opts.Value = c.f.WithdrawalFee.Get()

		txInfo, err := c.f.Withdraw(amount, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for Withdraw: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
