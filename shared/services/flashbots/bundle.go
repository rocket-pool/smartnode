package flashbots

import (
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

type Bundle struct {
	transactions      []*types.Transaction
	targetBlocknumber uint64
	minTimestamp      int64
	maxTimestamp      int64
	revertingTxHashes []string
	replacementUuid   string
	uuidAlreadySend   bool
	builders          []string

	bundleHash common.Hash
	isSmart    bool
}

func NewBundle() *Bundle {
	return &Bundle{
		replacementUuid:   uuid.New().String(),
		targetBlocknumber: 0,
	}
}

func NewBundleWithTransactions(transactions []*types.Transaction) *Bundle {
	return &Bundle{
		replacementUuid:   uuid.New().String(),
		transactions:      transactions,
		targetBlocknumber: 0,
	}
}

func (b *Bundle) Transactions() []*types.Transaction {
	return b.transactions
}

func (b *Bundle) TargetBlockNumber() uint64 {
	return b.targetBlocknumber
}

func (b *Bundle) MinTimestamp() int64 {
	return b.minTimestamp
}

func (b *Bundle) MaxTimestamp() int64 {
	return b.maxTimestamp
}

func (b *Bundle) RevertingTxHashes() []string {
	return b.revertingTxHashes
}

func (b *Bundle) ReplacementUuid() string {
	return b.replacementUuid
}

func (b *Bundle) BundleHash() common.Hash {
	return b.bundleHash
}

func (b *Bundle) Builders() []string {
	return b.builders
}

func (b *Bundle) AddTransaction(tx *types.Transaction) {
	b.transactions = append(b.transactions, tx)
}

func (b *Bundle) AddTransactions(txs []*types.Transaction) {
	b.transactions = append(b.transactions, txs...)
}

// SetBlockNumber sets the block number for which this bundle is valid
// If set to 0, the bundle is valid for the next block
func (b *Bundle) SetTargetBlockNumber(blocknumber uint64) error {
	if b.targetBlocknumber != 0 {
		return errors.New("targetBlocknumber already set")
	}

	b.targetBlocknumber = blocknumber
	return nil
}

// SetMinTimestamp sets the minimum timestamp for which this bundle is valid, in seconds since the unix epoch
func (b *Bundle) SetMinTimestamp(minTimestamp int64) error {
	if b.maxTimestamp != 0 && minTimestamp > b.maxTimestamp {
		return errors.New("minTimestamp must be less than maxTimestamp")
	}

	b.minTimestamp = minTimestamp
	return nil
}

// SetMaxTimestamp sets the maximum timestamp for which this bundle is valid, in seconds since the unix epoch
func (b *Bundle) SetMaxTimestamp(maxTimestamp int64) error {
	if b.minTimestamp != 0 && maxTimestamp < b.minTimestamp {
		return errors.New("maxTimestamp must be greater than minTimestamp")
	}
	if maxTimestamp < time.Now().Unix() {
		return errors.New("maxTimestamp must be in the future")
	}

	b.maxTimestamp = maxTimestamp
	return nil
}

// SetRevertingTxHash sets one transaction hash that is allowed to revert
func (b *Bundle) SetRevertingTxHash(revertingTxHash string) {
	b.revertingTxHashes = append(b.revertingTxHashes, revertingTxHash)
}

// SetRevertingTxHashes sets the list of transaction hashes that are allowed to revert
func (b *Bundle) SetRevertingTxHashes(revertingTxHashes []string) {
	b.revertingTxHashes = revertingTxHashes
}

// SetReplacementUuid sets the replacement UUID for this bundle that can be used to cancel/replace this bundle
func (b *Bundle) SetReplacementUuid(replacementUuid string) error {
	if b.uuidAlreadySend {
		return errors.New("bundle already send to relay, cant change uuid")
	}

	b.replacementUuid = replacementUuid
	return nil
}

func (b *Bundle) MaximumGasFeePaid() (feePaid *big.Int) {
	feePaid = big.NewInt(0)
	for _, tx := range b.transactions {
		feePaid.Add(feePaid, tx.Cost())
	}

	return feePaid
}

func (b *Bundle) UseAllBuilders(networkId uint64) {
	if networkId == 1 {
		b.builders = AllBuilders
	}
}

func (b *Bundle) Copy() *Bundle {
	transactions := make([]*types.Transaction, len(b.transactions))
	copy(transactions, b.transactions)

	revertingTxHashes := make([]string, len(b.revertingTxHashes))
	copy(revertingTxHashes, b.revertingTxHashes)

	builders := make([]string, len(b.builders))
	copy(builders, b.builders)

	return &Bundle{
		transactions:      transactions,
		targetBlocknumber: b.targetBlocknumber,
		minTimestamp:      b.minTimestamp,
		maxTimestamp:      b.maxTimestamp,
		revertingTxHashes: revertingTxHashes,
		replacementUuid:   b.replacementUuid,
		builders:          builders,
		bundleHash:        b.bundleHash,
	}
}

func (b *Bundle) GetBundelsForNextNBlocks(n uint64) ([]*Bundle, error) {
	if b.targetBlocknumber == 0 {
		return nil, errors.New("targetBlocknumber not set")
	}

	bundles := make([]*Bundle, n)
	for i := uint64(0); i < n; i++ {
		bundles[i] = b.Copy()
		bundles[i].targetBlocknumber += (i + 1)
		// Each copy gets its own replacement UUID: relays treat a new eth_sendBundle with the
		// same UUID as a replacement of the previous submission, which would defeat the
		// per-block redundancy these copies provide.
		bundles[i].replacementUuid = uuid.New().String()
	}

	return bundles, nil
}
