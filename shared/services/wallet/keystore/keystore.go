package keystore

import (
    eth2types "github.com/wealdtech/go-eth2-types/v2"
)


// Validator keystore interface
type Keystore interface {
    StoreValidatorKey(key *eth2types.BLSPrivateKey, derivationPath string) error
}

