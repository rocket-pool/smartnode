package prysm

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "strconv"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
)


// Config
const (
    RequestEth2ConfigPath = "/eth/v1alpha1/beacon/config"
    RequestBeaconHeadPath = "/eth/v1alpha1/beacon/chainhead"
    RequestValidatorPath = "/eth/v1alpha1/validator"
)


// Beacon response types
type Eth2ConfigResponse struct {
    Config struct {
        GenesisForkVersion string       `json:"GenesisForkVersion"`
        BLSWithdrawalPrefixByte string  `json:"BLSWithdrawalPrefixByte"`
        DomainBeaconProposer string     `json:"DomainBeaconProposer"`
        DomainBeaconAttester string     `json:"DomainBeaconAttester"`
        DomainRandao string             `json:"DomainRandao"`
        DomainDeposit string            `json:"DomainDeposit"`
        DomainVoluntaryExit string      `json:"DomainVoluntaryExit"`
        SlotsPerEpoch string            `json:"SlotsPerEpoch"`
    } `json:"config"`
}
type BeaconHeadResponse struct {
    HeadEpoch string                    `json:"headEpoch"`
    FinalizedEpoch string               `json:"finalizedEpoch"`
    JustifiedEpoch string               `json:"justifiedEpoch"`
}
type ValidatorResponse struct {
    PublicKey string                    `json:"publicKey"`
    WithdrawalCredentials string        `json:"withdrawalCredentials"`
    EffectiveBalance string             `json:"effectiveBalance"`
    Slashed bool                        `json:"slashed"`
    ActivationEligibilityEpoch string   `json:"activationEligibilityEpoch"`
    ActivationEpoch string              `json:"activationEpoch"`
    ExitEpoch string                    `json:"exitEpoch"`
    WithdrawableEpoch string            `json:"withdrawableEpoch"`
}


// Prysm client
type Client struct {
    providerUrl string
}


// Create new prysm client
func NewClient(providerUrl string) *Client {
    return &Client{
        providerUrl: providerUrl,
    }
}


// Get the eth2 config
func (c *Client) GetEth2Config() (beacon.Eth2Config, error) {

    // Get config
    var config Eth2ConfigResponse
    if responseBody, _, err := c.getRequest(RequestEth2ConfigPath); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not get eth2 config: %w", err)
    } else if err := json.Unmarshal(responseBody, &config); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode eth2 config: %w", err)
    }

    // Create response
    response := beacon.Eth2Config{}

    // Decode data and update
    if genesisForkVersion, err := deserializeBytes(config.Config.GenesisForkVersion); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode genesis fork version: %w", err)
    } else {
        response.GenesisForkVersion = genesisForkVersion
    }
    if blsWithdrawalPrefixByteInt, err := strconv.ParseUint(config.Config.BLSWithdrawalPrefixByte, 10, 8); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode BLS withdrawal prefix byte: %w", err)
    } else {
        response.BLSWithdrawalPrefixByte = byte(blsWithdrawalPrefixByteInt)
    }
    if domainBeaconProposer, err := deserializeBytes(config.Config.DomainBeaconProposer); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode beacon proposer domain: %w", err)
    } else {
        response.DomainBeaconProposer = domainBeaconProposer
    }
    if domainBeaconAttester, err := deserializeBytes(config.Config.DomainBeaconAttester); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode beacon attester domain: %w", err)
    } else {
        response.DomainBeaconAttester = domainBeaconAttester
    }
    if domainRandao, err := deserializeBytes(config.Config.DomainRandao); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode randao domain: %w", err)
    } else {
        response.DomainRandao = domainRandao
    }
    if domainDeposit, err := deserializeBytes(config.Config.DomainDeposit); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode deposit domain: %w", err)
    } else {
        response.DomainDeposit = domainDeposit
    }
    if domainVoluntaryExit, err := deserializeBytes(config.Config.DomainVoluntaryExit); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode voluntary exit domain: %w", err)
    } else {
        response.DomainVoluntaryExit = domainVoluntaryExit
    }
    if slotsPerEpoch, err := strconv.ParseUint(config.Config.SlotsPerEpoch, 10, 64); err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not decode slots per epoch: %w", err)
    } else {
        response.SlotsPerEpoch = slotsPerEpoch
    }

    // Return
    return response, nil

}


// Get the beacon head
func (c *Client) GetBeaconHead() (beacon.BeaconHead, error) {

    // Get beacon head
    var head BeaconHeadResponse
    if responseBody, _, err := c.getRequest(RequestBeaconHeadPath); err != nil {
        return beacon.BeaconHead{}, fmt.Errorf("Could not get beacon head: %w", err)
    } else if err := json.Unmarshal(responseBody, &head); err != nil {
        return beacon.BeaconHead{}, fmt.Errorf("Could not decode beacon head: %w", err)
    }

    // Create response
    response := beacon.BeaconHead{}

    // Decode data and update
    if headEpoch, err := strconv.ParseUint(head.HeadEpoch, 10, 64); err != nil {
        return beacon.BeaconHead{}, fmt.Errorf("Could not decode head epoch: %w", err)
    } else {
        response.Epoch = headEpoch
    }
    if finalizedEpoch, err := strconv.ParseUint(head.FinalizedEpoch, 10, 64); err != nil {
        return beacon.BeaconHead{}, fmt.Errorf("Could not decode finalized epoch: %w", err)
    } else {
        response.FinalizedEpoch = finalizedEpoch
    }
    if justifiedEpoch, err := strconv.ParseUint(head.JustifiedEpoch, 10, 64); err != nil {
        return beacon.BeaconHead{}, fmt.Errorf("Could not decode justified epoch: %w", err)
    } else {
        response.JustifiedEpoch = justifiedEpoch
    }

    // Return
    return response, nil

}


// Get a validator's status
func (c *Client) GetValidatorStatus(pubkey []byte) (beacon.ValidatorStatus, error) {

    // Get request params
    params := url.Values{}
    params.Set("publicKey", base64.StdEncoding.EncodeToString(pubkey))

    // Get validator status
    var validator ValidatorResponse
    if responseBody, status, err := c.getRequest(RequestValidatorPath + "?" + params.Encode()); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not get validator status: %w", err)
    } else if status == 404 {
        return beacon.ValidatorStatus{Exists: false}, nil
    } else if err := json.Unmarshal(responseBody, &validator); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode validator status: %w", err)
    }

    // Create response
    response := beacon.ValidatorStatus{
        Slashed: validator.Slashed,
        Exists: true,
    }

    // Decode data and update
    // TODO: add validator balance
    if publicKey, err := base64.StdEncoding.DecodeString(validator.PublicKey); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode public key: %w", err)
    } else {
        response.Pubkey = publicKey
    }
    if withdrawalCredentials, err := base64.StdEncoding.DecodeString(validator.WithdrawalCredentials); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode withdrawal credentials: %w", err)
    } else {
        response.WithdrawalCredentials = withdrawalCredentials
    }
    if effectiveBalance, err := strconv.ParseUint(validator.EffectiveBalance, 10, 64); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode effective balance: %w", err)
    } else {
        response.EffectiveBalance = effectiveBalance
    }
    if activationEligibilityEpoch, err := strconv.ParseUint(validator.ActivationEligibilityEpoch, 10, 64); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode activation eligibility epoch: %w", err)
    } else {
        response.ActivationEligibilityEpoch = activationEligibilityEpoch
    }
    if activationEpoch, err := strconv.ParseUint(validator.ActivationEpoch, 10, 64); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode activation epoch: %w", err)
    } else {
        response.ActivationEpoch = activationEpoch
    }
    if exitEpoch, err := strconv.ParseUint(validator.ExitEpoch, 10, 64); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode exit epoch: %w", err)
    } else {
        response.ExitEpoch = exitEpoch
    }
    if withdrawableEpoch, err := strconv.ParseUint(validator.WithdrawableEpoch, 10, 64); err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not decode withdrawable epoch: %w", err)
    } else {
        response.WithdrawableEpoch = withdrawableEpoch
    }

    // Return
    return response, nil

}


// Make a GET request to the beacon node
func (c *Client) getRequest(requestPath string) ([]byte, int, error) {

    // Send request
    response, err := http.Get(c.providerUrl + requestPath)
    if err != nil {
        return []byte{}, 0, err
    }
    defer response.Body.Close()

    // Get response
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return []byte{}, 0, err
    }

    // Return
    return body, response.StatusCode, nil

}

