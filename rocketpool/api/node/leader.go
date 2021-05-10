package node

import (
    "math/big"
    "sort"

    "github.com/ethereum/go-ethereum/common"
    "golang.org/x/sync/errgroup"
    "github.com/urfave/cli"
    "go.uber.org/multierr"

    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/smartnode/rocketpool/api/minipool"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/types/api"
)

// Settings
const (
    NodeDetailsBatchSize = 10
    TopMinipoolCount = 1
)


func getLeader(c *cli.Context) (*api.NodeLeaderResponse, error) {
    // Get services
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeLeaderResponse{}

    nodeRanks, err := GetNodeLeader(rp, bc)
    if err != nil { return nil, err }

    response.Nodes = nodeRanks
    return &response, nil
}


func GetNodeLeader(rp *rocketpool.RocketPool, bc beacon.Client) ([]api.NodeRank, error) {

    minipools, err := minipool.GetAllMinipoolDetails(rp, bc)
    if err != nil { return nil, err }
    nodeRanks, err := GetNodeDetails(rp)
    if err != nil && nodeRanks == nil { return nil, err }

    // Get stating and has validator minipools
    // put minipools into map by address
    nodeToValMap := make(map[common.Address][]api.MinipoolDetails, len(minipools))
    for _, minipool := range minipools {
        // Add to status list
        address := minipool.Node.Address
        if _, ok := nodeToValMap[address]; !ok {
            nodeToValMap[address] = []api.MinipoolDetails{}
        }
        nodeToValMap[address] = append(nodeToValMap[address], minipool)
    }

    for i, nodeRank := range nodeRanks {
        vals, ok := nodeToValMap[nodeRank.Address]
        if !ok { continue }
        nodeRanks[i].Details = vals
        nodeRanks[i].Score = calculateNodeScore(vals)
    }

    sortFunc := func(m, n int) bool {
        if nodeRanks[m].Score == nil { return false }
        if nodeRanks[n].Score == nil { return true }
        return nodeRanks[m].Score.Cmp(nodeRanks[n].Score) > 0
    }
    sort.SliceStable(nodeRanks, sortFunc)
    k := 1
    for i := range nodeRanks {
        if (nodeRanks[i].Score != nil) {
            nodeRanks[i].Rank = k
            k++
        } else {
            nodeRanks[i].Rank = 999999999
        }
    }

    return nodeRanks, nil
}


func GetNodeDetails(rp *rocketpool.RocketPool) ([]api.NodeRank, error) {

    nodeAddresses, err := node.GetNodeAddresses(rp, nil)
    if err != nil { return nil, err }

    var merr error
    nodeRanks := make([]api.NodeRank, len(nodeAddresses))
    for bsi := 0; bsi < len(nodeAddresses); bsi += NodeDetailsBatchSize {

        // Get batch start & end index
        msi := bsi
        mei := bsi + NodeDetailsBatchSize
        if mei > len(nodeAddresses) { mei = len(nodeAddresses) }

        // Load details
        var wg errgroup.Group
        for mi := msi; mi < mei; mi++ {
            mi2 := mi
            wg.Go(func() error {
                address := nodeAddresses[mi2]
                nodeRanks[mi2] = api.NodeRank{ Address: address }
                details, err := node.GetNodeDetails(rp, address, nil)
                if err == nil {
                    nodeRanks[mi2].Registered = details.Exists
                    nodeRanks[mi2].TimezoneLocation = details.TimezoneLocation
                }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            merr = multierr.Append(merr, err)
        }
    }

    return nodeRanks, merr
}


func calculateNodeScore(vals []api.MinipoolDetails) *big.Int {
    // score formula: take the top N performing validators
    // sum up their profits or losses
    // profit is defined as: current balance - initial node deposit - user deposit
    // unless something is broken, this should be current balance - 32
    // unit is wei

    var prevMax *big.Int
    score := new(big.Int)

    // remove non-existing validators from scoring
    // use selection sort so we don't need to alloc more memory
    for j := 0; j < TopMinipoolCount && j < len(vals); j++ {
        var currMax *api.MinipoolDetails
        for k := 0; k < len(vals); k++ {
            if vals[k].Validator.Balance != nil &&
                (currMax == nil || vals[k].Validator.Balance.Cmp(currMax.Validator.Balance) > 0) &&
                (prevMax == nil || vals[k].Validator.Balance.Cmp(prevMax) < 0) {
                    currMax = &vals[k]
            }
        }

        if currMax == nil {
            break
        }

        score.Add(score, currMax.Validator.Balance)
        score.Sub(score, currMax.Node.DepositBalance)
        score.Sub(score, currMax.User.DepositBalance)

        prevMax = currMax.Validator.Balance
    }

    return score
}
