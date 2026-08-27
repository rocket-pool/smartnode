package node

import (
	"testing"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
)

func TestLoadAutoTxGas(t *testing.T) {
	cfg := &config.RocketPoolConfig{Smartnode: &config.SmartnodeConfig{}}
	cfg.Smartnode.AutoTxGasThreshold.ID = "autoTx"
	cfg.Smartnode.AutoTxGasThreshold.Value = 12.5
	cfg.Smartnode.ManualMaxFee.ID = "maxFee"
	cfg.Smartnode.ManualMaxFee.Value = 30.0
	cfg.Smartnode.PriorityFee.ID = "prio"
	cfg.Smartnode.PriorityFee.Value = 1.5

	logger := log.NewColorLogger(0)
	gas := loadAutoTxGas(cfg, &logger)
	if gas.thresholdGwei != 12.5 {
		t.Fatalf("threshold = %v, want 12.5", gas.thresholdGwei)
	}
	if gas.maxFee.Cmp(math.GweiToWei(30)) != 0 {
		t.Fatalf("maxFee = %s, want 30 gwei", gas.maxFee)
	}
	if gas.maxPriorityFee.Cmp(math.GweiToWei(1.5)) != 0 {
		t.Fatalf("maxPriorityFee = %s, want 1.5 gwei", gas.maxPriorityFee)
	}
}

func TestLoadAutoTxGasDefaults(t *testing.T) {
	cfg := &config.RocketPoolConfig{Smartnode: &config.SmartnodeConfig{}}
	cfg.Smartnode.AutoTxGasThreshold.Value = 0.0
	cfg.Smartnode.ManualMaxFee.Value = 0.0
	cfg.Smartnode.PriorityFee.Value = 0.0

	logger := log.NewColorLogger(0)
	gas := loadAutoTxGas(cfg, &logger)
	if gas.thresholdGwei != 0 {
		t.Fatalf("threshold = %v, want 0", gas.thresholdGwei)
	}
	if gas.maxFee != nil {
		t.Fatalf("maxFee = %s, want nil", gas.maxFee)
	}
	wantPrio := math.GweiToWei(rpgas.DefaultPriorityFeeGwei)
	if gas.maxPriorityFee.Cmp(wantPrio) != 0 {
		t.Fatalf("maxPriorityFee = %s, want default %s", gas.maxPriorityFee, wantPrio)
	}
}
