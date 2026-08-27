package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v2"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool/assets"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	defaultNetworksConfigPath = "networks-default.yml"
	extraNetworksConfigPath   = "networks-extra.yml"
	networksFileVersion       = 1
)

type networksConfigFile struct {
	Version  int                   `yaml:"version"`
	Networks []*config.NetworkInfo `yaml:"networks"`
}

// NetworksConfig is the merged set of official + extra networks.
type NetworksConfig struct {
	Networks []*config.NetworkInfo
	byName   map[config.Network]*config.NetworkInfo
}

func (nc *NetworksConfig) AllNetworks() []*config.NetworkInfo {
	if nc == nil {
		return nil
	}
	return nc.Networks
}

func (nc *NetworksConfig) GetNetwork(name config.Network) *config.NetworkInfo {
	if nc == nil {
		return nil
	}
	return nc.byName[name]
}

func (nc *NetworksConfig) DefaultNetwork() config.Network {
	if nc == nil || len(nc.Networks) == 0 {
		return config.Network_Unknown
	}
	for _, n := range nc.Networks {
		if n.Default {
			return n.ID()
		}
	}
	return nc.Networks[0].ID()
}

func (nc *NetworksConfig) ByChainID(chainID uint64) *config.NetworkInfo {
	if nc == nil {
		return nil
	}
	for _, n := range nc.Networks {
		if uint64(n.ChainID) == chainID {
			return n
		}
	}
	return nil
}

func (nc *NetworksConfig) rebuildLookup() {
	nc.byName = make(map[config.Network]*config.NetworkInfo, len(nc.Networks))
	for _, n := range nc.Networks {
		nc.byName[n.ID()] = n
	}
}

func (nc *NetworksConfig) merge(extra *NetworksConfig) {
	if extra == nil {
		return
	}
	for _, n := range extra.Networks {
		id := n.ID()
		replaced := false
		for i, existing := range nc.Networks {
			if existing.ID() == id {
				nc.Networks[i] = n
				replaced = true
				break
			}
		}
		if !replaced {
			nc.Networks = append(nc.Networks, n)
		}
	}
	nc.rebuildLookup()
}

// LoadNetworks loads official networks (embed, then on-disk default if present)
// and overlays ~/.rocketpool/networks-extra.yml when it exists.
func LoadNetworks(rpDir string) (*NetworksConfig, error) {
	networks, err := parseNetworksYAML(assets.NetworksDefaultYAML(), "embedded "+defaultNetworksConfigPath, true)
	if err != nil {
		return nil, err
	}

	if rpDir != "" {
		diskDefault := filepath.Join(rpDir, defaultNetworksConfigPath)
		if _, err := os.Stat(diskDefault); err == nil {
			diskNetworks, err := loadNetworksFile(diskDefault, true)
			if err != nil {
				return nil, fmt.Errorf("could not load %s: %w", diskDefault, err)
			}
			// On-disk official file fully replaces the embed (same names and any extras in that file).
			networks = diskNetworks
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("could not stat %s: %w", diskDefault, err)
		}

		extraPath := filepath.Join(rpDir, extraNetworksConfigPath)
		if _, err := os.Stat(extraPath); err == nil {
			extra, err := loadNetworksFile(extraPath, false)
			if err != nil {
				return nil, fmt.Errorf("could not load %s: %w", extraPath, err)
			}
			networks.merge(extra)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("could not stat %s: %w", extraPath, err)
		}
	}

	return networks, nil
}

func loadNetworksFile(path string, requireStorage bool) (*NetworksConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read networks file: %w", err)
	}
	return parseNetworksYAML(bytes, path, requireStorage)
}

func parseNetworksYAML(bytes []byte, source string, requireStorage bool) (*NetworksConfig, error) {
	var file networksConfigFile
	if err := yaml.Unmarshal(bytes, &file); err != nil {
		return nil, fmt.Errorf("could not unmarshal %s: %w", source, err)
	}
	if file.Version != networksFileVersion {
		return nil, fmt.Errorf("%s: unsupported version %d (expected %d)", source, file.Version, networksFileVersion)
	}
	if err := validateNetworks(file.Networks, source, requireStorage); err != nil {
		return nil, err
	}
	nc := &NetworksConfig{Networks: file.Networks}
	nc.rebuildLookup()
	return nc, nil
}

func validateNetworks(networks []*config.NetworkInfo, source string, requireStorage bool) error {
	seen := make(map[config.Network]struct{}, len(networks))
	defaults := 0
	for i, n := range networks {
		if n == nil {
			return fmt.Errorf("%s: network %d is empty", source, i)
		}
		if n.Name == "" || n.Name == string(config.Network_All) {
			return fmt.Errorf("%s: network %d has invalid name %q", source, i, n.Name)
		}
		id := n.ID()
		if _, dup := seen[id]; dup {
			return fmt.Errorf("%s: duplicate network name %q", source, n.Name)
		}
		seen[id] = struct{}{}
		if n.ChainID == 0 {
			return fmt.Errorf("%s: network %q is missing chainID", source, n.Name)
		}
		if n.Addresses.Storage != "" && !common.IsHexAddress(n.Addresses.Storage) {
			return fmt.Errorf("%s: network %q has invalid storage address %q", source, n.Name, n.Addresses.Storage)
		}
		if requireStorage && n.Addresses.Storage == "" {
			return fmt.Errorf("%s: network %q is missing a valid storage address", source, n.Name)
		}
		if n.Label == "" || n.Description == "" {
			return fmt.Errorf("%s: network %q is missing label or description", source, n.Name)
		}
		if n.BeaconNetwork == "" && n.CustomChainConfigDir == "" {
			return fmt.Errorf("%s: network %q needs beaconNetwork or customChainConfigDir", source, n.Name)
		}
		if n.ClientTagSet == "" {
			n.ClientTagSet = config.ClientTagSetTest
		}
		if n.ClientTagSet != config.ClientTagSetProduction && n.ClientTagSet != config.ClientTagSetTest {
			return fmt.Errorf("%s: network %q has invalid clientTagSet %q", source, n.Name, n.ClientTagSet)
		}
		if err := validateMevRelays(n.MevRelays, source, n.Name); err != nil {
			return err
		}
		if n.Default {
			defaults++
		}
	}
	if defaults > 1 {
		return fmt.Errorf("%s: multiple networks marked default", source)
	}
	if len(networks) == 0 {
		return fmt.Errorf("%s: no networks defined", source)
	}
	return nil
}

func validateMevRelays(relays map[string]string, source, networkName string) error {
	for id := range relays {
		if !knownMevRelayID(config.MevRelayID(id)) {
			return fmt.Errorf("%s: network %q has unknown mevRelays key %q", source, networkName, id)
		}
	}
	return nil
}

func knownMevRelayID(id config.MevRelayID) bool {
	switch id {
	case config.MevRelayID_Flashbots,
		config.MevRelayID_BloxrouteEthical,
		config.MevRelayID_BloxrouteRegulated,
		config.MevRelayID_Ultrasound,
		config.MevRelayID_UltrasoundFiltered,
		config.MevRelayID_Aestus,
		config.MevRelayID_TitanGlobal,
		config.MevRelayID_TitanRegional,
		config.MevRelayID_BTCSOfac:
		return true
	default:
		return false
	}
}

func hexAddresses(values []string) []common.Address {
	out := make([]common.Address, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		out = append(out, common.HexToAddress(v))
	}
	return out
}

func clientTagDefaults(networks *NetworksConfig, prod, test string) map[config.Network]interface{} {
	defaults := map[config.Network]interface{}{config.Network_All: test}
	if networks == nil {
		return defaults
	}
	for _, n := range networks.AllNetworks() {
		if n.ClientTagSet == config.ClientTagSetProduction {
			defaults[n.ID()] = prod
		} else {
			defaults[n.ID()] = test
		}
	}
	return defaults
}

func nethermindPruneThresholdDefaults(networks *NetworksConfig) map[config.Network]interface{} {
	defaults := map[config.Network]interface{}{config.Network_All: uint64(51200)}
	if networks == nil {
		return defaults
	}
	for _, n := range networks.AllNetworks() {
		if n.NethermindPruneThresholdMb != 0 {
			defaults[n.ID()] = n.NethermindPruneThresholdMb
		}
	}
	return defaults
}

func relayUrlMap(networks *NetworksConfig, id config.MevRelayID) config.UrlMap {
	urls := config.UrlMap{}
	if networks == nil {
		return urls
	}
	for _, n := range networks.AllNetworks() {
		if url := n.RelayURL(id); url != "" {
			urls[n.ID()] = url
		}
	}
	return urls
}
