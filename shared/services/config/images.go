package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool/assets"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	ImagesMainnetFile = "mainnet.env"
	ImagesTestnetFile = "testnet.env"
	ImagesDevnetFile  = "devnet.env"
	ComposeMainFile   = "compose.yml"

	ImageSmartnode = "RP_IMAGE_SMARTNODE"

	ImageGeth       = "RP_IMAGE_GETH"
	ImageNethermind = "RP_IMAGE_NETHERMIND"
	ImageBesu       = "RP_IMAGE_BESU"
	ImageReth       = "RP_IMAGE_RETH"
	ImageErigon     = "RP_IMAGE_ERIGON"

	ImageLighthouse = "RP_IMAGE_LIGHTHOUSE"
	ImageLodestar   = "RP_IMAGE_LODESTAR"
	ImageNimbusBn   = "RP_IMAGE_NIMBUS_BN"
	ImageNimbusVc   = "RP_IMAGE_NIMBUS_VC"
	ImageTeku       = "RP_IMAGE_TEKU"
	ImagePrysmBn    = "RP_IMAGE_PRYSM_BN"
	ImagePrysmVc    = "RP_IMAGE_PRYSM_VC"

	ImageMevBoost    = "RP_IMAGE_MEV_BOOST"
	ImageCommitBoost = "RP_IMAGE_COMMIT_BOOST"

	ImagePrometheus   = "RP_IMAGE_PROMETHEUS"
	ImageExporter     = "RP_IMAGE_EXPORTER"
	ImageAlertmanager = "RP_IMAGE_ALERTMANAGER"
	ImageGrafana      = "RP_IMAGE_GRAFANA"
	ImageGWW          = "RP_IMAGE_GWW"
	ImageCurl         = "RP_IMAGE_CURL"
	ImageAlpine       = "RP_IMAGE_ALPINE"

	ECImageTagDefault            = "EC_IMAGE_TAG_DEFAULT"
	ECImageTagOverride           = "EC_IMAGE_TAG_OVERRIDE"
	BNImageTagDefault            = "BN_IMAGE_TAG_DEFAULT"
	BNImageTagOverride           = "BN_IMAGE_TAG_OVERRIDE"
	VCImageTagDefault            = "VC_IMAGE_TAG_DEFAULT"
	VCImageTagOverride           = "VC_IMAGE_TAG_OVERRIDE"
	SmartnodeImageTagDefault     = "SMARTNODE_IMAGE_TAG_DEFAULT"
	SmartnodeImageTagOverride    = "SMARTNODE_IMAGE_TAG_OVERRIDE"
	MevBoostImageTagDefault      = "MEV_BOOST_IMAGE_TAG_DEFAULT"
	MevBoostImageTagOverride     = "MEV_BOOST_IMAGE_TAG_OVERRIDE"
	CommitBoostImageTagDefault   = "COMMIT_BOOST_IMAGE_TAG_DEFAULT"
	CommitBoostImageTagOverride  = "COMMIT_BOOST_IMAGE_TAG_OVERRIDE"
	PrometheusImageTagDefault    = "PROMETHEUS_IMAGE_TAG_DEFAULT"
	PrometheusImageTagOverride   = "PROMETHEUS_IMAGE_TAG_OVERRIDE"
	GrafanaImageTagDefault       = "GRAFANA_IMAGE_TAG_DEFAULT"
	GrafanaImageTagOverride      = "GRAFANA_IMAGE_TAG_OVERRIDE"
	ExporterImageTagDefault      = "EXPORTER_IMAGE_TAG_DEFAULT"
	ExporterImageTagOverride     = "EXPORTER_IMAGE_TAG_OVERRIDE"
	AlertmanagerImageTagDefault  = "ALERTMANAGER_IMAGE_TAG_DEFAULT"
	AlertmanagerImageTagOverride = "ALERTMANAGER_IMAGE_TAG_OVERRIDE"
	GWWImageTagDefault           = "GWW_IMAGE_TAG_DEFAULT"
	GWWImageTagOverride          = "GWW_IMAGE_TAG_OVERRIDE"
	CurlImageTagDefault          = "CURL_IMAGE_TAG_DEFAULT"
	CurlImageTagOverride         = "CURL_IMAGE_TAG_OVERRIDE"
)

func ImageTagRef(overrideKey, defaultKey string) string {
	return "${" + overrideKey + ":-${" + defaultKey + "}}"
}

// requiredImageKeys must be present in mainnet.env. Network overlays may omit keys.
var requiredImageKeys = []string{
	ImageSmartnode,
	ImageGeth, ImageNethermind, ImageBesu, ImageReth, ImageErigon,
	ImageLighthouse, ImageLodestar, ImageNimbusBn, ImageNimbusVc, ImageTeku, ImagePrysmBn, ImagePrysmVc,
	ImageMevBoost, ImageCommitBoost,
	ImagePrometheus, ImageExporter, ImageAlertmanager, ImageGrafana,
	ImageGWW, ImageCurl, ImageAlpine,
}

// ImageCatalog is the parsed official (or overlay) image env file.
type ImageCatalog struct {
	values map[string]string
}

func (c *ImageCatalog) Get(key string) (string, bool) {
	if c == nil {
		return "", false
	}
	v, ok := c.values[key]
	return v, ok
}

func (c *ImageCatalog) Must(key string) string {
	v, ok := c.Get(key)
	if !ok || v == "" {
		panic("missing image catalog key " + key)
	}
	return v
}

func (c *ImageCatalog) Set(key, value string) {
	if c == nil {
		return
	}
	if c.values == nil {
		c.values = map[string]string{}
	}
	c.values[key] = value
}

func (c *ImageCatalog) Map() map[string]string {
	if c == nil {
		return nil
	}
	out := make(map[string]string, len(c.values))
	for k, v := range c.values {
		out[k] = v
	}
	return out
}

func EnvRef(key string) string {
	return "${" + key + "}"
}

// ParseEnvFile parses KEY=VALUE lines. Comments and blank lines are ignored.
// Quoted values have a single pair of matching quotes stripped.
func ParseEnvFile(data []byte) (map[string]string, error) {
	out := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected KEY=VALUE", lineNo)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNo)
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		out[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func LoadEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseEnvFile(data)
}

func catalogFileName(network config.Network) string {
	if network == "" || network == config.Network_Unknown {
		return ""
	}
	return string(network) + ".env"
}

func loadCatalogBytes(name string, data []byte, requireAll bool) (*ImageCatalog, error) {
	values, err := ParseEnvFile(data)
	if err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", name, err)
	}
	if requireAll {
		if err := validateImageCatalog(name, values); err != nil {
			return nil, err
		}
	}
	return &ImageCatalog{values: values}, nil
}

func loadCatalogFromDisk(rpDir, name string, requireAll bool) (*ImageCatalog, error) {
	path := filepath.Join(rpDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadCatalogBytes(name, data, requireAll)
}

// LoadImageCatalogs loads mainnet.env (full catalog) and optional per-network
// overlays ({network}.env). Overlay keys replace mainnet; missing keys inherit.
func LoadImageCatalogs(rpDir string, networks *NetworksConfig) (mainnet *ImageCatalog, overlays map[config.Network]*ImageCatalog, err error) {
	mainnet, err = loadCatalogBytes(ImagesMainnetFile, assets.ImagesMainnetEnv(), true)
	if err != nil {
		return nil, nil, err
	}
	if rpDir != "" {
		disk, err := loadCatalogFromDisk(rpDir, ImagesMainnetFile, true)
		if err == nil {
			mainnet = disk
		} else if !os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("could not load %s: %w", filepath.Join(rpDir, ImagesMainnetFile), err)
		}
	}

	overlays = map[config.Network]*ImageCatalog{}
	if networks == nil {
		return mainnet, overlays, nil
	}
	for _, n := range networks.AllNetworks() {
		id := n.ID()
		if id == "mainnet" {
			continue
		}
		name := catalogFileName(id)
		var overlay *ImageCatalog
		if rpDir != "" {
			disk, err := loadCatalogFromDisk(rpDir, name, false)
			if err == nil {
				overlay = disk
			} else if !os.IsNotExist(err) {
				return nil, nil, fmt.Errorf("could not load %s: %w", filepath.Join(rpDir, name), err)
			}
		}
		if overlay == nil {
			if data, ok := assets.EmbeddedNetworkEnv(name); ok {
				parsed, err := loadCatalogBytes(name, data, false)
				if err != nil {
					return nil, nil, err
				}
				overlay = parsed
			}
		}
		if overlay != nil {
			overlays[id] = overlay
		}
	}
	return mainnet, overlays, nil
}

func validateImageCatalog(name string, values map[string]string) error {
	var missing []string
	for _, key := range requiredImageKeys {
		if v, ok := values[key]; !ok || v == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s is missing required keys: %s", name, strings.Join(missing, ", "))
	}
	return nil
}

// UpdateEnvFile sets keys in an existing env file in place, preserving comments and other lines.
func UpdateEnvFile(path string, updates map[string]string) error {
	if len(updates) == 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	remaining := make(map[string]string, len(updates))
	for k, v := range updates {
		remaining[k] = v
	}
	var b strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			k, _, ok := strings.Cut(trimmed, "=")
			k = strings.TrimSpace(k)
			if ok {
				if value, found := remaining[k]; found {
					b.WriteString(k)
					b.WriteByte('=')
					b.WriteString(value)
					b.WriteByte('\n')
					delete(remaining, k)
					continue
				}
			}
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	keys := make([]string, 0, len(remaining))
	for k := range remaining {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(remaining[k])
		b.WriteByte('\n')
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// WriteEnvFile writes KEY=VALUE lines (sorted).
func WriteEnvFile(path string, values map[string]string) error {
	var b strings.Builder
	// Stable-ish: iterate a sorted key list
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(values[k])
		b.WriteByte('\n')
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (cfg *RocketPoolConfig) imageMainnet(key string) string {
	return cfg.imagesMainnet.Must(key)
}

func (cfg *RocketPoolConfig) imageForNetwork(key string, network config.Network) string {
	if overlay := cfg.imagesByNetwork[network]; overlay != nil {
		if v, ok := overlay.Get(key); ok && v != "" {
			return v
		}
	}
	return cfg.imagesMainnet.Must(key)
}

func (cfg *RocketPoolConfig) imageTagDefaults(key string) map[config.Network]interface{} {
	defaults := map[config.Network]interface{}{config.Network_All: cfg.imageMainnet(key)}
	if cfg.networks == nil {
		return defaults
	}
	for _, n := range cfg.networks.AllNetworks() {
		defaults[n.ID()] = cfg.imageForNetwork(key, n.ID())
	}
	return defaults
}

func (cfg *RocketPoolConfig) overlayFileName() string {
	info := cfg.GetNetworkInfo()
	if info == nil || info.Name == "" || info.Name == "mainnet" {
		return ""
	}
	return catalogFileName(info.ID())
}

func (cfg *RocketPoolConfig) IsTestnet() bool {
	info := cfg.GetNetworkInfo()
	return info != nil && string(info.ID()) == "testnet"
}

func (cfg *RocketPoolConfig) IsDevnet() bool {
	info := cfg.GetNetworkInfo()
	return info != nil && string(info.ID()) == "devnet"
}

// TestnetOnly returns "" on testnet so a compose.tmpl line is active, otherwise "#".
func (cfg *RocketPoolConfig) TestnetOnly() string {
	if cfg.IsTestnet() {
		return ""
	}
	return "#"
}

// DevnetOnly returns "" on devnet so a compose.tmpl line is active, otherwise "#".
func (cfg *RocketPoolConfig) DevnetOnly() string {
	if cfg.IsDevnet() {
		return ""
	}
	return "#"
}

// ImagesEnvFiles is the on-disk catalog list (mainnet.env, then {network}.env).
// Compose interpolates runtime/image-tags.env, which is generated from these.
func (cfg *RocketPoolConfig) OverlayNetworkNames() []string {
	if cfg.networks == nil {
		return nil
	}
	var names []string
	for _, n := range cfg.networks.AllNetworks() {
		id := string(n.ID())
		if id == "" || id == "mainnet" || id == "testnet" || id == "devnet" {
			continue
		}
		names = append(names, id)
	}
	sort.Strings(names)
	return names
}

func (cfg *RocketPoolConfig) ImagesEnvFiles() []string {
	files := []string{ImagesMainnetFile}
	if name := cfg.overlayFileName(); name != "" {
		path := filepath.Join(cfg.RocketPoolDirectory, name)
		if cfg.RocketPoolDirectory == "" {
			if _, ok := assets.EmbeddedNetworkEnv(name); ok {
				files = append(files, name)
			}
			return files
		}
		if _, err := os.Stat(path); err == nil {
			files = append(files, name)
		} else if _, ok := assets.EmbeddedNetworkEnv(name); ok {
			files = append(files, name)
		}
	}
	return files
}

func (cfg *RocketPoolConfig) selectedECCatalogKey() (string, error) {
	if !cfg.ExecutionClientLocal() {
		return "", fmt.Errorf("Execution client is external, there is no container tag")
	}
	switch cfg.ExecutionClient.Value.(config.ExecutionClient) {
	case config.ExecutionClient_Geth:
		return ImageGeth, nil
	case config.ExecutionClient_Nethermind:
		return ImageNethermind, nil
	case config.ExecutionClient_Besu:
		return ImageBesu, nil
	case config.ExecutionClient_Reth:
		return ImageReth, nil
	case config.ExecutionClient_Erigon:
		return ImageErigon, nil
	}
	return "", fmt.Errorf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value.(config.ExecutionClient)))
}

func (cfg *RocketPoolConfig) selectedECTagParam() (*config.Parameter, error) {
	if !cfg.ExecutionClientLocal() {
		return nil, fmt.Errorf("Execution client is external, there is no container tag")
	}
	switch cfg.ExecutionClient.Value.(config.ExecutionClient) {
	case config.ExecutionClient_Geth:
		return &cfg.Geth.ContainerTag, nil
	case config.ExecutionClient_Nethermind:
		return &cfg.Nethermind.ContainerTag, nil
	case config.ExecutionClient_Besu:
		return &cfg.Besu.ContainerTag, nil
	case config.ExecutionClient_Reth:
		return &cfg.Reth.ContainerTag, nil
	case config.ExecutionClient_Erigon:
		return &cfg.Erigon.ContainerTag, nil
	}
	return nil, fmt.Errorf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value.(config.ExecutionClient)))
}

// GetECImageEnvRef is ${EC_IMAGE_TAG_OVERRIDE:-${EC_IMAGE_TAG_DEFAULT}}.
func (cfg *RocketPoolConfig) GetECImageEnvRef() (string, error) {
	if _, err := cfg.selectedECCatalogKey(); err != nil {
		return "", err
	}
	return ImageTagRef(ECImageTagOverride, ECImageTagDefault), nil
}

func (cfg *RocketPoolConfig) selectedBNCatalogKey() (string, error) {
	cCfg, err := cfg.GetSelectedConsensusClientConfig()
	if err != nil {
		return "", err
	}
	switch cCfg.(type) {
	case *LighthouseConfig:
		return ImageLighthouse, nil
	case *LodestarConfig:
		return ImageLodestar, nil
	case *NimbusConfig:
		return ImageNimbusBn, nil
	case *PrysmConfig:
		return ImagePrysmBn, nil
	case *TekuConfig:
		return ImageTeku, nil
	default:
		return "", fmt.Errorf("unknown consensus client config %T", cCfg)
	}
}

// GetBeaconImageEnvRef is ${BN_IMAGE_TAG_OVERRIDE:-${BN_IMAGE_TAG_DEFAULT}}.
func (cfg *RocketPoolConfig) GetBeaconImageEnvRef() (string, error) {
	if _, err := cfg.selectedBNCatalogKey(); err != nil {
		return "", err
	}
	return ImageTagRef(BNImageTagOverride, BNImageTagDefault), nil
}

func (cfg *RocketPoolConfig) selectedVCCatalogKey() (string, error) {
	mode := cfg.ConsensusClientMode.Value.(config.Mode)
	if mode == config.Mode_Local {
		cCfg, err := cfg.GetSelectedConsensusClientConfig()
		if err != nil {
			return "", err
		}
		switch cCfg.(type) {
		case *LighthouseConfig:
			return ImageLighthouse, nil
		case *LodestarConfig:
			return ImageLodestar, nil
		case *NimbusConfig:
			return ImageNimbusVc, nil
		case *PrysmConfig:
			return ImagePrysmVc, nil
		case *TekuConfig:
			return ImageTeku, nil
		default:
			return "", fmt.Errorf("unknown consensus client config %T", cCfg)
		}
	}
	client := cfg.ExternalConsensusClient.Value.(config.ConsensusClient)
	switch client {
	case config.ConsensusClient_Lighthouse:
		return ImageLighthouse, nil
	case config.ConsensusClient_Lodestar:
		return ImageLodestar, nil
	case config.ConsensusClient_Nimbus:
		return ImageNimbusVc, nil
	case config.ConsensusClient_Prysm:
		return ImagePrysmVc, nil
	case config.ConsensusClient_Teku:
		return ImageTeku, nil
	default:
		return "", fmt.Errorf("unknown external consensus client [%v]", client)
	}
}

func (cfg *RocketPoolConfig) selectedVCTagParam() *config.Parameter {
	mode := cfg.ConsensusClientMode.Value.(config.Mode)
	if mode == config.Mode_Local {
		switch cfg.ConsensusClient.Value.(config.ConsensusClient) {
		case config.ConsensusClient_Lighthouse:
			return &cfg.Lighthouse.ContainerTag
		case config.ConsensusClient_Lodestar:
			return &cfg.Lodestar.ContainerTag
		case config.ConsensusClient_Nimbus:
			return &cfg.Nimbus.VcContainerTag
		case config.ConsensusClient_Prysm:
			return &cfg.Prysm.VcContainerTag
		case config.ConsensusClient_Teku:
			return &cfg.Teku.ContainerTag
		}
		return nil
	}
	switch cfg.ExternalConsensusClient.Value.(config.ConsensusClient) {
	case config.ConsensusClient_Lighthouse:
		return &cfg.ExternalLighthouse.ContainerTag
	case config.ConsensusClient_Lodestar:
		return &cfg.ExternalLodestar.ContainerTag
	case config.ConsensusClient_Nimbus:
		return &cfg.ExternalNimbus.ContainerTag
	case config.ConsensusClient_Prysm:
		return &cfg.ExternalPrysm.ContainerTag
	case config.ConsensusClient_Teku:
		return &cfg.ExternalTeku.ContainerTag
	}
	return nil
}

func (cfg *RocketPoolConfig) selectedBNTagParam() *config.Parameter {
	if cfg.ConsensusClientMode.Value.(config.Mode) != config.Mode_Local {
		return nil
	}
	switch cfg.ConsensusClient.Value.(config.ConsensusClient) {
	case config.ConsensusClient_Lighthouse:
		return &cfg.Lighthouse.ContainerTag
	case config.ConsensusClient_Lodestar:
		return &cfg.Lodestar.ContainerTag
	case config.ConsensusClient_Nimbus:
		return &cfg.Nimbus.BnContainerTag
	case config.ConsensusClient_Prysm:
		return &cfg.Prysm.BnContainerTag
	case config.ConsensusClient_Teku:
		return &cfg.Teku.ContainerTag
	}
	return nil
}

// GetVCImageEnvRef is ${VC_IMAGE_TAG_OVERRIDE:-${VC_IMAGE_TAG_DEFAULT}}.
func (cfg *RocketPoolConfig) GetVCImageEnvRef() (string, error) {
	if _, err := cfg.selectedVCCatalogKey(); err != nil {
		return "", err
	}
	return ImageTagRef(VCImageTagOverride, VCImageTagDefault), nil
}

func (cfg *RocketPoolConfig) GetMevBoostImageEnvRef() string {
	return ImageTagRef(MevBoostImageTagOverride, MevBoostImageTagDefault)
}

func (cfg *RocketPoolConfig) GetCommitBoostImageEnvRef() string {
	return ImageTagRef(CommitBoostImageTagOverride, CommitBoostImageTagDefault)
}

func (cfg *RocketPoolConfig) tuiOverride(param *config.Parameter) (string, bool) {
	if param == nil || param.Value == nil {
		return "", false
	}
	value := fmt.Sprint(param.Value)
	if value == "" {
		return "", false
	}
	def, err := param.GetDefault(cfg.GetNetwork())
	if err != nil || value == fmt.Sprint(def) {
		return "", false
	}
	return value, true
}

// ComposeImageDefaults maps service-level *_IMAGE_TAG_DEFAULT vars to the
// catalog pin for the selected client/network.
func (cfg *RocketPoolConfig) ComposeImageDefaults() map[string]string {
	out := map[string]string{
		SmartnodeImageTagDefault:    cfg.ResolvedImage(ImageSmartnode),
		PrometheusImageTagDefault:   cfg.ResolvedImage(ImagePrometheus),
		GrafanaImageTagDefault:      cfg.ResolvedImage(ImageGrafana),
		ExporterImageTagDefault:     cfg.ResolvedImage(ImageExporter),
		AlertmanagerImageTagDefault: cfg.ResolvedImage(ImageAlertmanager),
		GWWImageTagDefault:          cfg.ResolvedImage(ImageGWW),
		CurlImageTagDefault:         cfg.ResolvedImage(ImageCurl),
		MevBoostImageTagDefault:     cfg.ResolvedImage(ImageMevBoost),
		CommitBoostImageTagDefault:  cfg.ResolvedImage(ImageCommitBoost),
	}
	if key, err := cfg.selectedECCatalogKey(); err == nil {
		out[ECImageTagDefault] = cfg.ResolvedImage(key)
	}
	if key, err := cfg.selectedBNCatalogKey(); err == nil {
		out[BNImageTagDefault] = cfg.ResolvedImage(key)
	}
	if key, err := cfg.selectedVCCatalogKey(); err == nil {
		out[VCImageTagDefault] = cfg.ResolvedImage(key)
	}
	return out
}

// ComposeEnvOverrides returns *_IMAGE_TAG_OVERRIDE vars for TUI values that
// differ from the catalog pin. Unset override vars fall back to *_DEFAULT.
func (cfg *RocketPoolConfig) ComposeEnvOverrides() map[string]string {
	out := map[string]string{}
	if param, err := cfg.selectedECTagParam(); err == nil {
		if v, ok := cfg.tuiOverride(param); ok {
			out[ECImageTagOverride] = v
		}
	}
	if v, ok := cfg.tuiOverride(cfg.selectedBNTagParam()); ok {
		out[BNImageTagOverride] = v
	}
	if v, ok := cfg.tuiOverride(cfg.selectedVCTagParam()); ok {
		out[VCImageTagOverride] = v
	}
	if v, ok := cfg.tuiOverride(&cfg.MevBoost.ContainerTag); ok {
		out[MevBoostImageTagOverride] = v
	}
	if v, ok := cfg.tuiOverride(&cfg.CommitBoost.ContainerTag); ok {
		out[CommitBoostImageTagOverride] = v
	}
	if v, ok := cfg.tuiOverride(&cfg.Prometheus.ContainerTag); ok {
		out[PrometheusImageTagOverride] = v
	}
	if v, ok := cfg.tuiOverride(&cfg.Grafana.ContainerTag); ok {
		out[GrafanaImageTagOverride] = v
	}
	if v, ok := cfg.tuiOverride(&cfg.Exporter.ContainerTag); ok {
		out[ExporterImageTagOverride] = v
	}
	if v, ok := cfg.tuiOverride(&cfg.Alertmanager.ContainerTag); ok {
		out[AlertmanagerImageTagOverride] = v
	}
	gww := cfg.GraffitiWallWriter.GetConfig().GetParameters()
	for _, p := range gww {
		if p.ID == "containerTag" {
			if v, ok := cfg.tuiOverride(p); ok {
				out[GWWImageTagOverride] = v
			}
			break
		}
	}
	return out
}

func validateEnvAssignment(key, value string) error {
	if strings.ContainsAny(key, " \t\n\r") {
		return fmt.Errorf("image env name %q contains whitespace", key)
	}
	if strings.ContainsAny(value, " \t\n\r") {
		return fmt.Errorf("image env %s value %q contains whitespace", key, value)
	}
	return nil
}

// ComposeImageEnv is the single env map Compose interpolates: service
// *_DEFAULT pins (mainnet + network overlay) plus *_OVERRIDE when the TUI
// customized a tag.
func (cfg *RocketPoolConfig) ComposeImageEnv() (map[string]string, error) {
	out := cfg.ComposeImageDefaults()
	for key, value := range cfg.ComposeEnvOverrides() {
		out[key] = value
	}
	for key, value := range out {
		if err := validateEnvAssignment(key, value); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// ComposeEnvAssignments returns KEY=value strings for Compose interpolation:
// *_IMAGE_TAG_DEFAULT from the catalog and *_IMAGE_TAG_OVERRIDE when the TUI
// customized a tag. Keys are sorted so command logs and diffs are reproducible.
func (cfg *RocketPoolConfig) ComposeEnvAssignments() ([]string, error) {
	env, err := cfg.ComposeImageEnv()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key+"="+env[key])
	}
	return out, nil
}

func (cfg *RocketPoolConfig) resolveParamImage(param *config.Parameter, key string) string {
	if param == nil {
		return ""
	}
	value := fmt.Sprint(param.Value)
	if cfg.Smartnode != nil && cfg.Smartnode.Network.Value != nil {
		network := cfg.Smartnode.Network.Value.(config.Network)
		if def, err := param.GetDefault(network); err == nil && value != fmt.Sprint(def) {
			return value
		}
	}
	if v := cfg.ResolvedImage(key); v != "" {
		return v
	}
	return value
}

func (cfg *RocketPoolConfig) imageFromEnvRef(ref, tuiValue string) string {
	key := strings.TrimSuffix(strings.TrimPrefix(ref, "${"), "}")
	official := cfg.ResolvedImage(key)
	if tuiValue != "" && official != "" && tuiValue != official {
		return tuiValue
	}
	if official != "" {
		return official
	}
	return tuiValue
}

func (cfg *RocketPoolConfig) ResolvedImage(key string) string {
	network := config.Network_Unknown
	if cfg.Smartnode != nil && cfg.Smartnode.Network.Value != nil {
		network = cfg.Smartnode.Network.Value.(config.Network)
	}
	if overlay := cfg.imagesByNetwork[network]; overlay != nil {
		if v, ok := overlay.Get(key); ok && v != "" {
			return v
		}
	}
	if cfg.imagesMainnet != nil {
		if v, ok := cfg.imagesMainnet.Get(key); ok && v != "" {
			return v
		}
	}
	return ""
}
