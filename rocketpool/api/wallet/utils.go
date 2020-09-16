package wallet

import (
    "github.com/rocket-pool/rocketpool-go/types"

    "github.com/rocket-pool/smartnode/shared/services/wallet"
)


// Get the validator keys in a wallet
func getValidatorPubkeys(w *wallet.Wallet) ([]types.ValidatorPubkey, error) {

    // Get validator count
    validatorCount, err := w.GetValidatorKeyCount()
    if err != nil {
        return nil, err
    }

    // Get validator keys & return
    validatorKeys := make([]types.ValidatorPubkey, validatorCount)
    for vi := uint(0); vi < validatorCount; vi++ {
        validatorKey, err := w.GetValidatorKeyAt(vi)
        if err != nil {
            return nil, err
        }
        validatorKeys[vi] = types.BytesToValidatorPubkey(validatorKey.PublicKey().Marshal())
    }
    return validatorKeys, nil

}

