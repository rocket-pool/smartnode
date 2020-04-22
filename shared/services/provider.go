package services

import (
    "bytes"
    "errors"
    "fmt"
    "log"
    "os"

    "github.com/docker/docker/client"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/contracts"
    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/services/beacon/lighthouse"
    "github.com/rocket-pool/smartnode/shared/services/beacon/prysm"
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

    // Core
    DB                  bool
    PM                  bool
    AM                  bool
    KM                  bool

    // Eth
    Client              bool
    CM                  bool
    NodeContractAddress bool
    NodeContract        bool
    Uniswap             bool
    RPLExchangeAddress  bool
    RPLExchange         bool

    // Misc
    Publisher           bool
    Beacon              bool
    Docker              bool

    // Contracts
    LoadContracts       []string
    LoadAbis            []string

    // Sync
    ClientConn          bool
    ClientSync          bool
    RocketStorage       bool
    WaitPassword        bool
    WaitNodeAccount     bool
    WaitNodeRegistered  bool
    WaitClientConn      bool
    WaitClientSync      bool
    WaitRocketStorage   bool
    PasswordOptional    bool
    NodeAccountOptional bool

}


// Service provider
type Provider struct {

    // I/O
    Input               *os.File
    Output              *os.File
    Log                 *log.Logger

    // Core
    DB                  *database.Database
    PM                  *passwords.PasswordManager
    AM                  *accounts.AccountManager
    KM                  *validators.KeyManager

    // Eth
    Client              *ethclient.Client
    CM                  *rocketpool.ContractManager
    NodeContractAddress *common.Address
    NodeContract        *bind.BoundContract
    Uniswap             *contracts.UniswapFactory
    RPLExchangeAddress  *common.Address
    RPLExchange         *contracts.UniswapExchange

    // Misc
    Publisher           *messaging.Publisher
    Beacon              beacon.Client
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
    if opts.RPLExchange {
        opts.RPLExchangeAddress = true
    } // RPL Exchange contract requires RPL Exchange address
    if opts.RPLExchangeAddress {
        opts.Uniswap = true
        opts.CM = true
    } // RPL Exchange address requires uniswap & RP contract manager
    if opts.Uniswap {
        opts.Client = true
    } // Uniswap requires eth client
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

    // Initialise input source
    if inputPath := c.GlobalString("input"); inputPath != "" {
        if inputFile, err := os.Open(inputPath); err != nil {
            return nil, errors.New("Error opening CLI input file: " + err.Error())
        } else {
            p.Input = inputFile
        }
    } else {
        p.Input = os.Stdin
    }

    // Initialise output file
    if outputPath := c.GlobalString("output"); outputPath != "" {
        if outputFile, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
            return nil, errors.New("Error opening CLI output file: " + err.Error())
        } else {
            p.Output = outputFile
        }
    } else {
        p.Output = os.Stdout
    }
    p.Log = log.New(p.Output, log.Prefix(), log.Flags())

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
        } else if !opts.PasswordOptional && !p.PM.PasswordExists() {
            return nil, errors.New("Node password is not set, please initialize with `rocketpool node init`")
        }

    }

    // Initialise account manager
    if opts.AM {

        // Initialise
        p.AM = accounts.NewAccountManager(c.GlobalString("keychainPow"), p.PM)

        // Check or wait for node account
        if opts.WaitNodeAccount {
            sync.WaitNodeAccountSet(p.AM)
        } else if !opts.NodeAccountOptional && !p.AM.NodeAccountExists() {
            return nil, errors.New("Node account does not exist, please initialize with `rocketpool node init`")
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

    // Check or wait for ethereum client connection
    if opts.WaitClientConn {
        sync.WaitClientConnection(p.Client)
    } else if opts.ClientConn && !sync.ClientIsConnected(p.Client) {
        return nil, errors.New("Not connected to ethereum client")
    }

    // Check or wait for RocketStorage contract
    if opts.WaitRocketStorage {
        sync.WaitContractLoaded(p.Client, "RocketStorage", common.HexToAddress(c.GlobalString("storageAddress")))
    } else if opts.RocketStorage && !sync.ContractIsLoaded(p.Client, common.HexToAddress(c.GlobalString("storageAddress"))) {
        return nil, errors.New("RocketStorage contract not loaded")
    }

    // Check or wait for ethereum client sync
    if opts.WaitClientSync {
        if err := eth.WaitSync(p.Client, false, true); err != nil {
            return nil, err
        }
    } else if opts.ClientSync && !eth.IsSynced(p.Client) {
        return nil, errors.New("Ethereum client not synced")
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
        if err := p.CM.LoadContracts(opts.LoadContracts); err != nil { return nil, err }
        if err := p.CM.LoadABIs(opts.LoadAbis); err != nil { return nil, err }
    }

    // Wait until node is registered
    if opts.WaitNodeRegistered {

        // Check rocketNodeAPI contract is loaded
        if _, ok := p.CM.Contracts["rocketNodeAPI"]; !ok { return nil, errors.New("RocketNodeAPI contract is required to wait for node registration") }

        // Wait until node is registered
        sync.WaitNodeRegistered(p.AM, p.CM)

    }

    // Initialise node contract address
    if opts.NodeContractAddress {

        // Check rocketNodeAPI contract is loaded
        if _, ok := p.CM.Contracts["rocketNodeAPI"]; !ok { return nil, errors.New("RocketNodeAPI contract is required for node contract address") }

        // Get node contract address
        nodeAccount, _ := p.AM.GetNodeAccount()
        nodeContractAddress := new(common.Address)
        if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
            return nil, errors.New("Error checking node registration: " + err.Error())
        } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
            return nil, errors.New("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
        } else {
            p.NodeContractAddress = nodeContractAddress
        }

    }

    // Initialise node contract
    if opts.NodeContract {

        // Check rocketNodeContract ABI is loaded
        if _, ok := p.CM.Abis["rocketNodeContract"]; !ok { return nil, errors.New("RocketNodeContract ABI is required for node contract") }

        // Initialise contract
        if nodeContract, err := p.CM.NewContract(p.NodeContractAddress, "rocketNodeContract"); err != nil {
            return nil, errors.New("Error initialising node contract: " + err.Error())
        } else {
            p.NodeContract = nodeContract
        }

    }

    // Initialise uniswap contract
    if opts.Uniswap {
        if uniswap, err := contracts.NewUniswapFactory(common.HexToAddress(c.GlobalString("uniswapAddress")), p.Client); err != nil {
            return nil, errors.New("Error initialising Uniswap contract: " + err.Error())
        } else {
            p.Uniswap = uniswap
        }
    }

    // Initialise RPL exchange address
    if opts.RPLExchangeAddress {
        if rplTokenAddress, ok := p.CM.Addresses["rocketPoolToken"]; !ok {
            return nil, errors.New("Error retrieving RPL exchange address: RPL contract address not loaded")
        } else if rplExchangeAddress, err := p.Uniswap.GetExchange(nil, *rplTokenAddress); err != nil {
            return nil, errors.New("Error retrieving RPL exchange address: " + err.Error())
        } else {
            p.RPLExchangeAddress = &rplExchangeAddress
        }
    }

    // Initialise RPL exchange contract
    if opts.RPLExchange {
        if rplExchange, err := contracts.NewUniswapExchange(*(p.RPLExchangeAddress), p.Client); err != nil {
            return nil, errors.New("Error initialising RPL exchange contract: " + err.Error())
        } else {
            p.RPLExchange = rplExchange
        }
    }

    // Initialise publisher
    if opts.Publisher {
        p.Publisher = messaging.NewPublisher()
    }

    // Initialise beacon chain client
    if opts.Beacon {
        switch c.GlobalString("beaconApiMode") {
            case "lighthouse":
                p.Beacon = lighthouse.NewClient(c.GlobalString("providerBeacon"))
            case "prysm":
                p.Beacon = prysm.NewClient(c.GlobalString("providerBeacon"))
            default:
                return nil, errors.New(fmt.Sprintf("Invalid Eth 2.0 API mode '%s' specified", c.GlobalString("beaconApiMode")))
        }
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


/**
 * Cleanup service provider (close resources)
 */
func (p *Provider) Cleanup() {
    if p.Input != os.Stdin { p.Input.Close() }
    if p.Output != os.Stdout { p.Output.Close() }
}

