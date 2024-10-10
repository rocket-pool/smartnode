package node

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func TestApplyTo(t *testing.T) {
	tests := []struct {
		name              string
		maxPriorityFee    *big.Int
		maxFee            *big.Int
		expectedGasFeeCap *big.Int
		expectedGasTipCap *big.Int
	}{
		{
			name:              "maxPriorityFee less than maxFee",
			maxPriorityFee:    big.NewInt(50),
			maxFee:            big.NewInt(100),
			expectedGasFeeCap: big.NewInt(100),
			expectedGasTipCap: big.NewInt(50),
		},
		{
			name:              "maxPriorityFee equal to maxFee",
			maxPriorityFee:    big.NewInt(100),
			maxFee:            big.NewInt(100),
			expectedGasFeeCap: big.NewInt(100),
			expectedGasTipCap: big.NewInt(25),
		},
		{
			name:              "maxPriorityFee greater than maxFee",
			maxPriorityFee:    big.NewInt(150),
			maxFee:            big.NewInt(100),
			expectedGasFeeCap: big.NewInt(100),
			expectedGasTipCap: big.NewInt(25),
		},
		{
			name:              "maxPriorityFee less than 25%% of maxFee",
			maxPriorityFee:    big.NewInt(20),
			maxFee:            big.NewInt(100),
			expectedGasFeeCap: big.NewInt(100),
			expectedGasTipCap: big.NewInt(20),
		},
		{
			name:              "maxPriorityFee equal to 25%% of maxFee",
			maxPriorityFee:    big.NewInt(25),
			maxFee:            big.NewInt(100),
			expectedGasFeeCap: big.NewInt(100),
			expectedGasTipCap: big.NewInt(25),
		},
		{
			name:              "maxPriorityFee less than maxFee but greater than 25%% of maxFee",
			maxPriorityFee:    big.NewInt(30),
			maxFee:            big.NewInt(100),
			expectedGasFeeCap: big.NewInt(100),
			expectedGasTipCap: big.NewInt(30),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			g := &GasSettings{
				maxFee:         testCase.maxFee,
				maxPriorityFee: testCase.maxPriorityFee,
			}

			opts := &bind.TransactOpts{}

			g.ApplyTo(opts)

			if opts.GasFeeCap.Cmp(testCase.expectedGasFeeCap) != 0 {
				t.Errorf("expected GasFeeCap %s, got %s", testCase.expectedGasFeeCap.String(), opts.GasFeeCap.String())
			}
			if opts.GasTipCap.Cmp(testCase.expectedGasTipCap) != 0 {
				t.Errorf("expected GasTipCap %s, got %s", testCase.expectedGasTipCap.String(), opts.GasTipCap.String())
			}
		})
	}
}
