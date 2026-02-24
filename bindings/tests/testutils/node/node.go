package node

import (
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	trustednodesettings "github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/rocket-pool/smartnode/bindings/tokens"

	"github.com/rocket-pool/smartnode/bindings/tests/testutils/accounts"
	rplutils "github.com/rocket-pool/smartnode/bindings/tests/testutils/tokens/rpl"
)

// Trusted node counter
var trustedNodeIndex = 0

// Register a trusted node
// NOTE: This function is commented out because trustednodedao.BootstrapMember is no longer available
// If you need this functionality, you'll need to update it to use the new API
/*
func RegisterTrustedNode(rp *rocketpool.RocketPool, ownerAccount *accounts.Account, trustedNodeAccount *accounts.Account) error {

	// Register node
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", trustedNodeAccount.GetTransactor()); err != nil {
		return err
	}

	// Bootstrap trusted node DAO member
	if _, err := trustednodedao.BootstrapMember(rp, fmt.Sprintf("tn%d", trustedNodeIndex), fmt.Sprintf("tn%d@rocketpool.net", trustedNodeIndex), trustedNodeAccount.Address, ownerAccount.GetTransactor()); err != nil {
		return err
	}

	// Mint trusted node RPL bond
	if err := MintTrustedNodeBond(rp, ownerAccount, trustedNodeAccount); err != nil {
		return err
	}

	// Join trusted node DAO
	if _, err := trustednodedao.Join(rp, trustedNodeAccount.GetTransactor()); err != nil {
		return err
	}

	// Increment trusted node counter & return
	trustedNodeIndex++
	return nil

}
*/

// Mint trusted node DAO RPL bond to a node account and approve it for spending
func MintTrustedNodeBond(rp *rocketpool.RocketPool, ownerAccount *accounts.Account, trustedNodeAccount *accounts.Account) error {

	// Get RPL bond amount
	rplBondAmount, err := trustednodesettings.GetRPLBond(rp, nil)
	if err != nil {
		return err
	}

	// Get RocketDAONodeTrustedActions contract address
	rocketDAONodeTrustedActionsAddress, err := rp.GetAddress("rocketDAONodeTrustedActions", nil)
	if err != nil {
		return err
	}

	// Mint RPL to node & allow trusted node DAO contract to spend it
	if err := rplutils.MintRPL(rp, ownerAccount, trustedNodeAccount, rplBondAmount); err != nil {
		return err
	}
	if _, err := tokens.ApproveRPL(rp, *rocketDAONodeTrustedActionsAddress, rplBondAmount, trustedNodeAccount.GetTransactor()); err != nil {
		return err
	}

	// Return
	return nil

}

// Suppress unused import warnings
var _ = node.RegisterNode
var _ = trustednodesettings.GetRPLBond
