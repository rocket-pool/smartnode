package flashbots

// Flashbots relay URLs per chain ID, used as a fallback when no relay URL is
// passed to NewClient explicitly (the Smart Node passes cfg.GetFlashbotsRelayUrl()).
var FlashbotsUrlPerNetwork = map[uint64]string{
	1: "https://relay.flashbots.net",
}

const (
	JsonRpcParseError     = -32700
	JsonRpcInvalidRequest = -32600
	JsonRpcMethodNotFound = -32601
	JsonRpcInvalidParams  = -32602
	JsonRpcInternalError  = -32603
)

// AllBuilders is the list used by UseAllBuilders for broader inclusion chances on supported networks.
var AllBuilders = []string{
	"flashbots",
	"f1b.io",
	"rsync",
	"beaverbuild.org",
	"builder0x69",
	"Titan",
	"EigenPhi",
	"boba-builder",
	"Gambit Labs",
	"payload",
	"Loki",
	"BuildAI",
	"JetBuilder",
	"tbuilder",
	"penguinbuild",
	"bobthebuilder",
	"BTCS",
	"bloXroute",
	"Blockbeelder",
}
