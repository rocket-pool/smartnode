package key_recovery_manager

import (
	"fmt"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"golang.org/x/exp/maps"
	"strings"
)

type DryRunKeyRecoveryManager struct {
	manager *validator.ValidatorManager
}

func NewDryRunKeyRecoveryManager(m *validator.ValidatorManager) *DryRunKeyRecoveryManager {
	return &DryRunKeyRecoveryManager{
		manager: m,
	}
}

func (d *DryRunKeyRecoveryManager) RecoverMinipoolKeys() ([]beacon.ValidatorPubkey, map[beacon.ValidatorPubkey]error, error) {
	status, err := d.manager.GetWalletStatus()
	if err != nil {
		return []beacon.ValidatorPubkey{}, map[beacon.ValidatorPubkey]error{}, err
	}

	rpNode, mpMgr, err := d.manager.InitializeBindings(status)
	if err != nil {
		return []beacon.ValidatorPubkey{}, map[beacon.ValidatorPubkey]error{}, err
	}

	publicKeys, err := d.manager.GetMinipools(rpNode, mpMgr)
	if err != nil {
		return []beacon.ValidatorPubkey{}, map[beacon.ValidatorPubkey]error{}, err
	}

	recoveredCustomPublicKeys, unrecoverableCustomPublicKeys, _ := d.checkForAndRecoverCustomKeys(publicKeys)
	recoveredPublicKeys, unrecoverablePublicKeys := d.recoverConventionalKeys(publicKeys)

	allRecoveredPublicKeys := []beacon.ValidatorPubkey{}
	allRecoveredPublicKeys = append(allRecoveredPublicKeys, maps.Keys(recoveredCustomPublicKeys)...)
	allRecoveredPublicKeys = append(allRecoveredPublicKeys, recoveredPublicKeys...)

	for publicKey, err := range unrecoverablePublicKeys {
		unrecoverableCustomPublicKeys[publicKey] = err
	}

	return allRecoveredPublicKeys, unrecoverablePublicKeys, nil
}

func (d *DryRunKeyRecoveryManager) checkForAndRecoverCustomKeys(publicKeys map[beacon.ValidatorPubkey]bool) (map[beacon.ValidatorPubkey]bool, map[beacon.ValidatorPubkey]error, error) {

	recoveredKeys := make(map[beacon.ValidatorPubkey]bool)
	recoveryFailures := make(map[beacon.ValidatorPubkey]error)
	var passwords map[string]string

	keyFiles, err := d.manager.LoadFiles()
	if err != nil {
		return recoveredKeys, recoveryFailures, err
	}

	if len(keyFiles) > 0 {
		passwords, err = d.manager.LoadCustomKeyPasswords()
		if err != nil {
			return recoveredKeys, recoveryFailures, err
		}

		for _, file := range keyFiles {
			keystore, err := d.manager.ReadCustomKeystore(file)
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

			privateKey, err := d.manager.DecryptCustomKeystore(keystore, password)
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

			recoveredKeys[reconstructedPublicKey] = true
		}
	}

	return recoveredKeys, recoveryFailures, nil
}

func (d *DryRunKeyRecoveryManager) recoverConventionalKeys(publicKeys map[beacon.ValidatorPubkey]bool) ([]beacon.ValidatorPubkey, map[beacon.ValidatorPubkey]error) {
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

		keys, err := d.manager.GetValidatorKeys(bucketStart, bucketEnd-bucketStart)
		if err != nil {
			continue
		}

		for _, validatorKey := range keys {
			if exists := publicKeys[validatorKey.PublicKey]; exists {
				delete(publicKeys, validatorKey.PublicKey)
				recoveredPublicKeys = append(recoveredPublicKeys, validatorKey.PublicKey)
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
