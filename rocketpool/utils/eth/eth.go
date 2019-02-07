package eth

import (
    "math/big"

    "github.com/ethereum/go-ethereum/crypto/sha3"
)


// Conversion factor from wei to eth
const WEI_PER_ETH = 1000000000000000000


// Convert wei to eth
func WeiToEth(wei *big.Int) big.Int {
    var eth big.Int
    eth.Quo(wei, big.NewInt(WEI_PER_ETH))
    return eth
}


// Convert eth to wei
func EthToWei(eth *big.Int) big.Int {
    var wei big.Int
    wei.Mul(eth, big.NewInt(WEI_PER_ETH))
    return wei
}


// Make a keccak256 hash of a source byte slice and return as a 32-byte array
func KeccakBytes(src []byte) [32]byte {

    // Hash source data
    hash := sha3.NewKeccak256()
    hash.Write(src[:])

    // Copy hashed data to byte array
    var bytes [32]byte
    copy(bytes[:], hash.Sum(nil))

    // Return
    return bytes

}


// Make a keccak256 hash of a source string and return as a 32-byte array
func KeccakStr(src string) [32]byte {
    return KeccakBytes([]byte(src))
}

