package services

import (
	"bytes"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/urfave/cli.v1"

	"github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
	beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
	"github.com/rocket-pool/smartnode-cli/rocketpool/services/database"
	"github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
	"github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
	"github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)

// Service provider options
type ProviderOpts struct {
	DB                  bool
	AM                  bool
	Client              bool
	CM                  bool
	NodeContractAddress bool
	NodeContract        bool
	Publisher           bool
	Beacon              bool
	VM                  bool
	LoadContracts       []string
	LoadAbis            []string
}

// Service provider
type Provider struct {
	DB                  *database.Database
	AM                  *accounts.AccountManager
	Client              *ethclient.Client
	CM                  *rocketpool.ContractManager
	NodeContractAddress *common.Address
	NodeContract        *bind.BoundContract
	Publisher           *messaging.Publisher
	Beacon              *beaconchain.Client
	VM                  *node.ValidatorManager
}

/**
 * Create service provider
 */
func NewProvider(c *cli.Context, opts ProviderOpts) (*Provider, error) {

	// Process options
	if opts.VM {
		opts.AM = true
		opts.CM = true
	} // Validator manager requires node account manager & RP contract manager
	if opts.Beacon {
		opts.Publisher = true
	} // Beacon chain client requires publisher
	if opts.NodeContract {
		opts.NodeContractAddress = true
	} // Node contract requires node contract address
	if opts.NodeContractAddress {
		opts.AM = true
		opts.CM = true
	} // Node contract address requires node account manager & RP contract manager
	if len(opts.LoadContracts)+len(opts.LoadAbis) > 0 {
		opts.CM = true
	} // Contracts & ABIs require RP contract manager
	if opts.CM {
		opts.Client = true
	} // RP contract manager requires eth client

	// Service provider
	p := &Provider{}

	// Initialise database
	if opts.DB {
		p.DB = database.NewDatabase(c.GlobalString("database"))
	}

	// Initialise account manager
	if opts.AM {

		// Initialise
		p.AM = accounts.NewAccountManager(c.GlobalString("keychain"))

		// Check node account
		if !p.AM.NodeAccountExists() {
			return nil, errors.New("Node account does not exist, please initialize with `rocketpool node init`")
		}

	}

	// Initialise ethereum client
	if opts.Client {
		if client, err := ethclient.Dial(c.GlobalString("providerPow")); err != nil {
			return nil, errors.New("Error connecting to ethereum node: " + err.Error())
		} else {
			p.Client = client
		}
	}

	// Initialise Rocket Pool contract manager
	if opts.CM {
		if cm, err := rocketpool.NewContractManager(p.Client, c.GlobalString("storageAddress")); err != nil {
			return nil, err
		} else {
			p.CM = cm
		}
	}

	// Load contracts & ABIs
	if len(opts.LoadContracts)+len(opts.LoadAbis) > 0 {

		// Loading channels
		successChannel := make(chan bool)
		errorChannel := make(chan error)

		// Load Rocket Pool contracts
		go (func() {
			if err := p.CM.LoadContracts(opts.LoadContracts); err != nil {
				errorChannel <- err
			} else {
				successChannel <- true
			}
		})()
		go (func() {
			if err := p.CM.LoadABIs(opts.LoadAbis); err != nil {
				errorChannel <- err
			} else {
				successChannel <- true
			}
		})()

		// Await loading
		for received := 0; received < 2; {
			select {
			case <-successChannel:
				received++
			case err := <-errorChannel:
				return nil, err
			}
		}

	}

	// Initialise node contract address
	if opts.NodeContractAddress {
		nodeContractAddress := new(common.Address)
		if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", p.AM.GetNodeAccount().Address); err != nil {
			return nil, errors.New("Error checking node registration: " + err.Error())
		} else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
			return nil, errors.New("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
		} else {
			p.NodeContractAddress = nodeContractAddress
		}
	}

	// Initialise node contract
	if opts.NodeContract {
		if nodeContract, err := p.CM.NewContract(p.NodeContractAddress, "rocketNodeContract"); err != nil {
			return nil, errors.New("Error initialising node contract: " + err.Error())
		} else {
			p.NodeContract = nodeContract
		}
	}

	// Initialise publisher
	if opts.Publisher {
		p.Publisher = messaging.NewPublisher()
	}

	// Initialise beacon chain client
	if opts.Beacon {
		p.Beacon = beaconchain.NewClient(c.GlobalString("providerBeacon"), p.Publisher)
	}

	// Initialise validator manager
	if opts.VM {
		p.VM = node.NewValidatorManager(p.AM, p.CM)
	}

	// Return
	return p, nil

}
