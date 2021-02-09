package trustednode

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Config
const ProposalsSettingsContractName = "rocketDAONodeTrustedSettingsProposals"


// Get contracts
var proposalsSettingsContractLock sync.Mutex
func getProposalsSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    proposalsSettingsContractLock.Lock()
    defer proposalsSettingsContractLock.Unlock()
    return rp.GetContract(ProposalsSettingsContractName)
}

