package key_recovery_manager

import (
	"fmt"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"golang.org/x/exp/maps"
	"strings"
)

type PartialRecoveryManager struct {
	manager *validator.ValidatorManager
}

func NewPartialRecoveryManager(m *validator.ValidatorManager) *PartialRecoveryManager {
	return &PartialRecoveryManager{
		manager: m,
	}
}

func (p PartialRecoveryManager) RecoverMinipoolKeys() ([]beacon.ValidatorPubkey, map[beacon.ValidatorPubkey]error, error) {
	status, err := p.manager.GetWalletStatus()
	if err != nil {
		return []beacon.ValidatorPubkey{}, map[beacon.ValidatorPubkey]error{}, err
	}

	rpNode, mpMgr, err := p.manager.InitializeBindings(status)
	if err != nil {
		return []beacon.ValidatorPubkey{}, map[beacon.ValidatorPubkey]error{}, err
	}

	publicKeys, err := p.manager.GetMinipools(rpNode, mpMgr)
	if err != nil {
		return []beacon.ValidatorPubkey{}, map[beacon.ValidatorPubkey]error{}, err
	}

	recoveredCustomPublicKeys, unrecoverableCustomPublicKeys, _ := p.checkForAndRecoverCustomKeys(publicKeys)
	recoveredPublicKeys, unrecoverablePublicKeys := p.recoverConventionalKeys(publicKeys)

	allRecoveredPublicKeys := []beacon.ValidatorPubkey{}
	allRecoveredPublicKeys = append(allRecoveredPublicKeys, maps.Keys(recoveredCustomPublicKeys)...)
	allRecoveredPublicKeys = append(allRecoveredPublicKeys, recoveredPublicKeys...)

	for publicKey, err := range unrecoverablePublicKeys {
		unrecoverableCustomPublicKeys[publicKey] = err
	}

	return allRecoveredPublicKeys, unrecoverablePublicKeys, nil
}

func (p PartialRecoveryManager) checkForAndRecoverCustomKeys(publicKeys map[beacon.ValidatorPubkey]bool,
) (map[beacon.ValidatorPubkey]bool, map[beacon.ValidatorPubkey]error, error) {

	recoveredKeys := make(map[beacon.ValidatorPubkey]bool)
	recoveryFailures := make(map[beacon.ValidatorPubkey]error)
	var passwords map[string]string

	keyFiles, err := p.manager.LoadFiles()
	if err != nil {
		return recoveredKeys, recoveryFailures, err
	}

	if len(keyFiles) > 0 {
		passwords, err = p.manager.LoadCustomKeyPasswords()
		if err != nil {
			return recoveredKeys, recoveryFailures, err
		}

		for _, file := range keyFiles {
			keystore, err := p.manager.ReadCustomKeystore(file)
			if err != nil {
				continue
			}

			if _, exists := publicKeys[keystore.Pubkey]; !exists {
				err := fmt.Errorf("custom keystore for pubkey %s not found in minipool keyset", keystore.Pubkey.Hex())
				recoveryFailures[keystore.Pubkey] = err
				continue
			}

			formattedPublicKey := strings.ToUpper(utils.RemovePrefix(keystore.Pubkey.Hex()))
			password, exists := passwords[formattedPublicKey]
			if !exists {
				err := fmt.Errorf("custom keystore for pubkey %s needs a password, but none was provided", keystore.Pubkey.Hex())
				recoveryFailures[keystore.Pubkey] = err
				continue
			}

			privateKey, err := p.manager.DecryptCustomKeystore(keystore, password)
			if err != nil {
				err := fmt.Errorf("error recreating private key for validator %s: %w", keystore.Pubkey.Hex(), err)
				recoveryFailures[keystore.Pubkey] = err
				continue
			}

			reconstructedPublicKey := beacon.ValidatorPubkey(privateKey.PublicKey().Marshal())
			if reconstructedPublicKey != keystore.Pubkey {
				err := fmt.Errorf("private keystore file %s claims to be for validator %s but it's for validator %s", file.Name(), keystore.Pubkey.Hex(), reconstructedPublicKey.Hex())
				recoveryFailures[keystore.Pubkey] = err
				continue
			}

			if err := p.manager.StoreValidatorKey(&privateKey, keystore.Path); err != nil {
				recoveryFailures[reconstructedPublicKey] = fmt.Errorf("error storing private keystore for %s: %w", reconstructedPublicKey.Hex(), err)
			} else {
				recoveredKeys[reconstructedPublicKey] = true
			}

			delete(publicKeys, keystore.Pubkey)
		}
	}

	return recoveredKeys, recoveryFailures, nil
}

func (p *PartialRecoveryManager) recoverConventionalKeys(publicKeys map[beacon.ValidatorPubkey]bool) ([]beacon.ValidatorPubkey, map[beacon.ValidatorPubkey]error) {
	recoveredPublicKeys := []beacon.ValidatorPubkey{}
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

		keys, err := p.manager.GetValidatorKeys(bucketStart, bucketEnd-bucketStart)
		if err != nil {
			continue
		}

		for _, validatorKey := range keys {
			delete(publicKeys, validatorKey.PublicKey)
			if exists := publicKeys[validatorKey.PublicKey]; exists {
				if err := p.manager.SaveValidatorKey(validatorKey); err != nil {
					unrecoverablePublicKeys[validatorKey.PublicKey] = err
				} else {
					recoveredPublicKeys = append(recoveredPublicKeys, validatorKey.PublicKey)
				}
			} else {
				err := fmt.Errorf("keystore for pubkey %s not found in minipool keyset", validatorKey.PublicKey)
				unrecoverablePublicKeys[validatorKey.PublicKey] = err
				continue
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
