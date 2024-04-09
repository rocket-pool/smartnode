package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/node-manager-core/wallet"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	"gopkg.in/yaml.v3"
)

const (
	bucketSize      uint64 = 20
	bucketLimit     uint64 = 2000
	pubkeyBatchSize int    = 500
)

func (m *ValidatorManager) RecoverMinipoolKeys(testOnly bool) ([]beacon.ValidatorPubkey, error) {
	status, err := m.wallet.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("error getting wallet status: %w", err)
	}
	if wallet.IsWalletReady(status) {
		return nil, fmt.Errorf("cannot recover minipool keys without a wallet keystore and matching password loaded")
	}

	// Create the bindings
	address := status.Wallet.WalletAddress
	node, err := node.NewNode(m.rp, address)
	if err != nil {
		return nil, fmt.Errorf("error creating node binding: %w", err)
	}
	mpMgr, err := minipool.NewMinipoolManager(m.rp)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool manager binding: %w", err)
	}

	// Get node's validating pubkey count
	err = m.queryMgr.Query(nil, nil, node.ValidatingMinipoolCount)
	if err != nil {
		return nil, fmt.Errorf("error getting node's validating minipool count: %w", err)
	}

	// Get the minipools
	addresses, err := node.GetValidatingMinipoolAddresses(node.ValidatingMinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node's validating minipool addresses: %w", err)
	}
	mps, err := mpMgr.CreateMinipoolsFromAddresses(addresses, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating bindings for node's validating minipools: %w", err)
	}

	// Get the pubkeys
	err = m.queryMgr.BatchQuery(len(mps), pubkeyBatchSize, func(mc *batch.MultiCaller, i int) error {
		mps[i].Common().Pubkey.AddToQuery(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node's validating minipool pubkeys: %w", err)
	}

	// Remove zero pubkeys
	zeroPubkey := beacon.ValidatorPubkey{}
	pubkeys := []beacon.ValidatorPubkey{}
	pubkeyMap := map[beacon.ValidatorPubkey]bool{}
	for _, mp := range mps {
		pubkey := mp.Common().Pubkey.Get()
		if pubkey != zeroPubkey {
			pubkeyMap[pubkey] = true
			pubkeys = append(pubkeys, pubkey)
		}
	}

	pubkeyMap, err = m.CheckForAndRecoverCustomMinipoolKeys(pubkeyMap, testOnly)
	if err != nil {
		return nil, fmt.Errorf("error checking for or recovering custom validator keys: %w", err)
	}

	// Recover conventionally generated keys
	bucketStart := uint64(0)
	for {
		if bucketStart >= bucketLimit {
			return nil, fmt.Errorf("attempt limit exceeded (%d keys)", bucketLimit)
		}
		bucketEnd := bucketStart + bucketSize
		if bucketEnd > bucketLimit {
			bucketEnd = bucketLimit
		}

		// Get the keys for this bucket
		keys, err := m.GetValidatorKeys(bucketStart, bucketEnd-bucketStart)
		if err != nil {
			return nil, err
		}
		for _, validatorKey := range keys {
			_, exists := pubkeyMap[validatorKey.PublicKey]
			if exists {
				// Found one!
				delete(pubkeyMap, validatorKey.PublicKey)
				if !testOnly {
					err := m.SaveValidatorKey(validatorKey)
					if err != nil {
						return nil, fmt.Errorf("error recovering validator keys: %w", err)
					}
				}
			}
		}

		if len(pubkeyMap) == 0 {
			// All keys recovered!
			break
		}

		// Run another iteration with the next bucket
		bucketStart = bucketEnd
	}

	return pubkeys, nil
}

func (m *ValidatorManager) CheckForAndRecoverCustomMinipoolKeys(pubkeyMap map[beacon.ValidatorPubkey]bool, testOnly bool) (map[beacon.ValidatorPubkey]bool, error) {
	// Load custom validator keys
	customKeyDir := m.cfg.GetCustomKeyPath()
	info, err := os.Stat(customKeyDir)
	if !os.IsNotExist(err) && info.IsDir() {

		// Get the custom keystore files
		files, err := os.ReadDir(customKeyDir)
		if err != nil {
			return nil, fmt.Errorf("error enumerating custom keystores: %w", err)
		}

		// Initialize the BLS library
		err = eth2types.InitBLS()
		if err != nil {
			return nil, fmt.Errorf("error initializing BLS: %w", err)
		}

		if len(files) > 0 {
			// Deserialize the password file
			passwordFile := m.cfg.GetCustomKeyPasswordFilePath()
			fileBytes, err := os.ReadFile(passwordFile)
			if err != nil {
				return nil, fmt.Errorf("%d custom keystores were found but the password file could not be loaded: %w", len(files), err)
			}
			passwords := map[string]string{}
			err = yaml.Unmarshal(fileBytes, &passwords)
			if err != nil {
				return nil, fmt.Errorf("error unmarshalling custom keystore password file: %w", err)
			}

			// Process every custom key
			for _, file := range files {
				// Read the file
				bytes, err := os.ReadFile(filepath.Join(customKeyDir, file.Name()))
				if err != nil {
					return nil, fmt.Errorf("error reading custom keystore %s: %w", file.Name(), err)
				}

				// Deserialize it
				keystore := api.ValidatorKeystore{}
				err = json.Unmarshal(bytes, &keystore)
				if err != nil {
					return nil, fmt.Errorf("error deserializing custom keystore %s: %w", file.Name(), err)
				}

				// Check if it's one of the pubkeys for the minipool
				_, exists := pubkeyMap[keystore.Pubkey]
				if !exists {
					// This pubkey isn't for any of this node's minipools so ignore it
					continue
				}

				// Get the password for it
				formattedPubkey := strings.ToUpper(utils.RemovePrefix(keystore.Pubkey.Hex()))
				password, exists := passwords[formattedPubkey]
				if !exists {
					return nil, fmt.Errorf("custom keystore for pubkey %s needs a password, but none was provided", keystore.Pubkey.Hex())
				}

				// Get the encryption function it uses
				kdf, exists := keystore.Crypto["kdf"]
				if !exists {
					return nil, fmt.Errorf("error processing custom keystore %s: \"crypto\" didn't contain a subkey named \"kdf\"", file.Name())
				}
				kdfMap := kdf.(map[string]interface{})
				function, exists := kdfMap["function"]
				if !exists {
					return nil, fmt.Errorf("error processing custom keystore %s: \"crypto.kdf\" didn't contain a subkey named \"function\"", file.Name())
				}
				functionString := function.(string)

				// Decrypt the private key
				encryptor := eth2ks.New(eth2ks.WithCipher(functionString))
				decryptedKey, err := encryptor.Decrypt(keystore.Crypto, password)
				if err != nil {
					return nil, fmt.Errorf("error decrypting keystore for validator %s: %w", keystore.Pubkey.Hex(), err)
				}
				privateKey, err := eth2types.BLSPrivateKeyFromBytes(decryptedKey)
				if err != nil {
					return nil, fmt.Errorf("error recreating private key for validator %s: %w", keystore.Pubkey.Hex(), err)
				}

				// Verify the private key matches the public key
				reconstructedPubkey := beacon.ValidatorPubkey(privateKey.PublicKey().Marshal())
				if reconstructedPubkey != keystore.Pubkey {
					return nil, fmt.Errorf("private keystore file %s claims to be for validator %s but it's for validator %s", file.Name(), keystore.Pubkey.Hex(), reconstructedPubkey.Hex())
				}

				// Store the key
				if !testOnly {
					err = m.StoreValidatorKey(privateKey, keystore.Path)
					if err != nil {
						return nil, fmt.Errorf("error storing private keystore for %s: %w", reconstructedPubkey.Hex(), err)
					}
				}

				// Remove the pubkey from pending minipools to handle
				delete(pubkeyMap, reconstructedPubkey)
			}
		}
	}

	return pubkeyMap, nil
}
