package rocketpool

import (
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getVoteDirectionString(direction types.VoteDirection) string {
	switch direction {
	case types.VoteDirection_Abstain:
		return "abstain"
	case types.VoteDirection_For:
		return "for"
	case types.VoteDirection_Against:
		return "against"
	case types.VoteDirection_AgainstWithVeto:
		return "veto"
	}
	return ""
}

// Get protocol DAO proposals
func (c *Client) PDAOProposals() (api.PDAOProposalsResponse, error) {
	return c.callAPI[api.PDAOProposalsResponse]("GET", "/api/pdao/proposals", nil, "Could not get protocol DAO proposals")
}

// Get protocol DAO proposal details
func (c *Client) PDAOProposalDetails(proposalID uint64) (api.PDAOProposalResponse, error) {
	return c.callAPI[api.PDAOProposalResponse]("GET", "/api/pdao/proposal-details", url.Values{"id": {strconv.FormatUint(proposalID, 10)}}, "Could not get protocol DAO proposal")
}

// Check whether the node can vote on a proposal
func (c *Client) PDAOCanVoteProposal(proposalID uint64, voteDirection types.VoteDirection) (api.CanVoteOnPDAOProposalResponse, error) {
	return c.callAPI[api.CanVoteOnPDAOProposalResponse]("GET", "/api/pdao/can-vote-proposal", url.Values{
		"id":            {strconv.FormatUint(proposalID, 10)},
		"voteDirection": {getVoteDirectionString(voteDirection)},
	}, "Could not get protocol DAO can-vote-proposal")
}

// Vote on a proposal
func (c *Client) PDAOVoteProposal(proposalID uint64, voteDirection types.VoteDirection) (api.VoteOnPDAOProposalResponse, error) {
	return c.callAPI[api.VoteOnPDAOProposalResponse]("POST", "/api/pdao/vote-proposal", url.Values{
		"id":            {strconv.FormatUint(proposalID, 10)},
		"voteDirection": {getVoteDirectionString(voteDirection)},
	}, "Could not get protocol DAO vote-proposal")
}

// Check whether the node can override the delegate's vote on a proposal
func (c *Client) PDAOCanOverrideVote(proposalID uint64, voteDirection types.VoteDirection) (api.CanVoteOnPDAOProposalResponse, error) {
	return c.callAPI[api.CanVoteOnPDAOProposalResponse]("GET", "/api/pdao/can-override-vote", url.Values{
		"id":            {strconv.FormatUint(proposalID, 10)},
		"voteDirection": {getVoteDirectionString(voteDirection)},
	}, "Could not get protocol DAO can-override-vote")
}

// Override the delegate's vote on a proposal
func (c *Client) PDAOOverrideVote(proposalID uint64, voteDirection types.VoteDirection) (api.VoteOnPDAOProposalResponse, error) {
	return c.callAPI[api.VoteOnPDAOProposalResponse]("POST", "/api/pdao/override-vote", url.Values{
		"id":            {strconv.FormatUint(proposalID, 10)},
		"voteDirection": {getVoteDirectionString(voteDirection)},
	}, "Could not get protocol DAO override-vote")
}

// Check whether the node can execute a proposal
func (c *Client) PDAOCanExecuteProposal(proposalID uint64) (api.CanExecutePDAOProposalResponse, error) {
	return c.callAPI[api.CanExecutePDAOProposalResponse]("GET", "/api/pdao/can-execute-proposal", url.Values{"id": {strconv.FormatUint(proposalID, 10)}}, "Could not get protocol DAO can-execute-proposal")
}

// Execute a proposal
func (c *Client) PDAOExecuteProposal(proposalID uint64) (api.ExecutePDAOProposalResponse, error) {
	return c.callAPI[api.ExecutePDAOProposalResponse]("POST", "/api/pdao/execute-proposal", url.Values{"id": {strconv.FormatUint(proposalID, 10)}}, "Could not get protocol DAO execute-proposal")
}

// Get protocol DAO settings
func (c *Client) PDAOGetSettings() (api.GetPDAOSettingsResponse, error) {
	return c.callAPI[api.GetPDAOSettingsResponse]("GET", "/api/pdao/get-settings", nil, "Could not get protocol DAO get-settings")
}

// Check whether the node can propose updating a PDAO setting
func (c *Client) PDAOCanProposeSetting(contract string, setting string, value string) (api.CanProposePDAOSettingResponse, error) {
	return c.callAPI[api.CanProposePDAOSettingResponse]("GET", "/api/pdao/can-propose-setting", url.Values{
		"contract": {contract},
		"setting":  {setting},
		"value":    {value},
	}, "Could not get protocol DAO can-propose-setting")
}

// Check whether the node can propose updating multiple PDAO settings
func (c *Client) PDAOCanProposeSettingMulti(settings []api.PDAOBatchSetting, customMessage string) (api.CanProposePDAOSettingMultiResponse, error) {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return api.CanProposePDAOSettingMultiResponse{}, fmt.Errorf("Could not encode multi-setting proposal: %w", err)
	}
	return c.callAPI[api.CanProposePDAOSettingMultiResponse]("POST", "/api/pdao/can-propose-setting-multi", url.Values{
		"settings":      {string(settingsJSON)},
		"customMessage": {customMessage},
	}, "Could not get protocol DAO can-propose-setting-multi")
}

// Propose updating multiple PDAO settings
func (c *Client) PDAOProposeSettingMulti(settings []api.PDAOBatchSetting, customMessage string, blockNumber uint32) (api.ProposePDAOSettingMultiResponse, error) {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return api.ProposePDAOSettingMultiResponse{}, fmt.Errorf("Could not encode multi-setting proposal: %w", err)
	}
	return c.callAPI[api.ProposePDAOSettingMultiResponse]("POST", "/api/pdao/propose-setting-multi", url.Values{
		"settings":      {string(settingsJSON)},
		"customMessage": {customMessage},
		"blockNumber":   {strconv.FormatUint(uint64(blockNumber), 10)},
	}, "Could not get protocol DAO propose-setting-multi")
}

// Propose updating a PDAO setting
func (c *Client) PDAOProposeSetting(contract string, setting string, value string, blockNumber uint32) (api.ProposePDAOSettingResponse, error) {
	return c.callAPI[api.ProposePDAOSettingResponse]("POST", "/api/pdao/propose-setting", url.Values{
		"contract":    {contract},
		"setting":     {setting},
		"value":       {value},
		"blockNumber": {strconv.FormatUint(uint64(blockNumber), 10)},
	}, "Could not get protocol DAO propose-setting")
}

// Get the allocation percentages of RPL rewards
func (c *Client) PDAOGetRewardsPercentages() (api.PDAOGetRewardsPercentagesResponse, error) {
	return c.callAPI[api.PDAOGetRewardsPercentagesResponse]("GET", "/api/pdao/get-rewards-percentages", nil, "Could not get protocol DAO get-rewards-percentages")
}

// Check whether the node can propose new RPL rewards allocation percentages
func (c *Client) PDAOCanProposeRewardsPercentages(node *big.Int, odao *big.Int, pdao *big.Int) (api.PDAOCanProposeRewardsPercentagesResponse, error) {
	return c.callAPI[api.PDAOCanProposeRewardsPercentagesResponse]("GET", "/api/pdao/can-propose-rewards-percentages", url.Values{
		"node": {node.String()},
		"odao": {odao.String()},
		"pdao": {pdao.String()},
	}, "Could not get protocol DAO can-propose-rewards-percentages")
}

// Propose new RPL rewards allocation percentages
func (c *Client) PDAOProposeRewardsPercentages(node *big.Int, odao *big.Int, pdao *big.Int, blockNumber uint32) (api.ProposePDAOSettingResponse, error) {
	return c.callAPI[api.ProposePDAOSettingResponse]("POST", "/api/pdao/propose-rewards-percentages", url.Values{
		"node":        {node.String()},
		"odao":        {odao.String()},
		"pdao":        {pdao.String()},
		"blockNumber": {strconv.FormatUint(uint64(blockNumber), 10)},
	}, "Could not get protocol DAO propose-rewards-percentages")
}

// Check whether the node can propose a one-time spend of the Protocol DAO's treasury
func (c *Client) PDAOCanProposeOneTimeSpend(invoiceID string, recipient common.Address, amount *big.Int, customMessage string) (api.PDAOCanProposeOneTimeSpendResponse, error) {
	return c.callAPI[api.PDAOCanProposeOneTimeSpendResponse]("GET", "/api/pdao/can-propose-one-time-spend", url.Values{
		"invoiceId":     {invoiceID},
		"recipient":     {recipient.Hex()},
		"amount":        {amount.String()},
		"customMessage": {customMessage},
	}, "Could not get protocol DAO can-propose-one-time-spend")
}

// Propose a one-time spend of the Protocol DAO's treasury
func (c *Client) PDAOProposeOneTimeSpend(invoiceID string, recipient common.Address, amount *big.Int, blockNumber uint32, customMessage string) (api.PDAOProposeOneTimeSpendResponse, error) {
	return c.callAPI[api.PDAOProposeOneTimeSpendResponse]("POST", "/api/pdao/propose-one-time-spend", url.Values{
		"invoiceId":     {invoiceID},
		"recipient":     {recipient.Hex()},
		"amount":        {amount.String()},
		"blockNumber":   {strconv.FormatUint(uint64(blockNumber), 10)},
		"customMessage": {customMessage},
	}, "Could not get protocol DAO propose-one-time-spend")
}

// Check whether the node can propose a recurring spend of the Protocol DAO's treasury
func (c *Client) PDAOCanProposeRecurringSpend(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64, customMessage string) (api.PDAOCanProposeRecurringSpendResponse, error) {
	return c.callAPI[api.PDAOCanProposeRecurringSpendResponse]("GET", "/api/pdao/can-propose-recurring-spend", url.Values{
		"contractName":    {contractName},
		"recipient":       {recipient.Hex()},
		"amountPerPeriod": {amountPerPeriod.String()},
		"periodLength":    {periodLength.String()},
		"startTime":       {strconv.FormatInt(startTime.Unix(), 10)},
		"numberOfPeriods": {strconv.FormatUint(numberOfPeriods, 10)},
		"customMessage":   {customMessage},
	}, "Could not get protocol DAO can-propose-recurring-spend")
}

// Propose a recurring spend of the Protocol DAO's treasury
func (c *Client) PDAOProposeRecurringSpend(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64, blockNumber uint32, customMessage string) (api.PDAOProposeRecurringSpendResponse, error) {
	return c.callAPI[api.PDAOProposeRecurringSpendResponse]("POST", "/api/pdao/propose-recurring-spend", url.Values{
		"contractName":    {contractName},
		"recipient":       {recipient.Hex()},
		"amountPerPeriod": {amountPerPeriod.String()},
		"periodLength":    {periodLength.String()},
		"startTime":       {strconv.FormatInt(startTime.Unix(), 10)},
		"numberOfPeriods": {strconv.FormatUint(numberOfPeriods, 10)},
		"blockNumber":     {strconv.FormatUint(uint64(blockNumber), 10)},
		"customMessage":   {customMessage},
	}, "Could not get protocol DAO propose-recurring-spend")
}

// Check whether the node can propose an update to an existing recurring spend plan
func (c *Client) PDAOCanProposeRecurringSpendUpdate(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, customMessage string) (api.PDAOCanProposeRecurringSpendUpdateResponse, error) {
	return c.callAPI[api.PDAOCanProposeRecurringSpendUpdateResponse]("GET", "/api/pdao/can-propose-recurring-spend-update", url.Values{
		"contractName":    {contractName},
		"recipient":       {recipient.Hex()},
		"amountPerPeriod": {amountPerPeriod.String()},
		"periodLength":    {periodLength.String()},
		"numberOfPeriods": {strconv.FormatUint(numberOfPeriods, 10)},
		"customMessage":   {customMessage},
	}, "Could not get protocol DAO can-propose-recurring-spend-update")
}

// Propose an update to an existing recurring spend plan
func (c *Client) PDAOProposeRecurringSpendUpdate(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, blockNumber uint32, customMessage string) (api.PDAOProposeRecurringSpendUpdateResponse, error) {
	return c.callAPI[api.PDAOProposeRecurringSpendUpdateResponse]("POST", "/api/pdao/propose-recurring-spend-update", url.Values{
		"contractName":    {contractName},
		"recipient":       {recipient.Hex()},
		"amountPerPeriod": {amountPerPeriod.String()},
		"periodLength":    {periodLength.String()},
		"numberOfPeriods": {strconv.FormatUint(numberOfPeriods, 10)},
		"blockNumber":     {strconv.FormatUint(uint64(blockNumber), 10)},
		"customMessage":   {customMessage},
	}, "Could not get protocol DAO propose-recurring-spend-update")
}

// Check whether the node can invite someone to the security council
func (c *Client) PDAOCanProposeInviteToSecurityCouncil(id string, address common.Address) (api.PDAOCanProposeInviteToSecurityCouncilResponse, error) {
	return c.callAPI[api.PDAOCanProposeInviteToSecurityCouncilResponse]("GET", "/api/pdao/can-propose-invite-to-security-council", url.Values{
		"id":      {id},
		"address": {address.Hex()},
	}, "Could not get protocol DAO can-propose-invite-to-security-council")
}

// Propose inviting someone to the security council
func (c *Client) PDAOProposeInviteToSecurityCouncil(id string, address common.Address, blockNumber uint32) (api.PDAOProposeInviteToSecurityCouncilResponse, error) {
	return c.callAPI[api.PDAOProposeInviteToSecurityCouncilResponse]("POST", "/api/pdao/propose-invite-to-security-council", url.Values{
		"id":          {id},
		"address":     {address.Hex()},
		"blockNumber": {strconv.FormatUint(uint64(blockNumber), 10)},
	}, "Could not get protocol DAO propose-invite-to-security-council")
}

// Check whether the node can kick someone from the security council
func (c *Client) PDAOCanProposeKickFromSecurityCouncil(address common.Address) (api.PDAOCanProposeKickFromSecurityCouncilResponse, error) {
	return c.callAPI[api.PDAOCanProposeKickFromSecurityCouncilResponse]("GET", "/api/pdao/can-propose-kick-from-security-council", url.Values{"address": {address.Hex()}}, "Could not get protocol DAO can-propose-kick-from-security-council")
}

// Propose kicking someone from the security council
func (c *Client) PDAOProposeKickFromSecurityCouncil(address common.Address, blockNumber uint32) (api.PDAOProposeKickFromSecurityCouncilResponse, error) {
	return c.callAPI[api.PDAOProposeKickFromSecurityCouncilResponse]("POST", "/api/pdao/propose-kick-from-security-council", url.Values{
		"address":     {address.Hex()},
		"blockNumber": {strconv.FormatUint(uint64(blockNumber), 10)},
	}, "Could not get protocol DAO propose-kick-from-security-council")
}

// Check whether the node can kick multiple members from the security council
func (c *Client) PDAOCanProposeKickMultiFromSecurityCouncil(addresses []common.Address) (api.PDAOCanProposeKickMultiFromSecurityCouncilResponse, error) {
	addressStrings := make([]string, len(addresses))
	for i, address := range addresses {
		addressStrings[i] = address.Hex()
	}
	return c.callAPI[api.PDAOCanProposeKickMultiFromSecurityCouncilResponse]("GET", "/api/pdao/can-propose-kick-multi-from-security-council", url.Values{"addresses": {strings.Join(addressStrings, ",")}}, "Could not get protocol DAO can-propose-kick-multi-from-security-council")
}

// Propose kicking multiple members from the security council
func (c *Client) PDAOProposeKickMultiFromSecurityCouncil(addresses []common.Address, blockNumber uint32) (api.PDAOProposeKickMultiFromSecurityCouncilResponse, error) {
	addressStrings := make([]string, len(addresses))
	for i, address := range addresses {
		addressStrings[i] = address.Hex()
	}
	return c.callAPI[api.PDAOProposeKickMultiFromSecurityCouncilResponse]("POST", "/api/pdao/propose-kick-multi-from-security-council", url.Values{
		"addresses":   {strings.Join(addressStrings, ",")},
		"blockNumber": {strconv.FormatUint(uint64(blockNumber), 10)},
	}, "Could not get protocol DAO propose-kick-multi-from-security-council")
}

// Check whether the node can propose replacing someone on the security council
func (c *Client) PDAOCanProposeReplaceMemberOfSecurityCouncil(existingAddress common.Address, newID string, newAddress common.Address) (api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse, error) {
	return c.callAPI[api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse]("GET", "/api/pdao/can-propose-replace-member-of-security-council", url.Values{
		"existingAddress": {existingAddress.Hex()},
		"newId":           {newID},
		"newAddress":      {newAddress.Hex()},
	}, "Could not get protocol DAO can-propose-replace-member-of-security-council")
}

// Propose replacing someone on the security council
func (c *Client) PDAOProposeReplaceMemberOfSecurityCouncil(existingAddress common.Address, newID string, newAddress common.Address, blockNumber uint32) (api.PDAOProposeReplaceMemberOfSecurityCouncilResponse, error) {
	return c.callAPI[api.PDAOProposeReplaceMemberOfSecurityCouncilResponse]("POST", "/api/pdao/propose-replace-member-of-security-council", url.Values{
		"existingAddress": {existingAddress.Hex()},
		"newId":           {newID},
		"newAddress":      {newAddress.Hex()},
		"blockNumber":     {strconv.FormatUint(uint64(blockNumber), 10)},
	}, "Could not get protocol DAO propose-replace-member-of-security-council")
}

// Get the list of proposals with claimable / rewardable bonds
func (c *Client) PDAOGetClaimableBonds() (api.PDAOGetClaimableBondsResponse, error) {
	return c.callAPI[api.PDAOGetClaimableBondsResponse]("GET", "/api/pdao/get-claimable-bonds", nil, "Could not get protocol DAO get-claimable-bonds")
}

// Check whether the node can claim / unlock bonds from a proposal
func (c *Client) PDAOCanClaimBonds(proposalID uint64, indices []uint64) (api.PDAOCanClaimBondsResponse, error) {
	indicesStrings := make([]string, len(indices))
	for i, index := range indices {
		indicesStrings[i] = strconv.FormatUint(index, 10)
	}
	return c.callAPI[api.PDAOCanClaimBondsResponse]("GET", "/api/pdao/can-claim-bonds", url.Values{
		"proposalId": {strconv.FormatUint(proposalID, 10)},
		"indices":    {strings.Join(indicesStrings, ",")},
	}, "Could not get protocol DAO can-claim-bonds")
}

// Claim / unlock bonds from a proposal
func (c *Client) PDAOClaimBonds(isProposer bool, proposalID uint64, indices []uint64) (api.PDAOClaimBondsResponse, error) {
	indicesStrings := make([]string, len(indices))
	for i, index := range indices {
		indicesStrings[i] = strconv.FormatUint(index, 10)
	}
	isProposerStr := "false"
	if isProposer {
		isProposerStr = "true"
	}
	return c.callAPI[api.PDAOClaimBondsResponse]("POST", "/api/pdao/claim-bonds", url.Values{
		"isProposer": {isProposerStr},
		"proposalId": {strconv.FormatUint(proposalID, 10)},
		"indices":    {strings.Join(indicesStrings, ",")},
	}, "Could not get protocol DAO claim-bonds")
}

// Check whether the node can defeat a proposal
func (c *Client) PDAOCanDefeatProposal(proposalID uint64, index uint64) (api.PDAOCanDefeatProposalResponse, error) {
	return c.callAPI[api.PDAOCanDefeatProposalResponse]("GET", "/api/pdao/can-defeat-proposal", url.Values{
		"id":    {strconv.FormatUint(proposalID, 10)},
		"index": {strconv.FormatUint(index, 10)},
	}, "Could not get protocol DAO can-defeat-proposal")
}

// Defeat a proposal
func (c *Client) PDAODefeatProposal(proposalID uint64, index uint64) (api.PDAODefeatProposalResponse, error) {
	return c.callAPI[api.PDAODefeatProposalResponse]("POST", "/api/pdao/defeat-proposal", url.Values{
		"id":    {strconv.FormatUint(proposalID, 10)},
		"index": {strconv.FormatUint(index, 10)},
	}, "Could not get protocol DAO defeat-proposal")
}

// Check whether the node can finalize a proposal
func (c *Client) PDAOCanFinalizeProposal(proposalID uint64) (api.PDAOCanFinalizeProposalResponse, error) {
	return c.callAPI[api.PDAOCanFinalizeProposalResponse]("GET", "/api/pdao/can-finalize-proposal", url.Values{"id": {strconv.FormatUint(proposalID, 10)}}, "Could not get protocol DAO can-finalize-proposal")
}

// Finalize a proposal
func (c *Client) PDAOFinalizeProposal(proposalID uint64) (api.PDAOFinalizeProposalResponse, error) {
	return c.callAPI[api.PDAOFinalizeProposalResponse]("POST", "/api/pdao/finalize-proposal", url.Values{"id": {strconv.FormatUint(proposalID, 10)}}, "Could not get protocol DAO finalize-proposal")
}

// EstimateSetVotingDelegateGas estimates the gas required to set an on-chain voting delegate
func (c *Client) EstimateSetVotingDelegateGas(address common.Address) (api.PDAOCanSetVotingDelegateResponse, error) {
	return c.callAPI[api.PDAOCanSetVotingDelegateResponse]("GET", "/api/pdao/estimate-set-voting-delegate-gas", url.Values{"address": {address.Hex()}}, "could not call estimate-set-voting-delegate-gas")
}

// SetVotingDelegate sets an on-chain voting delegate for the node
func (c *Client) SetVotingDelegate(address common.Address) (api.PDAOSetVotingDelegateResponse, error) {
	return c.callAPI[api.PDAOSetVotingDelegateResponse]("POST", "/api/pdao/set-voting-delegate", url.Values{"address": {address.Hex()}}, "could not call set-voting-delegate")
}

// GetCurrentVotingDelegate gets the node current on-chain voting delegate
func (c *Client) GetCurrentVotingDelegate() (api.PDAOCurrentVotingDelegateResponse, error) {
	return c.callAPI[api.PDAOCurrentVotingDelegateResponse]("GET", "/api/pdao/get-current-voting-delegate", nil, "could not request get-current-voting-delegate")
}

// CanSetSignallingAddress fetches gas info and if a node can set the signalling address
func (c *Client) CanSetSignallingAddress(signallingAddress common.Address, signature string) (api.PDAOCanSetSignallingAddressResponse, error) {
	return c.callAPI[api.PDAOCanSetSignallingAddressResponse]("GET", "/api/pdao/can-set-signalling-address", url.Values{
		"address":   {signallingAddress.Hex()},
		"signature": {signature},
	}, "could not call can-set-signalling-address")
}

// SetSignallingAddress sets the node's signalling address
func (c *Client) SetSignallingAddress(signallingAddress common.Address, signature string) (api.PDAOSetSignallingAddressResponse, error) {
	return c.callAPI[api.PDAOSetSignallingAddressResponse]("POST", "/api/pdao/set-signalling-address", url.Values{
		"address":   {signallingAddress.Hex()},
		"signature": {signature},
	}, "could not call set-signalling-address")
}

// CanClearSignallingAddress fetches gas info and if a node can clear a signalling address
func (c *Client) CanClearSignallingAddress() (api.PDAOCanClearSignallingAddressResponse, error) {
	return c.callAPI[api.PDAOCanClearSignallingAddressResponse]("GET", "/api/pdao/can-clear-signalling-address", nil, "could not call can-clear-signalling-address")
}

// ClearSignallingAddress clears the node's signalling address
func (c *Client) ClearSignallingAddress() (api.PDAOSetSignallingAddressResponse, error) {
	return c.callAPI[api.PDAOSetSignallingAddressResponse]("POST", "/api/pdao/clear-signalling-address", nil, "could not call clear-signalling-address")
}

// Check whether the node can propose a list of addresses that can update commission share parameters
func (c *Client) PDAOCanProposeAllowListedControllers(addressList string) (api.PDAOACanProposeAllowListedControllersResponse, error) {
	return c.callAPI[api.PDAOACanProposeAllowListedControllersResponse]("GET", "/api/pdao/can-propose-allow-listed-controllers", url.Values{"addressList": {addressList}}, "Could not get protocol DAO can-propose-allow-listed-controllers")
}

// Propose a list of addresses that can update commission share parameters
func (c *Client) PDAOProposeAllowListedControllers(addressList string, blockNumber uint32) (api.PDAOProposeAllowListedControllersResponse, error) {
	return c.callAPI[api.PDAOProposeAllowListedControllersResponse]("POST", "/api/pdao/propose-allow-listed-controllers", url.Values{
		"addressList": {addressList},
		"blockNumber": {strconv.FormatUint(uint64(blockNumber), 10)},
	}, "Could not get protocol DAO propose-allow-listed-controllers")
}

// Get PDAO Status
func (c *Client) PDAOStatus() (api.PDAOStatusResponse, error) {
	return c.callAPI[api.PDAOStatusResponse]("GET", "/api/pdao/status", nil, "could not call get pdao status")
}
