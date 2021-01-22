package tokens

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
    //"github.com/rocket-pool/rocketpool-go/tokens"
)


// Mint an amount of rETH to an account
func MintRETH(rp *rocketpool.RocketPool, toAccount *accounts.Account, amount *big.Int) error {

    // Get rETH exchange rate
    // TODO: implement

    // Deposit from account to mint rETH
    opts := toAccount.GetTransactor()
    opts.Value = amount
    _, err := deposit.Deposit(rp, opts)
    return err

}

