package config

import (
	"fmt"
	"os"
	"path"

	"github.com/rocket-pool/smartnode/shared/types/config"
	"gopkg.in/yaml.v2"
)

const (
	defaultNetworksConfigPath = "networks-default.yml"
	extraNetworksConfigPath   = "networks-extra.yml"
)

type NetworksConfig struct {
	Networks      []*config.NetworkInfo `yaml:"networks"`
	networkLookup map[config.Network]*config.NetworkInfo
}

func NewNetworksConfig() *NetworksConfig {
	return &NetworksConfig{
		Networks:      make([]*config.NetworkInfo, 0),
		networkLookup: make(map[config.Network]*config.NetworkInfo),
	}
}

func LoadNetworksFromFile(configPath string) (*NetworksConfig, error) {
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("config path does not exist: %s", configPath)
	}

	filePath := path.Join(configPath, defaultNetworksConfigPath)
	defaultNetworks, err := loadNetworksFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not load default networks file: %w", err)
	}

	filePath = path.Join(configPath, extraNetworksConfigPath)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		// if the file didn't exist, we just use the default networks
		return defaultNetworks, nil
	}
	extraNetworks, err := loadNetworksFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not load extra networks file: %w", err)
	}
	defaultNetworks.Networks = append(defaultNetworks.Networks, extraNetworks.Networks...)

	// save the networks for lookup
	defaultNetworks.networkLookup = make(map[config.Network]*config.NetworkInfo) // Update the type of networkLookup
	for _, network := range defaultNetworks.Networks {
		defaultNetworks.networkLookup[config.Network(network.Name)] = network // Update the assignment statement with type assertion
	}

	return defaultNetworks, nil
}

func loadNetworksFile(filePath string) (*NetworksConfig, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read default networks file: %w", err)
	}

	networks := NetworksConfig{}
	err = yaml.Unmarshal(bytes, &networks)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal default networks file: %w", err)
	}

	return &networks, nil
}

func (nc *NetworksConfig) AllNetworks() []*config.NetworkInfo {
	return nc.Networks
}

func (nc *NetworksConfig) GetNetwork(name config.Network) *config.NetworkInfo {
	network, ok := nc.networkLookup[name]
	if !ok {
		return nil
	}
	return network
}
