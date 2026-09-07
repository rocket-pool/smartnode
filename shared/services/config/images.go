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
	ImagesEnvFile = "images.env"

	ImageSmartnode = "RP_IMAGE_SMARTNODE"

	ImageGethProd       = "RP_IMAGE_GETH_PROD"
	ImageGethTest       = "RP_IMAGE_GETH_TEST"
	ImageNethermindProd = "RP_IMAGE_NETHERMIND_PROD"
	ImageNethermindTest = "RP_IMAGE_NETHERMIND_TEST"
	ImageBesuProd       = "RP_IMAGE_BESU_PROD"
	ImageBesuTest       = "RP_IMAGE_BESU_TEST"
	ImageRethProd       = "RP_IMAGE_RETH_PROD"
	ImageRethTest       = "RP_IMAGE_RETH_TEST"
	ImageErigonProd     = "RP_IMAGE_ERIGON_PROD"
	ImageErigonTest     = "RP_IMAGE_ERIGON_TEST"

	ImageLighthouseProd = "RP_IMAGE_LIGHTHOUSE_PROD"
	ImageLighthouseTest = "RP_IMAGE_LIGHTHOUSE_TEST"
	ImageLodestarProd   = "RP_IMAGE_LODESTAR_PROD"
	ImageLodestarTest   = "RP_IMAGE_LODESTAR_TEST"
	ImageNimbusBnProd   = "RP_IMAGE_NIMBUS_BN_PROD"
	ImageNimbusBnTest   = "RP_IMAGE_NIMBUS_BN_TEST"
	ImageNimbusVcProd   = "RP_IMAGE_NIMBUS_VC_PROD"
	ImageNimbusVcTest   = "RP_IMAGE_NIMBUS_VC_TEST"
	ImageTekuProd       = "RP_IMAGE_TEKU_PROD"
	ImageTekuTest       = "RP_IMAGE_TEKU_TEST"
	ImagePrysmBnProd    = "RP_IMAGE_PRYSM_BN_PROD"
	ImagePrysmBnTest    = "RP_IMAGE_PRYSM_BN_TEST"
	ImagePrysmVcProd    = "RP_IMAGE_PRYSM_VC_PROD"
	ImagePrysmVcTest    = "RP_IMAGE_PRYSM_VC_TEST"

	ImageMevBoostProd    = "RP_IMAGE_MEV_BOOST_PROD"
	ImageMevBoostTest    = "RP_IMAGE_MEV_BOOST_TEST"
	ImageCommitBoostProd = "RP_IMAGE_COMMIT_BOOST_PROD"
	ImageCommitBoostTest = "RP_IMAGE_COMMIT_BOOST_TEST"

	ImagePrometheus   = "RP_IMAGE_PROMETHEUS"
	ImageExporter     = "RP_IMAGE_EXPORTER"
	ImageAlertmanager = "RP_IMAGE_ALERTMANAGER"
	ImageGrafana      = "RP_IMAGE_GRAFANA"
	ImageGWW          = "RP_IMAGE_GWW"
	ImageCurl         = "RP_IMAGE_CURL"
	ImageAlpine       = "RP_IMAGE_ALPINE"
)

// requiredImageKeys must be present in the official catalog.
var requiredImageKeys = []string{
	ImageSmartnode,
	ImageGethProd, ImageGethTest,
	ImageNethermindProd, ImageNethermindTest,
	ImageBesuProd, ImageBesuTest,
	ImageRethProd, ImageRethTest,
	ImageErigonProd, ImageErigonTest,
	ImageLighthouseProd, ImageLighthouseTest,
	ImageLodestarProd, ImageLodestarTest,
	ImageNimbusBnProd, ImageNimbusBnTest,
	ImageNimbusVcProd, ImageNimbusVcTest,
	ImageTekuProd, ImageTekuTest,
	ImagePrysmBnProd, ImagePrysmBnTest,
	ImagePrysmVcProd, ImagePrysmVcTest,
	ImageMevBoostProd, ImageMevBoostTest,
	ImageCommitBoostProd, ImageCommitBoostTest,
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

// LoadImages loads the official catalog: embed, then on-disk images.env if present.
func LoadImages(rpDir string) (*ImageCatalog, error) {
	values, err := ParseEnvFile(assets.ImagesEnv())
	if err != nil {
		return nil, fmt.Errorf("could not parse embedded %s: %w", ImagesEnvFile, err)
	}
	if rpDir != "" {
		diskPath := filepath.Join(rpDir, ImagesEnvFile)
		if _, err := os.Stat(diskPath); err == nil {
			diskValues, err := LoadEnvFile(diskPath)
			if err != nil {
				return nil, fmt.Errorf("could not load %s: %w", diskPath, err)
			}
			values = diskValues
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("could not stat %s: %w", diskPath, err)
		}
	}
	if err := validateImageCatalog(values); err != nil {
		return nil, err
	}
	return &ImageCatalog{values: values}, nil
}

func validateImageCatalog(values map[string]string) error {
	var missing []string
	for _, key := range requiredImageKeys {
		if v, ok := values[key]; !ok || v == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s is missing required keys: %s", ImagesEnvFile, strings.Join(missing, ", "))
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

func (cfg *RocketPoolConfig) imageDefault(key string) string {
	return cfg.images.Must(key)
}

func (cfg *RocketPoolConfig) clientImageKey(prodKey, testKey string) string {
	info := cfg.GetNetworkInfo()
	if info != nil && info.ClientTagSet == config.ClientTagSetProduction {
		return prodKey
	}
	return testKey
}

func (cfg *RocketPoolConfig) LoadedImages() *ImageCatalog {
	return cfg.images
}

func (cfg *RocketPoolConfig) ImageEnvPath() string {
	return filepath.Join(cfg.RocketPoolDirectory, ImagesEnvFile)
}

// GetECImageEnvRef returns the compose interpolation for the selected execution client.
func (cfg *RocketPoolConfig) GetECImageEnvRef() (string, error) {
	if !cfg.ExecutionClientLocal() {
		return "", fmt.Errorf("Execution client is external, there is no container tag")
	}
	switch cfg.ExecutionClient.Value.(config.ExecutionClient) {
	case config.ExecutionClient_Geth:
		return EnvRef(cfg.clientImageKey(ImageGethProd, ImageGethTest)), nil
	case config.ExecutionClient_Nethermind:
		return EnvRef(cfg.clientImageKey(ImageNethermindProd, ImageNethermindTest)), nil
	case config.ExecutionClient_Besu:
		return EnvRef(cfg.clientImageKey(ImageBesuProd, ImageBesuTest)), nil
	case config.ExecutionClient_Reth:
		return EnvRef(cfg.clientImageKey(ImageRethProd, ImageRethTest)), nil
	case config.ExecutionClient_Erigon:
		return EnvRef(cfg.clientImageKey(ImageErigonProd, ImageErigonTest)), nil
	}
	return "", fmt.Errorf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value.(config.ExecutionClient)))
}

// GetBeaconImageEnvRef returns the compose interpolation for the selected beacon client.
func (cfg *RocketPoolConfig) GetBeaconImageEnvRef() (string, error) {
	cCfg, err := cfg.GetSelectedConsensusClientConfig()
	if err != nil {
		return "", err
	}
	switch cCfg.(type) {
	case *LighthouseConfig:
		return EnvRef(cfg.clientImageKey(ImageLighthouseProd, ImageLighthouseTest)), nil
	case *LodestarConfig:
		return EnvRef(cfg.clientImageKey(ImageLodestarProd, ImageLodestarTest)), nil
	case *NimbusConfig:
		return EnvRef(cfg.clientImageKey(ImageNimbusBnProd, ImageNimbusBnTest)), nil
	case *PrysmConfig:
		return EnvRef(cfg.clientImageKey(ImagePrysmBnProd, ImagePrysmBnTest)), nil
	case *TekuConfig:
		return EnvRef(cfg.clientImageKey(ImageTekuProd, ImageTekuTest)), nil
	default:
		return "", fmt.Errorf("unknown consensus client config %T", cCfg)
	}
}

// GetVCImageEnvRef returns the compose interpolation for the selected validator client.
func (cfg *RocketPoolConfig) GetVCImageEnvRef() (string, error) {
	mode := cfg.ConsensusClientMode.Value.(config.Mode)
	if mode == config.Mode_Local {
		cCfg, err := cfg.GetSelectedConsensusClientConfig()
		if err != nil {
			return "", err
		}
		switch cCfg.(type) {
		case *LighthouseConfig:
			return EnvRef(cfg.clientImageKey(ImageLighthouseProd, ImageLighthouseTest)), nil
		case *LodestarConfig:
			return EnvRef(cfg.clientImageKey(ImageLodestarProd, ImageLodestarTest)), nil
		case *NimbusConfig:
			return EnvRef(cfg.clientImageKey(ImageNimbusVcProd, ImageNimbusVcTest)), nil
		case *PrysmConfig:
			return EnvRef(cfg.clientImageKey(ImagePrysmVcProd, ImagePrysmVcTest)), nil
		case *TekuConfig:
			return EnvRef(cfg.clientImageKey(ImageTekuProd, ImageTekuTest)), nil
		default:
			return "", fmt.Errorf("unknown consensus client config %T", cCfg)
		}
	}
	client := cfg.ExternalConsensusClient.Value.(config.ConsensusClient)
	switch client {
	case config.ConsensusClient_Lighthouse:
		return EnvRef(cfg.clientImageKey(ImageLighthouseProd, ImageLighthouseTest)), nil
	case config.ConsensusClient_Lodestar:
		return EnvRef(cfg.clientImageKey(ImageLodestarProd, ImageLodestarTest)), nil
	case config.ConsensusClient_Nimbus:
		return EnvRef(cfg.clientImageKey(ImageNimbusVcProd, ImageNimbusVcTest)), nil
	case config.ConsensusClient_Prysm:
		return EnvRef(cfg.clientImageKey(ImagePrysmVcProd, ImagePrysmVcTest)), nil
	case config.ConsensusClient_Teku:
		return EnvRef(cfg.clientImageKey(ImageTekuProd, ImageTekuTest)), nil
	default:
		return "", fmt.Errorf("unknown external consensus client [%v]", client)
	}
}

func (cfg *RocketPoolConfig) GetMevBoostImageEnvRef() string {
	return EnvRef(cfg.clientImageKey(ImageMevBoostProd, ImageMevBoostTest))
}

func (cfg *RocketPoolConfig) GetCommitBoostImageEnvRef() string {
	return EnvRef(cfg.clientImageKey(ImageCommitBoostProd, ImageCommitBoostTest))
}

type imageOverrideParam struct {
	param   *config.Parameter
	prodKey string
	testKey string
}

func (cfg *RocketPoolConfig) imageOverrideParams() []imageOverrideParam {
	gwwTag := cfg.GraffitiWallWriter.GetConfig().GetParameters()
	var gwwContainer *config.Parameter
	for _, p := range gwwTag {
		if p.ID == "containerTag" {
			gwwContainer = p
			break
		}
	}
	params := []imageOverrideParam{
		{&cfg.Geth.ContainerTag, ImageGethProd, ImageGethTest},
		{&cfg.Nethermind.ContainerTag, ImageNethermindProd, ImageNethermindTest},
		{&cfg.Besu.ContainerTag, ImageBesuProd, ImageBesuTest},
		{&cfg.Reth.ContainerTag, ImageRethProd, ImageRethTest},
		{&cfg.Erigon.ContainerTag, ImageErigonProd, ImageErigonTest},
		{&cfg.Lighthouse.ContainerTag, ImageLighthouseProd, ImageLighthouseTest},
		{&cfg.Lodestar.ContainerTag, ImageLodestarProd, ImageLodestarTest},
		{&cfg.Nimbus.BnContainerTag, ImageNimbusBnProd, ImageNimbusBnTest},
		{&cfg.Nimbus.VcContainerTag, ImageNimbusVcProd, ImageNimbusVcTest},
		{&cfg.Prysm.BnContainerTag, ImagePrysmBnProd, ImagePrysmBnTest},
		{&cfg.Prysm.VcContainerTag, ImagePrysmVcProd, ImagePrysmVcTest},
		{&cfg.Teku.ContainerTag, ImageTekuProd, ImageTekuTest},
		{&cfg.ExternalLighthouse.ContainerTag, ImageLighthouseProd, ImageLighthouseTest},
		{&cfg.ExternalLodestar.ContainerTag, ImageLodestarProd, ImageLodestarTest},
		{&cfg.ExternalNimbus.ContainerTag, ImageNimbusVcProd, ImageNimbusVcTest},
		{&cfg.ExternalPrysm.ContainerTag, ImagePrysmVcProd, ImagePrysmVcTest},
		{&cfg.ExternalTeku.ContainerTag, ImageTekuProd, ImageTekuTest},
		{&cfg.MevBoost.ContainerTag, ImageMevBoostProd, ImageMevBoostTest},
		{&cfg.CommitBoost.ContainerTag, ImageCommitBoostProd, ImageCommitBoostTest},
		{&cfg.Prometheus.ContainerTag, ImagePrometheus, ""},
		{&cfg.Grafana.ContainerTag, ImageGrafana, ""},
		{&cfg.Exporter.ContainerTag, ImageExporter, ""},
		{&cfg.Alertmanager.ContainerTag, ImageAlertmanager, ""},
	}
	if gwwContainer != nil {
		params = append(params, imageOverrideParam{gwwContainer, ImageGWW, ""})
	}
	return params
}

func (cfg *RocketPoolConfig) skipPersistingImageTags() {
	for _, item := range cfg.imageOverrideParams() {
		item.param.SkipUserSettings = true
	}
}

// tuiImageValues maps the current TUI container-tag values onto images.env keys
// for the selected network (_PROD vs _TEST).
func (cfg *RocketPoolConfig) tuiImageValues() map[string]string {
	out := map[string]string{}
	for _, item := range cfg.imageOverrideParams() {
		if item.param.Value == nil {
			continue
		}
		value := fmt.Sprint(item.param.Value)
		if value == "" {
			continue
		}
		key := item.prodKey
		if item.testKey != "" {
			key = cfg.clientImageKey(item.prodKey, item.testKey)
		}
		out[key] = value
	}
	return out
}

// SaveImagesEnv writes current TUI container tags into images.env.
// A later `rocketpool service install` replaces that file with the official catalog.
func (cfg *RocketPoolConfig) SaveImagesEnv() error {
	if cfg.RocketPoolDirectory == "" {
		return nil
	}
	path := cfg.ImageEnvPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	updates := cfg.tuiImageValues()
	if err := UpdateEnvFile(path, updates); err != nil {
		return fmt.Errorf("could not update %s: %w", path, err)
	}
	network := cfg.GetNetwork()
	for _, item := range cfg.imageOverrideParams() {
		if item.param.Value == nil {
			continue
		}
		value := fmt.Sprint(item.param.Value)
		if value == "" {
			continue
		}
		key := item.prodKey
		if item.testKey != "" {
			key = cfg.clientImageKey(item.prodKey, item.testKey)
			item.param.Default[network] = value
		} else {
			item.param.Default[config.Network_All] = value
		}
		cfg.images.Set(key, value)
	}
	return nil
}

func (cfg *RocketPoolConfig) resolveParamImage(param *config.Parameter, prodKey, testKey string) string {
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
	key := prodKey
	if testKey != "" {
		key = cfg.clientImageKey(prodKey, testKey)
	}
	if v := cfg.ResolvedImage(key); v != "" {
		return v
	}
	return value
}

func (cfg *RocketPoolConfig) imageFromEnvRef(ref, tuiValue string) string {
	key := strings.TrimSuffix(strings.TrimPrefix(ref, "${"), "}")
	if cfg.images != nil {
		if catalog, ok := cfg.images.Get(key); ok && tuiValue != "" && tuiValue != catalog {
			return tuiValue
		}
	}
	if v := cfg.ResolvedImage(key); v != "" {
		return v
	}
	return tuiValue
}

func (cfg *RocketPoolConfig) ResolvedImage(key string) string {
	if cfg.images != nil {
		if v, ok := cfg.images.Get(key); ok && v != "" {
			return v
		}
	}
	return ""
}
