package odao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/node"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

var joinSwapFlag *cli.BoolFlag = &cli.BoolFlag{
	Name:    "swap",
	Aliases: []string{"s"},
	Usage:   "Automatically confirm swapping old RPL before joining",
}

func join(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get node status
	status, err := rp.Api.Node.Status()
	if err != nil {
		return err
	}

	// Check for fixed-supply RPL balance
	if status.Data.NodeBalances.Fsrpl.Cmp(big.NewInt(0)) > 0 {
		// Confirm swapping RPL
		if c.Bool(joinSwapFlag.Name) || utils.Confirm(fmt.Sprintf("The node has a balance of %.6f old RPL. Would you like to swap it for new RPL before transferring your bond?", math.RoundDown(eth.WeiToEth(status.Data.NodeBalances.Fsrpl), 6))) {
			err = node.SwapRpl(c, rp, status.Data.NodeBalances.Fsrpl)
			if err != nil {
				return fmt.Errorf("error swapping legacy RPL: %w", err)
			}
		}
	}

	// Build the TX
	response, err := rp.Api.ODao.Join()
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanJoin {
		fmt.Println("Cannot join the oracle DAO:")
		if response.Data.ProposalExpired {
			fmt.Println("The proposal for you to join the Oracle DAO does not exist or has expired.")
		}
		if response.Data.AlreadyMember {
			fmt.Println("The node is already a member of the Oracle DAO.")
		}
		if response.Data.InsufficientRplBalance {
			fmt.Println("The node does not have enough RPL to pay the RPL bond.")
		}
		return nil
	}

	// Check if approval is required first
	if response.Data.ApproveTxInfo != nil {
		// Run the Approve TX
		validated, err := tx.HandleTx(c, rp, response.Data.ApproveTxInfo,
			"Do you want to let the Oracle DAO manager interact with your RPL? This is required to post your bond in order to join it.",
			"approving RPL for bond",
			"Approving RPL for joining the Oracle DAO...",
		)
		if err != nil {
			return err
		}
		if validated {
			fmt.Println("Successfully approved bond access to RPL.")
		}

		// Build the Join TX once approval is done
		response, err = rp.Api.ODao.Join()
		if err != nil {
			return err
		}
	}

	// Run the Join TX
	validated, err := tx.HandleTx(c, rp, response.Data.JoinTxInfo,
		"Are you sure you want to join the oracle DAO? Your RPL bond will be locked until you leave.",
		"joining Oracle DAO",
		"Joining the Oracle DAO...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully joined the Oracle DAO.")
	return nil
}
