package connectivity

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/alerting"
	cfg "github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

const (
	portCheckTimeout = 5 * time.Second
	dnsLookupTimeout = 5 * time.Second
)

// Well-known resolvers addressed by IP to
// avoid a bootstrap DNS lookup.
var publicIPResolvers = []struct {
	addr     string // host:port (IP literal)
	hostname string // special hostname that returns the caller's public IP
}{
	{"208.67.222.222:53", "myip.opendns.com"},       // OpenDNS primary
	{"208.67.220.220:53", "myip.opendns.com"},       // OpenDNS secondary
	{"216.239.32.10:53", "o-o.myaddr.l.google.com"}, // Google ns1
}

// Check port connectivity task
type CheckPortConnectivity struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *cfg.RocketPoolConfig

	// Track previous state to avoid flooding with repeated alerts
	wasEth1PortOpen  bool
	wasBeaconP2POpen bool

	// Function pointers for network and alerting actions (to facilitate testing)
	GetPublicIP                    func() (string, error)
	IsPortReachableNATReflection   func(string, uint16) bool
	IsPortReachableExternalService func(uint16) (bool, string, error)
	AlertEth1P2PPortNotOpen        func(*cfg.RocketPoolConfig, uint16) error
	AlertBeaconP2PPortNotOpen      func(*cfg.RocketPoolConfig, uint16) error
}

// Create check port connectivity task
func NewCheckPortConnectivity(c *cli.Context, config *cfg.RocketPoolConfig, logger log.ColorLogger) (*CheckPortConnectivity, error) {
	return &CheckPortConnectivity{
		c:   c,
		log: logger,
		cfg: config,
		// Assume ports are open at startup to avoid spurious alerts on first cycle
		wasEth1PortOpen:  true,
		wasBeaconP2POpen: true,

		// Default implementations
		GetPublicIP:                    getPublicIP,
		IsPortReachableNATReflection:   isPortReachableNATReflection,
		IsPortReachableExternalService: isPortReachableExternalService,
		AlertEth1P2PPortNotOpen:        alerting.AlertEth1P2PPortNotOpen,
		AlertBeaconP2PPortNotOpen:      alerting.AlertBeaconP2PPortNotOpen,
	}, nil
}

// Check whether the configured execution/consensus client P2P ports are
// reachable from the internet. Sends an alert the first time either port is detected as closed.
func (t *CheckPortConnectivity) Run() error {
	if t.cfg.Alertmanager.EnableAlerting.Value != true {
		return nil
	}
	if t.cfg.Alertmanager.AlertEnabled_PortConnectivityCheck.Value != true {
		return nil
	}
	t.log.Print("Checking port connectivity...")

	isLocalEc := t.cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local
	isLocalCc := t.cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local

	if !isLocalEc && !isLocalCc {
		return nil
	}

	ecOpen := false
	ccOpen := false
	ecP2PPort := t.cfg.ExecutionCommon.P2pPort.Value.(uint16)
	ccP2PPort := t.cfg.ConsensusCommon.P2pPort.Value.(uint16)
	publicIP, err := t.GetPublicIP()
	if err == nil {
		if isLocalEc {
			ecOpen = t.IsPortReachableNATReflection(publicIP, ecP2PPort)
		}
		if isLocalCc {
			ccOpen = t.IsPortReachableNATReflection(publicIP, ccP2PPort)
		}
	}

	if isLocalEc && !ecOpen {
		// Fallback to using an external service
		ecOpen, _, err = t.IsPortReachableExternalService(ecP2PPort)
		if err != nil {
			return fmt.Errorf("error checking port connectivity: %w", err)
		}
	}
	if isLocalEc {
		if ecOpen {
			t.log.Printf("Port %d is OPEN.", ecP2PPort)
			if !t.wasEth1PortOpen {
				t.log.Printlnf("Execution client P2P port %d is now accessible from the internet.", ecP2PPort)
			}
		} else {
			if t.wasEth1PortOpen {
				t.log.Printlnf("WARNING: Execution client P2P port %d is not accessible from the internet.", ecP2PPort)
			}
			if err := t.AlertEth1P2PPortNotOpen(t.cfg, ecP2PPort); err != nil {
				t.log.Printlnf("WARNING: Could not send Eth1P2PPortNotOpen alert: %s", err.Error())
			}
		}
		t.wasEth1PortOpen = ecOpen
	}

	if isLocalCc && !ccOpen {
		// Fallback to using an external service
		ccOpen, _, err = t.IsPortReachableExternalService(ccP2PPort)
		if err != nil {
			return fmt.Errorf("error checking port connectivity: %w", err)
		}
	}
	if isLocalCc {
		if ccOpen {
			t.log.Printf("Port %d is OPEN.", ccP2PPort)
			if !t.wasBeaconP2POpen {
				t.log.Printlnf("Consensus client P2P port %d is now accessible from the internet.", ccP2PPort)
			}
		} else {
			if t.wasBeaconP2POpen {
				t.log.Printlnf("WARNING: Consensus client P2P port %d is not accessible from the internet.", ccP2PPort)
			}
			if err := t.AlertBeaconP2PPortNotOpen(t.cfg, ccP2PPort); err != nil {
				t.log.Printlnf("WARNING: Could not send BeaconP2PPortNotOpen alert: %s", err.Error())
			}
		}
		t.wasBeaconP2POpen = ccOpen
	}

	return nil
}

// getPublicIP resolves the node's public IP by querying a well-known resolver for a
// special hostname that echoes back the caller's IP. Resolver addresses are hard-coded
// as IP literals to avoid a bootstrap DNS lookup. Falls back through the list on error.
func getPublicIP() (string, error) {
	var lastErr error
	for _, res := range publicIPResolvers {
		resolverAddr := res.addr
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "udp", resolverAddr)
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), dnsLookupTimeout)
		addrs, err := r.LookupHost(ctx, res.hostname)
		cancel()
		if err != nil {
			lastErr = err
			continue
		}
		if len(addrs) > 0 {
			return addrs[0], nil
		}
	}
	return "", fmt.Errorf("all public IP resolvers failed; last error: %w", lastErr)
}

// isPortReachableNATReflection attempts a TCP connection to host:port and returns true if
// the connection succeeds within portCheckTimeout.
func isPortReachableNATReflection(host string, port uint16) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, portCheckTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// isPortReachableExternalService checks if a given port is open using canyouseeme.org.
func isPortReachableExternalService(port uint16) (bool, string, error) {
	// Endpoint is the main page, as form submits to itself
	endpoint := "https://canyouseeme.org/"

	// Prepare form data
	data := url.Values{}
	data.Set("port", strconv.FormatUint(uint64(port), 10))
	// Some forms include the submit button; include it to mimic
	data.Set("submit", "Check")

	// Create request
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3") // Mimic browser

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}
	responseText := string(body)

	if strings.Contains(responseText, "Success:") {
		return true, extractResult(responseText), nil
	} else if strings.Contains(responseText, "Error:") {
		return false, extractResult(responseText), nil
	}

	return false, responseText, fmt.Errorf("unable to parse result; unexpected response")
}

// extractResult pulls the relevant message from the HTML (simple substring search)
func extractResult(html string) string {
	// Find the start of the result message
	startIdx := strings.Index(html, "<div id=\"result\">")
	if startIdx == -1 {
		return "" // Fallback if not found
	}
	startIdx += len("<div id=\"result\">")

	endIdx := strings.Index(html[startIdx:], "</div>")
	if endIdx == -1 {
		return html[startIdx:]
	}

	return strings.TrimSpace(html[startIdx : startIdx+endIdx])
}
