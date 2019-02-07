package eth

import (
    "github.com/ethereum/go-ethereum/crypto/sha3"
)


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

