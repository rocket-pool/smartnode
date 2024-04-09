package pdao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
	"github.com/urfave/cli/v2"
)

var scInviteIdFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "id",
	Aliases: []string{"i"},
	Usage:   "A descriptive ID of the entity being invited",
}

var scInviteAddressFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "address",
	Aliases: []string{"a"},
	Usage:   "The address of the entity being invited",
}

func proposeSecurityCouncilInvite(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the ID
	id := c.String(scInviteIdFlag.Name)
	if id == "" {
		id = cliutils.Prompt("Please enter an ID for the member you'd like to invite: (no spaces)", "^\\S+$", "Invalid ID")
	}
	id, err = utils.ValidateDaoMemberID("id", id)
	if err != nil {
		return err
	}

	// Get the address
	addressString := c.String(scInviteAddressFlag.Name)
	if addressString == "" {
		addressString = cliutils.Prompt("Please enter the member's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	address, err := input.ValidateAddress("address", addressString)
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.InviteToSecurityCouncil(id, address)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("Cannot submit proposal for invitation:")
		if response.Data.InsufficientRpl {
			fmt.Printf("You do not have enough unlocked RPL (proposals require locking %.6f RPL, but you only have %.6f RPL staked and unlocked).", eth.WeiToEth(response.Data.ProposalBond), eth.WeiToEth(big.NewInt(0).Sub(response.Data.StakedRpl, response.Data.LockedRpl)))
		}
		if response.Data.MemberAlreadyExists {
			fmt.Println("The address is already part of the security council.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to propose inviting %s (%s) to the security council?", id, address.Hex()),
		"security council invitation proposal",
		"Proposing invitation to security council...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil
}
