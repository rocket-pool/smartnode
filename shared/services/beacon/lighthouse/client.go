package lighthouse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/utils/eth2"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// Config
const (
	RequestUrlFormat   = "%s%s"
	RequestContentType = "application/json"

	RequestSyncStatusPath            = "/eth/v1/node/syncing"
	RequestEth2ConfigPath            = "/eth/v1/config/spec"
	RequestEth2DepositContractMethod = "/eth/v1/config/deposit_contract"
	RequestGenesisPath               = "/eth/v1/beacon/genesis"
	RequestFinalityCheckpointsPath   = "/eth/v1/beacon/states/%s/finality_checkpoints"
	RequestForkPath                  = "/eth/v1/beacon/states/%s/fork"
	RequestValidatorsPath            = "/eth/v1/beacon/states/%s/validators"
	RequestVoluntaryExitPath         = "/eth/v1/beacon/pool/voluntary_exits"
	RequestBeaconBlockPath           = "/eth/v1/beacon/blocks/%s"
	RequestValidatorSyncDuties       = "/eth/v1/validator/duties/sync/%s"
	RequestValidatorProposerDuties   = "/eth/v1/validator/duties/proposer/%s"

	MaxRequestValidatorsCount = 600
)

// Lighthouse client
type Client struct {
	providerAddress string
}

// Create new lighthouse client
func NewClient(providerAddress string) *Client {
	return &Client{
		providerAddress: providerAddress,
	}
}

// Close the client connection
func (c *Client) Close() error {
	return nil
}

// Get the beacon client type
func (c *Client) GetClientType() beacon.BeaconClientType {
	return beacon.SplitProcess
}

// Get the node's sync status
func (c *Client) GetSyncStatus() (beacon.SyncStatus, error) {

	// Get sync status
	syncStatus, err := c.getSyncStatus()
	if err != nil {
		return beacon.SyncStatus{}, err
	}

	// Calculate the progress
	progress := float64(syncStatus.Data.HeadSlot) / float64(syncStatus.Data.HeadSlot+syncStatus.Data.SyncDistance)

	// Return response
	return beacon.SyncStatus{
		Syncing:  syncStatus.Data.IsSyncing,
		Progress: progress,
	}, nil

}

// Get the eth2 config
func (c *Client) GetEth2Config() (beacon.Eth2Config, error) {

	// Data
	var wg errgroup.Group
	var eth2Config Eth2ConfigResponse
	var genesis GenesisResponse

	// Get eth2 config
	wg.Go(func() error {
		var err error
		eth2Config, err = c.getEth2Config()
		return err
	})

	// Get genesis
	wg.Go(func() error {
		var err error
		genesis, err = c.getGenesis()
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return beacon.Eth2Config{}, err
	}

	// Return response
	return beacon.Eth2Config{
		GenesisForkVersion:           genesis.Data.GenesisForkVersion,
		GenesisValidatorsRoot:        genesis.Data.GenesisValidatorsRoot,
		GenesisEpoch:                 0,
		GenesisTime:                  uint64(genesis.Data.GenesisTime),
		SecondsPerSlot:               uint64(eth2Config.Data.SecondsPerSlot),
		SlotsPerEpoch:                uint64(eth2Config.Data.SlotsPerEpoch),
		SecondsPerEpoch:              uint64(eth2Config.Data.SecondsPerSlot * eth2Config.Data.SlotsPerEpoch),
		EpochsPerSyncCommitteePeriod: uint64(eth2Config.Data.EpochsPerSyncCommitteePeriod),
	}, nil

}

// Get the eth2 deposit contract info
func (c *Client) GetEth2DepositContract() (beacon.Eth2DepositContract, error) {

	// Get the deposit contract
	depositContract, err := c.getEth2DepositContract()
	if err != nil {
		return beacon.Eth2DepositContract{}, err
	}

	// Return response
	return beacon.Eth2DepositContract{
		ChainID: uint64(depositContract.Data.ChainID),
		Address: depositContract.Data.Address,
	}, nil
}

// Get the beacon head
func (c *Client) GetBeaconHead() (beacon.BeaconHead, error) {

	// Data
	var wg errgroup.Group
	var eth2Config beacon.Eth2Config
	var finalityCheckpoints FinalityCheckpointsResponse

	// Get eth2 config
	wg.Go(func() error {
		var err error
		eth2Config, err = c.GetEth2Config()
		return err
	})

	// Get finality checkpoints
	wg.Go(func() error {
		var err error
		finalityCheckpoints, err = c.getFinalityCheckpoints("head")
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return beacon.BeaconHead{}, err
	}

	// Return response
	return beacon.BeaconHead{
		Epoch:                  eth2.EpochAt(eth2Config, uint64(time.Now().Unix())),
		FinalizedEpoch:         uint64(finalityCheckpoints.Data.Finalized.Epoch),
		JustifiedEpoch:         uint64(finalityCheckpoints.Data.CurrentJustified.Epoch),
		PreviousJustifiedEpoch: uint64(finalityCheckpoints.Data.PreviousJustified.Epoch),
	}, nil

}

// Get a validator's status
func (c *Client) GetValidatorStatus(pubkey types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {

	return c.getValidatorStatus(hexutil.AddPrefix(pubkey.Hex()), opts)

}
func (c *Client) GetValidatorStatusByIndex(index string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {

	return c.getValidatorStatus(index, opts)

}

func (c *Client) getValidatorStatus(pubkeyOrIndex string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	// Get validator
	validators, err := c.getValidatorsByOpts([]string{pubkeyOrIndex}, opts)
	if err != nil {
		return beacon.ValidatorStatus{}, err
	}
	if len(validators.Data) == 0 {
		return beacon.ValidatorStatus{}, nil
	}
	validator := validators.Data[0]

	// Return response
	return beacon.ValidatorStatus{
		Pubkey:                     types.BytesToValidatorPubkey(validator.Validator.Pubkey),
		Index:                      uint64(validator.Index),
		WithdrawalCredentials:      common.BytesToHash(validator.Validator.WithdrawalCredentials),
		Balance:                    uint64(validator.Balance),
		EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
		Slashed:                    validator.Validator.Slashed,
		ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
		ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
		ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
		WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
		Exists:                     true,
	}, nil
}

// Get multiple validators' statuses
func (c *Client) GetValidatorStatuses(pubkeys []types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (map[types.ValidatorPubkey]beacon.ValidatorStatus, error) {

	// Convert pubkeys into hex strings
	pubkeysHex := make([]string, len(pubkeys))
	for vi := 0; vi < len(pubkeys); vi++ {
		pubkeysHex[vi] = hexutil.AddPrefix(pubkeys[vi].Hex())
	}

	// Get validators
	validators, err := c.getValidatorsByOpts(pubkeysHex, opts)
	if err != nil {
		return map[types.ValidatorPubkey]beacon.ValidatorStatus{}, err
	}

	// Build validator status map
	statuses := make(map[types.ValidatorPubkey]beacon.ValidatorStatus)
	for _, validator := range validators.Data {

		// Get validator pubkey
		pubkey := types.BytesToValidatorPubkey(validator.Validator.Pubkey)

		// Add status
		statuses[pubkey] = beacon.ValidatorStatus{
			Pubkey:                     types.BytesToValidatorPubkey(validator.Validator.Pubkey),
			Index:                      uint64(validator.Index),
			WithdrawalCredentials:      common.BytesToHash(validator.Validator.WithdrawalCredentials),
			Balance:                    uint64(validator.Balance),
			EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
			Slashed:                    validator.Validator.Slashed,
			ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
			ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
			ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
			WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
			Exists:                     true,
		}

	}

	// Return
	return statuses, nil

}

// Get whether validators have sync duties to perform at given epoch
func (c *Client) GetValidatorSyncDuties(indices []uint64, epoch uint64) (map[uint64]bool, error) {

	// Convert incoming uint64 validator indices into an array of string for the request
	indicesStrings := make([]string, len(indices))

	for i, index := range indices {
		indicesStrings[i] = strconv.FormatUint(index, 10)
	}

	// Perform the post request
	responseBody, status, err := c.postRequest(fmt.Sprintf(RequestValidatorSyncDuties, strconv.FormatUint(epoch, 10)), indicesStrings)

	if err != nil {
		return nil, fmt.Errorf("Could not get validator sync duties: %w", err)
	} else if status != http.StatusOK {
		return nil, fmt.Errorf("Could not get validator sync duties: HTTP status %d; response body: '%s'", status, string(responseBody))
	}

	var response SyncDutiesResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("Could not decode validator sync duties data: %w", err)
	}

	// Map the results
	validatorMap := make(map[uint64]bool)

	for _, index := range indices {
		validatorMap[index] = false
		for _, duty := range response.Data {
			if uint64(duty.ValidatorIndex) == index {
				validatorMap[index] = true
				break
			}
		}
	}

	return validatorMap, nil
}

// Sums proposer duties per validators for a given epoch
func (c *Client) GetValidatorProposerDuties(indices []uint64, epoch uint64) (map[uint64]uint64, error) {

	// Perform the post request
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestValidatorProposerDuties, strconv.FormatUint(epoch, 10)))

	if err != nil {
		return nil, fmt.Errorf("Could not get validator proposer duties: %w", err)
	} else if status != http.StatusOK {
		return nil, fmt.Errorf("Could not get validator proposer duties: HTTP status %d; response body: '%s'", status, string(responseBody))
	}

	var response ProposerDutiesResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("Could not decode validator proposer duties data: %w", err)
	}

	// Map the results
	proposerMap := make(map[uint64]uint64)

	for _, index := range indices {
		proposerMap[index] = 0
		for _, duty := range response.Data {
			if uint64(duty.ValidatorIndex) == index {
				proposerMap[index]++
				break
			}
		}
	}

	return proposerMap, nil
}

// Get a validator's index
func (c *Client) GetValidatorIndex(pubkey types.ValidatorPubkey) (uint64, error) {

	// Get validator
	validators, err := c.getValidatorsByOpts([]string{pubkey.Hex()}, nil)
	if err != nil {
		return 0, err
	}
	if len(validators.Data) == 0 {
		return 0, fmt.Errorf("Validator %s index not found.", pubkey.Hex())
	}
	validator := validators.Data[0]

	// Return validator index
	return uint64(validator.Index), nil

}

// Get domain data for a domain type at a given epoch
func (c *Client) GetDomainData(domainType []byte, epoch uint64) ([]byte, error) {

	// Data
	var wg errgroup.Group
	var genesis GenesisResponse
	var fork ForkResponse

	// Get genesis
	wg.Go(func() error {
		var err error
		genesis, err = c.getGenesis()
		return err
	})

	// Get fork
	wg.Go(func() error {
		var err error
		fork, err = c.getFork("head")
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return []byte{}, err
	}

	// Get fork version
	var forkVersion []byte
	if epoch < uint64(fork.Data.Epoch) {
		forkVersion = fork.Data.PreviousVersion
	} else {
		forkVersion = fork.Data.CurrentVersion
	}

	// Compute & return domain
	var dt [4]byte
	copy(dt[:], domainType[:])
	return eth2types.Domain(dt, forkVersion, genesis.Data.GenesisValidatorsRoot), nil

}

// Perform a voluntary exit on a validator
func (c *Client) ExitValidator(validatorIndex, epoch uint64, signature types.ValidatorSignature) error {
	return c.postVoluntaryExit(VoluntaryExitRequest{
		Message: VoluntaryExitMessage{
			Epoch:          uinteger(epoch),
			ValidatorIndex: uinteger(validatorIndex),
		},
		Signature: signature.Bytes(),
	})
}

// Get the ETH1 data for the target beacon block
func (c *Client) GetEth1DataForEth2Block(blockId string) (beacon.Eth1Data, bool, error) {

	// Get the Beacon block
	block, exists, err := c.getBeaconBlock(blockId)
	if err != nil {
		return beacon.Eth1Data{}, false, err
	}
	if !exists {
		return beacon.Eth1Data{}, false, nil
	}

	// Convert the response to the eth1 data struct
	return beacon.Eth1Data{
		DepositRoot:  common.BytesToHash(block.Data.Message.Body.Eth1Data.DepositRoot),
		DepositCount: uint64(block.Data.Message.Body.Eth1Data.DepositCount),
		BlockHash:    common.BytesToHash(block.Data.Message.Body.Eth1Data.BlockHash),
	}, true, nil

}

func (c *Client) GetBeaconBlock(blockId string) (beacon.BeaconBlock, bool, error) {
	block, exists, err := c.getBeaconBlock(blockId)
	if !exists {
		return beacon.BeaconBlock{}, false, nil
	}
	if err != nil {
		return beacon.BeaconBlock{}, false, err
	}

	beaconBlock := beacon.BeaconBlock{
		Slot:          uint64(block.Data.Message.Slot),
		ProposerIndex: uint64(block.Data.Message.ProposerIndex),
	}

	// Execution payload only exists after the merge, so check for its existence
	if block.Data.Message.Body.ExecutionPayload == nil {
		beaconBlock.HasExecutionPayload = false
	} else {
		beaconBlock.HasExecutionPayload = true
		beaconBlock.FeeRecipient = common.BytesToAddress(block.Data.Message.Body.ExecutionPayload.FeeRecipient)
	}

	return beaconBlock, true, nil
}

// Get sync status
func (c *Client) getSyncStatus() (SyncStatusResponse, error) {
	responseBody, status, err := c.getRequest(RequestSyncStatusPath)
	if err != nil {
		return SyncStatusResponse{}, fmt.Errorf("Could not get node sync status: %w", err)
	} else if status != http.StatusOK {
		return SyncStatusResponse{}, fmt.Errorf("Could not get node sync status: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var syncStatus SyncStatusResponse
	if err := json.Unmarshal(responseBody, &syncStatus); err != nil {
		return SyncStatusResponse{}, fmt.Errorf("Could not decode node sync status: %w", err)
	}
	return syncStatus, nil
}

// Get the eth2 config
func (c *Client) getEth2Config() (Eth2ConfigResponse, error) {
	responseBody, status, err := c.getRequest(RequestEth2ConfigPath)
	if err != nil {
		return Eth2ConfigResponse{}, fmt.Errorf("Could not get eth2 config: %w", err)
	} else if status != http.StatusOK {
		return Eth2ConfigResponse{}, fmt.Errorf("Could not get eth2 config: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var eth2Config Eth2ConfigResponse
	if err := json.Unmarshal(responseBody, &eth2Config); err != nil {
		return Eth2ConfigResponse{}, fmt.Errorf("Could not decode eth2 config: %w", err)
	}
	return eth2Config, nil
}

// Get the eth2 deposit contract info
func (c *Client) getEth2DepositContract() (Eth2DepositContractResponse, error) {
	responseBody, status, err := c.getRequest(RequestEth2DepositContractMethod)
	if err != nil {
		return Eth2DepositContractResponse{}, fmt.Errorf("Could not get eth2 deposit contract: %w", err)
	} else if status != http.StatusOK {
		return Eth2DepositContractResponse{}, fmt.Errorf("Could not get eth2 deposit contract: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var eth2DepositContract Eth2DepositContractResponse
	if err := json.Unmarshal(responseBody, &eth2DepositContract); err != nil {
		return Eth2DepositContractResponse{}, fmt.Errorf("Could not decode eth2 deposit contract: %w", err)
	}
	return eth2DepositContract, nil
}

// Get genesis information
func (c *Client) getGenesis() (GenesisResponse, error) {
	responseBody, status, err := c.getRequest(RequestGenesisPath)
	if err != nil {
		return GenesisResponse{}, fmt.Errorf("Could not get genesis data: %w", err)
	} else if status != http.StatusOK {
		return GenesisResponse{}, fmt.Errorf("Could not get genesis data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var genesis GenesisResponse
	if err := json.Unmarshal(responseBody, &genesis); err != nil {
		return GenesisResponse{}, fmt.Errorf("Could not decode genesis: %w", err)
	}
	return genesis, nil
}

// Get finality checkpoints
func (c *Client) getFinalityCheckpoints(stateId string) (FinalityCheckpointsResponse, error) {
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestFinalityCheckpointsPath, stateId))
	if err != nil {
		return FinalityCheckpointsResponse{}, fmt.Errorf("Could not get finality checkpoints: %w", err)
	} else if status != http.StatusOK {
		return FinalityCheckpointsResponse{}, fmt.Errorf("Could not get finality checkpoints: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var finalityCheckpoints FinalityCheckpointsResponse
	if err := json.Unmarshal(responseBody, &finalityCheckpoints); err != nil {
		return FinalityCheckpointsResponse{}, fmt.Errorf("Could not decode finality checkpoints: %w", err)
	}
	return finalityCheckpoints, nil
}

// Get fork
func (c *Client) getFork(stateId string) (ForkResponse, error) {
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestForkPath, stateId))
	if err != nil {
		return ForkResponse{}, fmt.Errorf("Could not get fork data: %w", err)
	} else if status != http.StatusOK {
		return ForkResponse{}, fmt.Errorf("Could not get fork data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var fork ForkResponse
	if err := json.Unmarshal(responseBody, &fork); err != nil {
		return ForkResponse{}, fmt.Errorf("Could not decode fork data: %w", err)
	}
	return fork, nil
}

// Get validators
func (c *Client) getValidators(stateId string, pubkeys []string) (ValidatorsResponse, error) {
	var query string
	if len(pubkeys) > 0 {
		query = fmt.Sprintf("?id=%s", strings.Join(pubkeys, ","))
	}
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestValidatorsPath, stateId) + query)
	if err != nil {
		return ValidatorsResponse{}, fmt.Errorf("Could not get validators: %w", err)
	} else if status != http.StatusOK {
		return ValidatorsResponse{}, fmt.Errorf("Could not get validators: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var validators ValidatorsResponse
	if err := json.Unmarshal(responseBody, &validators); err != nil {
		return ValidatorsResponse{}, fmt.Errorf("Could not decode validators: %w", err)
	}
	return validators, nil
}

// Get validators by pubkeys and status options
func (c *Client) getValidatorsByOpts(pubkeysOrIndices []string, opts *beacon.ValidatorStatusOptions) (ValidatorsResponse, error) {

	// Get state ID
	var stateId string
	if opts == nil {
		stateId = "head"
	} else {

		// Get eth2 config
		eth2Config, err := c.getEth2Config()
		if err != nil {
			return ValidatorsResponse{}, err
		}

		// Get slot nuimber
		slot := opts.Epoch * uint64(eth2Config.Data.SlotsPerEpoch)
		stateId = strconv.FormatInt(int64(slot), 10)

	}

	// Load validator data in batches & return
	data := make([]Validator, 0, len(pubkeysOrIndices))
	for bsi := 0; bsi < len(pubkeysOrIndices); bsi += MaxRequestValidatorsCount {

		// Get batch start & end index
		vsi := bsi
		vei := bsi + MaxRequestValidatorsCount
		if vei > len(pubkeysOrIndices) {
			vei = len(pubkeysOrIndices)
		}

		// Get validator pubkeysOrIndices for batch request
		batch := make([]string, vei-vsi)
		for vi := vsi; vi < vei; vi++ {
			batch[vi-vsi] = pubkeysOrIndices[vi]
		}

		// Get & add validators
		validators, err := c.getValidators(stateId, batch)
		if err != nil {
			return ValidatorsResponse{}, err
		}
		data = append(data, validators.Data...)

	}
	return ValidatorsResponse{Data: data}, nil

}

// Send voluntary exit request
func (c *Client) postVoluntaryExit(request VoluntaryExitRequest) error {
	responseBody, status, err := c.postRequest(RequestVoluntaryExitPath, request)
	if err != nil {
		return fmt.Errorf("Could not broadcast exit for validator at index %d: %w", request.Message.ValidatorIndex, err)
	} else if status != http.StatusOK {
		return fmt.Errorf("Could not broadcast exit for validator at index %d: HTTP status %d; response body: '%s'", request.Message.ValidatorIndex, status, string(responseBody))
	}
	return nil
}

// Get the target beacon block
func (c *Client) getBeaconBlock(blockId string) (BeaconBlockResponse, bool, error) {
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestBeaconBlockPath, blockId))
	if err != nil {
		return BeaconBlockResponse{}, false, fmt.Errorf("Could not get beacon block data: %w", err)
	} else if status == http.StatusNotFound {
		return BeaconBlockResponse{}, false, nil
	} else if status != http.StatusOK {
		return BeaconBlockResponse{}, false, fmt.Errorf("Could not get beacon block data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var beaconBlock BeaconBlockResponse
	if err := json.Unmarshal(responseBody, &beaconBlock); err != nil {
		return BeaconBlockResponse{}, false, fmt.Errorf("Could not decode beacon block data: %w", err)
	}
	return beaconBlock, true, nil
}

// Make a GET request to the beacon node
func (c *Client) getRequest(requestPath string) ([]byte, int, error) {

	// Send request
	response, err := http.Get(fmt.Sprintf(RequestUrlFormat, c.providerAddress, requestPath))
	if err != nil {
		return []byte{}, 0, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// Get response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, 0, err
	}

	// Return
	return body, response.StatusCode, nil

}

// Make a POST request to the beacon node
func (c *Client) postRequest(requestPath string, requestBody interface{}) ([]byte, int, error) {

	// Get request body
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return []byte{}, 0, err
	}
	requestBodyReader := bytes.NewReader(requestBodyBytes)

	// Send request
	response, err := http.Post(fmt.Sprintf(RequestUrlFormat, c.providerAddress, requestPath), RequestContentType, requestBodyReader)
	if err != nil {
		return []byte{}, 0, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// Get response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, 0, err
	}

	// Return
	return body, response.StatusCode, nil

}
