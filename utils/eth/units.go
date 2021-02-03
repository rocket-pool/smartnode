package eth

import (
    "math/big"
    "strconv"
)


// Conversion factors
const (
    WeiPerEth float64 = 1e18
    WeiPerGwei float64 = 1e9
)


// Convert wei to eth
func WeiToEth(wei *big.Int) float64 {
    var weiFloat big.Float
    var eth big.Float
    weiFloat.SetInt(wei)
    eth.Quo(&weiFloat, big.NewFloat(WeiPerEth))
    eth64, _ := eth.Float64()
    return eth64
}


// Convert eth to wei
func EthToWei(eth float64) *big.Int {
    var ethFloat big.Float
    var weiFloat big.Float
    var wei big.Int
    ethFloat.SetString(strconv.FormatFloat(eth, 'f', -1, 64))
    weiFloat.Mul(&ethFloat, big.NewFloat(WeiPerEth))
    weiFloat.Int(&wei)
    return &wei
}


// Convert wei to gigawei
func WeiToGwei(wei *big.Int) float64 {
    var weiFloat big.Float
    var gwei big.Float
    weiFloat.SetInt(wei)
    gwei.Quo(&weiFloat, big.NewFloat(WeiPerGwei))
    gwei64, _ := gwei.Float64()
    return gwei64
}


// Convert gigawei to wei
func GweiToWei(gwei float64) *big.Int {
    var gweiFloat big.Float
    var weiFloat big.Float
    var wei big.Int
    gweiFloat.SetString(strconv.FormatFloat(gwei, 'f', -1, 64))
    weiFloat.Mul(&gweiFloat, big.NewFloat(WeiPerGwei))
    weiFloat.Int(&wei)
    return &wei
}

