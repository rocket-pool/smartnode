package validator

import (
	"fmt"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils"
	"golang.org/x/exp/maps"
	"strings"
)

type KeyRecoveryManager struct {
	manager        *ValidatorManager
	partialEnabled bool
	testOnly       bool
}

func NewKeyRecoveryManager(m *ValidatorManager, partialEnabled bool, testOnly bool) *KeyRecoveryManager {
	return &KeyRecoveryManager{
		manager:        m,
		partialEnabled: partialEnabled,
		testOnly:       testOnly,
	}
}

func (s *KeyRecoveryManager) RecoverMinipoolKeys() ([]beacon.ValidatorPubkey, map[beacon.ValidatorPubkey]error, error) {
	publicKeys, err := s.manager.GetMinipools()
	if err != nil {
		return []beacon.ValidatorPubkey{}, map[beacon.ValidatorPubkey]error{}, err
	}

	recoveredCustomPublicKeys, unrecoverableCustomPublicKeys, err := s.checkForAndRecoverCustomKeys(publicKeys)
	if err != nil && !s.partialEnabled {
		return maps.Keys(recoveredCustomPublicKeys), unrecoverableCustomPublicKeys, err
	}

	recoveredConventionalPublicKeys, unrecoveredPublicKeys := s.recoverConventionalKeys(publicKeys)

	var allRecoveredPublicKeys []beacon.ValidatorPubkey
	allRecoveredPublicKeys = append(allRecoveredPublicKeys, maps.Keys(recoveredCustomPublicKeys)...)
	allRecoveredPublicKeys = append(allRecoveredPublicKeys, recoveredConventionalPublicKeys...)

	for publicKey, err := range unrecoveredPublicKeys {
		unrecoverableCustomPublicKeys[publicKey] = err
	}

	return allRecoveredPublicKeys, unrecoveredPublicKeys, nil
}

func (s *KeyRecoveryManager) checkForAndRecoverCustomKeys(
	publicKeys map[beacon.ValidatorPubkey]bool,
) (map[beacon.ValidatorPubkey]bool, map[beacon.ValidatorPubkey]error, error) {

	recoveredKeys := make(map[beacon.ValidatorPubkey]bool)
	recoveryFailures := make(map[beacon.ValidatorPubkey]error)
	var passwords map[string]string

	keyFiles, err := s.manager.LoadFiles()
	if err != nil {
		return recoveredKeys, recoveryFailures, err
	}

	if len(keyFiles) == 0 {
		return recoveredKeys, recoveryFailures, nil
	}

	passwords, err = s.manager.LoadCustomKeyPasswords()
	if err != nil {
		return recoveredKeys, recoveryFailures, err
	}

	for _, file := range keyFiles {
		keystore, err := s.manager.ReadCustomKeystore(file)
		if err != nil {
			if s.partialEnabled {
				continue
			}
			return recoveredKeys, recoveryFailures, err
		}

		if _, exists := publicKeys[keystore.Pubkey]; !exists {
			err := fmt.Errorf("custom keystore for pubkey %s not found in minipool keyset", keystore.Pubkey.Hex())
			recoveryFailures[keystore.Pubkey] = err
			if s.partialEnabled {
				continue
			}
			return recoveredKeys, recoveryFailures, err
		}

		formattedPublicKey := strings.ToUpper(utils.RemovePrefix(keystore.Pubkey.Hex()))
		password, exists := passwords[formattedPublicKey]
		if !exists {
			err := fmt.Errorf("custom keystore for pubkey %s needs a password, but none was provided", keystore.Pubkey.Hex())
			recoveryFailures[keystore.Pubkey] = err
			if s.partialEnabled {
				continue
			}
			return recoveredKeys, recoveryFailures, err
		}

		privateKey, err := s.manager.DecryptCustomKeystore(keystore, password)
		if err != nil {
			err := fmt.Errorf("error recreating private key for validator %s: %w", keystore.Pubkey.Hex(), err)
			recoveryFailures[keystore.Pubkey] = err
			if s.partialEnabled {
				continue
			}
			return recoveredKeys, recoveryFailures, err
		}

		reconstructedPublicKey := beacon.ValidatorPubkey(privateKey.PublicKey().Marshal())
		if reconstructedPublicKey != keystore.Pubkey {
			err := fmt.Errorf("private keystore file %s claims to be for validator %s but it's for validator %s", file.Name(), keystore.Pubkey.Hex(), reconstructedPublicKey.Hex())
			recoveryFailures[keystore.Pubkey] = err
			if s.partialEnabled {
				continue
			}
			return recoveredKeys, recoveryFailures, err
		}

		if !s.testOnly {
			if err := s.manager.StoreValidatorKey(&privateKey, keystore.Path); err != nil {
				recoveryFailures[keystore.Pubkey] = err
				if s.partialEnabled {
					continue
				}
				return recoveredKeys, recoveryFailures, err
			}
		}
		recoveredKeys[reconstructedPublicKey] = true

		delete(publicKeys, keystore.Pubkey)
	}

	return recoveredKeys, recoveryFailures, nil
}

func (s *KeyRecoveryManager) recoverConventionalKeys(publicKeys map[beacon.ValidatorPubkey]bool) ([]beacon.ValidatorPubkey, map[beacon.ValidatorPubkey]error) {
	var recoveredPublicKeys []beacon.ValidatorPubkey
	unrecoverablePublicKeys := map[beacon.ValidatorPubkey]error{}

	bucketStart := uint64(0)
	for {
		if bucketStart >= bucketLimit {
			break
		}
		bucketEnd := bucketStart + bucketSize
		if bucketEnd > bucketLimit {
			bucketEnd = bucketLimit
		}

		keys, err := s.manager.GetValidatorKeys(bucketStart, bucketEnd-bucketStart)
		if err != nil {
			return recoveredPublicKeys, map[beacon.ValidatorPubkey]error{beacon.ValidatorPubkey{}: fmt.Errorf("error getting node's validator keys")}
		}

		for _, validatorKey := range keys {
			if exists := publicKeys[validatorKey.PublicKey]; exists {
				delete(publicKeys, validatorKey.PublicKey)
				if !s.testOnly {
					if err := s.manager.SaveValidatorKey(validatorKey); err != nil {
						unrecoverablePublicKeys[validatorKey.PublicKey] = err
						if s.partialEnabled {
							continue
						}
						return recoveredPublicKeys, unrecoverablePublicKeys
					}
				}
				recoveredPublicKeys = append(recoveredPublicKeys, validatorKey.PublicKey)

			} else {
				err := fmt.Errorf("keystore for pubkey %s not found in minipool keyset", validatorKey.PublicKey)
				unrecoverablePublicKeys[validatorKey.PublicKey] = err
				if !s.partialEnabled {
					return recoveredPublicKeys, unrecoverablePublicKeys
				}
			}
		}

		if len(publicKeys) == 0 {
			// All keys have been recovered.
			break
		}

		bucketStart = bucketEnd
	}

	return recoveredPublicKeys, unrecoverablePublicKeys
}
