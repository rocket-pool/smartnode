package key_recovery_manager

import (
	"github.com/rocket-pool/node-manager-core/beacon"
)

type KeyRecoveryManager interface {
	RecoverMinipoolKeys() ([]beacon.ValidatorPubkey, map[beacon.ValidatorPubkey]error, error)
}

const (
	bucketSize  uint64 = 20
	bucketLimit uint64 = 2000
)
