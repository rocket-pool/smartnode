package app

import (
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/contracts"
    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Add token liquidity
func AppAddTokenLiquidity(options AppOptions, token string, etherAmountWei *big.Int, tokenAmountWei *big.Int) error {

    // Initialise node password & account managers
    pm := passwords.NewPasswordManager(options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadContracts([]string{"rocketPoolToken"}); err != nil { return err }

    // Get uniswap factory
    uniswap, err := contracts.NewUniswapFactory(common.HexToAddress(options.UniswapAddress), client)
    if err != nil { return err }

    // Get exchange addresses
    rplExchangeAddress, err := uniswap.GetExchange(nil, *(cm.Addresses["rocketPoolToken"]))
    if err != nil { return err }

    // Get token properties
    var tokenContract string
    var tokenExchangeAddress *common.Address
    var tokenExchangeAbi abi.ABI
    switch token {
        case "RPL":
            tokenContract = "rocketPoolToken"
            tokenExchangeAddress = &rplExchangeAddress
            tokenExchangeAbi, _ = abi.JSON(strings.NewReader(contracts.UniswapExchangeABI))
    }

    // Initialise node account
    if err := pm.SetPassword("foobarbaz"); err != nil { return err }
    if _, err := am.CreateNodeAccount(); err != nil { return err }

    // Seed node account
    etherSeedAmountWei := big.NewInt(0)
    etherSeedAmountWei.Add(etherAmountWei, eth.EthToWei(1))
    if err := AppSeedNodeAccount(options, etherSeedAmountWei, tokenAmountWei); err != nil { return err }

    // Set exchange token allowance
    if txor, err := am.GetNodeAccountTransactor(); err != nil {
        return err
    } else if _, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses[tokenContract], cm.Abis[tokenContract], "approve", tokenExchangeAddress, tokenAmountWei); err != nil {
        return err
    }

    // Add liquidity to exchange
    if txor, err := am.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        txor.Value = etherAmountWei
        if _, err := eth.ExecuteContractTransaction(client, txor, tokenExchangeAddress, &tokenExchangeAbi, "addLiquidity", big.NewInt(0), tokenAmountWei, big.NewInt(5000000000)); err != nil { return err }
    }

    // Return
    return nil

}

