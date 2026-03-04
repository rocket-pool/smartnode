package connectivity

import (
	"testing"

	cfg "github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	log "github.com/rocket-pool/smartnode/shared/utils/log"
)

// TestCheckPortConnectivity_Run verifies that the port connectivity task correctly
// decides whether to perform network checks based on the client modes and global
// alerting settings.
func TestCheckPortConnectivity_Run(t *testing.T) {
	logger := log.NewColorLogger(0)
	tests := []struct {
		name                string
		ecMode              cfgtypes.Mode
		ccMode              cfgtypes.Mode
		enableAlerting      bool
		portAlertingEnabled bool
		expectNetCalls      bool
	}{
		{
			name:                "local EC + local CC -> performs checks",
			ecMode:              cfgtypes.Mode_Local,
			ccMode:              cfgtypes.Mode_Local,
			enableAlerting:      true,
			portAlertingEnabled: true,
			expectNetCalls:      true,
		},
		{
			name:                "external EC + external CC -> skips checks",
			ecMode:              cfgtypes.Mode_External,
			ccMode:              cfgtypes.Mode_External,
			enableAlerting:      true,
			portAlertingEnabled: true,
			expectNetCalls:      false,
		},
		{
			name:                "alerting disabled -> skips checks",
			ecMode:              cfgtypes.Mode_Local,
			ccMode:              cfgtypes.Mode_Local,
			enableAlerting:      false,
			portAlertingEnabled: true,
			expectNetCalls:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := cfg.NewRocketPoolConfig("", false)
			config.ExecutionClientMode.Value = tc.ecMode
			config.ConsensusClientMode.Value = tc.ccMode
			config.Alertmanager.EnableAlerting.Value = tc.enableAlerting
			config.Alertmanager.AlertEnabled_PortConnectivityCheck.Value = tc.portAlertingEnabled

			netCallsMade := false
			mockGetPublicIP := func() (string, error) {
				netCallsMade = true
				return "1.2.3.4", nil
			}
			mockIsPortReachable := func(host string, port uint16) bool {
				netCallsMade = true
				return true
			}
			mockExternalCheck := func(port uint16) (bool, string, error) {
				netCallsMade = true
				return true, "Success", nil
			}
			mockAlert := func(cfg *cfg.RocketPoolConfig, port uint16) error {
				return nil
			}

			task := &CheckPortConnectivity{
				cfg:                            config,
				log:                            logger,
				getPublicIP:                    mockGetPublicIP,
				isPortReachableNATReflection:   mockIsPortReachable,
				isPortReachableExternalService: mockExternalCheck,
				alertEth1P2PPortNotOpen:        mockAlert,
				alertBeaconP2PPortNotOpen:      mockAlert,
			}

			err := task.Run()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if netCallsMade != tc.expectNetCalls {
				t.Errorf("expected network calls: %v, got: %v", tc.expectNetCalls, netCallsMade)
			}
		})
	}
}
