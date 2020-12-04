package watchtower

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
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
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/services/wallet"
    "github.com/rocket-pool/smartnode/shared/utils/eth2"
    "github.com/rocket-pool/smartnode/shared/utils/log"
    "github.com/rocket-pool/smartnode/shared/utils/math"
    "github.com/rocket-pool/smartnode/shared/utils/rp"
)


// Settings
const MinipoolBalanceDetailsBatchSize = 20


// Submit network balances task
type submitNetworkBalances struct {
    c *cli.Context
    log log.ColorLogger
    w *wallet.Wallet
    ec *ethclient.Client
    rp *rocketpool.RocketPool
    bc beacon.Client
}


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


// Create submit network balances task
func newSubmitNetworkBalances(c *cli.Context, logger log.ColorLogger) (*submitNetworkBalances, error) {

    // Get services
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Return task
    return &submitNetworkBalances{
        c: c,
        log: logger,
        w: w,
        ec: ec,
        rp: rp,
        bc: bc,
    }, nil

}


// Submit network balances
func (t *submitNetworkBalances) run() error {

    // Wait for eth clients to sync
    if err := services.WaitEthClientSynced(t.c, true); err != nil {
        return err
    }
    if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := t.w.GetNodeAccount()
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
        nodeTrusted, err = node.GetNodeTrusted(t.rp, nodeAccount.Address, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        submitBalancesEnabled, err = settings.GetSubmitBalancesEnabled(t.rp, nil)
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

    // Log
    t.log.Println("Checking for network balance checkpoint...")

    // Get block to submit balances for
    blockNumber, err := t.getLatestReportableBlock()
    if err != nil {
        return err
    }

    // Check if balances for block can be submitted by node
    canSubmit, err := t.canSubmitBlockBalances(nodeAccount.Address, blockNumber)
    if err != nil {
        return err
    }
    if !canSubmit {
        return nil
    }

    // Log
    t.log.Printlnf("Calculating network balances for block %d...", blockNumber)

    // Get network balances at block
    balances, err := t.getNetworkBalances(blockNumber)
    if err != nil {
        return err
    }

    // Log
    t.log.Printlnf("Deposit pool balance: %.6f ETH", math.RoundDown(eth.WeiToEth(balances.DepositPool), 6))
    t.log.Printlnf("Total minipool user balance: %.6f ETH", math.RoundDown(eth.WeiToEth(balances.MinipoolsTotal), 6))
    t.log.Printlnf("Staking minipool user balance: %.6f ETH", math.RoundDown(eth.WeiToEth(balances.MinipoolsStaking), 6))
    t.log.Printlnf("rETH contract balance: %.6f ETH", math.RoundDown(eth.WeiToEth(balances.RETHContract), 6))
    t.log.Printlnf("rETH token supply: %.6f rETH", math.RoundDown(eth.WeiToEth(balances.RETHSupply), 6))

    // Submit balances
    if err := t.submitBalances(balances); err != nil {
        return fmt.Errorf("Could not submit network balances: %w", err)
    }

    // Return
    return nil

}


// Get the latest block number to report balances for
func (t *submitNetworkBalances) getLatestReportableBlock() (uint64, error) {

    // Data
    var wg errgroup.Group
    var currentBlock uint64
    var submitBalancesFrequency uint64

    // Get current block
    wg.Go(func() error {
        header, err := t.ec.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Uint64()
        }
        return err
    })

    // Get balance submission frequency
    wg.Go(func() error {
        var err error
        submitBalancesFrequency, err = settings.GetSubmitBalancesFrequency(t.rp, nil)
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
func (t *submitNetworkBalances) canSubmitBlockBalances(nodeAddress common.Address, blockNumber uint64) (bool, error) {

    // Data
    var wg errgroup.Group
    var currentBalancesBlock uint64
    var nodeSubmittedBlock bool

    // Get data
    wg.Go(func() error {
        var err error
        currentBalancesBlock, err = network.GetBalancesBlock(t.rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        blockNumberBuf := make([]byte, 32)
        big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
        nodeSubmittedBlock, err = t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.balances.submitted.node"), nodeAddress.Bytes(), blockNumberBuf))
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
func (t *submitNetworkBalances) getNetworkBalances(blockNumber uint64) (networkBalances, error) {

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
        depositPoolBalance, err = deposit.GetBalance(t.rp, opts)
        return err
    })

    // Get minipool balance details
    wg.Go(func() error {
        var err error
        minipoolBalanceDetails, err = t.getNetworkMinipoolBalanceDetails(opts)
        return err
    })

    // Get rETH contract balance
    wg.Go(func() error {
        rethContractAddress, err := t.rp.GetAddress("rocketETHToken")
        if err != nil {
            return err
        }
        rethContractBalance, err = t.ec.BalanceAt(context.Background(), *rethContractAddress, opts.BlockNumber)
        return err
    })

    // Get rETH token supply
    wg.Go(func() error {
        var err error
        rethTotalSupply, err = tokens.GetRETHTotalSupply(t.rp, opts)
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
func (t *submitNetworkBalances) getNetworkMinipoolBalanceDetails(opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var eth2Config beacon.Eth2Config
    var beaconHead beacon.BeaconHead
    var blockTime uint64

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetMinipoolAddresses(t.rp, opts)
        return err
    })

    // Get eth2 config
    wg1.Go(func() error {
        var err error
        eth2Config, err = t.bc.GetEth2Config()
        return err
    })

    // Get beacon head
    wg1.Go(func() error {
        var err error
        beaconHead, err = t.bc.GetBeaconHead()
        return err
    })

    // Get block time
    wg1.Go(func() error {
        header, err := t.ec.HeaderByNumber(context.Background(), opts.BlockNumber)
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
    blockEpoch := eth2.EpochAt(eth2Config, blockTime)
    if blockEpoch > beaconHead.Epoch {
        return []minipoolBalanceDetails{}, fmt.Errorf("Epoch %d at block %s is higher than current epoch %d", blockEpoch, opts.BlockNumber.String(), beaconHead.Epoch)
    }

    // Get minipool validator statuses
    validators, err := rp.GetMinipoolValidators(t.rp, t.bc, addresses, opts, &beacon.ValidatorStatusOptions{Epoch: blockEpoch})
    if err != nil {
        return []minipoolBalanceDetails{}, err
    }

    // Load details in batches
    details := make([]minipoolBalanceDetails, len(addresses))
    for bsi := 0; bsi < len(addresses); bsi += MinipoolBalanceDetailsBatchSize {

        // Get batch start & end index
        msi := bsi
        mei := bsi + MinipoolBalanceDetailsBatchSize
        if mei > len(addresses) { mei = len(addresses) }

        // Log
        //t.log.Printlnf("Calculating balances for minipools %d - %d of %d...", msi + 1, mei, len(addresses))

        // Load details
        var wg errgroup.Group
        for mi := msi; mi < mei; mi++ {
            mi := mi
            wg.Go(func() error {
                address := addresses[mi]
                validator := validators[address]
                mpDetails, err := t.getMinipoolBalanceDetails(address, opts, validator, eth2Config, blockEpoch)
                if err == nil { details[mi] = mpDetails }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            return []minipoolBalanceDetails{}, err
        }

    }

    // Return
    return details, nil

}


// Get minipool balance details
func (t *submitNetworkBalances) getMinipoolBalanceDetails(minipoolAddress common.Address, opts *bind.CallOpts, validator beacon.ValidatorStatus, eth2Config beacon.Eth2Config, blockEpoch uint64) (minipoolBalanceDetails, error) {

    // Create minipool
    mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
    if err != nil {
        return minipoolBalanceDetails{}, err
    }

    // Data
    var wg errgroup.Group
    var status types.MinipoolStatus
    var nodeFee float64
    var nodeDepositBalance *big.Int
    var userDepositBalance *big.Int
    var userDepositTime uint64
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
        nodeDepositBalance, err = mp.GetNodeDepositBalance(opts)
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
        withdrawalProcessed, err = minipool.GetMinipoolWithdrawalProcessed(t.rp, minipoolAddress, opts)
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

    // Use user deposit balance if validator not yet active on beacon chain at block
    if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
        return minipoolBalanceDetails{
            UserBalance: userDepositBalance,
        }, nil
    }

    // Get start epoch for node balance calculation
    startEpoch := eth2.EpochAt(eth2Config, userDepositTime)
    if startEpoch < validator.ActivationEpoch {
        startEpoch = validator.ActivationEpoch
    } else if startEpoch > blockEpoch {
        startEpoch = blockEpoch
    }

    // Get validator activation balance
    activationBalanceWei := new(big.Int)
    activationBalanceWei.Add(nodeDepositBalance, userDepositBalance)
    activationBalance := eth.WeiToGwei(activationBalanceWei)

    // Calculate approximate validator balance at start epoch & validator balance at block
    startBalance := eth.GweiToWei(activationBalance + (float64(validator.Balance) - activationBalance) * float64(startEpoch - validator.ActivationEpoch) / float64(blockEpoch - validator.ActivationEpoch))
    blockBalance := eth.GweiToWei(float64(validator.Balance))

    // Get node & user balance at block
    nodeBalance, err := minipool.GetMinipoolNodeRewardAmount(t.rp, nodeFee, userDepositBalance, startBalance, blockBalance, opts)
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
func (t *submitNetworkBalances) submitBalances(balances networkBalances) error {

    // Log
    t.log.Printlnf("Submitting network balances for block %d...", balances.Block)

    // Calculate total ETH balance
    totalEth := big.NewInt(0)
    totalEth.Add(totalEth, balances.DepositPool)
    totalEth.Add(totalEth, balances.MinipoolsTotal)
    totalEth.Add(totalEth, balances.RETHContract)

    // Get transactor
    opts, err := t.w.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Submit balances
    if _, err := network.SubmitBalances(t.rp, balances.Block, totalEth, balances.MinipoolsStaking, balances.RETHSupply, opts); err != nil {
        return err
    }

    // Log
    t.log.Printlnf("Successfully submitted network balances for block %d.", balances.Block)

    // Return
    return nil

}

