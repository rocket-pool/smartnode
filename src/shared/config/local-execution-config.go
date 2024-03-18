package config

import (
	"github.com/rocket-pool/node-manager-core/config"
)

// Create a new LocalExecutionConfig struct
func NewLocalExecutionConfig() *config.LocalExecutionConfig {
	cfg := config.NewLocalExecutionConfig()
	cfg.Besu.ContainerTag.Default[Network_Devnet] = cfg.Besu.ContainerTag.Default[config.Network_Holesky]
	cfg.Geth.ContainerTag.Default[Network_Devnet] = cfg.Geth.ContainerTag.Default[config.Network_Holesky]
	cfg.Nethermind.ContainerTag.Default[Network_Devnet] = cfg.Nethermind.ContainerTag.Default[config.Network_Holesky]
	cfg.Nethermind.FullPruningThresholdMb.Default[Network_Devnet] = cfg.Nethermind.FullPruningThresholdMb.Default[config.Network_Holesky]
	cfg.Reth.ContainerTag.Default[Network_Devnet] = cfg.Geth.ContainerTag.Default[config.Network_Holesky]
	return cfg
}
