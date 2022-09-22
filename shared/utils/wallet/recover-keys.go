package wallet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	"gopkg.in/yaml.v2"
)

func RecoverMinipoolKeys(c *cli.Context, rp *rocketpool.RocketPool, address common.Address, w *wallet.Wallet, testOnly bool) ([]types.ValidatorPubkey, error) {

	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Get node's validating pubkeys
	pubkeys, err := minipool.GetNodeValidatingMinipoolPubkeys(rp, address, nil)
	if err != nil {
		return nil, err
	}
	pubkeyMap := map[types.ValidatorPubkey]bool{}
	for _, pubkey := range pubkeys {
		pubkeyMap[pubkey] = true
	}

	// Load custom validator keys
	customKeyDir := cfg.Smartnode.GetCustomKeyPath()
	info, err := os.Stat(customKeyDir)
	if !os.IsNotExist(err) && info.IsDir() {

		// Get the custom keystore files
		files, err := ioutil.ReadDir(customKeyDir)
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
			passwordFile := cfg.Smartnode.GetCustomKeyPasswordFilePath()
			fileBytes, err := ioutil.ReadFile(passwordFile)
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
				bytes, err := ioutil.ReadFile(filepath.Join(customKeyDir, file.Name()))
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
				formattedPubkey := strings.ToUpper(hexutils.RemovePrefix(keystore.Pubkey.Hex()))
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
				reconstructedPubkey := types.BytesToValidatorPubkey(privateKey.PublicKey().Marshal())
				if reconstructedPubkey != keystore.Pubkey {
					return nil, fmt.Errorf("private keystore file %s claims to be for validator %s but it's for validator %s", file.Name(), keystore.Pubkey.Hex(), reconstructedPubkey.Hex())
				}

				// Store the key
				if !testOnly {
					err = w.StoreValidatorKey(privateKey, keystore.Path)
					if err != nil {
						return nil, fmt.Errorf("error storing private keystore for %s: %w", reconstructedPubkey.Hex(), err)
					}
				}

				// Remove the pubkey from pending minipools to handle
				delete(pubkeyMap, reconstructedPubkey)
			}
		}
	}

	// Create an array of remaining keys that preserves the minipool order (required for the recovery offset to work)
	var remainingPubkeys []types.ValidatorPubkey

	if len(pubkeyMap) == len(pubkeys) {
		// If there were no custom keys, run the whole thing on the original set
		remainingPubkeys = pubkeys
	} else {
		// Go through each original key in order; if it's still in the pubkey map, add it
		remainingPubkeys = make([]types.ValidatorPubkey, 0, len(pubkeyMap))
		for _, pubkey := range pubkeys {
			_, exists := pubkeyMap[pubkey]
			if exists {
				remainingPubkeys = append(remainingPubkeys, pubkey)
			}
		}
	}

	// Recover remaining validator keys normally
	index := uint(0)
	for _, pubkey := range remainingPubkeys {
		if testOnly {
			index, err = w.TestRecoverValidatorKey(pubkey, index)
			index++
		} else {
			index, err = w.RecoverValidatorKey(pubkey, index)
			index++
		}
		if err != nil {
			return nil, err
		}
	}

	return pubkeys, nil

}
