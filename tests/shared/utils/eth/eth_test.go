package eth

import (
    "bytes"
    "math/big"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Test wei to eth conversion
func TestWeiToEth(t *testing.T) {
    weiValue := big.NewInt(1000000000000000000)
    ethValue := eth.WeiToEth(weiValue)
    if ethValue != float64(1) { t.Errorf("Incorrect eth value: expected %.2f, got %.2f", float64(1), ethValue) }
}


// Test eth to wei conversion
func TestEthToWei(t *testing.T) {
    ethValue := float64(1)
    weiValue := eth.EthToWei(ethValue)
    if weiValue.String() != "1000000000000000000" { t.Errorf("Incorrect wei value: expected %s, got %s", "1000000000000000000", weiValue.String()) }
}


// Test wei to gwei conversion
func TestWeiToGwei(t *testing.T) {
    weiValue := big.NewInt(1000000000000000000)
    gweiValue := eth.WeiToGwei(weiValue)
    if gweiValue != float64(1000000000) { t.Errorf("Incorrect gwei value: expected %.2f, got %.2f", float64(1000000000), gweiValue) }
}


// Test gwei to wei conversion
func TestGweiToWei(t *testing.T) {
    gweiValue := float64(1)
    weiValue := eth.GweiToWei(gweiValue)
    if weiValue.String() != "1000000000" { t.Errorf("Incorrect wei value: expected %s, got %s", "1000000000", weiValue.String()) }
}


// Test keccak256 on bytes
func TestKeccakBytes(t *testing.T) {
    source := []byte{255,127,0,255,127,0,255,127,0,255,127,0}
    expectedHash := []byte{76,169,54,0,156,235,153,253,70,212,239,214,5,96,20,78,21,117,130,64,154,19,88,31,34,236,112,233,90,255,142,149}
    hash := eth.KeccakBytes(source)
    if !bytes.Equal(hash[:], expectedHash) { t.Error("Incorrect keccak hash from bytes") }
}


// Test keccak256 on string
func TestKeccakStr(t *testing.T) {
    source := "Lorem ipsum dolor sit amet"
    expectedHash := []byte{181,59,124,165,21,5,29,73,232,81,192,123,208,251,221,219,128,16,169,54,111,91,31,95,183,55,186,33,164,53,99,1}
    hash := eth.KeccakStr(source)
    if !bytes.Equal(hash[:], expectedHash) { t.Error("Incorrect keccak hash from string") }
}

