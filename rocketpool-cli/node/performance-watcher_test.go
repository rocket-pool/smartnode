package node

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/performancewatcher"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli/v3"
)

type watcherRoundTripper func(*http.Request) (*http.Response, error)

func (f watcherRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestWatcherValidatorDefaultNodeAndExplicitIndex(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.NewRocketPoolConfig(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	cfg.IsCLI = true
	cfg.Smartnode.DataPath.Value = filepath.Join(dir, "data")
	if err = cfg.Save(dir, rocketpool.SettingsFile); err != nil {
		t.Fatal(err)
	}
	previousDefaults, previousClient := rocketpool.Defaults, http.DefaultClient
	t.Cleanup(func() { rocketpool.SetDefaults(previousDefaults); http.DefaultClient = previousClient })
	rocketpool.SetDefaults(rocketpool.Globals{ConfigPath: dir})
	for _, tc := range []struct {
		args                 []string
		path, body, expected string
	}{
		{nil, "/api/node/performance-watcher/validators", `{"nodeAddress":"0x0000000000000000000000000000000000001234","validators":[]}`, "No validators for this node are currently tracked"},
		{[]string{"--json"}, "/api/node/performance-watcher/validators", `{"nodeAddress":"0x0000000000000000000000000000000000001234","validators":[{"validatorIndex":10},{"validatorIndex":12}]}`, `"validators"`},
		{[]string{"85453"}, "/api/node/performance-watcher/validator?validatorIndex=85453", `{"validator":{"validatorIndex":85453}}`, "Validator 85453"},
	} {
		requests := 0
		http.DefaultClient = &http.Client{Transport: watcherRoundTripper(func(r *http.Request) (*http.Response, error) {
			requests++
			if r.Method != http.MethodGet || r.URL.RequestURI() != tc.path {
				t.Fatalf("unexpected API request: %s %s", r.Method, r.URL.RequestURI())
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
		})}
		var output strings.Builder
		app := &cli.Command{Name: "rocketpool", Writer: &output, ErrWriter: io.Discard}
		RegisterCommands(app, "node", nil)
		err := app.Run(context.Background(), append([]string{"rocketpool", "node", "performance-watcher", "validator"}, tc.args...))
		if err != nil || requests != 1 || !strings.Contains(output.String(), tc.expected) {
			t.Fatalf("arguments %v: %s, %v (%d requests)", tc.args, output.String(), err, requests)
		}
		if len(tc.args) > 0 && tc.args[0] == "--json" {
			var r api.PerformanceWatcherNodeValidatorsResponse
			if err := json.Unmarshal([]byte(output.String()), &r); err != nil || len(r.Validators) != 2 || r.Validators[1].Index != 12 {
				t.Fatalf("invalid node validator JSON: %s (%v)", output.String(), err)
			}
		}
	}
}

func TestWatcherCandidatesRejectsValidatorArgument(t *testing.T) {
	app := &cli.Command{Name: "rocketpool", Writer: io.Discard, ErrWriter: io.Discard}
	RegisterCommands(app, "node", nil)
	err := app.Run(context.Background(), []string{"rocketpool", "node", "performance-watcher", "candidates", "85453"})
	if err == nil || !strings.Contains(err.Error(), "candidates takes no arguments") || !strings.Contains(err.Error(), "rocketpool node performance-watcher candidate <validator-index>") {
		t.Fatalf("expected an error explaining the single-validator command, got %v", err)
	}
}

func TestWatcherValidatorArguments(t *testing.T) {
	for _, args := range [][]string{{"abc"}, {"85453", "85454"}} {
		app := &cli.Command{Name: "rocketpool", Writer: io.Discard, ErrWriter: io.Discard}
		RegisterCommands(app, "node", nil)
		err := app.Run(context.Background(), append([]string{"rocketpool", "node", "performance-watcher", "validator"}, args...))
		if err == nil || !strings.Contains(err.Error(), "validator index") {
			t.Fatalf("invalid arguments %v: %v", args, err)
		}
	}
}

func TestWatcherNodeValidatorsOutput(t *testing.T) {
	r := api.PerformanceWatcherNodeValidatorsResponse{}
	if output := formatWatcherNodeValidators(r); !strings.Contains(output, "No validators for this node are currently tracked") {
		t.Fatal(output)
	}
	r.Validators = []performancewatcher.ValidatorPerformance{
		{Validator: performancewatcher.Validator{Index: 10, Kind: "minipool"}},
		{Validator: performancewatcher.Validator{Index: 12, Kind: "megapool"}},
	}
	output := formatWatcherNodeValidators(r)
	for _, expected := range []string{r.NodeAddress.Hex(), "2 tracked validators", "Validator 10 (minipool", "Validator 12 (megapool"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("missing %q in %s", expected, output)
		}
	}
}

func TestWatcherValidatorOutputDistinguishesCoverage(t *testing.T) {
	start, end, observed, rate := uint64(10), uint64(109), uint64(109), float64(100)
	v := &performancewatcher.ValidatorPerformance{
		Validator:  performancewatcher.Validator{Index: 85453},
		StartEpoch: &start, EndEpoch: &end, ExpectedEpochs: 100,
		ObservedEpochs: 1, TimelyEpochs: 1, UnknownEpochs: 99, CoveragePercent: 1,
		ObservedPerformancePercent: &rate, FirstObservedEpoch: &observed, LastObservedEpoch: &observed,
	}
	output := formatWatcherValidator(v)
	for _, expected := range []string{"Validator 85453", "Coverage: 1/100 epochs (1.00%); unknown: 99", "Observed performance: 100.00% (collected epochs only; unknown epochs excluded)"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("missing %q in %s", expected, output)
		}
	}
	v.ObservedPerformancePercent = nil
	if output = formatWatcherValidator(v); !strings.Contains(output, "unavailable (no collected observations)") {
		t.Fatal(output)
	}
}
