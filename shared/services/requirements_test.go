package services

import (
	"context"
	"errors"
	"math/big"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/passwords"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

type registrationTestWallet struct {
	wallet.Wallet
	address common.Address
}

func (w registrationTestWallet) GetInitialized() (bool, error) { return true, nil }
func (w registrationTestWallet) GetNodeAccount() (accounts.Account, error) {
	return accounts.Account{Address: w.address}, nil
}

func TestWaitNodeRegisteredWalletSelection(t *testing.T) {
	// Service caches are process-wide; these subtests must remain sequential.
	dir := t.TempDir()
	testCfg, err := config.NewRocketPoolConfig(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	testCfg.Smartnode.DataPath.Value = dir
	realAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	observedAddress := common.HexToAddress("0x2222222222222222222222222222222222222222")
	pm := passwords.NewPasswordManager(filepath.Join(dir, "password"))
	if err := pm.SetPassword("registration-test-password"); err != nil {
		t.Fatal(err)
	}
	initCfg.Do(func() { cfg = testCfg })
	initPasswordManager.Do(func() { passwordManager = pm })
	initNodeWallet.Do(func() { nodeWallet = registrationTestWallet{address: realAddress} })
	initStaticState.Do(func() { staticState = (&state.NetworkState{}).ToIndexedNetworkState() })
	t.Cleanup(func() {
		cfg, passwordManager, addressManager, nodeWallet, staticState = nil, nil, nil, nil, nil
		initCfg, initPasswordManager, initAddressManager, initNodeWallet, initStaticState = sync.Once{}, sync.Once{}, sync.Once{}, sync.Once{}, sync.Once{}
	})

	command := &cli.Command{
		Flags: []cli.Flag{&cli.StringFlag{Name: "network-state", Value: "test-snapshot"}},
		Action: func(_ context.Context, c *cli.Command) error {
			for _, tt := range []struct {
				name               string
				observe            bool
				useObservedWallet  bool
				realRegistered     bool
				observedRegistered bool
				wantReady          bool
			}{
				{"observed node registered with unregistered real wallet", true, true, false, true, true},
				{"observed node must be registered", true, true, true, false, false},
				{"normal startup uses real wallet", false, false, true, false, true},
				{"ordinary masquerade cannot bypass real registration", false, false, false, true, false},
				{"explicit HD check ignores registered observed node", true, false, false, true, false},
				{"explicit HD check ignores unregistered observed node", true, false, true, false, true},
			} {
				t.Run(tt.name, func(t *testing.T) {
					am := wallet.NewAddressManager(testCfg.Smartnode.GetNodeAddressPath())
					if err := am.SetAndSaveAddress(observedAddress, tt.observe); err != nil {
						t.Fatal(err)
					}
					staticState.NodeDetails = nil
					if tt.realRegistered {
						staticState.NodeDetails = append(staticState.NodeDetails, rpstate.NativeNodeDetails{NodeAddress: realAddress, DistributorBalance: new(big.Int)})
					}
					if tt.observedRegistered {
						staticState.NodeDetails = append(staticState.NodeDetails, rpstate.NativeNodeDetails{NodeAddress: observedAddress, DistributorBalance: new(big.Int)})
					}
					ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
					defer cancel()
					var err error
					if tt.useObservedWallet {
						// Use the actual address-file wallet, as both daemons do in observe mode.
						var w wallet.Wallet
						w, err = GetWallet(c)
						if err != nil {
							t.Fatal(err)
						}
						err = WaitNodeRegisteredWithWallet(ctx, c, w, false)
					} else {
						err = WaitNodeRegistered(ctx, c, false)
					}
					if tt.wantReady && err != nil {
						t.Fatalf("registered wallet did not become ready: %v", err)
					}
					if !tt.wantReady && !errors.Is(err, context.DeadlineExceeded) {
						t.Fatalf("unregistered wallet should keep waiting until cancellation, got %v", err)
					}
				})
			}
			return nil
		},
	}
	if err := command.Run(context.Background(), []string{"test"}); err != nil {
		t.Fatal(err)
	}
}
