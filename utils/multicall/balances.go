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
	"golang.org/x/sync/errgroup"
)

const (
	balanceBatchSize int = 1000
	threadLimit      int = 6
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

	// Sync
	count := len(addresses)
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	balances := make([]*big.Int, count)

	// Run the getters in batches
	for i := 0; i < count; i += balanceBatchSize {
		i := i
		max := i + balanceBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			subAddresses := addresses[i:max]
			tokens := []common.Address{
				{}, // Empty token for ETH balance
			}
			callData, err := b.ABI.Pack("balances", subAddresses, tokens)
			if err != nil {
				return fmt.Errorf("error creating calldata for balances: %w", err)
			}

			response, err := b.Client.CallContract(context.Background(), ethereum.CallMsg{To: &b.ContractAddress, Data: callData}, opts.BlockNumber)
			if err != nil {
				return fmt.Errorf("error calling balances: %w", err)
			}

			var subBalances []*big.Int
			err = b.ABI.UnpackIntoInterface(&subBalances, "balances", response)
			if err != nil {
				return fmt.Errorf("error unpacking balances response: %w", err)
			}

			if len(subBalances) != len(subAddresses) {
				return fmt.Errorf("received %d balances which mismatches query batch size %d", len(subBalances), len(subAddresses))
			}
			for j, balance := range subBalances {
				if balance == nil {
					return fmt.Errorf("received nil balance for address %s", subAddresses[j].String())
				}
				balances[i+j] = balance
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting balances: %w", err)
	}

	return balances, nil
}
