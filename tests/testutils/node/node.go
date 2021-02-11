package node

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
)


// Trusted node counter
var trustedNodeIndex = 0


// Register a trusted node
func RegisterTrustedNode(rp *rocketpool.RocketPool, ownerAccount *accounts.Account, trustedNodeAccount *accounts.Account) error {
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", trustedNodeAccount.GetTransactor()); err != nil { return err }
    if _, err := trustednode.BootstrapMember(rp, fmt.Sprintf("tn%d", trustedNodeIndex), fmt.Sprintf("tn%d@rocketpool.net", trustedNodeIndex), trustedNodeAccount.Address, ownerAccount.GetTransactor()); err != nil { return err }
    trustedNodeIndex++
    return nil
}

