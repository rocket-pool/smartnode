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
    vc pb.BeaconNodeValidatorClient
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
    vc := pb.NewBeaconNodeValidatorClient(conn)

    // Return client
    return &Client{
        conn: conn,
        bc: bc,
        nc: nc,
        vc: vc,
    }, nil

}


// Close the client connection
func (c *Client) Close() {
    c.conn.Close()
}


// Get the beacon client type
func (c *Client) GetClientType() (beacon.BeaconClientType) {
    return beacon.SplitProcess;
}


// Get the node's sync status
func (c *Client) GetSyncStatus() (beacon.SyncStatus, error) {

    // Get sync status
    syncStatus, err := c.nc.GetSyncStatus(context.Background(), &pbtypes.Empty{})
    if err != nil {
        return beacon.SyncStatus{}, fmt.Errorf("Could not get node sync status: %w", err)
    }

    // Return
    return beacon.SyncStatus{
        Syncing: syncStatus.Syncing,
    }, nil

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
    genesisEpoch, err := getConfigUint(cfg, "GenesisEpoch")
    if err != nil { return beacon.Eth2Config{}, err }
    secondsPerSlot, err := getConfigUint(cfg, "SecondsPerSlot")
    if err != nil { return beacon.Eth2Config{}, err }
    slotsPerEpoch, err := getConfigUint(cfg, "SlotsPerEpoch")
    if err != nil { return beacon.Eth2Config{}, err }

    // Return response
    return beacon.Eth2Config{
        GenesisForkVersion: genesisForkVersion,
        GenesisValidatorsRoot: genesis.GenesisValidatorsRoot,
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

    // Get validator status
    validators, err := c.bc.ListValidators(context.Background(), validatorsRequest)
    if err != nil {
        return beacon.ValidatorStatus{}, fmt.Errorf("Could not get validator %s status: %w", pubkey.Hex(), err)
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


// Get multiple validators' statuses
func (c *Client) GetValidatorStatuses(pubkeys []types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (map[types.ValidatorPubkey]beacon.ValidatorStatus, error) {

    // Return if no pubkeys defined
    if len(pubkeys) == 0 {
        return map[types.ValidatorPubkey]beacon.ValidatorStatus{}, nil
    }

    // Build validator statuses request
    validatorsRequest := &pb.ListValidatorsRequest{
        PublicKeys: make([][]byte, len(pubkeys)),
    }
    for ki, pubkey := range pubkeys {
        validatorsRequest.PublicKeys[ki] = pubkey.Bytes()
    }
    if opts != nil {
        validatorsRequest.QueryFilter = &pb.ListValidatorsRequest_Epoch{Epoch: opts.Epoch}
    }

    // Load validator statuses in pages
    validators := make([]*pb.Validators_ValidatorContainer, 0, len(pubkeys))
    for {

        // Get & add validators
        response, err := c.bc.ListValidators(context.Background(), validatorsRequest)
        if err != nil {
            return map[types.ValidatorPubkey]beacon.ValidatorStatus{}, fmt.Errorf("Could not get validator statuses: %w", err)
        }
        validators = append(validators, response.ValidatorList...)

        // Update request page token; break on last page
        if response.NextPageToken == "" { break }
        validatorsRequest.PageToken = response.NextPageToken

    }

    // Return if no validators found
    if len(validators) == 0 {
        return map[types.ValidatorPubkey]beacon.ValidatorStatus{}, nil
    }

    // Build validator balances request
    balancesRequest := &pb.ListValidatorBalancesRequest{
        PublicKeys: make([][]byte, len(validators)),
    }
    for vi, validator := range validators {
        balancesRequest.PublicKeys[vi] = validator.Validator.PublicKey
    }
    if opts != nil {
        balancesRequest.QueryFilter = &pb.ListValidatorBalancesRequest_Epoch{Epoch: opts.Epoch}
    }

    // Load validator balances in pages
    balances := make([]*pb.ValidatorBalances_Balance, 0, len(pubkeys))
    for {

        // Get & add balances
        response, err := c.bc.ListValidatorBalances(context.Background(), balancesRequest)
        if err != nil {
            return map[types.ValidatorPubkey]beacon.ValidatorStatus{}, fmt.Errorf("Could not get validator balances: %w", err)
        }
        balances = append(balances, response.Balances...)

        // Update request page token; break on last page
        if response.NextPageToken == "" { break }
        balancesRequest.PageToken = response.NextPageToken

    }

    // Check validator balances count
    if len(validators) != len(balances) {
        return map[types.ValidatorPubkey]beacon.ValidatorStatus{}, fmt.Errorf("Validator status and balance result counts do not match")
    }

    // Build & return status map
    statuses := make(map[types.ValidatorPubkey]beacon.ValidatorStatus)
    for vi := 0; vi < len(validators); vi++ {

        // Get validator status, balance & pubkey
        validator := validators[vi].Validator
        validatorBalance := balances[vi].Balance
        pubkey := types.BytesToValidatorPubkey(validator.PublicKey)

        // Add status
        statuses[pubkey] = beacon.ValidatorStatus{
            Pubkey: pubkey,
            WithdrawalCredentials: common.BytesToHash(validator.WithdrawalCredentials),
            Balance: validatorBalance,
            EffectiveBalance: validator.EffectiveBalance,
            Slashed: validator.Slashed,
            ActivationEligibilityEpoch: validator.ActivationEligibilityEpoch,
            ActivationEpoch: validator.ActivationEpoch,
            ExitEpoch: validator.ExitEpoch,
            WithdrawableEpoch: validator.WithdrawableEpoch,
            Exists: true,
        }

    }
    return statuses, nil

}


// Get a validator's index
func (c *Client) GetValidatorIndex(pubkey types.ValidatorPubkey) (uint64, error) {
    validatorIndex, err := c.vc.ValidatorIndex(context.Background(), &pb.ValidatorIndexRequest{PublicKey: pubkey.Bytes()})
    if err != nil {
        return 0, fmt.Errorf("Could not get validator %s index: %w", pubkey.Hex(), err)
    }
    return validatorIndex.Index, nil
}


// Get domain data for a domain type at a given epoch
func (c *Client) GetDomainData(domainType []byte, epoch uint64) ([]byte, error) {
    domainData, err := c.vc.DomainData(context.Background(), &pb.DomainRequest{Domain: domainType, Epoch: epoch})
    if err != nil {
        return []byte{}, fmt.Errorf("Could not get domain data for epoch %d: %w", epoch, err)
    }
    return domainData.SignatureDomain, nil
}


// Perform a voluntary exit on a validator
func (c *Client) ExitValidator(validatorIndex, epoch uint64, signature types.ValidatorSignature) error {

    // Build signed exit message
    signedExitMessage := &pb.SignedVoluntaryExit{
        Exit: &pb.VoluntaryExit{
            Epoch: epoch,
            ValidatorIndex: validatorIndex,
        },
        Signature: signature.Bytes(),
    }

    // Propose exit
    _, err := c.vc.ProposeExit(context.Background(), signedExitMessage)
    if err != nil {
        return fmt.Errorf("Could not propose exit for validator at index %d: %w", validatorIndex, err)
    }

    // Return
    return nil

}

