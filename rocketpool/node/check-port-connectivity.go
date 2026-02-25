package node

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/alerting"
	"github.com/rocket-pool/smartnode/shared/services/config"
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
	{"208.67.222.222:53", "myip.opendns.com"},             // OpenDNS primary (IPv4)
	{"208.67.220.220:53", "myip.opendns.com"},             // OpenDNS secondary (IPv4)
	{"216.239.32.10:53", "o-o.myaddr.l.google.com"},       // Google ns1 (IPv4)
	{"[2620:119:35::35]:53", "myip.opendns.com"},          // OpenDNS primary (IPv6)
	{"[2620:119:53::53]:53", "myip.opendns.com"},          // OpenDNS secondary (IPv6)
	{"[2001:4860:4860::8888]:53", "o-o.myaddr.l.google.com"}, // Google ns1 (IPv6)
}

// Check port connectivity task
type checkPortConnectivity struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig

	// Track previous state to avoid flooding with repeated alerts
	wasEth1PortOpen  bool
	wasBeaconP2POpen bool
}

// Create check port connectivity task
func newCheckPortConnectivity(c *cli.Context, logger log.ColorLogger) (*checkPortConnectivity, error) {
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	return &checkPortConnectivity{
		c:   c,
		log: logger,
		cfg: cfg,
		// Assume ports are open at startup to avoid spurious alerts on first cycle
		wasEth1PortOpen:  true,
		wasBeaconP2POpen: true,
	}, nil
}

// Check whether the configured execution client P2P port and beacon chain P2P port are
// reachable from the internet. Sends an alert the first time either port is detected as closed.
func (t *checkPortConnectivity) run() error {
	if t.cfg.Alertmanager.EnableAlerting.Value != true {
		return nil
	}
	if t.cfg.Alertmanager.AlertEnabled_PortConnectivityCheck.Value != true {
		return nil
	}
	t.log.Println("Checking port connectivity...")

	publicIP, err := getPublicIP()
	if err != nil {
		return fmt.Errorf("error getting public IP for port connectivity check: %w", err)
	}

	eth1P2PPort := t.cfg.ExecutionCommon.P2pPort.Value.(uint16)
	eth1Open := isPortReachable(publicIP, eth1P2PPort)
	if eth1Open {
		if !t.wasEth1PortOpen {
			t.log.Printlnf("Execution client P2P port %d is now accessible from the internet.", eth1P2PPort)
		}
	} else {
		if t.wasEth1PortOpen {
			t.log.Printlnf("WARNING: Execution client P2P port %d is not accessible from the internet.", eth1P2PPort)
		}
		if err := alerting.AlertEth1P2PPortNotOpen(t.cfg, eth1P2PPort); err != nil {
			t.log.Printlnf("WARNING: Could not send Eth1P2PPortNotOpen alert: %s", err.Error())
		}
	}
	t.wasEth1PortOpen = eth1Open

	beaconP2PPort := t.cfg.ConsensusCommon.P2pPort.Value.(uint16)
	beaconOpen := isPortReachable(publicIP, beaconP2PPort)
	if beaconOpen {
		if !t.wasBeaconP2POpen {
			t.log.Printlnf("Beacon chain P2P port %d is now accessible from the internet.", beaconP2PPort)
		}
	} else {
		if t.wasBeaconP2POpen {
			t.log.Printlnf("WARNING: Beacon chain P2P port %d is not accessible from the internet.", beaconP2PPort)
		}
		if err := alerting.AlertBeaconP2PPortNotOpen(t.cfg, beaconP2PPort); err != nil {
			t.log.Printlnf("WARNING: Could not send BeaconP2PPortNotOpen alert: %s", err.Error())
		}
	}
	t.wasBeaconP2POpen = beaconOpen

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

// isPortReachable attempts a TCP connection to host:port and returns true if
// the connection succeeds within portCheckTimeout.
func isPortReachable(host string, port uint16) bool {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, portCheckTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
