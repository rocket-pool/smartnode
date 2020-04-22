package prysm

import (
    "github.com/rocket-pool/smartnode/shared/services/beacon"
)


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

