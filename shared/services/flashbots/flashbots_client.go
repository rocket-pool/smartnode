package flashbots

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthRpc is the minimal execution client surface needed by the Flashbots client.
// Both *ethclient.Client and the Smart Node's ExecutionClientManager satisfy it.
type EthRpc interface {
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type ErrorType struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   ErrorType       `json:"error"`
}

type Request struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RawSimulationResultTransaction struct {
	CoinbaseDiff      string `json:"coinbaseDiff"`
	EthSentToCoinbase string `json:"ethSentToCoinbase"`
	FromAddress       string `json:"fromAddress"`
	GasFees           string `json:"gasFees"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           uint64 `json:"gasUsed"`
	ToAddress         string `json:"toAddress"`
	TxHash            string `json:"txHash"`
	Value             string `json:"value"`

	// revert related
	Error        string `json:"error"`
	RevertReason string `json:"revert"`
}

type RawSimulationResultBundle struct {
	BundleGasPrice    string
	BundleHash        string
	CoinbaseDiff      string
	EthSentToCoinbase string
	GasFees           string
	Results           []RawSimulationResultTransaction
	StateBlockNumber  uint64
	TotalGasUsed      uint64
	FirstRevert       string
}

type SimulationResultTransaction struct {
	CoinbaseDiff      *big.Int       `json:"coinbaseDiff"`
	EthSentToCoinbase *big.Int       `json:"ethSentToCoinbase"`
	FromAddress       common.Address `json:"fromAddress"`
	GasFees           *big.Int       `json:"gasFees"`
	GasPrice          *big.Int       `json:"gasPrice"`
	GasUsed           uint64         `json:"gasUsed"`
	ToAddress         common.Address `json:"toAddress"`
	TxHash            common.Hash    `json:"txHash"`
	Value             *big.Int       `json:"value"`

	// revert related
	Error        string `json:"error"`
	RevertReason string `json:"revert"`
}

type SimulationResultBundle struct {
	BundleGasPrice    *big.Int
	BundleHash        common.Hash
	CoinbaseDiff      *big.Int
	EthSentToCoinbase *big.Int
	GasFees           *big.Int
	Results           []SimulationResultTransaction
	StateBlockNumber  uint64
	TotalGasUsed      uint64
	FirstRevert       common.Hash
}

type FlashbotsClient struct {
	url                string
	logger             *slog.Logger
	client             http.Client
	ethRpc             EthRpc
	searcherSecret     *ecdsa.PrivateKey
	searcherAddress    common.Address
	nextRequestId      int
	nextRequestIdMutex sync.Mutex
}

func NewClientRpcString(logger *slog.Logger, ethRpcStr string, relayUrl string, searcherSecret *ecdsa.PrivateKey) (*FlashbotsClient, error) {
	if ethRpcStr == "" {
		return nil, errors.New("ethRpc is required")
	}

	rpcClient, err := ethclient.Dial(ethRpcStr)
	if err != nil {
		return nil, errors.Join(errors.New("error connecting to rpc client"), err)
	}

	return NewClient(logger, rpcClient, relayUrl, searcherSecret)
}

// NewClient creates a Flashbots client using the given execution client for on-chain queries.
// If relayUrl is empty, the relay is resolved from FlashbotsUrlPerNetwork using the chain ID.
func NewClient(logger *slog.Logger, ethRpc EthRpc, relayUrl string, searcherSecret *ecdsa.PrivateKey) (*FlashbotsClient, error) {
	if logger != nil {
		logger = logger.With(slog.String("module", "flashbots_client"))
	}

	url := relayUrl
	if url == "" {
		ctx := context.Background()
		timeoutContext, cancel := context.WithTimeout(ctx, 2*time.Second)

		chainId, err := ethRpc.ChainID(timeoutContext)
		cancel()

		if err != nil {
			return nil, errors.Join(errors.New("error calling rpc client"), err)
		}

		var ok bool
		url, ok = FlashbotsUrlPerNetwork[chainId.Uint64()]
		if !ok {
			return nil, fmt.Errorf("network %s not supported", chainId)
		}
	}

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	// Generate a random secret if searcherSecret is nil
	if searcherSecret == nil {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate random secret: %v", err)
		}
		searcherSecret = privateKey
	}

	searcherPublicKey := searcherSecret.Public()
	searcherPublicKeyECDSA, ok := searcherPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}
	searcherAddress := crypto.PubkeyToAddress(*searcherPublicKeyECDSA)

	fbClient := FlashbotsClient{
		url:                url,
		logger:             logger,
		client:             client,
		ethRpc:             ethRpc,
		searcherSecret:     searcherSecret,
		searcherAddress:    searcherAddress,
		nextRequestId:      0,
		nextRequestIdMutex: sync.Mutex{},
	}

	return &fbClient, nil
}

func (client *FlashbotsClient) Call(method string, params ...interface{}) (json.RawMessage, error) {
	return client.CallWithAdditionalHeaders(method, map[string]string{}, params...)
}

func (client *FlashbotsClient) CallWithAdditionalHeaders(method string, headers map[string]string, params ...interface{}) (json.RawMessage, error) {
	request := Request{
		ID:      client.getNextRequestId(),
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", client.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// add flashbots auth header
	hashedBody := crypto.Keccak256Hash([]byte(body)).Hex()
	signature, err := crypto.Sign(
		accounts.TextHash([]byte(hashedBody)),
		client.searcherSecret,
	)
	if err != nil {
		return nil, errors.Join(errors.New("error signing payload"), err)
	}
	req.Header.Add("X-Flashbots-Signature", client.searcherAddress.Hex()+":"+hexutil.Encode(signature))

	response, err := client.client.Do(req)
	if response != nil {
		defer func() {
			if cerr := response.Body.Close(); cerr != nil && client.logger != nil {
				client.logger.Warn("error closing flashbots response body", slog.String("error", cerr.Error()))
			}
		}()
	}
	if err != nil {
		return nil, errors.Join(errors.New("error calling flashbots API"), err)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Join(errors.New("error reading response body"), err)
	}

	resp := new(Response)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, errors.Join(errors.New("error parsing response"), err)
	}

	if resp.Error.Code != 0 {
		return nil, parseError(resp.Error)
	}

	return resp.Result, nil
}

func (client *FlashbotsClient) SendBundle(bundle *Bundle) (common.Hash, bool, error) {
	hexEncodedBlocknumber, sendTargetBlockNumber, err := client.hexEncodeBlocknumbeInTheFuture(bundle.targetBlocknumber)
	if err != nil {
		return common.Hash{}, false, errors.Join(errors.New("error encoding block number"), err)
	}

	bundle.targetBlocknumber = sendTargetBlockNumber

	rawTransactions, err := convertTransactionsToRawStrings(bundle.transactions)
	if err != nil {
		return common.Hash{}, false, errors.Join(errors.New("error converting transactions to raw strings"), err)
	}

	params := map[string]interface{}{
		"txs":         rawTransactions,
		"blockNumber": hexEncodedBlocknumber,
	}

	if bundle.minTimestamp != 0 {
		params["minTimestamp"] = bundle.minTimestamp
	}

	if bundle.maxTimestamp != 0 {
		params["maxTimestamp"] = bundle.maxTimestamp
	}

	if len(bundle.revertingTxHashes) > 0 {
		params["revertingTxHashes"] = bundle.revertingTxHashes
	}

	if bundle.replacementUuid != "" {
		params["replacementUuid"] = bundle.replacementUuid
	}

	if len(bundle.builders) > 0 {
		params["builders"] = bundle.builders
	}

	res, err := client.Call("eth_sendBundle", params)
	if err != nil {
		return common.Hash{}, false, errors.Join(errors.New("error calling flashbots"), err)
	}

	var response struct {
		BundleHash string `json:"bundleHash"`
		Smart      bool   `json:"smart"`
	}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return common.Hash{}, false, errors.Join(errors.New("error parsing response"), err)
	}

	bundle.bundleHash = common.HexToHash(response.BundleHash)
	bundle.isSmart = response.Smart
	bundle.uuidAlreadySend = true
	return common.HexToHash(response.BundleHash), bundle.isSmart, nil
}

func (client *FlashbotsClient) SendBundleAndWait(ctx context.Context, bundle *Bundle) (bool, error) {
	_, _, err := client.SendBundle(bundle)
	if err != nil {
		return false, errors.Join(errors.New("error sending bundle"), err)
	}

	return client.WaitForBundleInclusion(ctx, bundle)
}

func (client *FlashbotsClient) SendBundleNTimes(originalBundle *Bundle, n uint64) (bundlesToSend []*Bundle, hash common.Hash, smart bool, err error) {
	bundlesToSend = []*Bundle{originalBundle}

	// create n-1 followup bundles
	if n > 1 {
		nextBundles, err := originalBundle.GetBundelsForNextNBlocks(n - 1)
		if err != nil {
			return bundlesToSend, common.Hash{}, false, errors.Join(errors.New("error getting bundles for next n blocks"), err)
		}
		bundlesToSend = append(bundlesToSend, nextBundles...)
	}

	for _, bundle := range bundlesToSend {
		if client.logger != nil {
			client.logger.Debug("sending bundle", slog.Uint64("targetBlock", bundle.targetBlocknumber))
		}

		hash, smart, err = client.SendBundle(bundle)
		if err != nil {
			return bundlesToSend, common.Hash{}, false, errors.Join(errors.New("error sending bundle"), err)
		}
	}

	return bundlesToSend, hash, smart, nil
}

func (client *FlashbotsClient) SendNBundleAndWait(ctx context.Context, bundle *Bundle, n uint64) (bool, error) {
	bundles, _, _, err := client.SendBundleNTimes(bundle, n)
	if err != nil {
		return false, errors.Join(errors.New("error sending bundle"), err)
	}

	var success bool
	for _, nextBundle := range bundles {
		if client.logger != nil {
			client.logger.Debug("start waiting for bundle inclusion", slog.Uint64("targetBlock", nextBundle.targetBlocknumber))
		}

		success, err = client.WaitForBundleInclusion(ctx, nextBundle)
		if err != nil {
			if client.logger != nil {
				client.logger.Warn("error waiting for bundle inclusion - this does not affect the remaining bundle", slog.String("error", err.Error()))
			}

			// do not return, wait for other bundles
			continue
		}

		if success {
			break
		}
	}

	// cancel all bundles if one of them succeeded (each has its own replacement UUID)
	if success {
		for _, nextBundle := range bundles {
			err = client.CancelBundle(nextBundle.replacementUuid)
			if err != nil && client.logger != nil {
				client.logger.Warn("error canceling bundle - this does not affect the bundle", slog.String("error", err.Error()))
			}
		}
	}

	return success, nil
}

// SimulateBundle simulates the execution of a bundle
// The stateBlocknumber parameter is the block number at which the simulation should start, 0 for the current block
func (client *FlashbotsClient) SimulateBundle(bundle *Bundle, stateBlocknumber uint64) (*SimulationResultBundle, bool, error) {
	hexEncodedBlocknumber, _, err := client.hexEncodeBlocknumber(bundle.targetBlocknumber)
	if err != nil {
		return nil, false, errors.Join(errors.New("error encoding block number"), err)
	}

	var hexEncodedBlocknumberState string
	if stateBlocknumber == bundle.targetBlocknumber {
		hexEncodedBlocknumberState = hexEncodedBlocknumber
	} else {
		hexEncodedBlocknumberState, _, err = client.hexEncodeBlocknumber(stateBlocknumber)
		if err != nil {
			return nil, false, errors.Join(errors.New("error encoding state block number"), err)
		}
	}

	rawTransactions, err := convertTransactionsToRawStrings(bundle.transactions)
	if err != nil {
		return nil, false, errors.Join(errors.New("error converting transactions to raw strings"), err)
	}

	params := map[string]interface{}{
		"txs":              rawTransactions,
		"blockNumber":      hexEncodedBlocknumber,
		"stateBlockNumber": hexEncodedBlocknumberState,
	}

	if bundle.minTimestamp != 0 {
		params["minTimestamp"] = bundle.minTimestamp
	}

	result, err := client.Call("eth_callBundle", params)
	if err != nil {
		return nil, false, errors.Join(errors.New("error calling flashbots"), err)
	}

	var rawSimulationResult RawSimulationResultBundle
	err = json.Unmarshal(result, &rawSimulationResult)
	if err != nil {
		return nil, false, errors.Join(errors.New("error parsing simulation result"), err)
	}

	simulationResult := parseSimulationResultBundle(rawSimulationResult)

	// eval first revert
	for _, tx := range simulationResult.Results {
		if tx.Error != "" {
			simulationResult.FirstRevert = tx.TxHash
			break
		}
	}

	return &simulationResult, simulationResult.FirstRevert == common.Hash{}, nil
}

type BundleStats struct {
	IsHighPriority       bool   `json:"isHighPriority"`
	IsSimulated          bool   `json:"isSimulated"`
	SimulatedAt          string `json:"simulatedAt"`
	ReceivedAt           string `json:"receivedAt"`
	ConsideredByBuilders []struct {
		Pubkey    string `json:"pubkey"`
		Timestamp string `json:"timestamp"`
	} `json:"consideredByBuildersAt"`
	SealedByBuilders []struct {
		Pubkey    string `json:"pubkey"`
		Timestamp string `json:"timestamp"`
	} `json:"sealedByBuildersAt"`
}

func (client *FlashbotsClient) GetBundleStats(bundle *Bundle) (*BundleStats, error) {
	if bundle.isSmart {
		return client.GetSbundleStats(bundle)
	}
	bundleV2, err := client.GetBundleStatsV2(bundle)
	if err != nil {
		return nil, err
	}

	return &BundleStats{
		IsHighPriority:       bundleV2.IsHighPriority,
		IsSimulated:          bundleV2.IsSimulated,
		SimulatedAt:          bundleV2.SimulatedAt,
		ReceivedAt:           bundleV2.ReceivedAt,
		ConsideredByBuilders: bundleV2.ConsideredByBuilders,
		SealedByBuilders:     bundleV2.SealedByBuilders,
	}, nil
}

type BundleStatsV2 struct {
	IsHighPriority       bool   `json:"isHighPriority"`
	IsSimulated          bool   `json:"isSimulated"`
	SimulatedAt          string `json:"simulatedAt"`
	ReceivedAt           string `json:"receivedAt"`
	ConsideredByBuilders []struct {
		Pubkey    string `json:"pubkey"`
		Timestamp string `json:"timestamp"`
	} `json:"consideredByBuildersAt"`
	SealedByBuilders []struct {
		Pubkey    string `json:"pubkey"`
		Timestamp string `json:"timestamp"`
	} `json:"sealedByBuildersAt"`
}

func (client *FlashbotsClient) GetBundleStatsV2(bundle *Bundle) (*BundleStatsV2, error) {
	if bundle.bundleHash == (common.Hash{}) {
		return nil, errors.New("bundle hash not set")
	}

	if bundle.isSmart {
		return nil, errors.New("smart bundles are not supported by 'GetBundleStatsV2', use 'GetSbundleStats'")
	}

	params := map[string]interface{}{
		"bundleHash":  bundle.bundleHash.Hex(),
		"blockNumber": fmt.Sprintf("0x%x", bundle.targetBlocknumber),
	}

	result, err := client.Call("flashbots_getBundleStatsV2", params)
	if err != nil {
		return nil, errors.Join(errors.New("error calling flashbots"), err)
	}

	var response BundleStatsV2
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, errors.Join(errors.New("error parsing response"), err)
	}

	return &response, nil
}

func (client *FlashbotsClient) GetSbundleStats(bundle *Bundle) (*BundleStats, error) {
	if bundle.bundleHash == (common.Hash{}) {
		return nil, errors.New("bundle hash not set")
	}

	if !bundle.isSmart {
		return nil, errors.New("non-smart bundles are not supported by 'GetSbundleStats', use 'GetBundleStatsV2'")
	}

	params := map[string]interface{}{
		"bundleHash":  bundle.bundleHash.Hex(),
		"blockNumber": fmt.Sprintf("0x%x", bundle.targetBlocknumber),
	}

	result, err := client.Call("flashbots_getSbundleStats", params)
	if err != nil {
		return nil, errors.Join(errors.New("error calling flashbots"), err)
	}

	var response BundleStats
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, errors.Join(errors.New("error parsing response"), err)
	}

	return &response, nil
}

func (client *FlashbotsClient) CancelBundle(uuid string) error {
	params := map[string]interface{}{
		"replacementUuid": uuid,
	}

	_, err := client.Call("eth_cancelBundle", params)
	if err != nil {
		return errors.Join(errors.New("error calling flashbots"), err)
	}

	return nil
}

// WaitForBundleInclusion waits for a bundle to be included in a block
func (client *FlashbotsClient) WaitForBundleInclusion(ctx context.Context, bundle *Bundle) (bool, error) {
	targetBlock := bundle.TargetBlockNumber()

	firstTime := true
	var lastBlockChecked uint64
	txs := bundle.Transactions()
	for {
		// 1. Check if the context has been canceled or timed out
		select {
		case <-ctx.Done():
			return false, nil
		default:
			// continue
		}

		// 2. Get current on-chain block
		currentBlock, err := client.getBlocknumber(ctx)
		if err != nil {
			return false, errors.Join(errors.New("error getting current block number"), err)
		}

		if lastBlockChecked != currentBlock {
			wasIncluded, err := client.transactionsIncluded(ctx, txs)
			if err != nil {
				return false, errors.Join(errors.New("error checking transaction inclusion"), err)
			}

			if wasIncluded {
				return true, nil
			}

			lastBlockChecked = currentBlock
		}

		if currentBlock >= targetBlock {
			return false, nil
		}

		// 5. Not past the sealed block yet; sleep and poll again
		time.Sleep(1 * time.Second)

		if client.logger != nil {
			if client.logger.Enabled(ctx, slog.LevelDebug) {
				stats, err := client.GetBundleStats(bundle)
				if err != nil {
					client.logger.Warn("failed to get bundle stats", slog.String("error", err.Error()))
					continue
				}

				if !stats.IsSimulated {
					client.logger.Debug("Bundle not yet seen by relay", slog.Uint64("targetBlock", targetBlock))
				} else {
					if firstTime {
						client.logger.Debug("Bundle received and simulated",
							slog.Uint64("targetBlock", targetBlock),
							slog.String("receivedAt", stats.ReceivedAt),
							slog.String("simulatedAt", stats.SimulatedAt),
						)
						firstTime = false
					}

					client.logger.Debug("Bundle considered or sealed by builders",
						slog.Uint64("targetBlock", targetBlock),
						slog.Int("consideredByBuilders", len(stats.ConsideredByBuilders)),
						slog.Int("sealedByBuilders", len(stats.SealedByBuilders)),
					)
				}
			}
		}
	}
}

func (client *FlashbotsClient) CheckBundleIncusion(ctx context.Context, bundle *Bundle) (bool, error) {
	return client.transactionsIncluded(ctx, bundle.Transactions())
}

func (client *FlashbotsClient) hexEncodeBlocknumber(blocknumber uint64) (string, uint64, error) {
	if blocknumber == 0 {
		n, err := client.getBlocknumber(context.Background())
		if err != nil {
			return "", 0, errors.Join(errors.New("error getting blocknumber"), err)
		}

		blocknumber = n
	}

	return fmt.Sprintf("0x%x", blocknumber), blocknumber, nil
}

func (client *FlashbotsClient) hexEncodeBlocknumbeInTheFuture(blocknumber uint64) (string, uint64, error) {
	if blocknumber == 0 {
		n, err := client.getBlocknumber(context.Background())
		if err != nil {
			return "", 0, errors.Join(errors.New("error getting blocknumber"), err)
		}

		// use next block
		blocknumber = n + 1
	}

	return fmt.Sprintf("0x%x", blocknumber), blocknumber, nil
}

func (client *FlashbotsClient) getBlocknumber(ctx context.Context) (uint64, error) {
	timeoutContext, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return client.ethRpc.BlockNumber(timeoutContext)
}

func (client *FlashbotsClient) getNextRequestId() int {
	client.nextRequestIdMutex.Lock()
	defer client.nextRequestIdMutex.Unlock()

	client.nextRequestId++
	return client.nextRequestId
}

func (client *FlashbotsClient) transactionsIncluded(ctx context.Context, txs []*types.Transaction) (bool, error) {
	for _, tx := range txs {
		txHash := tx.Hash()

		receipt, err := client.ethRpc.TransactionReceipt(ctx, txHash)
		if err != nil {
			if errors.Is(err, ethereum.NotFound) {
				return false, nil
			}
			return false, errors.Join(errors.New("error getting transaction receipt"), err)
		}

		if receipt == nil {
			return false, nil
		}
	}

	return true, nil
}

func convertTransactionsToRawStrings(txs []*types.Transaction) ([]string, error) {
	var txsString []string
	for _, tx := range txs {
		binary, err := tx.MarshalBinary()
		if err != nil {
			return nil, errors.Join(errors.New("error marshalling transaction"), err)
		}

		txHex := "0x" + hex.EncodeToString(binary)

		txsString = append(txsString, txHex)
	}

	return txsString, nil
}

func parseSimulationResultBundle(raw RawSimulationResultBundle) SimulationResultBundle {
	var results []SimulationResultTransaction
	for _, rawTx := range raw.Results {
		tx := SimulationResultTransaction{
			CoinbaseDiff:      new(big.Int),
			EthSentToCoinbase: new(big.Int),
			FromAddress:       common.HexToAddress(rawTx.FromAddress),
			GasFees:           new(big.Int),
			GasPrice:          new(big.Int),
			GasUsed:           rawTx.GasUsed,
			ToAddress:         common.HexToAddress(rawTx.ToAddress),
			TxHash:            common.HexToHash(rawTx.TxHash),
			Value:             new(big.Int),

			Error:        rawTx.Error,
			RevertReason: rawTx.RevertReason,
		}

		tx.CoinbaseDiff.SetString(rawTx.CoinbaseDiff, 10)
		tx.EthSentToCoinbase.SetString(rawTx.EthSentToCoinbase, 10)
		tx.GasFees.SetString(rawTx.GasFees, 10)
		tx.GasPrice.SetString(rawTx.GasPrice, 10)
		tx.Value.SetString(rawTx.Value, 10)

		results = append(results, tx)
	}

	bundle := SimulationResultBundle{
		BundleGasPrice:    new(big.Int),
		BundleHash:        common.HexToHash(raw.BundleHash),
		CoinbaseDiff:      new(big.Int),
		EthSentToCoinbase: new(big.Int),
		GasFees:           new(big.Int),
		Results:           results,
		StateBlockNumber:  raw.StateBlockNumber,
		TotalGasUsed:      raw.TotalGasUsed,
		FirstRevert:       common.HexToHash(raw.FirstRevert),
	}

	bundle.BundleGasPrice.SetString(raw.BundleGasPrice, 10)
	bundle.CoinbaseDiff.SetString(raw.CoinbaseDiff, 10)
	bundle.EthSentToCoinbase.SetString(raw.EthSentToCoinbase, 10)
	bundle.GasFees.SetString(raw.GasFees, 10)

	return bundle
}

func (client *FlashbotsClient) UpdateFeeRefundRecipient(newFeeRefundRecipient common.Address) error {
	res, err := client.Call("flashbots_setFeeRefundRecipient", client.searcherAddress.Hex(), newFeeRefundRecipient.Hex())
	if err != nil {
		return errors.Join(errors.New("error calling flashbots"), err)
	}

	var response struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return errors.Join(errors.New("error parsing response"), err)
	}

	parseFromAddress := common.HexToAddress(response.From)
	parseToAddress := common.HexToAddress(response.To)

	if client.searcherAddress.Cmp(parseFromAddress) != 0 {
		return errors.New("from address not correctly updated")
	}

	if newFeeRefundRecipient.Cmp(parseToAddress) != 0 {
		return errors.New("to address not correctly updated")
	}

	return nil
}

func parseError(err ErrorType) error {
	switch err.Code {
	case JsonRpcParseError:
		return fmt.Errorf("flashbots: failed to parse request. %d: %s", err.Code, err.Message)
	case JsonRpcInvalidRequest:
		return fmt.Errorf("flashbots: invalid request. %d: %s", err.Code, err.Message)
	case JsonRpcMethodNotFound:
		return fmt.Errorf("flashbots: method not found. %d: %s", err.Code, err.Message)
	case JsonRpcInvalidParams:
		return fmt.Errorf("flashbots: invalid params. %d: %s", err.Code, err.Message)
	case JsonRpcInternalError:
		return fmt.Errorf("flashbots: internal error. %d: %s", err.Code, err.Message)
	default:
		return fmt.Errorf("flashbots: error (%d): %s", err.Code, err.Message)
	}
}
