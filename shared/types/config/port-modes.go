package config

import "fmt"

type RPCMode string

// Enum to describe the mode for the RPC port.
// Closed will not allow any connections to the RPC port.
// OpenLocalhost will allow connections from the same host.
// OpenExternal will allow connections from external hosts.
const (
	RPC_Closed        RPCMode = "closed"
	RPC_OpenLocalhost RPCMode = "localhost"
	RPC_OpenExternal  RPCMode = "external"
)

func (rpcMode RPCMode) String() string {
	return string(rpcMode)
}

func (rpcMode RPCMode) Open() bool {
	return rpcMode == RPC_OpenLocalhost || rpcMode == RPC_OpenExternal
}

func (rpcMode RPCMode) DockerPortMapping(port uint16) string {
	ports := fmt.Sprintf("%d:%d/tcp", port, port)

	if rpcMode == RPC_OpenExternal {
		return ports
	}

	if rpcMode == RPC_OpenLocalhost {
		return fmt.Sprintf("127.0.0.1:%s", ports)
	}

	return ""
}

func PortModes(warningOverride string) []ParameterOption {
	if warningOverride == "" {
		warningOverride = "Allow connections from external hosts. This is safe if you're running your node on your local network. If you're a VPS user, this would expose your node to the internet"
	}

	return []ParameterOption{{
		Name:        "Closed",
		Description: "Do not allow connections to the port",
		Value:       RPC_Closed,
	}, {
		Name:        "Open to Localhost",
		Description: "Allow connections from this host only",
		Value:       RPC_OpenLocalhost,
	}, {
		Name:        "Open to External hosts",
		Description: warningOverride,
		Value:       RPC_OpenExternal,
	}}
}

func RestrictedPortModes() []ParameterOption {

	return []ParameterOption{{
		Name:        "Closed",
		Description: "Do not allow connections to the port",
		Value:       RPC_Closed,
	}, {
		Name:        "Open to Localhost",
		Description: "Allow connections from this host only",
		Value:       RPC_OpenLocalhost,
	}}
}
