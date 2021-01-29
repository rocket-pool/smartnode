package node

import (
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
)


// Register a trusted node
func RegisterTrustedNode(rp *rocketpool.RocketPool, ownerAccount *accounts.Account, trustedNodeAccount *accounts.Account) error {
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", trustedNodeAccount.GetTransactor()); err != nil { return err }
    if _, err := node.SetNodeTrusted(rp, trustedNodeAccount.Address, true, ownerAccount.GetTransactor()); err != nil { return err }
    return nil
}

