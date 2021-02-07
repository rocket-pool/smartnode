package nimbus

import (
    "bytes"
    "fmt"
    "strconv"
    "time"

    "github.com/ethereum/go-ethereum/common"
    rpc "github.com/ethereum/go-ethereum/rpc"
    "github.com/rocket-pool/rocketpool-go/types"
    eth2types "github.com/wealdtech/go-eth2-types/v2"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/utils/eth2"
    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// Config
const (
    RequestSyncStatusMethod          = "get_v1_node_syncing"
    RequestEth2ConfigMethod          = "get_v1_config_spec"
    RequestGenesisMethod             = "get_v1_beacon_genesis"
    RequestFinalityCheckpointsMethod = "get_v1_beacon_states_finality_checkpoints"
    RequestForkMethod                = "get_v1_beacon_states_fork"
    RequestValidatorsMethod          = "get_v1_beacon_states_stateId_validators_validatorId"
    RequestVoluntaryExitMethod       = "get_v1_beacon_pool_voluntary_exits"

    MaxRequestValidatorsCount = 600
)

// Nimbus client
type Client struct {
    client *rpc.Client
}

// Create new Nimbus client
func NewClient(providerAddress string) (*Client, error) {

    // Start the RPC connection
    client, err := rpc.DialHTTP("http://" + providerAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not connect to Nimbus RPC server: %s", err)
    }
    return &Client{
        client: client,
    }, nil
}

// Close the client connection
func (c *Client) Close() {
    c.client.Close()
}

// Get the beacon client type
func (c *Client) GetClientType() beacon.BeaconClientType {
    return beacon.SingleProcess
}

// Get the node's sync status
func (c *Client) GetSyncStatus() (beacon.SyncStatus, error) {

    // Get sync status
    syncStatus, err := c.getSyncStatus()
    if err != nil {
        return beacon.SyncStatus{}, err
    }

    // Return response
    isSyncing := (syncStatus.SyncDistance != 0)
    return beacon.SyncStatus{
        Syncing: isSyncing,
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
        GenesisForkVersion:    genesis.GenesisForkVersion,
        GenesisValidatorsRoot: genesis.GenesisValidatorsRoot,
        GenesisEpoch:          0,
        GenesisTime:           uint64(genesis.GenesisTime),
        SecondsPerEpoch:       uint64(eth2Config.SecondsPerSlot * eth2Config.SlotsPerEpoch),
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
        FinalizedEpoch:         uint64(finalityCheckpoints.Finalized.Epoch),
        JustifiedEpoch:         uint64(finalityCheckpoints.CurrentJustified.Epoch),
        PreviousJustifiedEpoch: uint64(finalityCheckpoints.PreviousJustified.Epoch),
    }, nil

}

// Get a validator's status
func (c *Client) GetValidatorStatus(pubkey types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {

    // Get validator
    validators, err := c.getValidatorsByOpts([]types.ValidatorPubkey{pubkey}, opts)
    if err != nil {
        return beacon.ValidatorStatus{}, err
    }
    if len(validators) == 0 {
        return beacon.ValidatorStatus{}, nil
    }
    validator := validators[0]

    // Return response
    return beacon.ValidatorStatus{
        Pubkey:                     types.BytesToValidatorPubkey(validator.Validator.Pubkey),
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

    // Get validators
    validators, err := c.getValidatorsByOpts(pubkeys, opts)
    if err != nil {
        return map[types.ValidatorPubkey]beacon.ValidatorStatus{}, err
    }

    // Build validator status map
    statuses := make(map[types.ValidatorPubkey]beacon.ValidatorStatus)
    for _, validator := range validators {

        // Get validator pubkey
        pubkey := types.BytesToValidatorPubkey(validator.Validator.Pubkey)

        // Add status
        statuses[pubkey] = beacon.ValidatorStatus{
            Pubkey:                     types.BytesToValidatorPubkey(validator.Validator.Pubkey),
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

// Get a validator's index
func (c *Client) GetValidatorIndex(pubkey types.ValidatorPubkey) (uint64, error) {

    // Get validator
    validators, err := c.getValidatorsByOpts([]types.ValidatorPubkey{pubkey}, nil)
    if err != nil {
        return 0, err
    }
    if len(validators) == 0 {
        return 0, fmt.Errorf("Validator %s index not found.", pubkey.Hex())
    }
    validator := validators[0]

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
    if epoch < uint64(fork.Epoch) {
        forkVersion = fork.PreviousVersion
    } else {
        forkVersion = fork.CurrentVersion
    }

    // Compute & return domain
    var dt [4]byte
    copy(dt[:], domainType[:])
    return eth2types.Domain(dt, forkVersion, genesis.GenesisValidatorsRoot), nil

}

// Perform a voluntary exit on a validator
func (c *Client) ExitValidator(validatorIndex, epoch uint64, signature types.ValidatorSignature) error {
    return c.postVoluntaryExit(VoluntaryExitRequest{
        Message: VoluntaryExitMessage{
            Epoch:          epoch,
            ValidatorIndex: validatorIndex,
        },
        Signature: signature.Bytes(),
    })
}

// Get sync status
func (c *Client) getSyncStatus() (SyncStatusResponse, error) {
    var syncStatus SyncStatusResponse
    if err := c.client.Call(&syncStatus, RequestSyncStatusMethod); err != nil {
        return SyncStatusResponse{}, fmt.Errorf("Could not get node sync status: %w", err)
    }
    return syncStatus, nil
}

// Get the eth2 config
func (c *Client) getEth2Config() (Eth2ConfigResponse, error) {
    var eth2Config Eth2ConfigResponse
    if err := c.client.Call(&eth2Config, RequestEth2ConfigMethod); err != nil {
        return Eth2ConfigResponse{}, fmt.Errorf("Could not get eth2 config: %w", err)
    }
    return eth2Config, nil
}

// Get genesis information
func (c *Client) getGenesis() (GenesisResponse, error) {
    var genesis GenesisResponse
    if err := c.client.Call(&genesis, RequestGenesisMethod); err != nil {
        return GenesisResponse{}, fmt.Errorf("Could not get genesis data: %w", err)
    }
    return genesis, nil
}

// Get finality checkpoints
func (c *Client) getFinalityCheckpoints(stateId string) (FinalityCheckpointsResponse, error) {
    var finalityCheckpoints FinalityCheckpointsResponse
    if err := c.client.Call(&finalityCheckpoints, RequestFinalityCheckpointsMethod, stateId); err != nil {
        return FinalityCheckpointsResponse{}, fmt.Errorf("Could not get finality checkpoints: %w", err)
    }
    return finalityCheckpoints, nil
}

// Get fork
func (c *Client) getFork(stateId string) (ForkResponse, error) {
    var fork ForkResponse
    if err := c.client.Call(&fork, RequestForkMethod, stateId); err != nil {
        return ForkResponse{}, fmt.Errorf("Could not get fork data: %w", err)
    }
    return fork, nil
}

// Get validators
func (c *Client) getValidators(stateId string, pubkeys []string) ([]Validator, error) {
    var validators []Validator
    params := append([]string{stateId}, pubkeys...)
    if err := c.client.Call(&validators, RequestValidatorsMethod, params); err != nil {
        return []Validator{}, fmt.Errorf("Could not get validators: %w", err)
    }
    return validators, nil
}

// Get validators by pubkeys and status options
func (c *Client) getValidatorsByOpts(pubkeys []types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) ([]Validator, error) {

    // Get state ID
    var stateId string
    if opts == nil {
        stateId = "head"
    } else {

        // Get eth2 config
        eth2Config, err := c.getEth2Config()
        if err != nil {
            return []Validator{}, err
        }

        // Get slot nuimber
        slot := opts.Epoch * uint64(eth2Config.SlotsPerEpoch)
        stateId = strconv.FormatInt(int64(slot), 10)

    }

    // Get validators
    if len(pubkeys) <= MaxRequestValidatorsCount {

        // Get validator pubkeys
        pubkeysHex := make([]string, len(pubkeys))
        for ki, pubkey := range pubkeys {
            pubkeysHex[ki] = hexutil.AddPrefix(pubkey.Hex())
        }

        // Get & return validators
        return c.getValidators(stateId, pubkeysHex)

    } else {

        // Get all validators
        validators, err := c.getValidators(stateId, []string{})
        if err != nil {
            return []Validator{}, err
        }

        // Filter validator set by pubkeys and return
        response := []Validator{}
        for _, validator := range validators {
            var found bool
            for _, pubkey := range pubkeys {
                if bytes.Equal(validator.Validator.Pubkey, pubkey.Bytes()) {
                    found = true
                    break
                }
            }
            if !found {
                continue
            }
            response = append(response, validator)
        }
        return response, nil

    }

}

// Send voluntary exit request
func (c *Client) postVoluntaryExit(request VoluntaryExitRequest) error {
    if err := c.client.Call(nil, RequestVoluntaryExitMethod, request); err != nil {
        return fmt.Errorf("Could not broadcast exit for validator at index %d: %w", request.Message.ValidatorIndex, err)
    }
    return nil
}
