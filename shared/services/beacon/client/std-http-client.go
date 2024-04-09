package client

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
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

	RequestSyncStatusPath                  = "/eth/v1/node/syncing"
	RequestEth2ConfigPath                  = "/eth/v1/config/spec"
	RequestEth2DepositContractMethod       = "/eth/v1/config/deposit_contract"
	RequestGenesisPath                     = "/eth/v1/beacon/genesis"
	RequestCommitteePath                   = "/eth/v1/beacon/states/%s/committees"
	RequestFinalityCheckpointsPath         = "/eth/v1/beacon/states/%s/finality_checkpoints"
	RequestForkPath                        = "/eth/v1/beacon/states/%s/fork"
	RequestValidatorsPath                  = "/eth/v1/beacon/states/%s/validators"
	RequestVoluntaryExitPath               = "/eth/v1/beacon/pool/voluntary_exits"
	RequestAttestationsPath                = "/eth/v1/beacon/blocks/%s/attestations"
	RequestBeaconBlockPath                 = "/eth/v2/beacon/blocks/%s"
	RequestBeaconBlockHeaderPath           = "/eth/v1/beacon/headers/%s"
	RequestValidatorSyncDuties             = "/eth/v1/validator/duties/sync/%s"
	RequestValidatorProposerDuties         = "/eth/v1/validator/duties/proposer/%s"
	RequestWithdrawalCredentialsChangePath = "/eth/v1/beacon/pool/bls_to_execution_changes"

	MaxRequestValidatorsCount     = 600
	threadLimit               int = 12
)

// Beacon client using the standard Beacon HTTP REST API (https://ethereum.github.io/beacon-APIs/)
type StandardHttpClient struct {
	providerAddress string
}

// Create a new client instance
func NewStandardHttpClient(providerAddress string) *StandardHttpClient {
	return &StandardHttpClient{
		providerAddress: providerAddress,
	}
}

// Close the client connection
func (c *StandardHttpClient) Close() error {
	return nil
}

// Get the client's process configuration type
func (c *StandardHttpClient) GetClientType() (beacon.BeaconClientType, error) {
	return beacon.SplitProcess, nil
}

// Get the node's sync status
func (c *StandardHttpClient) GetSyncStatus() (beacon.SyncStatus, error) {

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
func (c *StandardHttpClient) GetEth2Config() (beacon.Eth2Config, error) {

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
func (c *StandardHttpClient) GetEth2DepositContract() (beacon.Eth2DepositContract, error) {

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
func (c *StandardHttpClient) GetBeaconHead() (beacon.BeaconHead, error) {

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
func (c *StandardHttpClient) GetValidatorStatus(pubkey types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {

	return c.getValidatorStatus(hexutil.AddPrefix(pubkey.Hex()), opts)

}
func (c *StandardHttpClient) GetValidatorStatusByIndex(index string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {

	return c.getValidatorStatus(index, opts)

}

func (c *StandardHttpClient) getValidatorStatus(pubkeyOrIndex string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {

	// Return zero status for null pubkeyOrIndex
	if pubkeyOrIndex == "" {
		return beacon.ValidatorStatus{}, nil
	}

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
		Index:                      validator.Index,
		WithdrawalCredentials:      common.BytesToHash(validator.Validator.WithdrawalCredentials),
		Balance:                    uint64(validator.Balance),
		EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
		Status:                     beacon.ValidatorState(validator.Status),
		Slashed:                    validator.Validator.Slashed,
		ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
		ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
		ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
		WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
		Exists:                     true,
	}, nil

}

// Get multiple validators' statuses
func (c *StandardHttpClient) GetValidatorStatuses(pubkeys []types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (map[types.ValidatorPubkey]beacon.ValidatorStatus, error) {

	// The null validator pubkey
	nullPubkey := types.ValidatorPubkey{}

	// Filter out null, invalid and duplicate pubkeys
	realPubkeys := []types.ValidatorPubkey{}
	for _, pubkey := range pubkeys {
		if bytes.Equal(pubkey.Bytes(), types.ValidatorPubkey{}.Bytes()) {
			continue
		}
		isDuplicate := false
		for _, pk := range realPubkeys {
			if bytes.Equal(pubkey.Bytes(), pk.Bytes()) {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}
		// Teku doesn't like invalid pubkeys, so filter them out to make it consistent with other clients
		_, err := bls.PublicKeyFromBytes(pubkey.Bytes())

		if err == nil {
			realPubkeys = append(realPubkeys, pubkey)
		}
	}

	// Convert pubkeys into hex strings
	pubkeysHex := make([]string, len(realPubkeys))
	for vi := 0; vi < len(realPubkeys); vi++ {
		pubkeysHex[vi] = hexutil.AddPrefix(realPubkeys[vi].Hex())
	}

	// Get validators
	validators, err := c.getValidatorsByOpts(pubkeysHex, opts)
	if err != nil {
		return nil, err
	}

	// Build validator status map
	statuses := make(map[types.ValidatorPubkey]beacon.ValidatorStatus)
	for _, validator := range validators.Data {

		// Ignore empty pubkeys
		if bytes.Equal(validator.Validator.Pubkey, nullPubkey[:]) {
			continue
		}

		// Get validator pubkey
		pubkey := types.BytesToValidatorPubkey(validator.Validator.Pubkey)

		// Add status
		statuses[pubkey] = beacon.ValidatorStatus{
			Pubkey:                     types.BytesToValidatorPubkey(validator.Validator.Pubkey),
			Index:                      validator.Index,
			WithdrawalCredentials:      common.BytesToHash(validator.Validator.WithdrawalCredentials),
			Balance:                    uint64(validator.Balance),
			EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
			Status:                     beacon.ValidatorState(validator.Status),
			Slashed:                    validator.Validator.Slashed,
			ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
			ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
			ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
			WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
			Exists:                     true,
		}

	}

	// Put an empty status in for null pubkeys
	statuses[nullPubkey] = beacon.ValidatorStatus{}

	// Return
	return statuses, nil

}

// Get whether validators have sync duties to perform at given epoch
func (c *StandardHttpClient) GetValidatorSyncDuties(indices []string, epoch uint64) (map[string]bool, error) {

	// Perform the post request
	responseBody, status, err := c.postRequest(fmt.Sprintf(RequestValidatorSyncDuties, strconv.FormatUint(epoch, 10)), indices)

	if err != nil {
		return nil, fmt.Errorf("Could not get validator sync duties: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Could not get validator sync duties: HTTP status %d; response body: '%s'", status, string(responseBody))
	}

	var response SyncDutiesResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("Could not decode validator sync duties data: %w", err)
	}

	// Map the results
	validatorMap := make(map[string]bool)

	for _, index := range indices {
		validatorMap[index] = false
		for _, duty := range response.Data {
			if duty.ValidatorIndex == index {
				validatorMap[index] = true
				break
			}
		}
	}

	return validatorMap, nil
}

// Sums proposer duties per validators for a given epoch
func (c *StandardHttpClient) GetValidatorProposerDuties(indices []string, epoch uint64) (map[string]uint64, error) {

	// Perform the post request
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestValidatorProposerDuties, strconv.FormatUint(epoch, 10)))

	if err != nil {
		return nil, fmt.Errorf("Could not get validator proposer duties: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Could not get validator proposer duties: HTTP status %d; response body: '%s'", status, string(responseBody))
	}

	var response ProposerDutiesResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("Could not decode validator proposer duties data: %w", err)
	}

	// Map the results
	proposerMap := make(map[string]uint64)

	for _, index := range indices {
		proposerMap[index] = 0
		for _, duty := range response.Data {
			if duty.ValidatorIndex == index {
				proposerMap[index]++
				break
			}
		}
	}

	return proposerMap, nil
}

// Get a validator's index
func (c *StandardHttpClient) GetValidatorIndex(pubkey types.ValidatorPubkey) (string, error) {

	// Get validator
	pubkeyString := hexutil.AddPrefix(pubkey.Hex())
	validators, err := c.getValidatorsByOpts([]string{pubkeyString}, nil)
	if err != nil {
		return "", err
	}
	if len(validators.Data) == 0 {
		return "", fmt.Errorf("Validator %s index not found.", pubkeyString)
	}
	validator := validators.Data[0]

	// Return validator index
	return validator.Index, nil

}

// Get domain data for a domain type at a given epoch
func (c *StandardHttpClient) GetDomainData(domainType []byte, epoch uint64, useGenesisFork bool) ([]byte, error) {

	// Data
	var wg errgroup.Group
	var genesis GenesisResponse
	var eth2Config Eth2ConfigResponse

	// Get genesis
	wg.Go(func() error {
		var err error
		genesis, err = c.getGenesis()
		return err
	})

	// Get the BN spec as we need the CAPELLA_FORK_VERSION
	wg.Go(func() error {
		var err error
		eth2Config, err = c.getEth2Config()
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return []byte{}, err
	}

	// Get fork version
	var forkVersion []byte
	if useGenesisFork {
		// Used to compute the domain for credential changes
		forkVersion = genesis.Data.GenesisForkVersion
	} else {
		// According to EIP-7044 (https://eips.ethereum.org/EIPS/eip-7044) the CAPELLA_FORK_VERSION should always be used to compute the domain for voluntary exits signatures.
		forkVersion = eth2Config.Data.CapellaForkVersion
	}

	// Compute & return domain
	var dt [4]byte
	copy(dt[:], domainType[:])
	return eth2types.ComputeDomain(dt, forkVersion, genesis.Data.GenesisValidatorsRoot)

}

// Perform a voluntary exit on a validator
func (c *StandardHttpClient) ExitValidator(validatorIndex string, epoch uint64, signature types.ValidatorSignature) error {
	return c.postVoluntaryExit(VoluntaryExitRequest{
		Message: VoluntaryExitMessage{
			Epoch:          uinteger(epoch),
			ValidatorIndex: validatorIndex,
		},
		Signature: signature.Bytes(),
	})
}

// Get the ETH1 data for the target beacon block
func (c *StandardHttpClient) GetEth1DataForEth2Block(blockId string) (beacon.Eth1Data, bool, error) {

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

func (c *StandardHttpClient) GetAttestations(blockId string) ([]beacon.AttestationInfo, bool, error) {
	attestations, exists, err := c.getAttestations(blockId)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	// Add attestation info
	attestationInfo := make([]beacon.AttestationInfo, len(attestations.Data))
	for i, attestation := range attestations.Data {
		bitString := hexutil.RemovePrefix(attestation.AggregationBits)
		attestationInfo[i].SlotIndex = uint64(attestation.Data.Slot)
		attestationInfo[i].CommitteeIndex = uint64(attestation.Data.Index)
		attestationInfo[i].AggregationBits, err = hex.DecodeString(bitString)
		if err != nil {
			return nil, false, fmt.Errorf("Error decoding aggregation bits for attestation %d of block %s: %w", i, blockId, err)
		}
	}

	return attestationInfo, true, nil
}

func (c *StandardHttpClient) GetBeaconBlock(blockId string) (beacon.BeaconBlock, bool, error) {
	block, exists, err := c.getBeaconBlock(blockId)
	if err != nil {
		return beacon.BeaconBlock{}, false, err
	}
	if !exists {
		return beacon.BeaconBlock{}, false, nil
	}

	beaconBlock := beacon.BeaconBlock{
		Slot:          uint64(block.Data.Message.Slot),
		ProposerIndex: block.Data.Message.ProposerIndex,
	}

	// Execution payload only exists after the merge, so check for its existence
	if block.Data.Message.Body.ExecutionPayload == nil {
		beaconBlock.HasExecutionPayload = false
	} else {
		beaconBlock.HasExecutionPayload = true
		beaconBlock.FeeRecipient = common.BytesToAddress(block.Data.Message.Body.ExecutionPayload.FeeRecipient)
		beaconBlock.ExecutionBlockNumber = uint64(block.Data.Message.Body.ExecutionPayload.BlockNumber)
	}

	// Add attestation info
	for i, attestation := range block.Data.Message.Body.Attestations {
		bitString := hexutil.RemovePrefix(attestation.AggregationBits)
		info := beacon.AttestationInfo{
			SlotIndex:      uint64(attestation.Data.Slot),
			CommitteeIndex: uint64(attestation.Data.Index),
		}
		info.AggregationBits, err = hex.DecodeString(bitString)
		if err != nil {
			return beacon.BeaconBlock{}, false, fmt.Errorf("Error decoding aggregation bits for attestation %d of block %s: %w", i, blockId, err)
		}
		beaconBlock.Attestations = append(beaconBlock.Attestations, info)
	}

	return beaconBlock, true, nil
}

func (c *StandardHttpClient) GetBeaconBlockHeader(blockId string) (beacon.BeaconBlockHeader, bool, error) {
	block, exists, err := c.getBeaconBlockHeader(blockId)
	if err != nil {
		fmt.Printf("Error getting beacon block header: %s\n", err.Error())
		return beacon.BeaconBlockHeader{}, false, err
	}
	if !exists {
		return beacon.BeaconBlockHeader{}, false, nil
	}
	beaconBlock := beacon.BeaconBlockHeader{
		Slot:          uint64(block.Data.Header.Message.Slot),
		ProposerIndex: block.Data.Header.Message.ProposerIndex,
	}
	return beaconBlock, true, nil
}

// Get the attestation committees for the given epoch, or the current epoch if nil
func (c *StandardHttpClient) GetCommitteesForEpoch(epoch *uint64) (beacon.Committees, error) {
	response, err := c.getCommittees("head", epoch)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Perform a withdrawal credentials change on a validator
func (c *StandardHttpClient) ChangeWithdrawalCredentials(validatorIndex string, fromBlsPubkey types.ValidatorPubkey, toExecutionAddress common.Address, signature types.ValidatorSignature) error {
	return c.postWithdrawalCredentialsChange(BLSToExecutionChangeRequest{
		Message: BLSToExecutionChangeMessage{
			ValidatorIndex:     validatorIndex,
			FromBLSPubkey:      fromBlsPubkey[:],
			ToExecutionAddress: toExecutionAddress[:],
		},
		Signature: signature.Bytes(),
	})
}

// Get sync status
func (c *StandardHttpClient) getSyncStatus() (SyncStatusResponse, error) {
	responseBody, status, err := c.getRequest(RequestSyncStatusPath)
	if err != nil {
		return SyncStatusResponse{}, fmt.Errorf("Could not get node sync status: %w", err)
	}
	if status != http.StatusOK {
		return SyncStatusResponse{}, fmt.Errorf("Could not get node sync status: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var syncStatus SyncStatusResponse
	if err := json.Unmarshal(responseBody, &syncStatus); err != nil {
		return SyncStatusResponse{}, fmt.Errorf("Could not decode node sync status: %w", err)
	}
	return syncStatus, nil
}

// Get the eth2 config
func (c *StandardHttpClient) getEth2Config() (Eth2ConfigResponse, error) {
	responseBody, status, err := c.getRequest(RequestEth2ConfigPath)
	if err != nil {
		return Eth2ConfigResponse{}, fmt.Errorf("Could not get eth2 config: %w", err)
	}
	if status != http.StatusOK {
		return Eth2ConfigResponse{}, fmt.Errorf("Could not get eth2 config: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var eth2Config Eth2ConfigResponse
	if err := json.Unmarshal(responseBody, &eth2Config); err != nil {
		return Eth2ConfigResponse{}, fmt.Errorf("Could not decode eth2 config: %w", err)
	}
	return eth2Config, nil
}

// Get the eth2 deposit contract info
func (c *StandardHttpClient) getEth2DepositContract() (Eth2DepositContractResponse, error) {
	responseBody, status, err := c.getRequest(RequestEth2DepositContractMethod)
	if err != nil {
		return Eth2DepositContractResponse{}, fmt.Errorf("Could not get eth2 deposit contract: %w", err)
	}
	if status != http.StatusOK {
		return Eth2DepositContractResponse{}, fmt.Errorf("Could not get eth2 deposit contract: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var eth2DepositContract Eth2DepositContractResponse
	if err := json.Unmarshal(responseBody, &eth2DepositContract); err != nil {
		return Eth2DepositContractResponse{}, fmt.Errorf("Could not decode eth2 deposit contract: %w", err)
	}
	return eth2DepositContract, nil
}

// Get genesis information
func (c *StandardHttpClient) getGenesis() (GenesisResponse, error) {
	responseBody, status, err := c.getRequest(RequestGenesisPath)
	if err != nil {
		return GenesisResponse{}, fmt.Errorf("Could not get genesis data: %w", err)
	}
	if status != http.StatusOK {
		return GenesisResponse{}, fmt.Errorf("Could not get genesis data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var genesis GenesisResponse
	if err := json.Unmarshal(responseBody, &genesis); err != nil {
		return GenesisResponse{}, fmt.Errorf("Could not decode genesis: %w", err)
	}
	return genesis, nil
}

// Get finality checkpoints
func (c *StandardHttpClient) getFinalityCheckpoints(stateId string) (FinalityCheckpointsResponse, error) {
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestFinalityCheckpointsPath, stateId))
	if err != nil {
		return FinalityCheckpointsResponse{}, fmt.Errorf("Could not get finality checkpoints: %w", err)
	}
	if status != http.StatusOK {
		return FinalityCheckpointsResponse{}, fmt.Errorf("Could not get finality checkpoints: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var finalityCheckpoints FinalityCheckpointsResponse
	if err := json.Unmarshal(responseBody, &finalityCheckpoints); err != nil {
		return FinalityCheckpointsResponse{}, fmt.Errorf("Could not decode finality checkpoints: %w", err)
	}
	return finalityCheckpoints, nil
}

// Get fork
func (c *StandardHttpClient) getFork(stateId string) (ForkResponse, error) {
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestForkPath, stateId))
	if err != nil {
		return ForkResponse{}, fmt.Errorf("Could not get fork data: %w", err)
	}
	if status != http.StatusOK {
		return ForkResponse{}, fmt.Errorf("Could not get fork data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var fork ForkResponse
	if err := json.Unmarshal(responseBody, &fork); err != nil {
		return ForkResponse{}, fmt.Errorf("Could not decode fork data: %w", err)
	}
	return fork, nil
}

// Get validators
func (c *StandardHttpClient) getValidators(stateId string, pubkeys []string) (ValidatorsResponse, error) {
	var query string
	if len(pubkeys) > 0 {
		query = fmt.Sprintf("?id=%s", strings.Join(pubkeys, ","))
	}
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestValidatorsPath, stateId) + query)
	if err != nil {
		return ValidatorsResponse{}, fmt.Errorf("Could not get validators: %w", err)
	}
	if status != http.StatusOK {
		return ValidatorsResponse{}, fmt.Errorf("Could not get validators: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var validators ValidatorsResponse
	if err := json.Unmarshal(responseBody, &validators); err != nil {
		return ValidatorsResponse{}, fmt.Errorf("Could not decode validators: %w", err)
	}
	return validators, nil
}

// Get validators by pubkeys and status options
func (c *StandardHttpClient) getValidatorsByOpts(pubkeysOrIndices []string, opts *beacon.ValidatorStatusOptions) (ValidatorsResponse, error) {

	// Get state ID
	var stateId string
	if opts == nil {
		stateId = "head"
	} else if opts.Slot != nil {
		stateId = strconv.FormatInt(int64(*opts.Slot), 10)
	} else if opts.Epoch != nil {

		// Get eth2 config
		eth2Config, err := c.getEth2Config()
		if err != nil {
			return ValidatorsResponse{}, err
		}

		// Get slot nuimber
		slot := *opts.Epoch * uint64(eth2Config.Data.SlotsPerEpoch)
		stateId = strconv.FormatInt(int64(slot), 10)

	} else {
		return ValidatorsResponse{}, fmt.Errorf("must specify a slot or epoch when calling getValidatorsByOpts")
	}

	count := len(pubkeysOrIndices)
	data := make([]Validator, count)
	validFlags := make([]bool, count)
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	for i := 0; i < count; i += MaxRequestValidatorsCount {
		i := i
		max := i + MaxRequestValidatorsCount
		if max > count {
			max = count
		}

		wg.Go(func() error {
			// Get & add validators
			batch := pubkeysOrIndices[i:max]
			validators, err := c.getValidators(stateId, batch)
			if err != nil {
				return fmt.Errorf("error getting validator statuses: %w", err)
			}
			for j, responseData := range validators.Data {
				data[i+j] = responseData
				validFlags[i+j] = true
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return ValidatorsResponse{}, fmt.Errorf("error getting validators by opts: %w", err)
	}

	// Clip all of the empty responses so only the valid pubkeys get returned
	trueData := make([]Validator, 0, count)
	for i, valid := range validFlags {
		if valid {
			trueData = append(trueData, data[i])
		}
	}

	return ValidatorsResponse{Data: trueData}, nil
}

// Send voluntary exit request
func (c *StandardHttpClient) postVoluntaryExit(request VoluntaryExitRequest) error {
	responseBody, status, err := c.postRequest(RequestVoluntaryExitPath, request)
	if err != nil {
		return fmt.Errorf("Could not broadcast exit for validator at index %s: %w", request.Message.ValidatorIndex, err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("Could not broadcast exit for validator at index %s: HTTP status %d; response body: '%s'", request.Message.ValidatorIndex, status, string(responseBody))
	}
	return nil
}

// Get the target beacon block
func (c *StandardHttpClient) getAttestations(blockId string) (AttestationsResponse, bool, error) {
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestAttestationsPath, blockId))
	if err != nil {
		return AttestationsResponse{}, false, fmt.Errorf("Could not get attestations data for slot %s: %w", blockId, err)
	}
	if status == http.StatusNotFound {
		return AttestationsResponse{}, false, nil
	}
	if status != http.StatusOK {
		return AttestationsResponse{}, false, fmt.Errorf("Could not get attestations data for slot %s: HTTP status %d; response body: '%s'", blockId, status, string(responseBody))
	}
	var attestations AttestationsResponse
	if err := json.Unmarshal(responseBody, &attestations); err != nil {
		return AttestationsResponse{}, false, fmt.Errorf("Could not decode attestations data for slot %s: %w", blockId, err)
	}
	return attestations, true, nil
}

// Get the target beacon block
func (c *StandardHttpClient) getBeaconBlock(blockId string) (BeaconBlockResponse, bool, error) {
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestBeaconBlockPath, blockId))
	if err != nil {
		return BeaconBlockResponse{}, false, fmt.Errorf("Could not get beacon block data: %w", err)
	}
	if status == http.StatusNotFound {
		return BeaconBlockResponse{}, false, nil
	}
	if status != http.StatusOK {
		return BeaconBlockResponse{}, false, fmt.Errorf("Could not get beacon block data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var beaconBlock BeaconBlockResponse
	if err := json.Unmarshal(responseBody, &beaconBlock); err != nil {
		return BeaconBlockResponse{}, false, fmt.Errorf("Could not decode beacon block data: %w", err)
	}
	return beaconBlock, true, nil
}

// Get the specified beacon block header
func (c *StandardHttpClient) getBeaconBlockHeader(blockId string) (BeaconBlockHeaderResponse, bool, error) {
	responseBody, status, err := c.getRequest(fmt.Sprintf(RequestBeaconBlockHeaderPath, blockId))
	if err != nil {
		return BeaconBlockHeaderResponse{}, false, fmt.Errorf("could not get beacon block header data: %w", err)
	}
	if status == http.StatusNotFound {
		return BeaconBlockHeaderResponse{}, false, nil
	}
	if status != http.StatusOK {
		return BeaconBlockHeaderResponse{}, false, fmt.Errorf("could not get beacon block header data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var beaconBlock BeaconBlockHeaderResponse
	if err := json.Unmarshal(responseBody, &beaconBlock); err != nil {
		return BeaconBlockHeaderResponse{}, false, fmt.Errorf("could not decode beacon block header data: %w", err)
	}
	return beaconBlock, true, nil
}

type committeesDecoder struct {
	decoder       *json.Decoder
	currentReader *io.ReadCloser
}

// Read will be called by the json decoder to request more bytes of data from
// the beacon node's committees response. Since the decoder is reused, we
// need to avoid sending it io.EOF, or it will enter an unusable state and can
// not be reused later.
//
// On subsequent calls to Decode, the decoder resets its internal buffer, which
// means any data it reads between the last json token and EOF is correctly
// discarded.
func (c *committeesDecoder) Read(p []byte) (int, error) {
	n, err := (*c.currentReader).Read(p)
	if err == io.EOF {
		return n, nil
	}

	return n, err
}

var committeesDecoderPool sync.Pool = sync.Pool{
	New: func() any {
		var out committeesDecoder

		out.decoder = json.NewDecoder(&out)
		return &out
	},
}

// Get the committees for the epoch
func (c *StandardHttpClient) getCommittees(stateId string, epoch *uint64) (CommitteesResponse, error) {
	var committees CommitteesResponse

	query := ""
	if epoch != nil {
		query = fmt.Sprintf("?epoch=%d", *epoch)
	}

	// Committees responses are large, so let the json decoder read it in a buffered fashion
	reader, status, err := c.getRequestReader(fmt.Sprintf(RequestCommitteePath, stateId) + query)
	if err != nil {
		return CommitteesResponse{}, fmt.Errorf("Could not get committees: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	if status != http.StatusOK {
		body, _ := io.ReadAll(reader)
		return CommitteesResponse{}, fmt.Errorf("Could not get committees: HTTP status %d; response body: '%s'", status, string(body))
	}

	d := committeesDecoderPool.Get().(*committeesDecoder)
	defer func() {
		d.currentReader = nil
		committeesDecoderPool.Put(d)
	}()

	d.currentReader = &reader

	// Begin decoding
	if err := d.decoder.Decode(&committees); err != nil {
		return CommitteesResponse{}, fmt.Errorf("Could not decode committees: %w", err)
	}

	return committees, nil
}

// Send withdrawal credentials change request
func (c *StandardHttpClient) postWithdrawalCredentialsChange(request BLSToExecutionChangeRequest) error {
	requestArray := []BLSToExecutionChangeRequest{request} // This route must be wrapped in an array
	responseBody, status, err := c.postRequest(RequestWithdrawalCredentialsChangePath, requestArray)
	if err != nil {
		return fmt.Errorf("Could not broadcast withdrawal credentials change for validator %d: %w", request.Message.ValidatorIndex, err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("Could not broadcast withdrawal credentials change for validator %d: HTTP status %d; response body: '%s'", request.Message.ValidatorIndex, status, string(responseBody))
	}
	return nil
}

// Make a GET request but do not read its body yet (allows buffered decoding)
func (c *StandardHttpClient) getRequestReader(requestPath string) (io.ReadCloser, int, error) {

	// Send request
	response, err := http.Get(fmt.Sprintf(RequestUrlFormat, c.providerAddress, requestPath))
	if err != nil {
		return nil, 0, err
	}

	return response.Body, response.StatusCode, nil
}

// Make a GET request to the beacon node and read the body of the response
func (c *StandardHttpClient) getRequest(requestPath string) ([]byte, int, error) {

	// Send request
	reader, status, err := c.getRequestReader(requestPath)
	if err != nil {
		return []byte{}, 0, err
	}
	defer func() {
		_ = reader.Close()
	}()

	// Get response
	body, err := io.ReadAll(reader)
	if err != nil {
		return []byte{}, 0, err
	}

	// Return
	return body, status, nil
}

// Make a POST request to the beacon node
func (c *StandardHttpClient) postRequest(requestPath string, requestBody interface{}) ([]byte, int, error) {

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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, 0, err
	}

	// Return
	return body, response.StatusCode, nil

}
