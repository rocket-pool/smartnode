/*
* This code was derived from https://github.com/depocket/multicall-go
 */

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

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/units"
)

type Call struct {
	Method   string         `json:"method"`
	Target   common.Address `json:"target"`
	CallData []byte         `json:"call_data"`
	Contract *rocketpool.Contract
	output   interface{}
}

type CallResponse struct {
	Method        string
	Status        bool
	ReturnDataRaw []byte `json:"returnData"`
}

type Result struct {
	Success bool `json:"success"`
	Output  interface{}
}

func (call Call) GetMultiCall() MultiCall {
	return MultiCall{Target: call.Target, CallData: call.CallData}
}

type MultiCaller struct {
	Client          rocketpool.ExecutionClient
	ABI             abi.ABI
	ContractAddress common.Address
	calls           []Call
}

func NewMultiCaller(client rocketpool.ExecutionClient, multicallerAddress common.Address) (*MultiCaller, error) {
	mcAbi, err := abi.JSON(strings.NewReader(MulticallABI))
	if err != nil {
		return nil, err
	}

	return &MultiCaller{
		Client:          client,
		ABI:             mcAbi,
		ContractAddress: multicallerAddress,
		calls:           []Call{},
	}, nil
}

func (caller *MultiCaller) AddCall(contract *rocketpool.Contract, output interface{}, method string, args ...interface{}) error {
	// iterate args and convert any units.Wei to *big.Int
	for i, arg := range args {
		if wei, ok := arg.(units.Wei); ok {
			args[i] = wei.Decimal.BigInt()
		}
	}
	callData, err := contract.ABI.Pack(method, args...)
	if err != nil {
		return fmt.Errorf("error adding call [%s]: %w", method, err)
	}
	call := Call{
		Method:   method,
		Target:   *contract.Address,
		CallData: callData,
		Contract: contract,
		output:   output,
	}
	caller.calls = append(caller.calls, call)
	return nil
}

func (caller *MultiCaller) Execute(requireSuccess bool, opts *bind.CallOpts) ([]CallResponse, error) {
	var multiCalls = make([]MultiCall, 0, len(caller.calls))
	for _, call := range caller.calls {
		multiCalls = append(multiCalls, call.GetMultiCall())
	}
	callData, err := caller.ABI.Pack("tryAggregate", requireSuccess, multiCalls)
	if err != nil {
		return nil, err
	}

	resp, err := caller.Client.CallContract(context.Background(), ethereum.CallMsg{To: &caller.ContractAddress, Data: callData}, opts.BlockNumber)
	if err != nil {
		return nil, err
	}

	responses, err := caller.ABI.Unpack("tryAggregate", resp)

	if err != nil {
		return nil, err
	}

	results := make([]CallResponse, len(caller.calls))
	for i, response := range responses[0].([]struct {
		Success    bool   `json:"success"`
		ReturnData []byte `json:"returnData"`
	}) {
		results[i].Method = caller.calls[i].Method
		results[i].ReturnDataRaw = response.ReturnData
		results[i].Status = response.Success
	}
	return results, nil
}

func (caller *MultiCaller) FlexibleCall(requireSuccess bool, opts *bind.CallOpts) ([]Result, error) {
	res := make([]Result, len(caller.calls))
	results, err := caller.Execute(requireSuccess, opts)
	if err != nil {
		caller.calls = []Call{}
		return nil, err
	}
	for i, call := range caller.calls {
		callSuccess := results[i].Status
		if callSuccess {
			var local *big.Int
			dst := call.output
			// If dst is a *units.Wei, use a local *big.Int to unpack into
			if _, ok := dst.(*units.Wei); ok {
				dst = &local
			}
			err := call.Contract.ABI.UnpackIntoInterface(dst, call.Method, results[i].ReturnDataRaw)
			if err != nil {
				caller.calls = []Call{}
				return nil, err
			}
			if local != nil {
				*(call.output.(*units.Wei)) = units.NewWei(local)
			}
		}
		res[i].Success = callSuccess
		res[i].Output = call.output
	}
	caller.calls = []Call{}
	return res, err
}
