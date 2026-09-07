package node

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const cancelTxGasLimit uint64 = 21000

func canCancelNodeTransaction(c *cli.Command, nonce uint64) (*api.CanCancelNodeTransactionResponse, error) {
	// Require node wallet
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

	// Verify not in observe/masquerade mode
	if _, err := w.GetNodePrivateKeyBytes(); err != nil {
		return nil, fmt.Errorf("node is in observe mode; cannot sign transactions: %w", err)
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	nodeAddress := nodeAccount.Address

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Verify nonce >= latestNonce
	latestNonce, err := ec.NonceAt(ctx, nodeAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting latest on-chain nonce: %w", err)
	}
	if nonce < latestNonce {
		return nil, fmt.Errorf("nonce %d has already been mined (latest on-chain nonce is %d)", nonce, latestNonce)
	}

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
	// Require node wallet
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

	if _, err := w.GetNodePrivateKeyBytes(); err != nil {
		return nil, fmt.Errorf("node is in observe mode; cannot sign transactions: %w", err)
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	nodeAddress := nodeAccount.Address

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify nonce is still valid
	latestNonce, err := ec.NonceAt(ctx, nodeAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting latest on-chain nonce: %w", err)
	}
	if nonce < latestNonce {
		return nil, fmt.Errorf("nonce %d has already been mined (latest on-chain nonce is %d)", nonce, latestNonce)
	}

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

	// Prepare a 0-ETH dynamic fee self-transfer
	chainID := w.GetChainID()
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  opts.GasTipCap,
		GasFeeCap:  opts.GasFeeCap,
		Gas:        cancelTxGasLimit,
		To:         &nodeAddress,
		Value:      big.NewInt(0),
		Data:       []byte{},
		AccessList: []types.AccessTuple{},
	})

	if opts.Signer == nil {
		return nil, fmt.Errorf("transactor signer is not configured")
	}

	signedTx, err := opts.Signer(opts.From, tx)
	if err != nil {
		return nil, fmt.Errorf("error signing cancellation transaction: %w", err)
	}

	if err := ec.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("error broadcasting cancellation transaction: %w", err)
	}

	return &api.CancelNodeTransactionResponse{
		TxHash: signedTx.Hash(),
	}, nil
}

// calculateReplacementFees calculates appropriate tip and max fee for a replacement tx
func calculateReplacementFees(ctx context.Context, ec *services.ExecutionClientManager, nodeAddress common.Address, nonce uint64) (*big.Int, *big.Int) {
	// Base values from current network
	suggestedTip, err := ec.SuggestGasTipCap(ctx)
	if err != nil || suggestedTip == nil {
		suggestedTip = math.GweiToWei(2.0)
	}
	// Default minimum 2 gwei tip
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
	var poolContent txPoolContentFromResponse
	_ = ec.RawCallContext(ctx, &poolContent, "txpool_contentFrom", nodeAddress.Hex())

	nonceStr := strconv.FormatUint(nonce, 10)
	var existingTx map[string]interface{}
	if txMap, ok := poolContent.Pending[nonceStr]; ok {
		existingTx = txMap
	} else if txMap, ok := poolContent.Queued[nonceStr]; ok {
		existingTx = txMap
	}

	if existingTx != nil {
		var oldTip *big.Int
		var oldFee *big.Int

		if p := parseHexBigInt(existingTx["maxPriorityFeePerGas"]); p != nil {
			oldTip = p
		}
		if f := parseHexBigInt(existingTx["maxFeePerGas"]); f != nil {
			oldFee = f
		} else if gp := parseHexBigInt(existingTx["gasPrice"]); gp != nil {
			oldFee = gp
		}

		if oldTip != nil {
			// Apply 15% bump over old tip
			bumpedTip := new(big.Int).Mul(oldTip, big.NewInt(115))
			bumpedTip.Div(bumpedTip, big.NewInt(100))
			if bumpedTip.Cmp(suggestedTip) > 0 {
				suggestedTip = bumpedTip
			}
		}

		if oldFee != nil {
			// Apply 15% bump over old fee
			bumpedFee := new(big.Int).Mul(oldFee, big.NewInt(115))
			bumpedFee.Div(bumpedFee, big.NewInt(100))
			if bumpedFee.Cmp(marketMaxFee) > 0 {
				marketMaxFee = bumpedFee
			}
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
