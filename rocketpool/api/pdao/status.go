package pdao

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	node131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/node"
	"github.com/urfave/cli"
	"github.com/wealdtech/go-ens/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
	updateCheck "github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func getStatus(c *cli.Context) (*api.PDAOStatusResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	reg, err := services.GetRocketSignerRegistry(c)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("Error getting the signer registry on network [%v].", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}

	// Check if Saturn is already deployed
	saturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOStatusResponse{}
	response.NodeRPLLocked = big.NewInt(0)

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address
	response.AccountAddressFormatted = formatResolvedAddress(c, response.AccountAddress)

	// Sync
	var wg errgroup.Group
	var blockNumber uint64

	// Get the node onchain voting delegate
	wg.Go(func() error {
		var err error
		response.OnchainVotingDelegate, err = network.GetCurrentVotingDelegate(rp, nodeAccount.Address, nil)
		if err == nil {
			response.OnchainVotingDelegateFormatted = formatResolvedAddress(c, response.OnchainVotingDelegate)
		}
		return err
	})

	// Get latest block number
	wg.Go(func() error {
		_blockNumber, err := ec.BlockNumber(context.Background())
		if err != nil {
			return fmt.Errorf("Error getting block number: %w", err)
		}
		blockNumber = _blockNumber
		return nil
	})

	// Check if node is opted into pdao proposal checking duty
	wg.Go(func() error {
		var err error
		response.VerifyEnabled = cfg.Smartnode.VerifyProposals.Value.(bool)
		if err != nil {
			return fmt.Errorf("Error loading configuration: %w", err)
		}
		return nil
	})

	// Check whether RPL locking is allowed for the node
	wg.Go(func() error {
		var err error
		response.IsRPLLockingAllowed, err = node.GetRPLLockedAllowed(rp, nodeAccount.Address, nil)
		return err
	})

	if saturnDeployed {
		// Get the node's locked RPL
		wg.Go(func() error {
			var err error
			response.NodeRPLLocked, err = node.GetNodeLockedRPL(rp, nodeAccount.Address, nil)
			return err
		})
	} else {
		// Get the node's locked RPL
		wg.Go(func() error {
			var err error
			response.NodeRPLLocked, err = node131.GetNodeRPLLocked(rp, nodeAccount.Address, nil)
			return err
		})
	}

	// Check if Node is registered
	wg.Go(func() error {
		var err error
		response.IsNodeRegistered, err = node.GetNodeExists(rp, nodeAccount.Address, nil)
		return err
	})

	// Get active and past votes from Snapshot, but treat errors as non-Fatal
	if reg != nil {
		wg.Go(func() error {
			var err error
			r := &response.SnapshotResponse
			if cfg.Smartnode.GetRocketSignerRegistryAddress() != "" {
				response.SignallingAddress, err = reg.NodeToSigner(&bind.CallOpts{}, nodeAccount.Address)
				if err != nil {
					r.Error = err.Error()
					return nil
				}
				blankAddress := common.Address{}
				if response.SignallingAddress != blankAddress {
					response.SignallingAddressFormatted = formatResolvedAddress(c, response.SignallingAddress)
				}
				votedProposals, err := GetSnapshotVotedProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), nodeAccount.Address, response.SignallingAddress)
				if err != nil {
					r.Error = err.Error()
					return nil
				}
				r.ProposalVotes = votedProposals.Data.Votes
			}
			snapshotResponse, err := GetSnapshotProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), "active")
			if err != nil {
				r.Error = err.Error()
				return nil
			}
			r.ActiveSnapshotProposals = snapshotResponse.Data.Proposals
			return nil
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Cast to uint32
	response.BlockNumber = uint32(blockNumber)

	// Get the proposal artifacts
	propMgr, err := proposals.NewProposalManager(nil, cfg, rp, bc)
	if err != nil {
		return nil, err
	}

	// Get the delegated voting power
	totalDelegatedVP, _, _, err := propMgr.GetArtifactsForVoting(response.BlockNumber, nodeAccount.Address)
	if err != nil {
		return nil, err
	}
	response.TotalDelegatedVp = totalDelegatedVP

	// Get the local tree
	votingTree, err := propMgr.GetNetworkTree(response.BlockNumber, nil)
	if err != nil {
		return nil, err
	}
	response.SumVotingPower = votingTree.Nodes[0].Sum

	// Get voting power
	response.VotingPower, err = network.GetVotingPower(rp, nodeAccount.Address, response.BlockNumber, nil)
	if err != nil {
		return nil, err
	}

	// Update & return response
	return &response, nil
}

func formatResolvedAddress(c *cli.Context, address common.Address) string {
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return address.Hex()
	}

	name, err := ens.ReverseResolve(rp.Client, address)
	if err != nil {
		return address.Hex()
	}
	return fmt.Sprintf("%s (%s)", name, address.Hex())
}

func GetSnapshotVotedProposals(apiDomain string, space string, nodeAddress common.Address, delegate common.Address) (*api.SnapshotVotedProposals, error) {
	client := getHttpClientWithTimeout()
	query := fmt.Sprintf(`query Votes{
		votes(
		  where: {
			space: "%s",
			voter_in: ["%s", "%s"],
			created_gte: 1727694646
		  },
		  orderBy: "created",
		  orderDirection: desc
		) {
		  choice
		  voter
		  proposal {id, state}
		}
	  }`, space, nodeAddress, delegate)
	url := fmt.Sprintf("https://%s/graphql?operationName=Votes&query=%s", apiDomain, url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var votedProposals api.SnapshotVotedProposals
	if err := json.Unmarshal(body, &votedProposals); err != nil {
		return nil, fmt.Errorf("could not decode snapshot response: %w", err)

	}

	return &votedProposals, nil
}

func GetSnapshotProposals(apiDomain string, space string, state string) (*api.SnapshotResponse, error) {
	client := getHttpClientWithTimeout()
	stateFilter := ""
	if state != "" {
		stateFilter = fmt.Sprintf(`, state: "%s"`, state)
	}
	query := fmt.Sprintf(`query Proposals {
	proposals(where: {space: "%s"%s, start_gte: 1727694646}, orderBy: "created", orderDirection: desc) {
	    id
	    title
	    choices
	    start
	    end
	    snapshot
	    state
	    author
		scores
		scores_total
		scores_updated
		quorum
		link
	  }
    }`, space, stateFilter)

	url := fmt.Sprintf("https://%s/graphql?operationName=Proposals&query=%s", apiDomain, url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var snapshotResponse api.SnapshotResponse
	if err := json.Unmarshal(body, &snapshotResponse); err != nil {
		return nil, fmt.Errorf("Could not decode snapshot response: %w", err)

	}

	return &snapshotResponse, nil
}

func getHttpClientWithTimeout() *http.Client {
	return &http.Client{
		Timeout: time.Second * 5,
	}
}

func GetSnapshotVotingPower(apiDomain string, space string, nodeAddress common.Address) (*api.SnapshotVotingPower, error) {
	client := getHttpClientWithTimeout()
	query := fmt.Sprintf(`query Vp{
		vp(
			space: "%s",
			voter: "%s",
		) {
			vp
		}
	}
	`, space, nodeAddress)
	url := fmt.Sprintf("https://%s/graphql?operationName=Vp&query=%s", apiDomain, url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var votingPower api.SnapshotVotingPower
	if err := json.Unmarshal(body, &votingPower); err != nil {
		return nil, fmt.Errorf("could not decode snapshot response: %w", err)

	}

	return &votingPower, nil
}
