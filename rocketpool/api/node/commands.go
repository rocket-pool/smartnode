package node

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the node",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's status",
                UsageText: "rocketpool api node status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getStatus(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "sync",
                Aliases:   []string{"y"},
                Usage:     "Get the sync progress of the eth1 and eth2 clients",
                UsageText: "rocketpool api node sync",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getSyncProgress(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-register",
                Usage:     "Check whether the node can be registered with Rocket Pool",
                UsageText: "rocketpool api node can-register timezone-location",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canRegisterNode(c, timezoneLocation))
                    return nil

                },
            },
            cli.Command{
                Name:      "register",
                Aliases:   []string{"r"},
                Usage:     "Register the node with Rocket Pool",
                UsageText: "rocketpool api node register timezone-location",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(registerNode(c, timezoneLocation))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-set-withdrawal-address",
                Usage:     "Checks if the node can set its withdrawal address",
                UsageText: "rocketpool api node can-set-withdrawal-address address confirm",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    withdrawalAddress, err := cliutils.ValidateAddress("withdrawal address", c.Args().Get(0))
                    if err != nil { return err }

                    confirm, err := cliutils.ValidateBool("confirm", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canSetWithdrawalAddress(c, withdrawalAddress, confirm))
                    return nil

                },
            },
            cli.Command{
                Name:      "set-withdrawal-address",
                Aliases:   []string{"w"},
                Usage:     "Set the node's withdrawal address",
                UsageText: "rocketpool api node set-withdrawal-address address confirm",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    withdrawalAddress, err := cliutils.ValidateAddress("withdrawal address", c.Args().Get(0))
                    if err != nil { return err }

                    confirm, err := cliutils.ValidateBool("confirm", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(setWithdrawalAddress(c, withdrawalAddress, confirm))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-confirm-withdrawal-address",
                Usage:     "Checks if the node can confirm its withdrawal address",
                UsageText: "rocketpool api node can-confirm-withdrawal-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canConfirmWithdrawalAddress(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "confirm-withdrawal-address",
                Usage:     "Confirms the node's withdrawal address if it was set back to the node address",
                UsageText: "rocketpool api node confirm-withdrawal-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(confirmWithdrawalAddress(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-set-timezone",
                Usage:     "Checks if the node can set its timezone location",
                UsageText: "rocketpool api node can-set-timezone timezone-location",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canSetTimezoneLocation(c, timezoneLocation))
                    return nil

                },
            },
            cli.Command{
                Name:      "set-timezone",
                Aliases:   []string{"t"},
                Usage:     "Set the node's timezone location",
                UsageText: "rocketpool api node set-timezone timezone-location",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(setTimezoneLocation(c, timezoneLocation))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-swap-rpl",
                Usage:     "Check whether the node can swap old RPL for new RPL",
                UsageText: "rocketpool api node can-swap-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeSwapRpl(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "swap-rpl-approve-rpl",
                Aliases:   []string{"p1"},
                Usage:     "Approve fixed-supply RPL for swapping to new RPL",
                UsageText: "rocketpool api node swap-rpl-approve-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(approveFsRpl(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "wait-and-swap-rpl",
                Aliases:   []string{"p2"},
                Usage:     "Swap old RPL for new RPL, waiting for the approval TX hash to be mined first",
                UsageText: "rocketpool api node wait-and-swap-rpl amount tx-hash",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
                    if err != nil { return err }
                    hash, err := cliutils.ValidateTxHash("swap amount", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(waitForApprovalAndSwapFsRpl(c, amountWei, hash))
                    return nil

                },
            },
            cli.Command{
                Name:      "get-swap-rpl-approval-gas",
                Usage:     "Estimate the gas cost of legacy RPL interaction approval",
                UsageText: "rocketpool api node get-swap-rpl-approval-gas",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("approve amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(getSwapApprovalGas(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "swap-rpl-allowance",
                Usage:     "Get the node's legacy RPL allowance for new RPL contract",
                UsageText: "rocketpool api node swap-allowance-rpl",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(allowanceFsRpl(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "swap-rpl",
                Aliases:   []string{"p3"},
                Usage:     "Swap old RPL for new RPL",
                UsageText: "rocketpool api node swap-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(swapRpl(c, amountWei))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-stake-rpl",
                Usage:     "Check whether the node can stake RPL",
                UsageText: "rocketpool api node can-stake-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("stake amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeStakeRpl(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "stake-rpl-approve-rpl",
                Aliases:   []string{"k1"},
                Usage:     "Approve RPL for staking against the node",
                UsageText: "rocketpool api node stake-rpl-approve-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("stake amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(approveRpl(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "wait-and-stake-rpl",
                Aliases:   []string{"k2"},
                Usage:     "Stake RPL against the node, waiting for approval tx-hash to be mined first",
                UsageText: "rocketpool api node wait-and-stake-rpl amount tx-hash",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("stake amount", c.Args().Get(0))
                    if err != nil { return err }
                    hash, err := cliutils.ValidateTxHash("tx-hash", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(waitForApprovalAndStakeRpl(c, amountWei, hash))
                    return nil

                },
            },
            cli.Command{
                Name:      "get-stake-rpl-approval-gas",
                Usage:     "Estimate the gas cost of new RPL interaction approval",
                UsageText: "rocketpool api node get-stake-rpl-approval-gas",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("approve amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(getStakeApprovalGas(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "stake-rpl-allowance",
                Usage:     "Get the node's RPL allowance for the staking contract",
                UsageText: "rocketpool api node stake-allowance-rpl",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(allowanceRpl(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "stake-rpl",
                Aliases:   []string{"k3"},
                Usage:     "Stake RPL against the node",
                UsageText: "rocketpool api node stake-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("stake amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(stakeRpl(c, amountWei))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-withdraw-rpl",
                Usage:     "Check whether the node can withdraw staked RPL",
                UsageText: "rocketpool api node can-withdraw-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeWithdrawRpl(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "withdraw-rpl",
                Aliases:   []string{"i"},
                Usage:     "Withdraw RPL staked against the node",
                UsageText: "rocketpool api node withdraw-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeWithdrawRpl(c, amountWei))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-deposit",
                Usage:     "Check whether the node can make a deposit",
                UsageText: "rocketpool api node can-deposit amount min-fee",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidateDepositWeiAmount("deposit amount", c.Args().Get(0))
                    if err != nil { return err }
                    minNodeFee, err := cliutils.ValidateFraction("minimum node fee", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeDeposit(c, amountWei, minNodeFee))
                    return nil

                },
            },
            cli.Command{
                Name:      "deposit",
                Aliases:   []string{"d"},
                Usage:     "Make a deposit and create a minipool",
                UsageText: "rocketpool api node deposit amount min-fee",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidateDepositWeiAmount("deposit amount", c.Args().Get(0))
                    if err != nil { return err }
                    minNodeFee, err := cliutils.ValidateFraction("minimum node fee", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeDeposit(c, amountWei, minNodeFee))
                    return nil

                },
            },
            cli.Command{
                Name:      "get-minipool-address",
                Aliases:   []string{"m"},
                Usage:     "Wait for a deposit to complete and get the resulting minipool address",
                UsageText: "rocketpool api node get-minipool-address tx-hash",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    hash, err := cliutils.ValidateTxHash("tx-hash", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(getMinipoolAddress(c, hash))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-send",
                Usage:     "Check whether the node can send ETH or tokens to an address",
                UsageText: "rocketpool api node can-send amount token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("send amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeSend(c, amountWei, token))
                    return nil

                },
            },
            cli.Command{
                Name:      "send",
                Aliases:   []string{"n"},
                Usage:     "Send ETH or tokens from the node account to an address",
                UsageText: "rocketpool api node send amount token to",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 3); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("send amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }
                    toAddress, err := cliutils.ValidateAddress("to address", c.Args().Get(2))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeSend(c, amountWei, token, toAddress))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-burn",
                Usage:     "Check whether the node can burn tokens for ETH",
                UsageText: "rocketpool api node can-burn amount token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("burn amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateBurnableTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeBurn(c, amountWei, token))
                    return nil

                },
            },
            cli.Command{
                Name:      "burn",
                Aliases:   []string{"b"},
                Usage:     "Burn tokens for ETH",
                UsageText: "rocketpool api node burn amount token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("burn amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateBurnableTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeBurn(c, amountWei, token))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-claim-rpl-rewards",
                Usage:     "Check whether the node has RPL rewards available to claim",
                UsageText: "rocketpool api node can-claim-rpl-rewards",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeClaimRpl(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "claim-rpl-rewards",
                Usage:     "Claim available RPL rewards",
                UsageText: "rocketpool api node claim-rpl-rewards",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(nodeClaimRpl(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "rewards",
                Usage:     "Get RPL rewards info",
                UsageText: "rocketpool api node rewards",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getRewards(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "deposit-contract-info",
                Usage:     "Get information about the deposit contract specified by Rocket Pool and the Beacon Chain client",
                UsageText: "rocketpool api node deposit-contract-info",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getDepositContractInfo(c))
                    return nil

                },
            },

        },
    })
}

