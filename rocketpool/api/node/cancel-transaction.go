package node

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const cancelTxGasLimit uint64 = 21000

func canCancelNodeTransaction(c *cli.Command, nonce uint64) (*api.CanCancelNodeTransactionResponse, error) {
	_, ec, nodeAddress, err := validateCancelPreflight(c, nonce)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Calculate suggested gas fees
	minPriorityFee, suggestedMaxFee := calculateReplacementFees(ctx, ec, nodeAddress, nonce)

	// Check ETH balance
	balance, err := ec.BalanceAt(ctx, nodeAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node account balance: %w", err)
	}

	requiredBalance := new(big.Int).Mul(suggestedMaxFee, big.NewInt(int64(cancelTxGasLimit)))
	if balance.Cmp(requiredBalance) < 0 {
		return &api.CanCancelNodeTransactionResponse{
			CanCancel:           false,
			Nonce:               nonce,
			GasLimit:            cancelTxGasLimit,
			MinPriorityFeeGwei:  math.WeiToGwei(minPriorityFee),
			SuggestedMaxFeeGwei: math.WeiToGwei(suggestedMaxFee),
		}, fmt.Errorf("insufficient ETH balance: have %.6f ETH, need at least %.6f ETH for gas", math.WeiToEth(balance), math.WeiToEth(requiredBalance))
	}

	return &api.CanCancelNodeTransactionResponse{
		CanCancel:           true,
		Nonce:               nonce,
		GasLimit:            cancelTxGasLimit,
		MinPriorityFeeGwei:  math.WeiToGwei(minPriorityFee),
		SuggestedMaxFeeGwei: math.WeiToGwei(suggestedMaxFee),
	}, nil
}

func cancelNodeTransaction(c *cli.Command, nonce uint64, t *snroute.TransactOpts) (*api.CancelNodeTransactionResponse, error) {
	w, ec, nodeAddress, err := validateCancelPreflight(c, nonce)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := t.Opts()
	// Fall back to suggested replacement fees if none provided
	if opts.GasTipCap == nil || opts.GasFeeCap == nil {
		suggTip, suggFee := calculateReplacementFees(ctx, ec, nodeAddress, nonce)
		if opts.GasTipCap == nil {
			opts.GasTipCap = suggTip
		}
		if opts.GasFeeCap == nil {
			opts.GasFeeCap = suggFee
		}
	}

	tx := buildCancelDynamicFeeTx(w.GetChainID(), nonce, nodeAddress, opts.GasTipCap, opts.GasFeeCap)

	txHash, err := broadcastCancelTx(ctx, ec, opts, tx)
	if err != nil {
		return nil, err
	}

	return &api.CancelNodeTransactionResponse{
		TxHash: txHash,
	}, nil
}

// validateCancelPreflight ensures the wallet is unlocked, not observing, and the target nonce is valid.
func validateCancelPreflight(c *cli.Command, nonce uint64) (wallet.Wallet, *services.ExecutionClientManager, common.Address, error) {
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, nil, common.Address{}, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, nil, common.Address{}, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, nil, common.Address{}, err
	}

	if _, err := w.GetNodePrivateKeyBytes(); err != nil {
		return nil, nil, common.Address{}, fmt.Errorf("node is in observe mode; cannot sign transactions: %w", err)
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, nil, common.Address{}, err
	}
	nodeAddress := nodeAccount.Address

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	latestNonce, err := ec.NonceAt(ctx, nodeAddress, nil)
	if err != nil {
		return nil, nil, common.Address{}, fmt.Errorf("error getting latest on-chain nonce: %w", err)
	}
	if nonce < latestNonce {
		return nil, nil, common.Address{}, fmt.Errorf("nonce %d has already been mined (latest on-chain nonce is %d)", nonce, latestNonce)
	}

	return w, ec, nodeAddress, nil
}

// buildCancelDynamicFeeTx constructs a 0-ETH dynamic fee transaction to self.
func buildCancelDynamicFeeTx(chainID *big.Int, nonce uint64, nodeAddress common.Address, gasTipCap, gasFeeCap *big.Int) *types.Transaction {
	return types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		Gas:        cancelTxGasLimit,
		To:         &nodeAddress,
		Value:      big.NewInt(0),
		Data:       []byte{},
		AccessList: []types.AccessTuple{},
	})
}

// broadcastCancelTx signs and transmits the cancellation transaction via the execution client.
func broadcastCancelTx(ctx context.Context, ec *services.ExecutionClientManager, opts *bind.TransactOpts, tx *types.Transaction) (common.Hash, error) {
	if opts.Signer == nil {
		return common.Hash{}, fmt.Errorf("transactor signer is not configured")
	}

	signedTx, err := opts.Signer(opts.From, tx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error signing cancellation transaction: %w", err)
	}

	if err := ec.SendTransaction(ctx, signedTx); err != nil {
		return common.Hash{}, fmt.Errorf("error broadcasting cancellation transaction: %w", err)
	}

	return signedTx.Hash(), nil
}

// calculateReplacementFees calculates appropriate tip and max fee for a replacement tx
func calculateReplacementFees(ctx context.Context, ec *services.ExecutionClientManager, nodeAddress common.Address, nonce uint64) (*big.Int, *big.Int) {
	suggestedTip, err := ec.SuggestGasTipCap(ctx)
	if err != nil || suggestedTip == nil {
		suggestedTip = math.GweiToWei(2.0)
	}
	minTip := math.GweiToWei(2.0)
	if suggestedTip.Cmp(minTip) < 0 {
		suggestedTip = minTip
	}

	baseFee := big.NewInt(0)
	if header, err := ec.HeaderByNumber(ctx, nil); err == nil && header != nil && header.BaseFee != nil {
		baseFee = header.BaseFee
	}

	// Market max fee = 2 * BaseFee + Tip
	marketMaxFee := new(big.Int).Mul(baseFee, big.NewInt(2))
	marketMaxFee.Add(marketMaxFee, suggestedTip)

	// Attempt txpool enrichment
	poolContent := fetchNodeMempool(ctx, ec, nodeAddress)
	oldTip, oldFee := poolContent.getTxGasFees(nonce)

	if oldTip != nil {
		bumpedTip := new(big.Int).Mul(oldTip, big.NewInt(115))
		bumpedTip.Div(bumpedTip, big.NewInt(100))
		if bumpedTip.Cmp(suggestedTip) > 0 {
			suggestedTip = bumpedTip
		}
	}

	if oldFee != nil {
		bumpedFee := new(big.Int).Mul(oldFee, big.NewInt(115))
		bumpedFee.Div(bumpedFee, big.NewInt(100))
		if bumpedFee.Cmp(marketMaxFee) > 0 {
			marketMaxFee = bumpedFee
		}
	}

	// Ensure maxFee >= suggestedTip
	if marketMaxFee.Cmp(suggestedTip) < 0 {
		marketMaxFee = new(big.Int).Add(suggestedTip, math.GweiToWei(1.0))
	}

	return suggestedTip, marketMaxFee
}

func canCancelTransactionHandler(ctx snroute.Context) {
	nonceStr := ctx.Request.URL.Query().Get("nonce")
	nonce, err := strconv.ParseUint(nonceStr, 10, 64)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, fmt.Errorf("invalid nonce: %w", err))
		return
	}
	resp, err := canCancelNodeTransaction(ctx.Command(), nonce)
	response.WriteResponse(ctx.Writer, resp, err)
}

func cancelTransactionHandler(ctx snroute.WriteContext) {
	nonceStr := ctx.Request.FormValue("nonce")
	nonce, err := strconv.ParseUint(nonceStr, 10, 64)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, fmt.Errorf("invalid nonce: %w", err))
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := cancelNodeTransaction(ctx.Command(), nonce, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
