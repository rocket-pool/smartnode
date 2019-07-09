package services

import (
    "bytes"
    "errors"

    "github.com/docker/docker/client"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/shared/services/accounts"
    beaconchain "github.com/rocket-pool/smartnode/shared/services/beacon-chain"
    "github.com/rocket-pool/smartnode/shared/services/database"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/validators"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
    "github.com/rocket-pool/smartnode/shared/utils/messaging"
    "github.com/rocket-pool/smartnode/shared/utils/sync"
)


// Config
const DOCKER_API_VERSION string = "1.39"


// Service provider options
type ProviderOpts struct {
    DB                  bool
    PM                  bool
    AM                  bool
    KM                  bool
    Client              bool
    CM                  bool
    NodeContractAddress bool
    NodeContract        bool
    Publisher           bool
    Beacon              bool
    Docker              bool
    LoadContracts       []string
    LoadAbis            []string
    WaitPassword        bool
    WaitNodeAccount     bool
    WaitNodeRegistered  bool
    WaitClientConn      bool
    WaitClientSync      bool
    WaitRocketStorage   bool
}


// Service provider
type Provider struct {
    DB                  *database.Database
    PM                  *passwords.PasswordManager
    AM                  *accounts.AccountManager
    KM                  *validators.KeyManager
    Client              *ethclient.Client
    CM                  *rocketpool.ContractManager
    NodeContractAddress *common.Address
    NodeContract        *bind.BoundContract
    Publisher           *messaging.Publisher
    Beacon              *beaconchain.Client
    Docker              *client.Client
}


/**
 * Create service provider
 */
func NewProvider(c *cli.Context, opts ProviderOpts) (*Provider, error) {

    // Process options
    if opts.WaitPassword {
        opts.PM = true
    } // Password requires password manager
    if opts.WaitNodeAccount {
        opts.AM = true
    } // Node account requires node account manager
    if opts.WaitNodeRegistered {
        opts.AM = true
        opts.CM = true
    } // Node registration requires node account manager & RP contract manager
    if opts.WaitClientConn || opts.WaitClientSync || opts.WaitRocketStorage {
        opts.Client = true
    } // Connected client, synced client and RS contract require eth client
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
    if len(opts.LoadContracts) + len(opts.LoadAbis) > 0 {
        opts.CM = true
    } // Contracts & ABIs require RP contract manager
    if opts.CM {
        opts.Client = true
    } // RP contract manager requires eth client
    if opts.AM || opts.KM {
        opts.PM = true
    } // Account & key managers require password manager

    // Service provider
    p := &Provider{}

    // Initialise database
    if opts.DB {
        p.DB = database.NewDatabase(c.GlobalString("database"))
    }

    // Initialise password manager
    if opts.PM {

        // Initialise
        p.PM = passwords.NewPasswordManager(c.GlobalString("password"))

        // Check or wait for password set
        if opts.WaitPassword {
            sync.WaitPasswordSet(p.PM)
        } else if !p.PM.PasswordExists() {
            return nil, errors.New("Node password is not set, please initialize with `rocketpool run node init`")
        }

    }

    // Initialise account manager
    if opts.AM {

        // Initialise
        p.AM = accounts.NewAccountManager(c.GlobalString("keychainPow"), p.PM)

        // Check or wait for node account
        if opts.WaitNodeAccount {
            sync.WaitNodeAccountSet(p.AM)
        } else if !p.AM.NodeAccountExists() {
            return nil, errors.New("Node account does not exist, please initialize with `rocketpool run node init`")
        }

    }

    // Initialise validator key manager
    if opts.KM {
        p.KM = validators.NewKeyManager(c.GlobalString("keychainBeacon"), p.PM)
    }

    // Initialise ethereum client
    if opts.Client {
        if client, err := ethclient.Dial(c.GlobalString("providerPow")); err != nil {
            return nil, errors.New("Error connecting to ethereum node: " + err.Error())
        } else {
            p.Client = client
        }
    }

    // Wait for ethereum client connection
    if opts.WaitClientConn {
        sync.WaitClientConnection(p.Client)
    }

    // Wait until RocketStorage contract is available
    if opts.WaitRocketStorage {
        sync.WaitContractLoaded(p.Client, "RocketStorage", common.HexToAddress(c.GlobalString("storageAddress")))
    }

    // Wait for ethereum client to sync
    if opts.WaitClientSync {
        if err := eth.WaitSync(p.Client, false, true); err != nil {
            return nil, err
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
    if len(opts.LoadContracts) + len(opts.LoadAbis) > 0 {

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

    // Wait until node is registered
    if opts.WaitNodeRegistered {
        sync.WaitNodeRegistered(p.AM, p.CM)
    }

    // Initialise node contract address
    if opts.NodeContractAddress {
        nodeContractAddress := new(common.Address)
        if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", p.AM.GetNodeAccount().Address); err != nil {
            return nil, errors.New("Error checking node registration: " + err.Error())
        } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
            return nil, errors.New("Node is not registered with Rocket Pool, please register with `rocketpool run node register`")
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

    // Initialise docker client
    if opts.Docker {
        if docker, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION)); err != nil {
            return nil, errors.New("Error initialising docker client: " + err.Error())
        } else {
            p.Docker = docker
        }
    }

    // Return
    return p, nil

}
