package node

import (
	"context"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpsvc "github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

type feeRecipientTestWallet struct {
	wallet.Wallet
	address      common.Address
	masquerading bool
}

func (w feeRecipientTestWallet) IsNodeMasquerading() bool { return w.masquerading }
func (w feeRecipientTestWallet) GetNodeAccount() (accounts.Account, error) {
	return accounts.Account{Address: w.address}, nil
}

type feeRecipientTestState struct {
	state.NetworkStateProvider
	requested common.Address
	snapshot  *state.NetworkStateIndex
	err       error
}

func (s *feeRecipientTestState) GetHeadStateForNode(address common.Address) (*state.NetworkStateIndex, error) {
	s.requested = address
	return s.snapshot, s.err
}

type feeRecipientTestBeacon struct{ beacon.Client }

func (feeRecipientTestBeacon) GetBeaconHead() (beacon.BeaconHead, error) {
	return beacon.BeaconHead{}, nil
}

func TestManageFeeRecipientUsesRealWalletWhileObserving(t *testing.T) {
	realAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	observedAddress := common.HexToAddress("0x2222222222222222222222222222222222222222")
	realRecipient := common.HexToAddress("0x3333333333333333333333333333333333333333")
	observedRecipient := common.HexToAddress("0x4444444444444444444444444444444444444444")
	for _, tt := range []struct {
		name                                  string
		masquerading, missingReal, fetchError bool
	}{
		{name: "real wallet state is fetched"},
		{name: "unregistered real wallet leaves recipient unchanged", missingReal: true},
		{name: "state failure leaves recipient unchanged", fetchError: true},
		{name: "masqueraded wallet is rejected", masquerading: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg, err := config.NewRocketPoolConfig(dir, true)
			if err != nil {
				t.Fatal(err)
			}
			cfg.Smartnode.DataPath.Value = dir
			am := wallet.NewAddressManager(cfg.Smartnode.GetNodeAddressPath())
			if err := am.SetAndSaveAddress(observedAddress, true); err != nil {
				t.Fatal(err)
			}
			path := cfg.Smartnode.GetGlobalFeeRecipientFilePath()
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := rpsvc.UpdateGlobalFeeRecipientFile(realRecipient, cfg); err != nil {
				t.Fatal(err)
			}
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			observedNode := &rpstate.NativeNodeDetails{NodeAddress: observedAddress, FeeDistributorAddress: observedRecipient, MinipoolCount: big.NewInt(1), SmoothingPoolRegistrationChanged: new(big.Int)}
			observedState := &state.NetworkStateIndex{NetworkState: &state.NetworkState{NetworkDetails: &rpstate.NetworkDetails{}}, NodeDetailsByAddress: map[common.Address]*rpstate.NativeNodeDetails{observedAddress: observedNode}}
			realState := &state.NetworkStateIndex{NetworkState: observedState.NetworkState, NodeDetailsByAddress: map[common.Address]*rpstate.NativeNodeDetails{observedAddress: observedNode}}
			if !tt.missingReal {
				realState.NodeDetailsByAddress[realAddress] = &rpstate.NativeNodeDetails{NodeAddress: realAddress, FeeDistributorAddress: realRecipient, MinipoolCount: big.NewInt(1), SmoothingPoolRegistrationChanged: new(big.Int)}
			}
			provider := &feeRecipientTestState{snapshot: realState}
			if tt.fetchError {
				provider.err = errors.New("state unavailable")
			}
			command := &cli.Command{
				Flags: []cli.Flag{&cli.StringFlag{Name: "network-state", Value: "test-snapshot"}},
				Action: func(_ context.Context, c *cli.Command) error {
					task := &manageFeeRecipient{c: c, cfg: cfg, log: log.NewColorLogger(ManageFeeRecipientColor), w: feeRecipientTestWallet{address: realAddress, masquerading: tt.masquerading}, bc: feeRecipientTestBeacon{}, stateManager: provider}
					err := task.run(observedState)
					wantError := tt.masquerading || tt.missingReal || tt.fetchError
					if (err != nil) != wantError {
						t.Fatalf("run error = %v, want error = %v", err, wantError)
					}
					if tt.masquerading && !errors.Is(err, wallet.ErrIsMasquerading) {
						t.Fatalf("expected masquerade rejection, got %v", err)
					}
					return nil
				},
			}
			if err := command.Run(context.Background(), []string{"test"}); err != nil {
				t.Fatal(err)
			}
			if !tt.masquerading && provider.requested != realAddress {
				t.Fatalf("fetched state for %s instead of real wallet", provider.requested)
			}
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(contents) != string(original) {
				t.Fatal("fee recipient changed while observing")
			}
		})
	}
}
