package validator

import (
    "github.com/rocket-pool/smartnode/shared/utils/bls"
)


// Prysm keystore settings
const PRYSM_KEYSTORE_PATH string = "prysm"
const PRYSM_KEY_FILENAME_PREFIX string = "validatorprivatekey"


// Read Prysm keys from keystore
func readPrysmKeys(keystorePath string) (map[string]*bls.Key, error) {
	return nil, nil
}


// Write Prysm key to keystore
func writePrysmKey(keystorePath string, key *bls.Key) error {
	return nil
}

