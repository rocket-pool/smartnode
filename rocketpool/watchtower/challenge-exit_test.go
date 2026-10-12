package watchtower

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/fatih/color"

	"github.com/rocket-pool/smartnode/bindings/contracts"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/flashbots"
)

const exitChallengeTestABI = `[{"type":"function","name":"challengeExit","inputs":[{"name":"challenges","type":"tuple[]","components":[{"name":"megapool","type":"address"},{"name":"validatorIds","type":"uint32[]"}]}],"outputs":[],"stateMutability":"nonpayable"}]`

type exitChallengeLocalClient struct {
	rocketpool.ExecutionClient
	storageABI abi.ABI
	encodedABI string
	manager    common.Address
	sent       int
	allowSend  bool
	received   *types.Transaction
}

func (c *exitChallengeLocalClient) CallContract(_ context.Context, msg ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	method, err := c.storageABI.MethodById(msg.Data)
	if err != nil {
		return nil, err
	}
	switch method.Name {
	case "getAddress":
		return method.Outputs.Pack(c.manager)
	case "getString":
		return method.Outputs.Pack(c.encodedABI)
	default:
		return nil, fmt.Errorf("unexpected storage call: %s", method.Name)
	}
}

func (c *exitChallengeLocalClient) PendingNonceAt(context.Context, common.Address) (uint64, error) {
	return 7, nil
}

func (c *exitChallengeLocalClient) SendTransaction(_ context.Context, tx *types.Transaction) error {
	c.sent++
	if !c.allowSend {
		return errors.New("transaction leaked to local execution client")
	}
	c.received = tx
	return nil
}

func (c *exitChallengeLocalClient) BlockNumber(context.Context) (uint64, error) {
	return 100, nil
}

func (c *exitChallengeLocalClient) TransactionReceipt(_ context.Context, hash common.Hash) (*types.Receipt, error) {
	return &types.Receipt{TxHash: hash, Status: types.ReceiptStatusSuccessful}, nil
}

type exitChallengeRelay struct {
	sendErr  error
	received chan *types.Transaction
}

func (r *exitChallengeRelay) SendBundle(args struct{ Txs []hexutil.Bytes }) (map[string]string, error) {
	if len(args.Txs) != 1 {
		return nil, fmt.Errorf("expected one transaction, got %d", len(args.Txs))
	}
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(args.Txs[0]); err != nil {
		return nil, err
	}
	r.received <- tx
	return map[string]string{"bundleHash": tx.Hash().Hex()}, r.sendErr
}

func (r *exitChallengeRelay) CancelBundle(_ struct{ ReplacementUUID string }) (bool, error) {
	return true, nil
}

func TestSubmitExitChallengesBundle(t *testing.T) {
	for _, test := range []struct {
		name        string
		sendErr     error
		missingURL  bool
		invalidURL  bool
		noSend      bool
		noSigner    bool
		wantError   string
		wantBundles int
		wantLocal   bool
	}{
		{name: "protected only", wantBundles: int(flashbots.DefaultBundleBlockCount)},
		{name: "rejected without public fallback", sendErr: errors.New("private submission rejected"), wantError: "private submission rejected", wantBundles: 1},
		{name: "missing endpoint", missingURL: true, wantLocal: true},
		{name: "invalid endpoint", invalidURL: true, wantError: "error submitting exit challenge bundle"},
		{name: "submission disabled", noSend: true, wantError: "transaction submission disabled"},
		{name: "signing failed", noSigner: true, wantError: "no signer"},
	} {
		t.Run(test.name, func(t *testing.T) {
			protected := &exitChallengeRelay{
				sendErr:  test.sendErr,
				received: make(chan *types.Transaction, flashbots.DefaultBundleBlockCount),
			}
			rpcServer := rpc.NewServer()
			if err := rpcServer.RegisterName("eth", protected); err != nil {
				t.Fatal(err)
			}
			defer rpcServer.Stop()
			server := httptest.NewServer(rpcServer)
			defer server.Close()

			cfg, err := config.NewRocketPoolConfig(t.TempDir(), true)
			if err != nil {
				t.Fatal(err)
			}
			cfg.GetNetworkInfo().FlashbotsRelayUrl = server.URL
			if test.missingURL {
				cfg.GetNetworkInfo().FlashbotsRelayUrl = ""
			}
			if test.invalidURL {
				cfg.GetNetworkInfo().FlashbotsRelayUrl = "://invalid"
			}

			storageABI, err := abi.JSON(strings.NewReader(contracts.RocketStorageABI))
			if err != nil {
				t.Fatal(err)
			}
			encodedABI, err := rocketpool.EncodeAbiStr(exitChallengeTestABI)
			if err != nil {
				t.Fatal(err)
			}
			local := &exitChallengeLocalClient{storageABI: storageABI, encodedABI: encodedABI, manager: common.Address{1}, allowSend: test.wantLocal}
			rp, err := rocketpool.NewRocketPool(local, common.Address{2})
			if err != nil {
				t.Fatal(err)
			}
			key, err := crypto.GenerateKey()
			if err != nil {
				t.Fatal(err)
			}
			opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1))
			if err != nil {
				t.Fatal(err)
			}
			opts.GasLimit = 500000
			opts.GasFeeCap = big.NewInt(20000000000)
			opts.GasTipCap = big.NewInt(2000000000)
			opts.NoSend = test.noSend
			originalContext := opts.Context
			if test.noSigner {
				opts.Signer = nil
			}
			batch := []megapool.ExitChallenge{
				{Megapool: common.Address{3}, ValidatorIds: []uint32{0, 10}},
				{Megapool: common.Address{4}, ValidatorIds: []uint32{20}},
			}
			task := &challengeValidatorsExiting{log: log.NewColorLogger(color.FgYellow), cfg: cfg, rp: rp, ec: local}
			tx, bundled, err := task.submitExitChallenges(batch, opts)
			if test.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("got error %v, want %q", err, test.wantError)
			}
			if !test.wantLocal && local.sent != 0 {
				t.Fatalf("broadcast %d transactions to local execution client", local.sent)
			}
			if test.wantLocal && local.sent != 1 {
				t.Fatalf("local execution client received %d transactions, want 1", local.sent)
			}
			if err == nil && bundled == test.wantLocal {
				t.Fatalf("bundled = %t, want %t", bundled, !test.wantLocal)
			}
			if opts.NoSend != test.noSend || opts.Context != originalContext {
				t.Fatal("modified caller's transaction options")
			}
			if len(protected.received) != test.wantBundles {
				t.Fatalf("relay received %d bundles, want %d", len(protected.received), test.wantBundles)
			}
			var got *types.Transaction
			switch {
			case test.wantBundles > 0:
				got = <-protected.received
			case test.wantLocal:
				got = local.received
			default:
				return
			}
			if err == nil && got.Hash() != tx.Hash() {
				t.Fatal("submission received a different transaction")
			}
			sender, err := types.Sender(types.LatestSignerForChainID(big.NewInt(1)), got)
			if err != nil || sender != opts.From {
				t.Fatalf("incorrect transaction signer: %s, %v", sender, err)
			}
			if got.To() == nil || *got.To() != local.manager || got.Nonce() != 7 || got.Gas() != opts.GasLimit || got.GasFeeCap().Cmp(opts.GasFeeCap) != 0 || got.GasTipCap().Cmp(opts.GasTipCap) != 0 {
				t.Fatal("incorrect transaction destination, nonce, or gas settings")
			}
			contractABI, err := abi.JSON(strings.NewReader(exitChallengeTestABI))
			if err != nil {
				t.Fatal(err)
			}
			wantData, err := contractABI.Pack("challengeExit", batch)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got.Data(), wantData) {
				t.Fatal("submission received different exit challenge batch calldata")
			}
		})
	}
}
