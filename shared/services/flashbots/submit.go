package flashbots

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

// Defaults for bundle submission: retry the bundle over the next 4 blocks (~48s on mainnet)
// and give inclusion polling a bit of margin on top of that.
const (
	DefaultBundleBlockCount  uint64 = 5
	DefaultSubmissionTimeout        = 70 * time.Second
)

// SubmitBundleAndWait bundles the given signed transactions, targets the next block, submits
// the bundle with redundancy across the next nBlocks blocks and waits for inclusion
// (cancelling the remaining bundles once one of them lands).
// If relayUrl is empty, the relay is resolved from the chain ID.
// Returns true if the bundle was included before ctx expired.
func SubmitBundleAndWait(ctx context.Context, logger *slog.Logger, ethRpc EthRpc, relayUrl string, txs []*types.Transaction, nBlocks uint64) (bool, error) {
	// A random searcher key is generated internally (standard for public bundles)
	client, err := NewClient(logger, ethRpc, relayUrl, nil)
	if err != nil {
		return false, errors.Join(errors.New("error creating flashbots client"), err)
	}

	bundle := NewBundleWithTransactions(txs)

	// Target the next block; on error leave the target at 0 and let the client resolve it.
	if blockNum, err := ethRpc.BlockNumber(ctx); err == nil {
		_ = bundle.SetTargetBlockNumber(blockNum + 1)
	}

	return client.SendNBundleAndWait(ctx, bundle, nBlocks)
}
