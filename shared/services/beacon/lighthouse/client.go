package lighthouse

import (
    "bytes"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strconv"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/types"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
    bytesutil "github.com/rocket-pool/smartnode/shared/utils/bytes"
    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)


// Config
const (
    RequestUrlFormat = "%s://%s%s"
    RequestProtocol = "http"
    RequestContentType = "application/json"

    RequestEth2ConfigPath = "/spec"
    RequestSlotsPerEpochPath = "/spec/slots_per_epoch"
    RequestBeaconHeadPath = "/beacon/head"
    RequestBeaconStateRootPath = "/beacon/state_root"
    RequestValidatorsPath = "/beacon/validators"
)


// Beacon request types
type ValidatorsRequest struct {
    StateRoot string                `json:"state_root,omitempty"`
    Pubkeys []string                `json:"pubkeys"`
}

// Beacon response types
type Eth2ConfigResponse struct {
    GenesisForkVersion string       `json:"genesis_fork_version"`
    DomainDeposit uint64            `json:"domain_deposit"`
    DomainVoluntaryExit uint64      `json:"domain_voluntary_exit"`
}
type BeaconHeadResponse struct {
    Slot uint64                     `json:"slot"`
    FinalizedSlot uint64            `json:"finalized_slot"`
    JustifiedSlot uint64            `json:"justified_slot"`
    PreviousJustifiedSlot uint64    `json:"previous_justified_slot"`
}
type ValidatorResponse struct {
    Balance uint64                  `json:"balance"`
    Validator struct {
        Pubkey string                       `json:"pubkey"`
        WithdrawalCredentials string        `json:"withdrawal_credentials"`
        EffectiveBalance uint64             `json:"effective_balance"`
        Slashed bool                        `json:"slashed"`
        ActivationEligibilityEpoch uint64   `json:"activation_eligibility_epoch"`
        ActivationEpoch uint64              `json:"activation_epoch"`
        ExitEpoch uint64                    `json:"exit_epoch"`
        WithdrawableEpoch uint64            `json:"withdrawable_epoch"`
    }                               `json:"validator"`
}


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
func (c *Client) Close() {}


// Get the eth2 config
func (c *Client) GetEth2Config() (beacon.Eth2Config, error) {

    // Data
    var wg errgroup.Group
    var config Eth2ConfigResponse
    var slotsPerEpoch uint64

    // Request eth2 config
    wg.Go(func() error {
        responseBody, err := c.getRequest(RequestEth2ConfigPath)
        if err != nil {
            return fmt.Errorf("Could not get eth2 config: %w", err)
        }
        if err := json.Unmarshal(responseBody, &config); err != nil {
            return fmt.Errorf("Could not decode eth2 config: %w", err)
        }
        return nil
    })

    // Request slots per epoch
    wg.Go(func() error {
        responseBody, err := c.getRequest(RequestSlotsPerEpochPath)
        if err != nil {
            return fmt.Errorf("Could not get slots per epoch: %w", err)
        }
        slotsPerEpoch, err = strconv.ParseUint(string(responseBody), 10, 64)
        if err != nil {
            return fmt.Errorf("Could not parse slots per epoch: %w", err)
        }
        return nil
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return beacon.Eth2Config{}, err
    }

    // Create response
    response := beacon.Eth2Config{
        DomainDeposit: bytesutil.Bytes4(config.DomainDeposit),
        DomainVoluntaryExit: bytesutil.Bytes4(config.DomainVoluntaryExit),
    }

    // Decode hex data and update
    if genesisForkVersion, err := hex.DecodeString(hexutil.RemovePrefix(config.GenesisForkVersion)); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode genesis fork version: %w", err)
    } else {
        response.GenesisForkVersion = genesisForkVersion
    }

    // Return
    return response, nil

}


// Get the beacon head
func (c *Client) GetBeaconHead() (beacon.BeaconHead, error) {

    // Data
    var wg errgroup.Group
    var head BeaconHeadResponse
    var slotsPerEpoch uint64

    // Request beacon head
    wg.Go(func() error {
        responseBody, err := c.getRequest(RequestBeaconHeadPath)
        if err != nil {
            return fmt.Errorf("Could not get beacon head: %w", err)
        }
        if err := json.Unmarshal(responseBody, &head); err != nil {
            return fmt.Errorf("Could not decode beacon head: %w", err)
        }
        return nil
    })

    // Request slots per epoch
    wg.Go(func() error {
        responseBody, err := c.getRequest(RequestSlotsPerEpochPath)
        if err != nil {
            return fmt.Errorf("Could not get slots per epoch: %w", err)
        }
        slotsPerEpoch, err = strconv.ParseUint(string(responseBody), 10, 64)
        if err != nil {
            return fmt.Errorf("Could not parse slots per epoch: %w", err)
        }
        return nil
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return beacon.BeaconHead{}, err
    }

    // Return response
    return beacon.BeaconHead{
        Slot: head.Slot,
        FinalizedSlot: head.FinalizedSlot,
        JustifiedSlot: head.JustifiedSlot,
        PreviousJustifiedSlot: head.PreviousJustifiedSlot,
        Epoch: head.Slot / slotsPerEpoch,
        FinalizedEpoch: head.FinalizedSlot / slotsPerEpoch,
        JustifiedEpoch: head.JustifiedSlot / slotsPerEpoch,
        PreviousJustifiedEpoch: head.PreviousJustifiedSlot / slotsPerEpoch,
    }, nil

}


// Get a validator's status
func (c *Client) GetValidatorStatus(pubkey types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {

    // Request
    responseBody, err := c.postRequest(RequestValidatorsPath, ValidatorsRequest{
        Pubkeys: []string{hexutil.AddPrefix(pubkey.Hex())},
    })
    if err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not get validator status: %w", err)
    }

    // Unmarshal response
    var validators []ValidatorResponse
    if err := json.Unmarshal(responseBody, &validators); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode validator status: %w", err)
    }
    validator := validators[0]

    // Check if validator exists
    // Pubkey is empty if validator is null in response
    if validator.Validator.Pubkey == "" {
        return beacon.ValidatorStatus{}, nil
    }

    // Create response
    response := beacon.ValidatorStatus{
        Balance: validator.Balance,
        EffectiveBalance: validator.Validator.EffectiveBalance,
        Slashed: validator.Validator.Slashed,
        ActivationEligibilityEpoch: validator.Validator.ActivationEligibilityEpoch,
        ActivationEpoch: validator.Validator.ActivationEpoch,
        ExitEpoch: validator.Validator.ExitEpoch,
        WithdrawableEpoch: validator.Validator.WithdrawableEpoch,
        Exists: true, 
    }

    // Decode hex data and update
    if pubkey, err := hex.DecodeString(hexutil.RemovePrefix(validator.Validator.Pubkey)); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode validator pubkey: %w", err)
    } else {
        response.Pubkey = types.BytesToValidatorPubkey(pubkey)
    }
    if withdrawalCredentials, err := hex.DecodeString(hexutil.RemovePrefix(validator.Validator.WithdrawalCredentials)); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode validator withdrawal credentials: %w", err)
    } else {
        response.WithdrawalCredentials = common.BytesToHash(withdrawalCredentials)
    }

    // Return
    return response, nil

}


// Make a GET request to the beacon node
func (c *Client) getRequest(requestPath string) ([]byte, error) {

    // Send request
    response, err := http.Get(fmt.Sprintf(RequestUrlFormat, RequestProtocol, c.providerAddress, requestPath))
    if err != nil {
        return []byte{}, err
    }
    defer response.Body.Close()

    // Get response
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return []byte{}, err
    }

    // Return
    return body, nil

}


// Make a POST request to the beacon node
func (c *Client) postRequest(requestPath string, requestBody interface{}) ([]byte, error) {

    // Get request body
    requestBodyBytes, err := json.Marshal(requestBody)
    if err != nil {
        return []byte{}, err
    }
    requestBodyReader := bytes.NewReader(requestBodyBytes)

    // Send request
    response, err := http.Post(fmt.Sprintf(RequestUrlFormat, RequestProtocol, c.providerAddress, requestPath), RequestContentType, requestBodyReader)
    if err != nil {
        return []byte{}, err
    }
    defer response.Body.Close()

    // Get response
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return []byte{}, err
    }

    // Return
    return body, nil

}

