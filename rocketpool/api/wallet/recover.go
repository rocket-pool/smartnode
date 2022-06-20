package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
)

const (
	findIterations  uint   = 100000
	passwordPattern string = "^(?P<address>0x[0-9a-fA-F]{96})=\"(?P<password>.*)\"$"
)

// Encrypted validator keystore following the EIP-2335 standard
// (https://eips.ethereum.org/EIPS/eip-2335)
type validatorKeystore struct {
	Crypto  map[string]interface{} `json:"crypto"`
	Version uint                   `json:"version"`
	UUID    uuid.UUID              `json:"uuid"`
	Path    string                 `json:"path"`
	Pubkey  types.ValidatorPubkey  `json:"pubkey"`
}

func recoverWallet(c *cli.Context, mnemonic string) (*api.RecoverWalletResponse, error) {

	// Get services
	if err := services.RequireNodePassword(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	var rp *rocketpool.RocketPool
	if !c.Bool("skip-validator-key-recovery") {
		if err := services.RequireRocketStorage(c); err != nil {
			return nil, err
		}
		rp, err = services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
	}

	// Response
	response := api.RecoverWalletResponse{}

	// Check if wallet is already initialized
	if w.IsInitialized() {
		return nil, errors.New("The wallet is already initialized")
	}

	// Get the derivation path
	path := c.String("derivation-path")
	switch path {
	case "":
		path = wallet.DefaultNodeKeyPath
	case "ledgerLive":
		path = wallet.LedgerLiveNodeKeyPath
	case "mew":
		path = wallet.MyEtherWalletNodeKeyPath
	}

	// Get the wallet index
	walletIndex := c.Uint("wallet-index")

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

	if !c.Bool("skip-validator-key-recovery") {
		recoverMinipoolKeys(c, rp, nodeAccount.Address, w)
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func searchAndRecoverWallet(c *cli.Context, mnemonic string, address common.Address) (*api.SearchAndRecoverWalletResponse, error) {

	// Get services
	if err := services.RequireNodePassword(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	var rp *rocketpool.RocketPool
	if !c.Bool("skip-validator-key-recovery") {
		if err := services.RequireRocketStorage(c); err != nil {
			return nil, err
		}
		rp, err = services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
	}

	// Response
	response := api.SearchAndRecoverWalletResponse{}

	// Check if wallet is already initialized
	if w.IsInitialized() {
		return nil, errors.New("The wallet is already initialized")
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
			recoveredWallet, err := wallet.NewWallet("", uint(w.GetChainID().Uint64()), nil, nil, 0, nil)
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
		return nil, fmt.Errorf("Exhausted all derivation paths and indices from 0 to %d, wallet not found.", findIterations)
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

	if !c.Bool("skip-validator-key-recovery") {
		recoverMinipoolKeys(c, rp, nodeAccount.Address, w)
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func recoverMinipoolKeys(c *cli.Context, rp *rocketpool.RocketPool, address common.Address, w *wallet.Wallet) ([]types.ValidatorPubkey, error) {

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

	// Handle passwords for custom validator keys
	passwordRegex := regexp.MustCompile(passwordPattern)
	passwordArgs := c.StringSlice("password")
	customKeyPasswords := map[types.ValidatorPubkey]string{}
	for _, passwordArg := range passwordArgs {
		match := passwordRegex.FindStringSubmatch(passwordArg)
		if match == nil {
			return nil, fmt.Errorf("custom password argument [%s] was not a valid format. Expect `--password 0xabcd...=\"some password\"", passwordArg)
		}
		pubkey, err := types.HexToValidatorPubkey(match[1])
		if err != nil {
			return nil, fmt.Errorf("can't parse pubkey [%s] in argument [%s]: %w", match[1], passwordArg, err)
		}
		customKeyPasswords[pubkey] = match[2]
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

		// Process every custom key
		for _, file := range files {
			// Read the file
			bytes, err := ioutil.ReadFile(filepath.Join(customKeyDir, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("error reading custom keystore %s: %w", file.Name(), err)
			}

			// Deserialize it
			keystore := validatorKeystore{}
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

			// Check if we have a password for it
			password, exists := customKeyPasswords[keystore.Pubkey]
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
			err = w.StoreValidatorKey(privateKey, keystore.Path)
			if err != nil {
				return nil, fmt.Errorf("error storing private keystore for %s: %w", reconstructedPubkey.Hex(), err)
			}

			// Remove the pubkey from pending minipools to handle
			delete(pubkeyMap, reconstructedPubkey)
		}
	}

	// Recover remaining validator keys normally
	for pubkey := range pubkeyMap {
		if err := w.RecoverValidatorKey(pubkey); err != nil {
			return nil, err
		}
	}

	return pubkeys, nil

}
