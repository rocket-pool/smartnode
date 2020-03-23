package beacon

import (
    "bytes"
    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "strconv"
)


// Beacon config
const REQUEST_CONTENT_TYPE string = "application/json"

// Beacon endpoints
const REQUEST_ETH2_CONFIG_PATH string = "/spec"
const REQUEST_SLOTS_PER_EPOCH_PATH string = "/spec/slots_per_epoch"
const REQUEST_BEACON_HEAD_PATH string = "/beacon/head"
const REQUEST_VALIDATORS_PATH string = "/beacon/validators"


// Beacon request types
type ValidatorsRequest struct {
    Pubkeys []string                `json:"pubkeys"`
}

// Beacon response types
type Eth2ConfigResponse struct {
    GenesisForkVersion string       `json:"genesis_fork_version"`
    BLSWithdrawalPrefixByte string  `json:"bls_withdrawal_prefix_byte"`
    DomainBeaconProposer uint64     `json:"domain_beacon_proposer"`
    DomainBeaonAttester uint64      `json:"domain_beacon_attester"`
    DomainRandao uint64             `json:"domain_randao"`
    DomainDeposit uint64            `json:"domain_deposit"`
    DomainVoluntaryExit uint64      `json:"domain_voluntary_exit"`
    SlotsPerEpoch uint64
}
type BeaconHeadResponse struct {
    Slot uint64                     `json:"slot"`
    FinalizedSlot uint64            `json:"finalized_slot"`
    JustifiedSlot uint64            `json:"justified_slot"`
    Epoch uint64
    FinalizedEpoch uint64
    JustifiedEpoch uint64
}
type ValidatorResponse struct {
    Pubkey string                   `json:"pubkey"`
    ValidatorIndex uint64           `json:"validator_index"`
    Balance uint64                  `json:"balance"`
    Validator struct {
        WithdrawalCredentials string        `json:"withdrawal_credentials"`
        EffectiveBalance uint64             `json:"effective_balance"`
        Slashed bool                        `json:"slashed"`
        ActivationEligibilityEpoch uint64   `json:"activation_eligibility_epoch"`
        ActivationEpoch uint64              `json:"activation_epoch"`
        ExitEpoch uint64                    `json:"exit_epoch"`
        WithdrawableEpoch uint64            `json:"withdrawable_epoch"`
    }                               `json:"validator"`
    Exists bool
}


// Client
type Client struct {
    providerUrl string
}


/**
 * Create client
 */
func NewClient(providerUrl string) *Client {
    return &Client{
        providerUrl: providerUrl,
    }
}


/**
 * Get the eth2 config
 */
func (c *Client) GetEth2Config() (*Eth2ConfigResponse, error) {

    // Data channels
    responseChannel := make(chan Eth2ConfigResponse)
    slotsPerEpochChannel := make(chan uint64)
    errorChannel := make(chan error)

    // Request eth2 config
    go (func() {
        var response Eth2ConfigResponse
        if responseBody, err := c.getRequest(REQUEST_ETH2_CONFIG_PATH); err != nil {
            errorChannel <- errors.New("Error retrieving eth2 config: " + err.Error())
        } else if err := json.Unmarshal(responseBody, &response); err != nil {
            errorChannel <- errors.New("Error unpacking eth2 config: " + err.Error())
        } else {
            responseChannel <- response
        }
    })()

    // Request slots per epoch
    go (func() {
        if responseBody, err := c.getRequest(REQUEST_SLOTS_PER_EPOCH_PATH); err != nil {
            errorChannel <- errors.New("Error retrieving slots per epoch: " + err.Error())
        } else if slotsPerEpoch, err := strconv.Atoi(string(responseBody)); err != nil {
            errorChannel <- errors.New("Error unpacking slots per epoch: " + err.Error())
        } else {
            slotsPerEpochChannel <- uint64(slotsPerEpoch)
        }
    })()

    // Receive data
    var response Eth2ConfigResponse
    var slotsPerEpoch uint64
    for received := 0; received < 2; {
        select {
            case response = <-responseChannel:
                received++
            case slotsPerEpoch = <-slotsPerEpochChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Update response & return
    response.SlotsPerEpoch = slotsPerEpoch
    return &response, nil

}


/**
 * Get the beacon head
 */
func (c *Client) GetBeaconHead() (*BeaconHeadResponse, error) {

    // Data channels
    responseChannel := make(chan BeaconHeadResponse)
    slotsPerEpochChannel := make(chan uint64)
    errorChannel := make(chan error)

    // Request beacon head
    go (func() {
        var response BeaconHeadResponse
        if responseBody, err := c.getRequest(REQUEST_BEACON_HEAD_PATH); err != nil {
            errorChannel <- errors.New("Error retrieving beacon head: " + err.Error())
        } else if err := json.Unmarshal(responseBody, &response); err != nil {
            errorChannel <- errors.New("Error unpacking beacon head: " + err.Error())
        } else {
            responseChannel <- response
        }
    })()

    // Request slots per epoch
    go (func() {
        if responseBody, err := c.getRequest(REQUEST_SLOTS_PER_EPOCH_PATH); err != nil {
            errorChannel <- errors.New("Error retrieving slots per epoch: " + err.Error())
        } else if slotsPerEpoch, err := strconv.Atoi(string(responseBody)); err != nil {
            errorChannel <- errors.New("Error unpacking slots per epoch: " + err.Error())
        } else {
            slotsPerEpochChannel <- uint64(slotsPerEpoch)
        }
    })()

    // Receive data
    var response BeaconHeadResponse
    var slotsPerEpoch uint64
    for received := 0; received < 2; {
        select {
            case response = <-responseChannel:
                received++
            case slotsPerEpoch = <-slotsPerEpochChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Update response & return
    response.Epoch = response.Slot / slotsPerEpoch
    response.FinalizedEpoch = response.FinalizedSlot / slotsPerEpoch
    response.JustifiedEpoch = response.JustifiedSlot / slotsPerEpoch
    return &response, nil

}


/**
 * Get a validator's status
 */
func (c *Client) GetValidatorStatus(pubkey string) (*ValidatorResponse, error) {

    // Request
    responseBody, err := c.postRequest(REQUEST_VALIDATORS_PATH, ValidatorsRequest{Pubkeys: []string{pubkey}})
    if err != nil {
        return nil, errors.New("Error retrieving validator status: " + err.Error())
    }

    // Unmarshal response
    var response []ValidatorResponse
    if err := json.Unmarshal(responseBody, &response); err != nil {
        return nil, errors.New("Error unpacking validator status: " + err.Error())
    }

    // Update response & return
    validator := response[0]
    validator.Exists = validator.Validator.ActivationEpoch != 0 // Set to default value of 0 only if validator is null in JSON response
    return &validator, nil

}


/**
 * Make GET request to beacon server
 */
func (c *Client) getRequest(requestPath string) ([]byte, error) {

    // Send request
    response, err := http.Get(c.providerUrl + requestPath)
    if err != nil { return nil, err }
    defer response.Body.Close()

    // Get response
    body, err := ioutil.ReadAll(response.Body)
    if err != nil { return nil, err }

    // Return
    return body, nil

}


/**
 * Make POST request to beacon server
 */
func (c *Client) postRequest(requestPath string, requestBody interface{}) ([]byte, error) {

    // Get request body
    requestBodyBytes, err := json.Marshal(requestBody)
    if err != nil { return nil, err }
    requestBodyReader := bytes.NewReader(requestBodyBytes)

    // Send request
    response, err := http.Post(c.providerUrl + requestPath, REQUEST_CONTENT_TYPE, requestBodyReader)
    if err != nil { return nil, err }
    defer response.Body.Close()

    // Get response
    body, err := ioutil.ReadAll(response.Body)
    if err != nil { return nil, err }

    // Return
    return body, nil

}

