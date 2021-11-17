package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v2/beacon-chain/core/signing"
	tndao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	tnsettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	prdeposit "github.com/prysmaticlabs/prysm/v2/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v2/proto/prysm/v1alpha1"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)


type minipoolCreated struct {
    Minipool common.Address
    Node common.Address
    Time *big.Int
}


func canNodeDeposit(c *cli.Context, amountWei *big.Int, minNodeFee float64, salt *big.Int) (*api.CanNodeDepositResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Get eth2 config
    eth2Config, err := bc.GetEth2Config()
    if err != nil {
        return nil, err
    }

    // Response
    response := api.CanNodeDepositResponse{}

    // Check if amount is zero
    amountIsZero := (amountWei.Cmp(big.NewInt(0)) == 0)

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Adjust the salt
    if salt.Cmp(big.NewInt(0)) == 0 {
        nonce, err := ec.NonceAt(context.Background(), nodeAccount.Address, nil)
        if err != nil {
            return nil, err
        }
        salt.SetUint64(nonce)
    }

    // Data
    var wg1 errgroup.Group
    var isTrusted bool
    var minipoolCount uint64
    var minipoolLimit uint64
    var minipoolAddress common.Address

    // Check node balance
    wg1.Go(func() error {
        ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
        if err == nil {
            response.InsufficientBalance = (amountWei.Cmp(ethBalanceWei) > 0)
        }
        return err
    })

    // Check node deposits are enabled
    wg1.Go(func() error {
        depositEnabled, err := protocol.GetNodeDepositEnabled(rp, nil)
        if err == nil {
            response.DepositDisabled = !depositEnabled
        }
        return err
    })

    // Get trusted status
    wg1.Go(func() error {
        var err error
        isTrusted, err = tndao.GetMemberExists(rp, nodeAccount.Address, nil)
        return err
    })

    // Get node staking information
    wg1.Go(func() error {
        var err error
        minipoolCount, err = minipool.GetNodeMinipoolCount(rp, nodeAccount.Address, nil)
        return err
    })
    wg1.Go(func() error {
        var err error
        minipoolLimit, err = node.GetNodeMinipoolLimit(rp, nodeAccount.Address, nil)
        return err
    })

    // Get consensus status
    wg1.Go(func() error {
        var err error
        inConsensus, err := network.InConsensus(rp, nil)
        response.InConsensus = inConsensus
        return err
    })

    // Get gas estimate
    wg1.Go(func() error {
        opts, err := w.GetNodeAccountTransactor()
        if err != nil { 
            return err 
        }
        opts.Value = amountWei

        // Get the deposit type
        depositType, err := node.GetDepositType(rp, amountWei, nil)
        if err != nil {
            return err
        }

        // Get the next validator key
        validatorKey, err := w.GetNextValidatorKey()
        if err != nil {
            return err
        }

        // Get the next minipool address and withdrawal credentials
        minipoolAddress, err = utils.GenerateAddress(rp, nodeAccount.Address, depositType, salt, nil)
        if err != nil {
            return err
        }
        withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, minipoolAddress, nil)
        if err != nil {
            return err
        }

        // Get validator deposit data and associated parameters
        depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config)
        if err != nil {
            return err
        }
        pubKey := rptypes.BytesToValidatorPubkey(depositData.PublicKey)
        signature := rptypes.BytesToValidatorSignature(depositData.Signature)

        // Do a final sanity check
        err = validateDepositInfo(eth2Config, opts.Value, pubKey, withdrawalCredentials, signature)
        if err != nil {
            return fmt.Errorf("Your deposit failed the validation safety check: %w\n" + 
                "For your safety, this deposit will not be submitted and your ETH will not be staked.\n" +
                "PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS.", err)
        }

        // Run the deposit gas estimator
        gasInfo, err := node.EstimateDepositGas(rp, minNodeFee, pubKey, signature, depositDataRoot, salt, minipoolAddress, opts)
        if err == nil {
            response.GasInfo = gasInfo
        }
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return nil, err
    }

    // Check data
    response.InsufficientRplStake = (minipoolCount >= minipoolLimit)
    response.MinipoolAddress = minipoolAddress
    response.InvalidAmount = (!isTrusted && amountIsZero)

    // Check oracle node unbonded minipool limit
    if isTrusted && amountIsZero {

        // Data
        var wg2 errgroup.Group
        var unbondedMinipoolCount uint64
        var unbondedMinipoolsMax uint64

        // Get unbonded minipool details
        wg2.Go(func() error {
            var err error
            unbondedMinipoolCount, err = tndao.GetMemberUnbondedValidatorCount(rp, nodeAccount.Address, nil)
            return err
        })
        wg2.Go(func() error {
            var err error
            unbondedMinipoolsMax, err = tnsettings.GetMinipoolUnbondedMax(rp, nil)
            return err
        })

        // Wait for data
        if err := wg2.Wait(); err != nil {
            return nil, err
        }

        // Check unbonded minipool limit
        response.UnbondedMinipoolsAtMax = (unbondedMinipoolCount >= unbondedMinipoolsMax)

    }

    // Update & return response
    response.CanDeposit = !(response.InsufficientBalance || response.InsufficientRplStake || response.InvalidAmount || response.UnbondedMinipoolsAtMax || response.DepositDisabled || !response.InConsensus)
    return &response, nil

}


func nodeDeposit(c *cli.Context, amountWei *big.Int, minNodeFee float64, salt *big.Int) (*api.NodeDepositResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Get eth2 config
    eth2Config, err := bc.GetEth2Config()
    if err != nil {
        return nil, err
    }

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Response
    response := api.NodeDepositResponse{}

    // Adjust the salt
    if salt.Cmp(big.NewInt(0)) == 0 {
        nonce, err := ec.NonceAt(context.Background(), nodeAccount.Address, nil)
        if err != nil {
            return nil, err
        }
        salt.SetUint64(nonce)
    }

    // Make sure ETH2 is on the correct chain
    depositContractInfo, err := getDepositContractInfo(c)
    if err != nil {
        return nil, err
    }
    if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
       depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
            return nil, fmt.Errorf("Beacon network mismatch! Expected %s on chain %d, but beacon is using %s on chain %d.",
                            depositContractInfo.RPDepositContract.Hex(),
                            depositContractInfo.RPNetwork,
                            depositContractInfo.BeaconDepositContract.Hex(),
                            depositContractInfo.BeaconNetwork)
    }

    // Get the scrub period
    scrubPeriodUnix, err := trustednode.GetScrubPeriod(rp, nil)
    if err != nil {
        return nil, err
    }
    scrubPeriod := time.Duration(scrubPeriodUnix) * time.Second
    response.ScrubPeriod = scrubPeriod

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }
    opts.Value = amountWei

    // Get the deposit type
    depositType, err := node.GetDepositType(rp, amountWei, nil)
    if err != nil {
        return nil, err
    }

    // Create and save a new validator key
    validatorKey, err := w.CreateValidatorKey()
    if err != nil {
        return nil, err
    }

    // Get the next minipool address and withdrawal credentials
    minipoolAddress, err := utils.GenerateAddress(rp, nodeAccount.Address, depositType, salt, nil)
    if err != nil {
        return nil, err
    }
    withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, minipoolAddress, nil)
    if err != nil {
        return nil, err
    }

    // Get validator deposit data and associated parameters
    depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config)
    if err != nil {
        return nil, err
    }
    pubKey := rptypes.BytesToValidatorPubkey(depositData.PublicKey)
    signature := rptypes.BytesToValidatorSignature(depositData.Signature)

    // Make sure a validator with this pubkey doesn't already exist
    status, err := bc.GetValidatorStatus(pubKey, nil)
    if err != nil {
        return nil, fmt.Errorf("Error checking for existing validator status: %w\nYour funds have not been deposited for your own safety.", err)
    }
    if status.Exists {
        return nil, fmt.Errorf("**** ALERT ****\n" +
            "Your minipool %s has the following as a validator pubkey:\n\t%s\n" +
            "This key is already in use by validator %d on the Beacon chain!\n"  +
            "Rocket Pool will not allow you to deposit this validator for your own safety so you do not get slashed.\n" +
            "PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS.\n" +
            "***************\n", minipoolAddress.Hex(), pubKey.Hex(), status.Index);
    }

    // Do a final sanity check
    err = validateDepositInfo(eth2Config, opts.Value, pubKey, withdrawalCredentials, signature)
    if err != nil {
        return nil, fmt.Errorf("Your deposit failed the validation safety check: %w\n" + 
            "For your safety, this deposit will not be submitted and your ETH will not be staked.\n" +
            "PLEASE REPORT THIS TO THE ROCKET POOL DEVELOPERS.", err)
    }

    // Override the provided pending TX if requested 
    err = eth1.CheckForNonceOverride(c, opts)
    if err != nil {
        return nil, fmt.Errorf("Error checking for nonce override: %w", err)
    }

    // Deposit
    hash, err := node.Deposit(rp, minNodeFee, pubKey, signature, depositDataRoot, salt, minipoolAddress, opts)
    if err != nil {
        return nil, err
    }

    // Save wallet
    if err := w.Save(); err != nil {
        return nil, err
    }

    response.TxHash = hash
    response.MinipoolAddress = minipoolAddress
    response.ValidatorPubkey = pubKey

    // Return response
    return &response, nil

}


func validateDepositInfo(eth2Config beacon.Eth2Config, depositAmountWei *big.Int, pubkey rptypes.ValidatorPubkey, withdrawalCredentials common.Hash, signature rptypes.ValidatorSignature) (error) {

    // Get the deposit domain based on the eth2 config 
    depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
    if err != nil {
        return err
    }
    
    // Convert the deposit amount to gwei
    weiPerGwei := big.NewInt(int64(eth.WeiPerGwei))
    depositAmountGwei := big.NewInt(0).Div(depositAmountWei, weiPerGwei)
    
    // Create the deposit struct
    depositData := new(ethpb.Deposit_Data)
    depositData.Amount = depositAmountGwei.Uint64()
    depositData.PublicKey = pubkey.Bytes()
    depositData.WithdrawalCredentials = withdrawalCredentials.Bytes()
    depositData.Signature = signature.Bytes()

    // Validate the signature
    err = prdeposit.VerifyDepositSignature(depositData, depositDomain)
    return err

}

