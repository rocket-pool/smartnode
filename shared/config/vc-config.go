package config

import (
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/shared/config/ids"
)

// Configuration for the Validator Client
type ValidatorClientConfig struct {
	VcCommon   *config.ValidatorClientCommonConfig
	Lighthouse *config.LighthouseVcConfig
	Lodestar   *config.LodestarVcConfig
	Nimbus     *config.NimbusVcConfig
	Prysm      *config.PrysmVcConfig
	Teku       *config.TekuVcConfig
}

// Generates a new Validator Client config
func NewValidatorClientConfig() *ValidatorClientConfig {
	cfg := &ValidatorClientConfig{
		VcCommon:   config.NewValidatorClientCommonConfig(),
		Lighthouse: config.NewLighthouseVcConfig(),
		Lodestar:   config.NewLodestarVcConfig(),
		Nimbus:     config.NewNimbusVcConfig(),
		Prysm:      config.NewPrysmVcConfig(),
		Teku:       config.NewTekuVcConfig(),
	}

	cfg.Lighthouse.ContainerTag.Default[Network_Devnet] = cfg.Lighthouse.ContainerTag.Default[config.Network_Holesky]
	cfg.Lodestar.ContainerTag.Default[Network_Devnet] = cfg.Lodestar.ContainerTag.Default[config.Network_Holesky]
	cfg.Nimbus.ContainerTag.Default[Network_Devnet] = cfg.Nimbus.ContainerTag.Default[config.Network_Holesky]
	cfg.Prysm.ContainerTag.Default[Network_Devnet] = cfg.Prysm.ContainerTag.Default[config.Network_Holesky]
	cfg.Teku.ContainerTag.Default[Network_Devnet] = cfg.Teku.ContainerTag.Default[config.Network_Holesky]

	return cfg
}

// The title for the config
func (cfg *ValidatorClientConfig) GetTitle() string {
	return "Validator Client"
}

// Get the parameters for this config
func (cfg *ValidatorClientConfig) GetParameters() []config.IParameter {
	return []config.IParameter{}
}

// Get the sections underneath this one
func (cfg *ValidatorClientConfig) GetSubconfigs() map[string]config.IConfigSection {
	return map[string]config.IConfigSection{
		ids.VcCommonID:   cfg.VcCommon,
		ids.LighthouseID: cfg.Lighthouse,
		ids.LodestarID:   cfg.Lodestar,
		ids.NimbusID:     cfg.Nimbus,
		ids.PrysmID:      cfg.Prysm,
		ids.TekuID:       cfg.Teku,
	}
}
