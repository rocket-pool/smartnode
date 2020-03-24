package minipool

import (
    "bytes"
    "context"
    "encoding/hex"
    "errors"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Minipool data
type Minipool struct {
    Address *common.Address
    Contract *bind.BoundContract
    Pubkey string
}


// Initialise and return minipool
func Initialise(p *services.Provider, minipoolAddressStr string) (*Minipool, error) {

    // Get minipool address
    minipoolAddress := common.HexToAddress(minipoolAddressStr)

    // Check contract code at minipool address
    if code, err := p.Client.CodeAt(context.Background(), minipoolAddress, nil); err != nil {
        return nil, errors.New("Error retrieving contract code at minipool address: " + err.Error())
    } else if len(code) == 0 {
        return nil, errors.New("No contract code found at minipool address")
    }

    // Initialise minipool contract
    minipoolContract, err := p.CM.NewContract(&minipoolAddress, "rocketMinipool")
    if err != nil {
        return nil, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Check minipool node owner
    nodeAccount, _ := p.AM.GetNodeAccount()
    nodeOwner := new(common.Address)
    if err := minipoolContract.Call(nil, nodeOwner, "getNodeOwner"); err != nil {
       return nil, errors.New("Error retrieving minipool node owner: " + err.Error())
    } else if !bytes.Equal(nodeOwner.Bytes(), nodeAccount.Address.Bytes()) {
        return nil, errors.New("Minipool is not owned by this node")
    }

    // Get minipool validator pubkey
    validatorPubkey := new([]byte)
    if err := minipoolContract.Call(nil, validatorPubkey, "getValidatorPubkey"); err != nil {
       return nil, errors.New("Error retrieving minipool validator pubkey: " + err.Error())
    }

    // Return
    return &Minipool{
        Address: &minipoolAddress,
        Contract: minipoolContract,
        Pubkey: hex.EncodeToString(*validatorPubkey),
    }, nil

}

