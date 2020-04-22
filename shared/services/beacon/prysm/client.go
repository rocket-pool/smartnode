package prysm

import (
    "github.com/rocket-pool/smartnode/shared/services/beacon"
)


// Beacon endpoints
const REQUEST_ETH2_CONFIG_PATH string = "/eth/v1alpha1/beacon/config"
const REQUEST_BEACON_HEAD_PATH string = "/eth/v1alpha1/beacon/chainhead"
const REQUEST_VALIDATOR_PATH string = "/eth/v1alpha1/validator"


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
func (c *Client) GetEth2Config() (*beacon.Eth2Config, error) {
    return &beacon.Eth2Config{}, nil
}


/**
 * Get the beacon head
 */
func (c *Client) GetBeaconHead() (*beacon.BeaconHead, error) {
    return &beacon.BeaconHead{}, nil
}


/**
 * Get a validator's status
 */
func (c *Client) GetValidatorStatus(pubkey string) (*beacon.ValidatorStatus, error) {
    return &beacon.ValidatorStatus{}, nil
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

