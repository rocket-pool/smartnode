package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func estimateSetSnapshotDelegateGas(c *cli.Context, address common.Address) (*api.EstimateSetSnapshotDelegateGasResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
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

	// Response
	response := api.EstimateSetSnapshotDelegateGasResponse{}

	// Get the snapshot address
	addressString := cfg.Smartnode.GetSnapshotDelegationAddress()
	if addressString == "" {
		return nil, fmt.Errorf("Network [%v] does not have a snapshot delegation contract.", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}
	snapshotDelegationAddress := common.HexToAddress(addressString)

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Create the snapshot delegation contract binding
	snapshotDelegationAbi, err := abi.JSON(strings.NewReader(contracts.SnapshotDelegationABI))
	if err != nil {
		return nil, err
	}
	contract := &rocketpool.Contract{
		Contract: bind.NewBoundContract(snapshotDelegationAddress, snapshotDelegationAbi, ec, ec, ec),
		Address:  &snapshotDelegationAddress,
		ABI:      &snapshotDelegationAbi,
		Client:   ec,
	}

	// Create the ID hash
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	// Get the gas info
	gasInfo, err := contract.GetTransactionGasInfo(opts, "setDelegate", idHash, address)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil

}

func setSnapshotDelegate(c *cli.Context, address common.Address) (*api.SetSnapshotDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	s, err := services.GetSnapshotDelegation(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetSnapshotDelegateResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Create the ID hash
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	// Set the delegate
	tx, err := s.SetDelegate(opts, idHash, address)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}

func estimateClearSnapshotDelegateGas(c *cli.Context) (*api.EstimateClearSnapshotDelegateGasResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
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

	// Response
	response := api.EstimateClearSnapshotDelegateGasResponse{}

	// Get the snapshot address
	addressString := cfg.Smartnode.GetSnapshotDelegationAddress()
	if addressString == "" {
		return nil, fmt.Errorf("Network [%v] does not have a snapshot delegation contract.", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}
	snapshotDelegationAddress := common.HexToAddress(addressString)

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Create the snapshot delegation contract binding
	snapshotDelegationAbi, err := abi.JSON(strings.NewReader(contracts.SnapshotDelegationABI))
	if err != nil {
		return nil, err
	}
	contract := &rocketpool.Contract{
		Contract: bind.NewBoundContract(snapshotDelegationAddress, snapshotDelegationAbi, ec, ec, ec),
		Address:  &snapshotDelegationAddress,
		ABI:      &snapshotDelegationAbi,
		Client:   ec,
	}

	// Create the ID hash
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	// Get the gas info
	gasInfo, err := contract.GetTransactionGasInfo(opts, "clearDelegate", idHash)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil

}

func clearSnapshotDelegate(c *cli.Context) (*api.ClearSnapshotDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	s, err := services.GetSnapshotDelegation(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ClearSnapshotDelegateResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Create the ID hash
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	// Set the delegate
	tx, err := s.ClearDelegate(opts, idHash)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var votingPower api.SnapshotVotingPower
	if err := json.Unmarshal(body, &votingPower); err != nil {
		return nil, fmt.Errorf("could not decode snapshot response: %w", err)

	}

	return &votingPower, nil
}

func GetSnapshotVotedProposals(apiDomain string, space string, nodeAddress common.Address, delegate common.Address) (*api.SnapshotVotedProposals, error) {
	client := getHttpClientWithTimeout()
	query := fmt.Sprintf(`query Votes{
		votes(
		  where: {
			space: "%s",
			voter_in: ["%s", "%s"],
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
	body, err := ioutil.ReadAll(resp.Body)
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
	proposals(where: {space: "%s"%s}, orderBy: "created", orderDirection: desc) {
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var snapshotResponse api.SnapshotResponse
	if err := json.Unmarshal(body, &snapshotResponse); err != nil {
		return nil, fmt.Errorf("Could not decode snapshot response: %w", err)

	}

	return &snapshotResponse, nil
}
