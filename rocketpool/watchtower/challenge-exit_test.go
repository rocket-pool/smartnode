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

	"github.com/rocket-pool/smartnode/bindings/contracts"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

const exitChallengeTestABI = `[{"type":"function","name":"challengeExit","inputs":[{"name":"challenges","type":"tuple[]","components":[{"name":"megapool","type":"address"},{"name":"validatorIds","type":"uint32[]"}]}],"outputs":[],"stateMutability":"nonpayable"}]`

type exitChallengeLocalClient struct {
	rocketpool.ExecutionClient
	storageABI abi.ABI
	encodedABI string
	manager    common.Address
	sent       int
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

func (c *exitChallengeLocalClient) SendTransaction(context.Context, *types.Transaction) error {
	c.sent++
	return errors.New("transaction leaked to local execution client")
}

type exitChallengeProtectedRPC struct {
	chainID  int64
	chainErr error
	sendErr  error
	received chan *types.Transaction
}

func (r *exitChallengeProtectedRPC) ChainId() (*hexutil.Big, error) {
	return (*hexutil.Big)(big.NewInt(r.chainID)), r.chainErr
}

func (r *exitChallengeProtectedRPC) SendRawTransaction(raw hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(raw); err != nil {
		return common.Hash{}, err
	}
	r.received <- tx
	return tx.Hash(), r.sendErr
}

func TestSubmitExitChallengesProtectedRPC(t *testing.T) {
	for _, test := range []struct {
		name       string
		chainID    int64
		chainErr   error
		sendErr    error
		missingURL bool
		invalidURL bool
		noSend     bool
		noSigner   bool
		wantError  string
		wantSent   bool
	}{
		{name: "protected only", chainID: 1, wantSent: true},
		{name: "rejected without public fallback", chainID: 1, sendErr: errors.New("private submission rejected"), wantError: "private submission rejected", wantSent: true},
		{name: "wrong chain", chainID: 2, wantError: "does not match"},
		{name: "chain lookup failed", chainErr: errors.New("chain lookup failed"), wantError: "chain lookup failed"},
		{name: "missing endpoint", missingURL: true, wantError: "flashbotsProtectUrl is not configured"},
		{name: "invalid endpoint", invalidURL: true, wantError: "error connecting to protected RPC"},
		{name: "submission disabled", noSend: true, wantError: "transaction submission disabled"},
		{name: "signing failed", chainID: 1, noSigner: true, wantError: "no signer"},
	} {
		t.Run(test.name, func(t *testing.T) {
			protected := &exitChallengeProtectedRPC{
				chainID: test.chainID, chainErr: test.chainErr, sendErr: test.sendErr,
				received: make(chan *types.Transaction, 1),
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
			cfg.GetNetworkInfo().FlashbotsProtectUrl = server.URL
			if test.missingURL {
				cfg.GetNetworkInfo().FlashbotsProtectUrl = ""
			}
			if test.invalidURL {
				cfg.GetNetworkInfo().FlashbotsProtectUrl = "://invalid"
			}

			storageABI, err := abi.JSON(strings.NewReader(contracts.RocketStorageABI))
			if err != nil {
				t.Fatal(err)
			}
			encodedABI, err := rocketpool.EncodeAbiStr(exitChallengeTestABI)
			if err != nil {
				t.Fatal(err)
			}
			local := &exitChallengeLocalClient{storageABI: storageABI, encodedABI: encodedABI, manager: common.Address{1}}
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
			task := &challengeValidatorsExiting{cfg: cfg, rp: rp, ec: local}
			tx, err := task.submitExitChallenges(batch, opts)
			if test.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("got error %v, want %q", err, test.wantError)
			}
			if local.sent != 0 {
				t.Fatalf("broadcast %d transactions to local execution client", local.sent)
			}
			if opts.NoSend != test.noSend || opts.Context != originalContext {
				t.Fatal("modified caller's transaction options")
			}
			if (len(protected.received) != 0) != test.wantSent {
				t.Fatalf("protected RPC received %d transactions, wantSent=%t", len(protected.received), test.wantSent)
			}
			if !test.wantSent {
				return
			}
			got := <-protected.received
			if err == nil && got.Hash() != tx.Hash() {
				t.Fatal("protected RPC received a different transaction")
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
				t.Fatal("protected RPC received different exit challenge batch calldata")
			}
		})
	}
}
