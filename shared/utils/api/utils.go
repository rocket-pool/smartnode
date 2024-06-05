package api

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

type EIP712Components struct {
	V uint8    `json:"v"`
	R [32]byte `json:"r"`
	S [32]byte `json:"s"`
}

// The fraction of the timeout period to trigger overdue transactions
const TimeoutSafetyFactor int = 2

// Print the gas price and cost of a TX
func PrintAndCheckGasInfo(gasInfo rocketpool.GasInfo, checkThreshold bool, gasThresholdGwei float64, logger *log.ColorLogger, maxFeeWei *big.Int, gasLimit uint64) bool {

	// Check the gas threshold if requested
	if checkThreshold {
		gasThresholdWei := math.RoundUp(gasThresholdGwei*eth.WeiPerGwei, 0)
		gasThreshold := new(big.Int).SetUint64(uint64(gasThresholdWei))
		if maxFeeWei.Cmp(gasThreshold) != -1 {
			logger.Printlnf("Current network gas price is %.2f Gwei, which is not lower than the set threshold of %.2f Gwei. "+
				"Aborting the transaction.", eth.WeiToGwei(maxFeeWei), gasThresholdGwei)
			return false
		}
	} else {
		logger.Println("This transaction does not check the gas threshold limit, continuing...")
	}

	// Print the total TX cost
	var gas *big.Int
	var safeGas *big.Int
	if gasLimit != 0 {
		gas = new(big.Int).SetUint64(gasLimit)
		safeGas = gas
	} else {
		gas = new(big.Int).SetUint64(gasInfo.EstGasLimit)
		safeGas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
	}
	totalGasWei := new(big.Int).Mul(maxFeeWei, gas)
	totalSafeGasWei := new(big.Int).Mul(maxFeeWei, safeGas)
	logger.Printlnf("This transaction will use a max fee of %.6f Gwei, for a total of up to %.6f - %.6f ETH.",
		eth.WeiToGwei(maxFeeWei),
		math.RoundDown(eth.WeiToEth(totalGasWei), 6),
		math.RoundDown(eth.WeiToEth(totalSafeGasWei), 6))

	return true
}

// Print a TX's details to the logger and waits for it to validated.
func PrintAndWaitForTransaction(cfg *config.RocketPoolConfig, hash common.Hash, ec rocketpool.ExecutionClient, logger *log.ColorLogger) error {

	txWatchUrl := cfg.Smartnode.GetTxWatchUrl()
	hashString := hash.String()

	logger.Printlnf("Transaction has been submitted with hash %s.", hashString)
	if txWatchUrl != "" {
		logger.Printlnf("You may follow its progress by visiting:")
		logger.Printlnf("%s/%s\n", txWatchUrl, hashString)
	}
	logger.Println("Waiting for the transaction to be validated...")

	// Wait for the TX to be included in a block
	if _, err := utils.WaitForTransaction(ec, hash); err != nil {
		return fmt.Errorf("Error waiting for transaction: %w", err)
	}

	return nil

}

// True if a transaction is due and needs to bypass the gas threshold
func IsTransactionDue(rp *rocketpool.RocketPool, startTime time.Time) (bool, time.Duration, error) {

	// Get the dissolve timeout
	timeout, err := protocol.GetMinipoolLaunchTimeout(rp, nil)
	if err != nil {
		return false, 0, err
	}

	dueTime := timeout / time.Duration(TimeoutSafetyFactor)
	isDue := time.Since(startTime) > dueTime
	timeUntilDue := time.Until(startTime.Add(dueTime))
	return isDue, timeUntilDue, nil

}

//  Expects a 129 byte 0x-prefixed EIP-712 signature and returns v/r/s as v uint8 and r, s [32]byte

func ParseEIP712(signature string) (*EIP712Components, error) {
	if len(signature) != 132 || signature[:2] != "0x" {
		return nil, fmt.Errorf("Invalid 129 byte 0x-prefixed EIP-712 signature while parsing: '%s'", signature)
	}
	signature = signature[2:]
	if !regexp.MustCompile("^[A-Fa-f0-9]+$").MatchString(signature) {
		return &EIP712Components{}, fmt.Errorf("Invalid 129 byte 0x-prefixed EIP-712 signature while parsing: '%s'", signature)
	}

	// Slice signature string into v, r, s component of a signature giving node permission to use the given signer
	str_v := signature[len(signature)-2:]
	str_r := signature[:64]
	str_s := signature[64:128]

	// Convert v to uint8 and v,s to [32]byte
	bytes_r, err := hex.DecodeString(str_r)
	if err != nil {
		return &EIP712Components{}, fmt.Errorf("error decoding r: %v", err)
	}
	bytes_s, err := hex.DecodeString(str_s)
	if err != nil {
		return &EIP712Components{}, fmt.Errorf("error decoding s: %v", err)
	}

	int_v, err := strconv.ParseUint(str_v, 16, 8)
	if err != nil {
		return &EIP712Components{}, fmt.Errorf("error parsing v: %v", err)
	}

	return &EIP712Components{uint8(int_v), ([32]byte)(bytes_r), ([32]byte)(bytes_s)}, nil
}
