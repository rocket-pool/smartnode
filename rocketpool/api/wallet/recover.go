package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	"gopkg.in/yaml.v2"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"

	hexutils "github.com/rocket-pool/smartnode/shared/hex"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	findIterations uint = 100000
	bucketSize     uint = 20
	bucketLimit    uint = 2000
)

func recoverWalletWithParams(c *cli.Command, mnemonic string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint) (*api.RecoverWalletResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	// Chain clients are only needed to discover and filter validator pubkeys
	var rp *rocketpool.RocketPool
	var bc beacon.Client
	if !skipValidatorKeyRecovery {
		if err := services.RequireRocketStorage(c); err != nil {
			return nil, err
		}
		rp, err = services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
		bc, err = services.GetBeaconClient(c)
		if err != nil {
			return nil, err
		}
	}

	// Response
	response := api.RecoverWalletResponse{}

	// Check if wallet is already initialized
	if w.IsInitialized() {
		return nil, errors.New("the wallet is already initialized")
	}

	// Get the derivation path
	path := derivationPath
	switch path {
	case "":
		path = wallet.DefaultNodeKeyPath
	case "ledgerLive":
		path = wallet.LedgerLiveNodeKeyPath
	case "mew":
		path = wallet.MyEtherWalletNodeKeyPath
	}

	// Recover wallet
	if err := w.Recover(path, walletIndex, mnemonic); err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	if !skipValidatorKeyRecovery {
		response.ValidatorKeys, err = recoverNodeKeys(c, rp, bc, nodeAccount.Address, w, false)
		if err != nil {
			return nil, err
		}
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func searchAndRecoverWalletWithParams(c *cli.Command, mnemonic string, address common.Address, skipValidatorKeyRecovery bool) (*api.SearchAndRecoverWalletResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	// Chain clients are only needed to discover and filter validator pubkeys
	var rp *rocketpool.RocketPool
	var bc beacon.Client
	if !skipValidatorKeyRecovery {
		if err := services.RequireRocketStorage(c); err != nil {
			return nil, err
		}
		rp, err = services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
		bc, err = services.GetBeaconClient(c)
		if err != nil {
			return nil, err
		}
	}

	// Response
	response := api.SearchAndRecoverWalletResponse{}

	// Check if wallet is already initialized
	if w.IsInitialized() {
		return nil, errors.New("the wallet is already initialized")
	}

	// Try each derivation path across all of the iterations
	paths := []string{
		wallet.DefaultNodeKeyPath,
		wallet.LedgerLiveNodeKeyPath,
		wallet.MyEtherWalletNodeKeyPath,
	}
	for i := uint(0); i < findIterations; i++ {
		for j := 0; j < len(paths); j++ {
			derivationPath := paths[j]
			recoveredWallet, err := wallet.NewWallet("", "", uint(w.GetChainID().Uint64()), nil, nil, 0, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("error generating new wallet: %w", err)
			}
			err = recoveredWallet.TestRecovery(derivationPath, i, mnemonic)
			if err != nil {
				return nil, fmt.Errorf("error recovering wallet with path [%s], index [%d]: %w", derivationPath, i, err)
			}

			// Get recovered account
			recoveredAccount, err := recoveredWallet.GetNodeAccount()
			if err != nil {
				return nil, fmt.Errorf("error getting recovered account: %w", err)
			}
			if recoveredAccount.Address == address {
				// We found the correct derivation path and index
				response.FoundWallet = true
				response.DerivationPath = derivationPath
				response.Index = i
				break
			}
		}
		if response.FoundWallet {
			break
		}
	}

	if !response.FoundWallet {
		return nil, fmt.Errorf("exhausted all derivation paths and indices from 0 to %d, wallet not found", findIterations)
	}

	// Recover wallet
	if err := w.Recover(response.DerivationPath, response.Index, mnemonic); err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	if !skipValidatorKeyRecovery {
		response.ValidatorKeys, err = recoverNodeKeys(c, rp, bc, nodeAccount.Address, w, false)
		if err != nil {
			return nil, err
		}
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func recoverNodeKeys(c *cli.Command, rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, w wallet.Wallet, testOnly bool) ([]types.ValidatorPubkey, error) {
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Get node's validating pubkeys
	pubkeys, err := minipool.GetNodeValidatingMinipoolPubkeys(rp, nodeAddress, nil)
	if err != nil {
		return nil, err
	}

	// Check if the node has a megapool
	megapoolDeployed, err := megapool.GetMegapoolDeployed(rp, nodeAddress, nil)
	if err != nil {
		return nil, err
	}

	if megapoolDeployed {
		// Get the megapool address
		megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAddress, nil)
		if err != nil {
			return nil, err
		}

		// Load the megapool
		mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
		if err != nil {
			return nil, err
		}

		megapoolPubkeys, err := mp.GetMegapoolPubkeys(nil)
		if err != nil {
			return nil, err
		}

		pubkeys = append(pubkeys, megapoolPubkeys...)
	}

	// Remove zero pubkeys
	zeroPubkey := types.ValidatorPubkey{}
	filteredPubkeys := []types.ValidatorPubkey{}
	for _, pubkey := range pubkeys {
		if !bytes.Equal(pubkey[:], zeroPubkey[:]) {
			filteredPubkeys = append(filteredPubkeys, pubkey)
		}
	}
	pubkeys = filteredPubkeys

	// Get validator statuses by pubkeys
	statuses, err := bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting validator statuses: %w", err)
	}

	// Filter out inactive validators
	filteredPubkeys = []types.ValidatorPubkey{}
	for _, pubkey := range pubkeys {
		if statuses[pubkey].Status == beacon.ValidatorState_ActiveOngoing ||
			statuses[pubkey].Status == beacon.ValidatorState_ActiveExiting ||
			statuses[pubkey].Status == beacon.ValidatorState_PendingInitialized ||
			statuses[pubkey].Status == beacon.ValidatorState_PendingQueued {
			filteredPubkeys = append(filteredPubkeys, pubkey)
		}
	}

	pubkeyMap := map[types.ValidatorPubkey]bool{}
	for _, pubkey := range pubkeys {
		pubkeyMap[pubkey] = true
	}

	// Publish progress so a second CLI invocation can report how far along this is
	keysTotal := len(pubkeyMap)
	activeRecovery.setKeysTotal(keysTotal)

	pubkeyMap, err = checkForAndRecoverCustomMinipoolKeys(cfg, pubkeyMap, w, testOnly)
	if err != nil {
		return nil, fmt.Errorf("error checking for or recovering custom validator keys: %w", err)
	}
	activeRecovery.setKeysFound(keysTotal - len(pubkeyMap))

	// Recover conventionally generated keys
	bucketStart := uint(0)
	for {
		if bucketStart >= bucketLimit {
			return nil, fmt.Errorf("attempt limit exceeded (%d keys)", bucketLimit)
		}
		bucketEnd := bucketStart + bucketSize
		if bucketEnd > bucketLimit {
			bucketEnd = bucketLimit
		}

		// Get the keys for this bucket
		keys, err := w.GetValidatorKeys(bucketStart, bucketEnd-bucketStart)
		if err != nil {
			return nil, err
		}
		for _, validatorKey := range keys {
			_, exists := pubkeyMap[validatorKey.PublicKey]
			if exists {
				// Found one!
				delete(pubkeyMap, validatorKey.PublicKey)
				if !testOnly {
					err := w.SaveValidatorKey(validatorKey)
					if err != nil {
						return nil, fmt.Errorf("error recovering validator keys: %w", err)
					}
				}
				activeRecovery.setKeysFound(keysTotal - len(pubkeyMap))
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

func checkForAndRecoverCustomMinipoolKeys(cfg *config.RocketPoolConfig, pubkeyMap map[types.ValidatorPubkey]bool, w wallet.Wallet, testOnly bool) (map[types.ValidatorPubkey]bool, error) {

	// Load custom validator keys
	customKeyDir := cfg.Smartnode.GetCustomKeyPath()
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
			passwordFile := cfg.Smartnode.GetCustomKeyPasswordFilePath()
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

	return pubkeyMap, nil

}
