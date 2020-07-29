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
)


// Settings
var submitNetworkBalancesInterval, _ = time.ParseDuration("1m")


// Network balance info
type networkBalances struct {
    Block int64
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

    // Submit network balances at interval
    go (func() {
        for {
            if err := submitNetworkBalances(c, am, rp); err != nil {
                log.Println(err)
            }
            time.Sleep(submitNetworkBalancesInterval)
        }
    })()

    // Return
    return nil

}


// Submit network balances
func submitNetworkBalances(c *cli.Context, am *accounts.AccountManager, rp *rocketpool.RocketPool) error {

    // Wait for eth client to sync
    if err := services.WaitClientSynced(c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil {
        return err
    }

    // Check node trusted status
    nodeTrusted, err := node.GetNodeTrusted(rp, nodeAccount.Address, nil)
    if err != nil {
        return err
    }
    if !nodeTrusted {
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
    balances, err := getNetworkBalances(rp, blockNumber)
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
func getLatestReportableBlock(rp *rocketpool.RocketPool) (int64, error) {

    // Data
    var wg errgroup.Group
    var currentBlock int64
    var submitBalancesFrequency int64

    // Get current block
    wg.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Int64()
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
func canSubmitBlockBalances(rp *rocketpool.RocketPool, nodeAddress common.Address, blockNumber int64) (bool, error) {

    // Data
    var wg errgroup.Group
    var submitBalancesEnabled bool
    var currentBalancesBlock int64
    var nodeSubmittedBlock bool

    // Get data
    wg.Go(func() error {
        var err error
        submitBalancesEnabled, err = settings.GetSubmitBalancesEnabled(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        currentBalancesBlock, err = network.GetBalancesBlock(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        nodeSubmittedBlock, err = rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.balances.submitted.node"), nodeAddress.Bytes(), big.NewInt(blockNumber).Bytes()))
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Return
    return (submitBalancesEnabled && blockNumber > currentBalancesBlock && !nodeSubmittedBlock), nil

}


// Get the network balances at a specific block
func getNetworkBalances(rp *rocketpool.RocketPool, blockNumber int64) (networkBalances, error) {

    // Initialize call options
    opts := &bind.CallOpts{
        BlockNumber: big.NewInt(blockNumber),
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
        minipoolBalanceDetails, err = getNetworkMinipoolBalanceDetails(rp, opts)
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
func getNetworkMinipoolBalanceDetails(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var blockTime int64

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetMinipoolAddresses(rp, opts)
        return err
    })

    // Get beacon chain config with: genesis time, genesis epoch & seconds per epoch
    // TODO: implement

    // Get block time
    wg1.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
        if err == nil {
            blockTime = int64(header.Time)
        }
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []minipoolBalanceDetails{}, err
    }

    // Data
    var wg2 errgroup.Group
    details := make([]minipoolBalanceDetails, len(addresses))

    // Load details
    for mi, address := range addresses {
        mi, address := mi, address
        wg2.Go(func() error {
            mpDetails, err := getMinipoolBalanceDetails(rp, address, opts, blockTime)
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
func getMinipoolBalanceDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts, blockTime int64) (minipoolBalanceDetails, error) {

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
    var userDepositTime int64
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
            userDepositTime = userDepositAssignedTime.Unix()
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

    // Get validator start balance at userDepositTime (or activation) epoch
    // Get validator current balance at blockTime epoch
    // If validator current balance not found
    // - Use minipool user deposit balance for staking minipools
    // - Error for withdrawable minipools
    // Get user share of validator balance
    // TODO: implement

    // Return
    return minipoolBalanceDetails{
        UserBalance: big.NewInt(0),
    }, nil

}


// Submit network balances
func submitBalances(am *accounts.AccountManager, rp *rocketpool.RocketPool, balances networkBalances) error {

    // Log
    log.Printf("Submitting network balances for block %d...\n", balances.Block)

    // Calculate ETH balances
    totalEth := big.NewInt(0)
    totalEth.Add(totalEth, balances.DepositPool)
    totalEth.Add(totalEth, balances.MinipoolsTotal)
    totalEth.Add(totalEth, balances.MinipoolsStaking)
    totalEth.Add(totalEth, balances.RETHContract)
    stakingEth := big.NewInt(0)
    stakingEth.Add(stakingEth, balances.MinipoolsStaking)

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Submit balances
    if _, err := network.SubmitBalances(rp, balances.Block, totalEth, stakingEth, balances.RETHSupply, opts); err != nil {
        return err
    }

    // Log
    log.Printf("Successfully submitted network balances for block %d.\n", balances.Block)

    // Return
    return nil

}

