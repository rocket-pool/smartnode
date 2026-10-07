package config

import (
	"errors"
	"fmt"
	"slices"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

func validateClientList[T ~string](name string, clients, known []T) error {
	seen := make(map[T]bool)
	for _, client := range clients {
		if !slices.Contains(known, client) || seen[client] {
			return fmt.Errorf("invalid or duplicate %s client %q", name, client)
		}
		seen[client] = true
	}
	return nil
}

func validateNetworkClients(n *config.NetworkInfo) error {
	return errors.Join(
		validateClientList("execution", n.ExecutionClients, []config.ExecutionClient{
			config.ExecutionClient_Geth, config.ExecutionClient_Nethermind, config.ExecutionClient_Besu,
			config.ExecutionClient_Reth, config.ExecutionClient_Erigon,
		}),
		validateClientList("consensus", n.ConsensusClients, []config.ConsensusClient{
			config.ConsensusClient_Lighthouse, config.ConsensusClient_Lodestar, config.ConsensusClient_Nimbus,
			config.ConsensusClient_Prysm, config.ConsensusClient_Teku,
		}),
	)
}

func supportsNetworkClient(n *config.NetworkInfo, client any) bool {
	if n == nil {
		return true
	}
	switch client := client.(type) {
	case config.ExecutionClient:
		return len(n.ExecutionClients) == 0 || slices.Contains(n.ExecutionClients, client)
	case config.ConsensusClient:
		return len(n.ConsensusClients) == 0 || slices.Contains(n.ConsensusClients, client)
	default:
		return false
	}
}

func (cfg *RocketPoolConfig) initializeNetworkClientDefaults() {
	cfg.clientOptions = make(map[*config.Parameter][]config.ParameterOption)
	for _, param := range []*config.Parameter{&cfg.ExecutionClient, &cfg.ConsensusClient, &cfg.ExternalConsensusClient, &cfg.Native.ConsensusClient} {
		cfg.clientOptions[param] = param.Options
		for _, network := range cfg.networks.AllNetworks() {
			if supportsNetworkClient(network, param.Default[config.Network_All]) {
				continue
			}
			for _, option := range param.Options {
				if supportsNetworkClient(network, option.Value) {
					param.Default[network.ID()] = option.Value
					break
				}
			}
		}
	}
	for _, network := range cfg.networks.AllNetworks() {
		if network.CheckpointSyncUrl != "" {
			cfg.ConsensusCommon.CheckpointSyncProvider.Default[network.ID()] = network.CheckpointSyncUrl
		}
		if !network.SupportsMevBoost {
			cfg.EnableMevBoost.Default[network.ID()] = false
		}
	}
}

// Refresh choices after loading or switching networks. Only an interactive network
// switch replaces an unsupported selection; loading a saved config preserves it.
func (cfg *RocketPoolConfig) refreshNetworkClientOptions(resetUnsupported bool) {
	network := cfg.GetNetworkInfo()
	for param, allOptions := range cfg.clientOptions {
		param.Options = nil
		for _, option := range allOptions {
			if supportsNetworkClient(network, option.Value) {
				param.Options = append(param.Options, option)
			}
		}
		if resetUnsupported && !supportsNetworkClient(network, param.Value) {
			_ = param.SetToDefault(cfg.GetNetwork())
		}
	}
}

// ValidateNetworkClients also covers saved/headless settings that bypass the TUI.
// External execution endpoints have no client-family setting to validate.
func (cfg *RocketPoolConfig) ValidateNetworkClients() error {
	network := cfg.GetNetworkInfo()
	if network == nil {
		return fmt.Errorf("unknown network %q", cfg.GetNetwork())
	}
	var params []*config.Parameter
	if cfg.IsNativeMode {
		params = append(params, &cfg.Native.ConsensusClient)
	} else {
		if cfg.ExecutionClientMode.Value == config.Mode_Local {
			params = append(params, &cfg.ExecutionClient)
		}
		if cfg.ConsensusClientMode.Value == config.Mode_Local {
			params = append(params, &cfg.ConsensusClient)
		} else {
			params = append(params, &cfg.ExternalConsensusClient)
		}
	}
	var problems []error
	for _, param := range params {
		if !supportsNetworkClient(network, param.Value) {
			problems = append(problems, fmt.Errorf("%s %q is not supported on %s", param.Name, param.Value, network.Label))
		}
	}
	if !network.SupportsMevBoost && (cfg.EnableMevBoost.Value == true || cfg.EnableCommitBoost.Value == true) {
		problems = append(problems, fmt.Errorf("MEV-Boost and Commit-Boost are not supported on %s", network.Label))
	}
	return errors.Join(problems...)
}

// HasRocketPoolContracts reports whether contract-dependent services can run.
func (cfg *RocketPoolConfig) HasRocketPoolContracts() bool {
	network := cfg.GetNetworkInfo()
	return network != nil && !network.ContractsPending && network.Addresses.Storage != ""
}

// RequireRocketPoolContracts prevents accidentally using the zero storage address.
func (cfg *RocketPoolConfig) RequireRocketPoolContracts() error {
	network := cfg.GetNetworkInfo()
	if network == nil {
		return fmt.Errorf("unknown network %q", cfg.GetNetwork())
	}
	if !cfg.HasRocketPoolContracts() {
		return fmt.Errorf("Rocket Pool contracts are not configured for %s; supply the deployment addresses and set contractsPending to false in the network definition before using contract-dependent services", network.Label)
	}
	return nil
}

// ValidateNetworkForStart permits client startup before Rocket Pool is deployed.
func (cfg *RocketPoolConfig) ValidateNetworkForStart() error {
	return cfg.ValidateNetworkClients()
}
