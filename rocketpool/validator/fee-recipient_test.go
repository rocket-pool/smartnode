package validator

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/moby/moby/client"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

type validatorTestBeacon struct {
	beacon.Client
	clientType beacon.BeaconClientType
}

func (b validatorTestBeacon) GetClientType() (beacon.BeaconClientType, error) {
	return b.clientType, nil
}

type validatorTestTransport func(*http.Request) (*http.Response, error)

func (f validatorTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestValidatorDockerRequests(t *testing.T) {
	for _, tt := range []struct {
		name             string
		clientType       beacon.BeaconClientType
		restart          bool
		missingContainer bool
		listError        bool
		operationStatus  int
		operationBody    string
		wantError        string
	}{
		{name: "restart split process", clientType: beacon.SplitProcess, restart: true},
		{name: "restart single process", clientType: beacon.SingleProcess, restart: true},
		{name: "pause split process", clientType: beacon.SplitProcess},
		{name: "pause single process", clientType: beacon.SingleProcess},
		{name: "restart missing container", restart: true, missingContainer: true, wantError: "Validator container not found"},
		{name: "pause missing container", missingContainer: true, wantError: "Validator container test_validator not found"},
		{name: "list failure", listError: true, wantError: "Could not get docker containers: Error response from daemon: list failed"},
		{name: "restart failure", restart: true, operationStatus: http.StatusInternalServerError, operationBody: `{"message":"restart failed"}`, wantError: "Could not restart validator container: Error response from daemon: restart failed"},
		{name: "pause failure", operationStatus: http.StatusInternalServerError, operationBody: `{"message":"pause failed"}`, wantError: "Could not stop validator container test_validator: Error response from daemon: pause failed"},
		{name: "pause already stopped", operationStatus: http.StatusConflict, operationBody: `{"message":"Container selected is not running"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.NewRocketPoolConfig(t.TempDir(), false)
			if err != nil {
				t.Fatal(err)
			}
			cfg.Smartnode.ProjectName.Value = "test"
			containerName := "/test_validator"
			if tt.clientType == beacon.SingleProcess {
				containerName = "/test_eth2"
			}
			operation, query := "pause", ""
			if tt.restart {
				operation, query = "restart", "t=5"
			}
			requests := 0
			transport := validatorTestTransport(func(req *http.Request) (*http.Response, error) {
				requests++
				status, body := http.StatusNoContent, ""
				switch requests {
				case 1:
					if req.Method != http.MethodGet || req.URL.Path != "/v1.44/containers/json" || req.URL.RawQuery != "all=1" {
						t.Fatalf("unexpected container list request: %s %s", req.Method, req.URL)
					}
					status = http.StatusOK
					body = fmt.Sprintf(`[{"Id":"unrelated","Names":["/other_validator"]},{"Id":"selected","Names":[%q]}]`, containerName)
					if tt.missingContainer {
						body = `[{"Id":"unrelated","Names":["/other_validator"]}]`
					}
					if tt.listError {
						status, body = http.StatusInternalServerError, `{"message":"list failed"}`
					}
				case 2:
					if tt.missingContainer || tt.listError {
						t.Fatal("container operation must not run when lookup fails")
					}
					if req.Method != http.MethodPost || req.URL.Path != "/v1.44/containers/selected/"+operation || req.URL.RawQuery != query {
						t.Fatalf("unexpected container operation: %s %s", req.Method, req.URL)
					}
					if tt.operationStatus != 0 {
						status, body = tt.operationStatus, tt.operationBody
					}
				default:
					t.Fatalf("unexpected extra Docker request: %s %s", req.Method, req.URL)
				}
				return &http.Response{
					StatusCode: status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    req,
				}, nil
			})
			d, err := client.New(client.WithHost("http://docker.test"), client.WithAPIVersion("1.44"), client.WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := d.Close(); err != nil {
					t.Error(err)
				}
			})
			bc := validatorTestBeacon{clientType: tt.clientType}
			if tt.restart {
				err = RestartValidator(cfg, bc, nil, d)
			} else {
				err = StopValidator(cfg, bc, nil, d)
			}
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("error = %v, want %q", err, tt.wantError)
			}
			wantRequests := 2
			if tt.missingContainer || tt.listError {
				wantRequests = 1
			}
			if requests != wantRequests {
				t.Errorf("got %d Docker requests, want %d", requests, wantRequests)
			}
		})
	}
}
