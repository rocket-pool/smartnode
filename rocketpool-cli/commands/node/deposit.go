package node

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

// Config
const (
	amountFlag                string  = "amount"
	maxSlippageFlag           string  = "max-slippage"
	saltFlag                  string  = "salt"
	defaultMaxNodeFeeSlippage float64 = 0.01 // 1% below current network fee
	depositWarningMessage     string  = "NOTE: by creating a new minipool, your node will automatically initialize voting power to itself. If you would like to delegate your on-chain voting power, you should run the command `rocketpool pdao initialize-voting` before creating a new minipool."
)

type deposit struct {
	amountWei  *big.Int
	minNodeFee float64
	salt       *big.Int
}

// Common deposit flow.
// If soloConversionPubkey is passed, this is treated as a vacant minipool
func newDepositPrompts(c *cli.Context, rp *client.Client, soloConversionPubkey *beacon.ValidatorPubkey) (*deposit, error) {

	// Make sure Beacon is on the correct chain
	depositContractInfo, err := rp.Api.Network.GetDepositContractInfo(true)
	if err != nil {
		return nil, err
	}
	if depositContractInfo.Data.PrintMismatch() {
		return nil, nil
	}

	// Get the node's registration status
	smoothie, err := rp.Api.Node.SetSmoothingPoolRegistrationState(true)
	if err != nil {
		return nil, err
	}
	if smoothie.Data.NodeRegistered {
		fmt.Println("Your node is currently opted into the smoothing pool.")
	} else if !smoothie.Data.NodeRegistered {
		fmt.Println("Your node is not opted into the smoothing pool.")
	}
	fmt.Println()

	// Post a warning about ETH only minipools
	if !(c.Bool("yes") || utils.Confirm(fmt.Sprintf("%sNOTE: We’re excited to announce that newly launched Saturn 0 minipools will feature a commission structure ranging from 5%% to 14%%.\n\n- 5%% base commission\n- 5%% dynamic commission boost until Saturn 1\n- Up to 4%% boost for staked RPL valued at ≥10%% of borrowed ETH\n\n- Smoothing pool participation is required to benefit from dynamic commission\n- Dynamic commission starts when reward tree v10 is released (currently in development)\n- Dynamic commission ends soon after Saturn 1 is released\n\nNewly launched minipools with no RPL staked receive 10%% commission while newly launched minipools with ≥10%% of borrowed ETH staked receive 14%% commission.\n\nTo learn more about Saturn 0 and how it affects newly launched minipools, visit: https://rpips.rocketpool.net/tokenomics-explainers/005-rework-prelude%s\nWould you like to continue?", terminal.ColorYellow, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil, err
	}

	// Post a final warning about the dynamic comission boost
	if !smoothie.Data.NodeRegistered {
		if !(c.Bool("yes") || utils.Confirm(fmt.Sprintf("%sWARNING: Your node is not opted into the smoothing pool, which means newly launched minipools will not benefit from the 5-9%% dynamic commission boost. You can join the smoothing pool using: 'rocketpool node join-smoothing-pool'.\n%sAre you sure you'd like to continue without opting into the smoothing pool?", terminal.ColorRed, terminal.ColorReset))) {
			fmt.Println("Cancelled.")
			return nil, err
		}
	}

	// If hotfix is live and voting isn't initialized, display a warning
	err = warnIfVotingUninitialized(rp, c, depositWarningMessage)
	if err != nil {
		return nil, err
	}

	// Check if the fee distributor has been initialized
	feeDistributorResponse, err := rp.Api.Node.InitializeFeeDistributor()
	if err != nil {
		return nil, err
	}
	if !feeDistributorResponse.Data.IsInitialized {
		fmt.Println("Your fee distributor has not been initialized yet so you cannot create a new minipool.\nPlease run `rocketpool node initialize-fee-distributor` to initialize it first.")
		return nil, nil
	}

	// Post a warning about fee distribution
	if !(c.Bool("yes") || utils.Confirm(fmt.Sprintf("%sNOTE: by creating a new minipool, your node will automatically claim and distribute any balance you have in your fee distributor contract. If you don't want to claim your balance at this time, you should not create a new minipool.%s\nWould you like to continue?", terminal.ColorYellow, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil, nil
	}

	if soloConversionPubkey != nil {
		// Print a notification about the pubkey
		fmt.Printf("You are about to convert the solo staker %s into a Rocket Pool minipool. This will convert your 32 ETH deposit into an 8 ETH deposit, and convert the remaining 24 ETH into a deposit from the Rocket Pool staking pool. The staking pool portion will be credited to your node's account, allowing you to create more validators without depositing additional ETH onto the Beacon Chain. Your excess balance (your existing Beacon rewards) will be preserved and not shared with the pool stakers.\n", soloConversionPubkey.Hex())
		fmt.Println()
		fmt.Println("Please thoroughly read our documentation at https://docs.rocketpool.net/guides/atlas/solo-staker-migration.html to learn about the process and its implications.")
		fmt.Println()
		fmt.Println("1. First, we'll create the new minipool.")
		fmt.Println("2. Next, we'll ask whether you want to import the validator's private key into your Smartnode's Validator Client, or keep running your own externally-managed validator.")
		fmt.Println("3. Finally, we'll help you migrate your validator's withdrawal credentials to the minipool address.")
		fmt.Println()
		fmt.Printf("%sNOTE: If you intend to use the credit balance to create additional validators, you will need to have enough RPL staked to support them.%s\n", terminal.ColorYellow, terminal.ColorReset)
		fmt.Println()
	}

	// Get deposit amount
	var amount float64
	if c.String(amountFlag) != "" {
		// Parse amount
		depositAmount, err := strconv.ParseFloat(c.String(amountFlag), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid deposit amount '%s': %w", c.String(amountFlag), err)
		}
		amount = depositAmount
	} else {
		// Post a warning about deposit size
		if !(c.Bool("yes") || utils.Confirm(fmt.Sprintf("%sNOTE: You are about to make an 8 ETH deposit.%s\nWould you like to continue?", terminal.ColorYellow, terminal.ColorReset))) {
			fmt.Println("Cancelled.")
			return nil, nil
		}
		amount = 8
	}

	amountWei := eth.EthToWei(amount)

	// Get network node fees
	nodeFeeResponse, err := rp.Api.Network.NodeFee()
	if err != nil {
		return nil, err
	}

	// Get minimum node fee
	var minNodeFee float64
	if c.String(maxSlippageFlag) == "auto" {
		// Use default max slippage
		minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.NodeFee) - defaultMaxNodeFeeSlippage
		if minNodeFee < eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee) {
			minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee)
		}
	} else if c.String(maxSlippageFlag) != "" {
		// Parse max slippage
		maxNodeFeeSlippagePerc, err := strconv.ParseFloat(c.String(maxSlippageFlag), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid maximum commission rate slippage '%s': %w", c.String(maxSlippageFlag), err)
		}
		maxNodeFeeSlippage := maxNodeFeeSlippagePerc / 100

		// Calculate min node fee
		minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.NodeFee) - maxNodeFeeSlippage
		if minNodeFee < eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee) {
			minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee)
		}
	} else {
		// Prompt for min node fee
		if nodeFeeResponse.Data.MinNodeFee.Cmp(nodeFeeResponse.Data.MaxNodeFee) == 0 {
			fmt.Printf("Your minipool will use the current base commission rate of %.2f%%.\n", eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee)*100)
			minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee)
		} else {
			minNodeFee = promptMinNodeFee(eth.WeiToEth(nodeFeeResponse.Data.NodeFee), eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee))
		}
	}

	// Get minipool salt
	var salt *big.Int
	if c.String(saltFlag) != "" {
		var success bool
		salt, success = big.NewInt(0).SetString(c.String(saltFlag), 0)
		if !success {
			return nil, fmt.Errorf("invalid minipool salt: %s", c.String(saltFlag))
		}
	} else {
		buffer := make([]byte, 32)
		_, err = rand.Read(buffer)
		if err != nil {
			return nil, fmt.Errorf("error generating random salt: %w", err)
		}
		salt = big.NewInt(0).SetBytes(buffer)
	}

	return &deposit{
		salt:       salt,
		minNodeFee: minNodeFee,
		amountWei:  amountWei,
	}, nil
}

func nodeDeposit(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	d, err := newDepositPrompts(c, rp, nil)
	if err != nil {
		return err
	}
	if d == nil {
		return nil
	}

	minNodeFee := d.minNodeFee
	amountWei := d.amountWei
	amount := eth.WeiToEth(amountWei)

	// Build the TX
	response, err := rp.Api.Node.Deposit(amountWei, minNodeFee, d.salt)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanDeposit {
		fmt.Println("Cannot make node deposit:")
		if response.Data.InsufficientBalanceWithoutCredit {
			nodeBalance := eth.WeiToEth(response.Data.NodeBalance)
			fmt.Printf("There is not enough ETH in the staking pool to use your credit balance (it needs at least 1 ETH but only has %.2f ETH) and you don't have enough ETH in your wallet (%.6f ETH) to cover the deposit amount yourself. If you want to continue creating a minipool, you will either need to wait for the staking pool to have more ETH deposited or add more ETH to your node wallet.", eth.WeiToEth(response.Data.DepositBalance), nodeBalance)
		}
		if response.Data.InsufficientBalance {
			nodeBalance := eth.WeiToEth(response.Data.NodeBalance)
			creditBalance := eth.WeiToEth(response.Data.CreditBalance)
			fmt.Printf("The node's balance of %.6f ETH and credit balance of %.6f ETH are not enough to create a minipool with a %.1f ETH bond.", nodeBalance, creditBalance, amount)
		}
		if response.Data.InsufficientRplStake {
			fmt.Printf("The node has not staked enough RPL to collateralize a new minipool with a bond of %d ETH (this also includes the RPL required to support any pending bond reductions).\n", int(amount))
		}
		if response.Data.InvalidAmount {
			fmt.Println("The deposit amount is invalid.")
		}
		if response.Data.UnbondedMinipoolsAtMax {
			fmt.Println("The node cannot create any more unbonded minipools.")
		}
		if response.Data.DepositDisabled {
			fmt.Println("Node deposits are currently disabled.")
		}
		return nil
	}

	// Print credit balance info
	fmt.Printf("You currently have %.2f ETH in your credit balance plus ETH staked on your behalf.\n", eth.WeiToEth(response.Data.CreditBalance))
	if response.Data.CreditBalance.Cmp(big.NewInt(0)) > 0 {
		if response.Data.CanUseCredit {
			// Get how much credit to use
			remainingAmount := big.NewInt(0).Sub(amountWei, response.Data.CreditBalance)
			if remainingAmount.Cmp(big.NewInt(0)) > 0 {
				fmt.Printf("This deposit will use all %.6f ETH from your credit balance plus ETH staked on your behalf and %.6f ETH from your node.\n\n", eth.WeiToEth(response.Data.CreditBalance), eth.WeiToEth(remainingAmount))
			} else {
				fmt.Printf("This deposit will use %.6f ETH from your credit balance plus ETH staked on your behalf and will not require any ETH from your node.\n\n", amount)
			}
		} else {
			fmt.Printf("%sNOTE: Your credit balance *cannot* currently be used to create a new minipool; there is not enough ETH in the staking pool to cover the initial deposit on your behalf (it needs at least 1 ETH but only has %.2f ETH).%s\nIf you want to continue creating this minipool now, you will have to pay for the full bond amount.\n\n", terminal.ColorYellow, eth.WeiToEth(response.Data.DepositBalance), terminal.ColorReset)

			// Prompt for confirmation
			if !(c.Bool("yes") || utils.Confirm("Would you like to continue?")) {
				fmt.Println("Cancelled.")
				return nil
			}
		}
	}

	// Print salt and minipool address info
	if c.String(saltFlag) != "" {
		fmt.Printf("Using custom salt %s, your minipool address will be %s.\n\n", c.String(saltFlag), response.Data.MinipoolAddress.Hex())
	}

	// Check to see if eth2 is synced
	syncResponse, err := rp.Api.Service.ClientStatus()
	if err != nil {
		fmt.Print(terminal.ColorRed)
		fmt.Println("**WARNING**: Can't verify the sync status of your consensus client.")
		fmt.Println("YOU WILL LOSE ETH if your minipool is activated before it is fully synced.")
		fmt.Printf("Reason: %v\n", err)
		fmt.Print(terminal.ColorReset)
	} else {
		if syncResponse.Data.BcManagerStatus.PrimaryClientStatus.IsSynced {
			fmt.Println("Your consensus client is synced, you may safely create a minipool.")
		} else if syncResponse.Data.BcManagerStatus.FallbackEnabled {
			if syncResponse.Data.BcManagerStatus.FallbackClientStatus.IsSynced {
				fmt.Println("Your fallback consensus client is synced, you may safely create a minipool.")
			} else {
				fmt.Print(terminal.ColorRed)
				fmt.Println("**WARNING**: neither your primary nor fallback consensus clients are fully synced.")
				fmt.Println("YOU WILL LOSE ETH if your minipool is activated before they are fully synced.")
				fmt.Print(terminal.ColorReset)
			}
		} else {
			fmt.Print(terminal.ColorRed)
			fmt.Println("**WARNING**: your primary consensus client is either not fully synced or offline and you do not have a fallback client configured.")
			fmt.Println("YOU WILL LOSE ETH if your minipool is activated before it is fully synced.")
			fmt.Print(terminal.ColorReset)
		}
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf(
			"You are about to deposit %.6f ETH to create a minipool with a minimum possible commission rate of %f%%.\n"+
				"%sARE YOU SURE YOU WANT TO DO THIS? Exiting this minipool and retrieving your capital cannot be done until your minipool has been *active* on the Beacon Chain for 256 epochs (approx. 27 hours).%s\n",
			math.RoundDown(eth.WeiToEth(amountWei), 6),
			minNodeFee*100,
			terminal.ColorYellow,
			terminal.ColorReset),
		"creating minipool",
		"Creating minipool...",
	)
	if err != nil {
		return err
	}
	if !validated {
		// Prompt for saving the key
		if c.Bool(utils.YesFlag.Name) || utils.Confirm("Would you like to save the private key for this validator to disk? You'll need to do this if you plan to submit that transaction later and want to be able to attest with this validator.") {
			_, err = rp.Api.Wallet.CreateValidatorKey(response.Data.ValidatorPubkey, response.Data.Index)
			if err != nil {
				return fmt.Errorf("error saving validator key to disk: %w", err)
			}
		}
		return nil
	}

	// Save the validator key to disk
	_, err = rp.Api.Wallet.CreateValidatorKey(response.Data.ValidatorPubkey, response.Data.Index)
	if err != nil {
		fmt.Printf("%sError saving validator key to disk: %s%s\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
		fmt.Println("Your deposit has not been submitted, and your ETH is still in your node wallet.")
		return nil
	}

	// Log & return
	fmt.Printf("The node deposit of %.6f ETH was made successfully!\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	fmt.Printf("Your new minipool's address is: %s\n", response.Data.MinipoolAddress)
	fmt.Printf("The validator pubkey is: %s\n\n", response.Data.ValidatorPubkey.HexWithPrefix())

	fmt.Println("Your minipool is now in Initialized status.")
	fmt.Println("Once the remaining ETH has been assigned to your minipool from the staking pool, it will move to Prelaunch status.")
	fmt.Printf("After that, it will move to Staking status once %s have passed.\n", response.Data.ScrubPeriod)
	fmt.Println("You can watch its progress using `rocketpool service logs node`.")

	fmt.Println()

	fmt.Println("Your Validator Client must be restarted in order to load the new validator key so it can begin attesting once it has been activated on the Beacon Chain.")
	if c.Bool(utils.YesFlag.Name) || utils.Confirm("Would you like to restart the Validator Client now?") {
		_, err := rp.Api.Service.RestartVc()
		if err != nil {
			fmt.Printf("%sWARNING: Error restarting Validator Client: %s%s\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
			fmt.Println("Please restart the Validator Client manually before your validator becomes active in order to load the new validator key.")
			fmt.Printf("%sIf you don't restart it, you will miss attestations and lose ETH!%s\n", terminal.ColorYellow, terminal.ColorReset)
		} else {
			fmt.Println("Successfully restarted the Validator Client. Your new validator key is now loaded.")
			return nil
		}
	} else {
		fmt.Println("Please restart the Validator Client manually before your validator becomes active in order to load the new validator key.")
		fmt.Printf("%sIf you don't restart it, you will miss attestations and lose ETH!%s\n", terminal.ColorYellow, terminal.ColorReset)
	}

	return nil
}
