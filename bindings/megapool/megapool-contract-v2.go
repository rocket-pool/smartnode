package megapool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

type MegapoolV2 interface {
	Megapool
	EstimateForceExitGas(validatorIds []uint32, feeLimit *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	ForceExit(validatorIds []uint32, feeLimit *big.Int, opts *bind.TransactOpts) (common.Hash, error)
}

// Megapool contract from delegate version 2, which adds support for EL-triggered forced exits
type megapoolV2 struct {
	megapoolV1
}

const (
	megapoolV2EncodedAbi string = "eJztWktv2zgQ/isLn4M9tLs99OYmbhHATVMn3j0EgUFLY5sITaokZcco9r/vUC/rZYuyJEOtc0pkksN58ZsZDp9+DggXfLcWvhp8XBCm4GpAuedr/Hz6if+68ApuakiD5IQ97jwYfBz4+P3u7w+DqwEna/MDQUJc47fOTvjvqj4tTfFPkdJzMuErLIknBPskuDsB13eQeDIfNoCMmH1/fQFvYK5vuSOBqN9ZxN/ahveAVPRu6HmMNpFRS9+CrQ1h1CVayFv3zHL+E+88VIou+UWIegM/fPAvQ1SqlGCbi5B1xDsz6/t3FpLipDMJ+kr1ZVgUBaV8eQGSjoXzchEmfdDkMgSdctbQpn1MjyawJdJV14zgzO5l48KFYYvybQSOt0lQgl61Sc+TQgsH82si2iRra1dMF7Skc/9AeMlY8viOs4Nuud+UYHK9ixJtnCh8HZHHKUoTDV99TeaUURw2vsA9siNzlpJk4XNHU8EtuEuH8NnRGJ5iMMiJP/vcVdX8WTK3p+4Q5vgMyaAKXAxzkQ0yO1Ur2hyR/cpSX7E4E40oJE47/NaIjjlMhwmUKn1DYXuSO9R21sRab2Y6o5lOObXOijAGfAkmfWwTV1JbmPjXDWkXPAkOUuolIrph6dxP3qJaN0nBOjJQEiO7ob8EPcQJKTmuI3CqwJu0zkr1dNq5zHIW3RIhbzkvqMaNNsEiw5S5jOwLL1HJ2kfbjV49KomZ9Bimg71Q2JgonaSdfeMtqOD6aMs72CZshf2UHz6VsAYbBs+jvDtTu7muBFWRK5FkUoGleKhFloy2+qSi78GdZZ+4alSOdMbVBHBSv+LOVIG8Jh7VhPWJpdCjesZYT0C0tWQzLdItX4icRI5Ye4IjGlttyjAGhpmVAfDHwmVN0lmopGLiACi9T9CaUYlgqTYRLGKEotqaibkQbL9a5e6Ig9HKVZDrititovx7VNDUXHaPgcnwWZ9Nz4S0qarPq1ts5tmrJtNHsVuWv8A+tMrY9cNf2e0+EUa4kzM+TrJYHm5bPAVm9XPhrhMzR9/Rf0zMIh3fwj/gqSRLGJMdHso/8we1ACva97BC6yGqDLl7789fIHs9+gYvb/DyBi99hxcafmUhpkQnOw2p2seLj7tOD/cMmkox6bhc7UmUYelfqleuJFvCriW4iG6UsIqSJdi/PL2MhzrtNcwR9I739o6cmZmvYBQe8UdqXNLi0GUNsbf3fbmr1aDwgCGFaF+CLZGM40UofkM0mQihy80Qt09S9x2D/lw+V6PNbBt7KHIz8oSzOgWzZhIcdO4D61NqEpoudm33PbpXUth/u+VftkBt9FO4Mpo5pucj92uTm6M6JmJnMtBnipT2IewXMpRiIkggkcu1Vy3ondBnccba3VwZvF0u3Lq11sSRgEsKfYkm7wZaC+RxWtmB1BuQygyf/c6pyaOjApYI5t4Aw1RPwyE8qUcQg1cjgvWf88TbTb2lJG7r7+0KEi6kWLcj2llfs430yrypcIA2eiFdzk6h6NTlxVzLMmHlN8atlJ56CAblYtmAEMbTOXFeDj9TyHhYN3gS9HaTg9OLFlLMT9BMhYqeTdYF8tZvqau7WEDwXKBvmkoc0Y6xjnRlkTryTKOyRBlJQU9NeUlVtzlbruArBY5UNK9S9GnsWSCEDHHTKkt5ei7PU3IPyIKJNk/QZguAMV3T40neQkinlUdgz/8DcNqKAw=="
)

// The decoded ABI for v2 megapools
var megapoolV2Abi *abi.ABI

// Create new megapool contract
func NewMegaPoolV2(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (MegapoolV2, error) {

	var contract *rocketpool.Contract
	var err error
	if megapoolV2Abi == nil {
		// Get contract
		contract, err = createMegapoolContractFromEncodedAbi(rp, address, megapoolV2EncodedAbi)
	} else {
		contract, err = createMegapoolContractFromAbi(rp, address, megapoolV2Abi)
	}
	if err != nil {
		return nil, err
	} else if megapoolV2Abi == nil {
		megapoolV2Abi = contract.ABI
	}

	// Create and return
	return &megapoolV2{
		megapoolV1: megapoolV1{
			Address:    address,
			Version:    2,
			Contract:   contract,
			RocketPool: rp,
		},
	}, nil
}

// Estimate the gas of ForceExit
func (mp *megapoolV2) EstimateForceExitGas(validatorIds []uint32, feeLimit *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "forceExit", validatorIds, feeLimit)
}

// Force exit megapool validators that failed to exit within the cooperative exit phase
func (mp *megapoolV2) ForceExit(validatorIds []uint32, feeLimit *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "forceExit", validatorIds, feeLimit)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error force exiting megapool %s validators: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}
