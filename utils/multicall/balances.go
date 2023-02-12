package multicall

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

const (
	balanceBatchSize int = 4000
)

type BalanceBatcher struct {
	Client          rocketpool.ExecutionClient
	ABI             abi.ABI
	ContractAddress common.Address
}

func NewBalanceBatcher(client rocketpool.ExecutionClient, address common.Address) (*BalanceBatcher, error) {
	abi, err := abi.JSON(strings.NewReader(BalancesABI))
	if err != nil {
		return nil, err
	}

	return &BalanceBatcher{
		Client:          client,
		ContractAddress: address,
		ABI:             abi,
	}, nil
}

func (b *BalanceBatcher) GetEthBalances(addresses []common.Address, opts *bind.CallOpts) ([]*big.Int, error) {

	tokens := make([]common.Address, len(addresses)) // Array of nils

	callData, err := b.ABI.Pack("balances", addresses, tokens)
	if err != nil {
		return nil, fmt.Errorf("error creating calldata for balances: %w", err)
	}

	response, err := b.Client.CallContract(context.Background(), ethereum.CallMsg{To: &b.ContractAddress, Data: callData}, opts.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("error getting balances: %w", err)
	}

	balances := new([]*big.Int)
	err = b.ABI.UnpackIntoInterface(&balances, "balances", response)
	if err != nil {
		return nil, fmt.Errorf("error unpacking balances response: %w", err)
	}

	return *balances, nil
}
