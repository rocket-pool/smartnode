/*
* This code was derived from https://github.com/depocket/multicall-go
 */

package multicall

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type Call struct {
	Key      string         `json:"key"`
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
	Abi             abi.ABI
	ContractAddress common.Address
	calls           []Call
}

func NewMultiCaller(client rocketpool.ExecutionClient, multicallerAddress common.Address) (*MultiCaller, error) {
	mcAbi, err := abi.JSON(strings.NewReader(MultiMetaData.ABI))
	if err != nil {
		return nil, err
	}

	return &MultiCaller{
		Client:          client,
		Abi:             mcAbi,
		ContractAddress: multicallerAddress,
		calls:           []Call{},
	}, nil
}

func (caller *MultiCaller) AddCall(callName string, contract *rocketpool.Contract, output interface{}, method string, args ...interface{}) error {
	callData, err := contract.ABI.Pack(method, args...)
	if err != nil {
		return fmt.Errorf("error adding call [%s]: %w", method, err)
	}
	call := Call{
		Method:   method,
		Target:   *contract.Address,
		Key:      callName,
		CallData: callData,
		Contract: contract,
		output:   output,
	}
	caller.calls = append(caller.calls, call)
	return nil
}

func (caller *MultiCaller) Execute(requireSuccess bool) (map[string]CallResponse, error) {
	var multiCalls = make([]MultiCall, 0, len(caller.calls))
	for _, call := range caller.calls {
		multiCalls = append(multiCalls, call.GetMultiCall())
	}
	callData, err := caller.Abi.Pack("tryAggregate", requireSuccess, multiCalls)
	if err != nil {
		return nil, err
	}

	resp, err := caller.Client.CallContract(context.Background(), ethereum.CallMsg{To: &caller.ContractAddress, Data: callData}, nil)
	if err != nil {
		return nil, err
	}

	responses, err := caller.Abi.Unpack("tryAggregate", resp)

	if err != nil {
		return nil, err
	}

	results := make(map[string]CallResponse)
	for i, response := range responses[0].([]struct {
		Success    bool   `json:"success"`
		ReturnData []byte `json:"returnData"`
	}) {
		results[caller.calls[i].Key] = CallResponse{
			Method:        caller.calls[i].Method,
			ReturnDataRaw: response.ReturnData,
			Status:        response.Success,
		}
	}
	return results, nil
}

func (caller *MultiCaller) FlexibleCall(requireSuccess bool) (map[string]Result, error) {
	res := make(map[string]Result)
	results, err := caller.Execute(requireSuccess)
	if err != nil {
		caller.calls = []Call{}
		return nil, err
	}
	for _, call := range caller.calls {
		callSuccess := results[call.Key].Status
		if callSuccess {
			err := call.Contract.ABI.UnpackIntoInterface(call.output, call.Method, results[call.Key].ReturnDataRaw)
			if err != nil {
				caller.calls = []Call{}
				return nil, err
			}
			res[call.Key] = Result{
				Success: results[call.Key].Status,
				Output:  call.output,
			}
		} else {
			res[call.Key] = Result{
				Success: results[call.Key].Status,
				Output:  call.output,
			}
		}
	}
	caller.calls = []Call{}
	return res, err
}
