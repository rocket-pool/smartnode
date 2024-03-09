package config

import (
	"github.com/rocket-pool/node-manager-core/config"
)

// Create a new LocalBeaconConfig struct
func NewLocalBeaconConfig() *config.LocalBeaconConfig {
	cfg := config.NewLocalBeaconConfig()
	cfg.Lighthouse.ContainerTag.Default[Network_Devnet] = cfg.Lighthouse.ContainerTag.Default[config.Network_Holesky]
	cfg.Lodestar.ContainerTag.Default[Network_Devnet] = cfg.Lodestar.ContainerTag.Default[config.Network_Holesky]
	cfg.Nimbus.ContainerTag.Default[Network_Devnet] = cfg.Nimbus.ContainerTag.Default[config.Network_Holesky]
	cfg.Prysm.ContainerTag.Default[Network_Devnet] = cfg.Prysm.ContainerTag.Default[config.Network_Holesky]
	cfg.Teku.ContainerTag.Default[Network_Devnet] = cfg.Teku.ContainerTag.Default[config.Network_Holesky]
	return cfg
}
