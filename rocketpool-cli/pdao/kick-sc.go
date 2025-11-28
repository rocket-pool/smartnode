package pdao

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func proposeSecurityCouncilKick(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the list of members
	membersResponse, err := rp.SecurityMembers()
	if err != nil {
		return fmt.Errorf("error getting list of security council members: %w", err)
	}

	// Get the address
	var addresses []common.Address
	addressesString := c.String("addresses")
	if addressesString == "" {
		// Print the members
		if len(membersResponse.Members) == 0 {
			fmt.Printf("There are no security council members.")
			return nil
		}
		for i, member := range membersResponse.Members {
			fmt.Printf("%d: %s (%s), joined %s\n", i+1, member.ID, member.Address, time.Unix(int64(member.JoinedTime), 0))
		}

		for {
			indexSelection := prompt.Prompt("Which members would you like to kick? Use a comma separated list (such as '1,2,3') to select multiple members.", "^\\d+(,\\d+)*$", "Invalid index selection")
			elements := strings.Split(indexSelection, ",")
			allValid := true
			indices := []uint64{}
			seenIndices := map[uint64]bool{}

			for _, element := range elements {
				index, err := strconv.ParseUint(element, 0, 64)
				if err != nil {
					allValid = false
					fmt.Printf("'%s' is not a valid index.\n", element)
					break
				}

				if index < 1 || index > uint64(len(membersResponse.Members)) {
					allValid = false
					fmt.Printf("'%s' must be between 1 and %d.\n", element, len(membersResponse.Members))
					break
				}

				// Ignore duplicates
				_, exists := seenIndices[index]
				if !exists {
					indices = append(indices, index)
					seenIndices[index] = true
				}
			}
			if allValid {
				for _, index := range indices {
					addresses = append(addresses, membersResponse.Members[index-1].Address)
				}
				break
			}
		}
	} else {
		addresses, err = cliutils.ValidateAddresses("addresses", addressesString)
		if err != nil {
			return err
		}
	}

	var hash common.Hash
	if len(addresses) == 1 {
		address := addresses[0]
		// Get the ID
		var id *string
		for _, member := range membersResponse.Members {
			if member.Address == address {
				id = &member.ID
				break
			}
		}
		if id == nil {
			return fmt.Errorf("address %s is not on the security council", address.Hex())
		}

		// Check submissions
		canResponse, err := rp.PDAOCanProposeKickFromSecurityCouncil(address)
		if err != nil {
			return err
		}
		if !canResponse.CanPropose {
			fmt.Println("Cannot propose kicking member:")
			if canResponse.IsRplLockingDisallowed {
				fmt.Println("Please enable RPL locking using the command 'rocketpool node allow-rpl-locking' to raise proposals.")
			}
			return nil
		}

		// Assign max fee
		err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
		if err != nil {
			return err
		}

		// Prompt for confirmation
		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to propose kicking %s (%s) from the security council?", *id, address.Hex()))) {
			fmt.Println("Cancelled.")
			return nil
		}

		// Submit
		response, err := rp.PDAOProposeKickFromSecurityCouncil(address, canResponse.BlockNumber)
		if err != nil {
			return err
		}
		hash = response.TxHash
	} else {
		ids := make([]string, len(addresses))
		for i, address := range addresses {
			// Get the ID
			var id *string
			for _, member := range membersResponse.Members {
				if member.Address == address {
					id = &member.ID
					break
				}
			}
			if id == nil {
				return fmt.Errorf("address %s is not on the security council", address.Hex())
			}
			ids[i] = *id
		}

		// Check submissions
		canResponse, err := rp.PDAOCanProposeKickMultiFromSecurityCouncil(addresses)
		if err != nil {
			return err
		}

		// Assign max fee
		err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
		if err != nil {
			return err
		}

		// Create the kick string
		var kickString strings.Builder
		for i, address := range addresses {
			kickString.WriteString(fmt.Sprintf("\t- %s (%s)\n", ids[i], address.Hex()))
		}

		// Prompt for confirmation
		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to propose kicking these members from the security council?\n%s", kickString.String()))) {
			fmt.Println("Cancelled.")
			return nil
		}

		// Submit
		response, err := rp.PDAOProposeKickMultiFromSecurityCouncil(addresses, canResponse.BlockNumber)
		if err != nil {
			return err
		}
		hash = response.TxHash
	}

	fmt.Printf("Proposing kick from security council...\n")
	cliutils.PrintTransactionHash(rp, hash)
	if _, err = rp.WaitForTransaction(hash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil

}
