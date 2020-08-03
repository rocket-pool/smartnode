package prysm

import (
    "context"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    pbtypes "github.com/gogo/protobuf/types"
    pb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
    "github.com/rocket-pool/rocketpool-go/types"
    "google.golang.org/grpc"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
)


// Prysm client
type Client struct {
    conn *grpc.ClientConn
    bc pb.BeaconChainClient
    nc pb.NodeClient
}


// Create new prysm client
func NewClient(providerAddress string) (*Client, error) {

    // Initialize gRPC connection
    conn, err := grpc.Dial(providerAddress, grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        return nil, fmt.Errorf("Could not connect to gRPC server: %w", err)
    }

    // Initialize clients
    bc := pb.NewBeaconChainClient(conn)
    nc := pb.NewNodeClient(conn)

    // Return client
    return &Client{
        conn: conn,
        bc: bc,
        nc: nc,
    }, nil

}


// Close the client connection
func (c *Client) Close() {
    c.conn.Close()
}


// Get the eth2 config
func (c *Client) GetEth2Config() (beacon.Eth2Config, error) {

    // Get beacon config
    config, err := c.bc.GetBeaconConfig(context.Background(), &pbtypes.Empty{})
    if err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not get beacon chain config: %w", err)
    }
    cfg := config.GetConfig()

    // Get genesis data
    genesis, err := c.nc.GetGenesis(context.Background(), &pbtypes.Empty{})
    if err != nil {
        return beacon.Eth2Config{}, fmt.Errorf("Could not get genesis data: %w", err)
    }

    // Get config settings
    genesisForkVersion, err := getConfigBytes(cfg, "GenesisForkVersion")
    if err != nil { return beacon.Eth2Config{}, err }
    domainDeposit, err := getConfigBytes(cfg, "DomainDeposit")
    if err != nil { return beacon.Eth2Config{}, err }
    domainVoluntaryExit, err := getConfigBytes(cfg, "DomainVoluntaryExit")
    if err != nil { return beacon.Eth2Config{}, err }
    genesisEpoch, err := getConfigUint(cfg, "GenesisEpoch")
    if err != nil { return beacon.Eth2Config{}, err }
    secondsPerSlot, err := getConfigUint(cfg, "SecondsPerSlot")
    if err != nil { return beacon.Eth2Config{}, err }
    slotsPerEpoch, err := getConfigUint(cfg, "SlotsPerEpoch")
    if err != nil { return beacon.Eth2Config{}, err }

    // Return response
    return beacon.Eth2Config{
        GenesisForkVersion: genesisForkVersion,
        DomainDeposit: domainDeposit,
        DomainVoluntaryExit: domainVoluntaryExit,
        GenesisEpoch: genesisEpoch,
        GenesisTime: uint64(genesis.GenesisTime.Seconds),
        SecondsPerEpoch: secondsPerSlot * slotsPerEpoch,
    }, nil

}


// Get the beacon head
func (c *Client) GetBeaconHead() (beacon.BeaconHead, error) {

    // Get chain head
    head, err := c.bc.GetChainHead(context.Background(), &pbtypes.Empty{})
    if err != nil {
        return beacon.BeaconHead{}, fmt.Errorf("Could not get beacon chain head: %w", err)
    }

    // Return response
    return beacon.BeaconHead{
        Slot: head.HeadSlot,
        FinalizedSlot: head.FinalizedSlot,
        JustifiedSlot: head.JustifiedSlot,
        PreviousJustifiedSlot: head.PreviousJustifiedSlot,
        Epoch: head.HeadEpoch,
        FinalizedEpoch: head.FinalizedEpoch,
        JustifiedEpoch: head.JustifiedEpoch,
        PreviousJustifiedEpoch: head.PreviousJustifiedEpoch,
    }, nil

}


// Get a validator's status
func (c *Client) GetValidatorStatus(pubkey types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {

    // Build validator requests
    validatorsRequest := &pb.ListValidatorsRequest{
        PublicKeys: [][]byte{pubkey.Bytes()},
    }
    balancesRequest := &pb.ListValidatorBalancesRequest{
        PublicKeys: [][]byte{pubkey.Bytes()},
    }
    if opts != nil {
        validatorsRequest.QueryFilter = &pb.ListValidatorsRequest_Epoch{Epoch: opts.Epoch}
        balancesRequest.QueryFilter = &pb.ListValidatorBalancesRequest_Epoch{Epoch: opts.Epoch}
    }

    // Get validator
    validators, err := c.bc.ListValidators(context.Background(), validatorsRequest)
    if err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not get validator %s: %w", pubkey.Hex(), err)
    }
    if len(validators.ValidatorList) == 0 {
        return beacon.ValidatorStatus{}, nil
    }
    validator := validators.ValidatorList[0].Validator

    // Get validator balance
    balances, err := c.bc.ListValidatorBalances(context.Background(), balancesRequest)
    if err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not get validator %s balance: %w", pubkey.Hex(), err)
    }
    if len(balances.Balances) == 0 {
        return beacon.ValidatorStatus{}, nil
    }
    validatorBalance := balances.Balances[0].Balance

    // Return response
    return beacon.ValidatorStatus{
        Pubkey: types.BytesToValidatorPubkey(validator.PublicKey),
        WithdrawalCredentials: common.BytesToHash(validator.WithdrawalCredentials),
        Balance: validatorBalance,
        EffectiveBalance: validator.EffectiveBalance,
        Slashed: validator.Slashed,
        ActivationEligibilityEpoch: validator.ActivationEligibilityEpoch,
        ActivationEpoch: validator.ActivationEpoch,
        ExitEpoch: validator.ExitEpoch,
        WithdrawableEpoch: validator.WithdrawableEpoch,
        Exists: true,
    }, nil

}

