package watchtower

import (
    "context"
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
)


// Settings
var submitNetworkBalancesInterval, _ = time.ParseDuration("1m")


// Network balance info
type networkBalances struct {
    Block uint64
    DepositPool *big.Int
    MinipoolsTotal *big.Int
    MinipoolsStaking *big.Int
    RETHContract *big.Int
    RETHSupply *big.Int
}
type minipoolBalanceDetails struct {
    IsStaking bool
    UserBalance *big.Int
}


// Start submit network balances task
func startSubmitNetworkBalances(c *cli.Context) error {

    // Get services
    if err := services.WaitNodeRegistered(c, true); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return err }

    // Submit network balances at interval
    go (func() {
        for {
            if err := submitNetworkBalances(c, am, rp, bc); err != nil {
                log.Println(err)
            }
            time.Sleep(submitNetworkBalancesInterval)
        }
    })()

    // Return
    return nil

}


// Submit network balances
func submitNetworkBalances(c *cli.Context, am *accounts.AccountManager, rp *rocketpool.RocketPool, bc beacon.Client) error {

    // Wait for eth clients to sync
    if err := services.WaitEthClientSynced(c, true); err != nil {
        return err
    }
    if err := services.WaitBeaconClientSynced(c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil {
        return err
    }

    // Data
    var wg errgroup.Group
    var nodeTrusted bool
    var submitBalancesEnabled bool

    // Get data
    wg.Go(func() error {
        var err error
        nodeTrusted, err = node.GetNodeTrusted(rp, nodeAccount.Address, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        submitBalancesEnabled, err = settings.GetSubmitBalancesEnabled(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    // Check node trusted status & settings
    if !(nodeTrusted && submitBalancesEnabled) {
        return nil
    }

    // Get block to submit balances for
    blockNumber, err := getLatestReportableBlock(rp)
    if err != nil {
        return err
    }

    // Check if balances for block can be submitted by node
    canSubmit, err := canSubmitBlockBalances(rp, nodeAccount.Address, blockNumber)
    if err != nil {
        return err
    }
    if !canSubmit {
        return nil
    }

    // Log
    log.Printf("Calculating network balances for block %d...\n", blockNumber)

    // Get network balances at block
    balances, err := getNetworkBalances(rp, bc, blockNumber)
    if err != nil {
        return err
    }

    // Log
    log.Printf("Deposit pool balance: %.2f ETH\n", eth.WeiToEth(balances.DepositPool))
    log.Printf("Total minipool user balance: %.2f ETH\n", eth.WeiToEth(balances.MinipoolsTotal))
    log.Printf("Staking minipool user balance: %.2f ETH\n", eth.WeiToEth(balances.MinipoolsStaking))
    log.Printf("rETH contract balance: %.2f ETH\n", eth.WeiToEth(balances.RETHContract))
    log.Printf("rETH token supply: %.2f rETH\n", eth.WeiToEth(balances.RETHSupply))

    // Submit balances
    if err := submitBalances(am, rp, balances); err != nil {
        return fmt.Errorf("Could not submit network balances: %w", err)
    }

    // Return
    return nil

}


// Get the latest block number to report balances for
func getLatestReportableBlock(rp *rocketpool.RocketPool) (uint64, error) {

    // Data
    var wg errgroup.Group
    var currentBlock uint64
    var submitBalancesFrequency uint64

    // Get current block
    wg.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Uint64()
        }
        return err
    })

    // Get balance submission frequency
    wg.Go(func() error {
        var err error
        submitBalancesFrequency, err = settings.GetSubmitBalancesFrequency(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return 0, err
    }

    // Calculate and return
    return (currentBlock / submitBalancesFrequency) * submitBalancesFrequency, nil

}


// Check whether balances for a block can be submitted by the node
func canSubmitBlockBalances(rp *rocketpool.RocketPool, nodeAddress common.Address, blockNumber uint64) (bool, error) {

    // Data
    var wg errgroup.Group
    var currentBalancesBlock uint64
    var nodeSubmittedBlock bool

    // Get data
    wg.Go(func() error {
        var err error
        currentBalancesBlock, err = network.GetBalancesBlock(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        nodeSubmittedBlock, err = rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.balances.submitted.node"), nodeAddress.Bytes(), big.NewInt(int64(blockNumber)).Bytes()))
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Return
    return (blockNumber > currentBalancesBlock && !nodeSubmittedBlock), nil

}


// Get the network balances at a specific block
func getNetworkBalances(rp *rocketpool.RocketPool, bc beacon.Client, blockNumber uint64) (networkBalances, error) {

    // Initialize call options
    opts := &bind.CallOpts{
        BlockNumber: big.NewInt(int64(blockNumber)),
    }

    // Data
    var wg errgroup.Group
    var depositPoolBalance *big.Int
    var minipoolBalanceDetails []minipoolBalanceDetails
    var rethContractBalance *big.Int
    var rethTotalSupply *big.Int

    // Get deposit pool balance
    wg.Go(func() error {
        var err error
        depositPoolBalance, err = deposit.GetBalance(rp, opts)
        return err
    })

    // Get minipool balance details
    wg.Go(func() error {
        var err error
        minipoolBalanceDetails, err = getNetworkMinipoolBalanceDetails(rp, bc, opts)
        return err
    })

    // Get rETH contract balance
    wg.Go(func() error {
        rethContractAddress, err := rp.GetAddress("rocketETHToken")
        if err != nil {
            return err
        }
        rethContractBalance, err = rp.Client.BalanceAt(context.Background(), *rethContractAddress, opts.BlockNumber)
        return err
    })

    // Get rETH token supply
    wg.Go(func() error {
        var err error
        rethTotalSupply, err = tokens.GetRETHTotalSupply(rp, opts)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return networkBalances{}, err
    }

    // Balances
    balances := networkBalances{
        Block: blockNumber,
        DepositPool: depositPoolBalance,
        MinipoolsTotal: big.NewInt(0),
        MinipoolsStaking: big.NewInt(0),
        RETHContract: rethContractBalance,
        RETHSupply: rethTotalSupply,
    }

    // Add minipool balances
    for _, mp := range minipoolBalanceDetails {
        balances.MinipoolsTotal.Add(balances.MinipoolsTotal, mp.UserBalance)
        if mp.IsStaking {
            balances.MinipoolsStaking.Add(balances.MinipoolsStaking, mp.UserBalance)
        }
    }

    // Return
    return balances, nil

}


// Get all minipool balance details
func getNetworkMinipoolBalanceDetails(rp *rocketpool.RocketPool, bc beacon.Client, opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var eth2Config beacon.Eth2Config
    var beaconHead beacon.BeaconHead
    var blockTime uint64

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetMinipoolAddresses(rp, opts)
        return err
    })

    // Get eth2 config
    wg1.Go(func() error {
        var err error
        eth2Config, err = bc.GetEth2Config()
        return err
    })

    // Get beacon head
    wg1.Go(func() error {
        var err error
        beaconHead, err = bc.GetBeaconHead()
        return err
    })

    // Get block time
    wg1.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
        if err == nil {
            blockTime = header.Time
        }
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []minipoolBalanceDetails{}, err
    }

    // Get & check epoch at block
    blockEpoch := epochAt(eth2Config, blockTime)
    if blockEpoch > beaconHead.Epoch {
        return []minipoolBalanceDetails{}, fmt.Errorf("Epoch %d at block %s is higher than current epoch %d", blockEpoch, opts.BlockNumber.String(), beaconHead.Epoch)
    }

    // Data
    var wg2 errgroup.Group
    details := make([]minipoolBalanceDetails, len(addresses))

    // Load details
    for mi, address := range addresses {
        mi, address := mi, address
        wg2.Go(func() error {
            mpDetails, err := getMinipoolBalanceDetails(rp, bc, address, opts, eth2Config, blockEpoch)
            if err == nil { details[mi] = mpDetails }
            return err
        })
    }

    // Wait for data
    if err := wg2.Wait(); err != nil {
        return []minipoolBalanceDetails{}, err
    }

    // Return
    return details, nil

}


// Get minipool balance details
func getMinipoolBalanceDetails(rp *rocketpool.RocketPool, bc beacon.Client, minipoolAddress common.Address, opts *bind.CallOpts, eth2Config beacon.Eth2Config, blockEpoch uint64) (minipoolBalanceDetails, error) {

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return minipoolBalanceDetails{}, err
    }

    // Data
    var wg errgroup.Group
    var status types.MinipoolStatus
    var nodeFee float64
    var userDepositBalance *big.Int
    var userDepositTime uint64
    var pubkey types.ValidatorPubkey
    var withdrawalProcessed bool

    // Load data
    wg.Go(func() error {
        var err error
        status, err = mp.GetStatus(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        nodeFee, err = mp.GetNodeFee(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        userDepositBalance, err = mp.GetUserDepositBalance(opts)
        return err
    })
    wg.Go(func() error {
        userDepositAssignedTime, err := mp.GetUserDepositAssignedTime(opts)
        if err == nil {
            userDepositTime = uint64(userDepositAssignedTime.Unix())
        }
        return err
    })
    wg.Go(func() error {
        var err error
        pubkey, err = minipool.GetMinipoolPubkey(rp, minipoolAddress, opts)
        return err
    })
    wg.Go(func() error {
        var err error
        withdrawalProcessed, err = minipool.GetMinipoolWithdrawalProcessed(rp, minipoolAddress, opts)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return minipoolBalanceDetails{}, err
    }

    // No balance if no user deposit assigned or withdrawal has been processed
    if userDepositBalance.Cmp(big.NewInt(0)) == 0 || withdrawalProcessed {
        return minipoolBalanceDetails{
            UserBalance: big.NewInt(0),
        }, nil
    }

    // Use user deposit balance if initialized or prelaunch
    if status == types.Initialized || status == types.Prelaunch {
        return minipoolBalanceDetails{
            UserBalance: userDepositBalance,
        }, nil
    }

    // Get validator status at block
    validator, err := bc.GetValidatorStatus(pubkey, &beacon.ValidatorStatusOptions{Epoch: blockEpoch})
    if err != nil {
        return minipoolBalanceDetails{}, err
    }

    // Use user deposit balance if validator not yet active on beacon chain at block
    if !validator.Exists || validator.ActivationEpoch > blockEpoch {
        return minipoolBalanceDetails{
            UserBalance: userDepositBalance,
        }, nil
    }

    // Get start epoch
    startEpoch := epochAt(eth2Config, userDepositTime)
    if startEpoch < validator.ActivationEpoch {
        startEpoch = validator.ActivationEpoch
    } else if startEpoch > blockEpoch {
        startEpoch = blockEpoch
    }

    // Get validator status at start epoch
    validatorStart, err := bc.GetValidatorStatus(pubkey, &beacon.ValidatorStatusOptions{Epoch: startEpoch})
    if err != nil {
        return minipoolBalanceDetails{}, err
    }
    if !validatorStart.Exists {
        return minipoolBalanceDetails{}, fmt.Errorf("Could not get validator %s balance at epoch %d", pubkey.Hex(), startEpoch)
    }

    // Get validator balances at start epoch and at block
    startBalance := eth.GweiToWei(float64(validatorStart.Balance))
    blockBalance := eth.GweiToWei(float64(validator.Balance))

    // Get node & user balance at block
    nodeBalance, err := minipool.GetMinipoolNodeRewardAmount(rp, nodeFee, userDepositBalance, startBalance, blockBalance, opts)
    if err != nil {
        return minipoolBalanceDetails{}, err
    }
    userBalance := big.NewInt(0)
    userBalance.Sub(blockBalance, nodeBalance)

    // Return
    return minipoolBalanceDetails{
        IsStaking: (validator.ExitEpoch > blockEpoch),
        UserBalance: userBalance,
    }, nil

}


// Submit network balances
func submitBalances(am *accounts.AccountManager, rp *rocketpool.RocketPool, balances networkBalances) error {

    // Log
    log.Printf("Submitting network balances for block %d...\n", balances.Block)

    // Calculate total ETH balance
    totalEth := big.NewInt(0)
    totalEth.Add(totalEth, balances.DepositPool)
    totalEth.Add(totalEth, balances.MinipoolsTotal)
    totalEth.Add(totalEth, balances.RETHContract)

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Submit balances
    if _, err := network.SubmitBalances(rp, balances.Block, totalEth, balances.MinipoolsStaking, balances.RETHSupply, opts); err != nil {
        return err
    }

    // Log
    log.Printf("Successfully submitted network balances for block %d.\n", balances.Block)

    // Return
    return nil

}

