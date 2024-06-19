package rocketpool

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/rocketpool-go/types"
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
	responseBytes, err := c.callAPI("pdao proposals")
	if err != nil {
		return api.PDAOProposalsResponse{}, fmt.Errorf("Could not get protocol DAO proposals: %w", err)
	}
	var response api.PDAOProposalsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposalsResponse{}, fmt.Errorf("Could not decode protocol DAO proposals response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposalsResponse{}, fmt.Errorf("Could not get protocol DAO proposals: %s", response.Error)
	}
	return response, nil
}

// Get protocol DAO proposal details
func (c *Client) PDAOProposalDetails(proposalID uint64) (api.PDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao proposal-details %d", proposalID))
	if err != nil {
		return api.PDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO proposal: %w", err)
	}
	var response api.PDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposalResponse{}, fmt.Errorf("Could not decode protocol DAO proposal response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can vote on a proposal
func (c *Client) PDAOCanVoteProposal(proposalID uint64, voteDirection types.VoteDirection) (api.CanVoteOnPDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-vote-proposal %d %s", proposalID, getVoteDirectionString(voteDirection)))
	if err != nil {
		return api.CanVoteOnPDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-vote-proposal: %w", err)
	}
	var response api.CanVoteOnPDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanVoteOnPDAOProposalResponse{}, fmt.Errorf("Could not decode protocol DAO can-vote-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.CanVoteOnPDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-vote-proposal: %s", response.Error)
	}
	return response, nil
}

// Vote on a proposal
func (c *Client) PDAOVoteProposal(proposalID uint64, voteDirection types.VoteDirection) (api.VoteOnPDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao vote-proposal %d %s", proposalID, getVoteDirectionString(voteDirection)))
	if err != nil {
		return api.VoteOnPDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO vote-proposal: %w", err)
	}
	var response api.VoteOnPDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.VoteOnPDAOProposalResponse{}, fmt.Errorf("Could not decode protocol DAO vote-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.VoteOnPDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO vote-proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can override the delegate's vote on a proposal
func (c *Client) PDAOCanOverrideVote(proposalID uint64, voteDirection types.VoteDirection) (api.CanVoteOnPDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-override-vote %d %s", proposalID, getVoteDirectionString(voteDirection)))
	if err != nil {
		return api.CanVoteOnPDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-override-vote: %w", err)
	}
	var response api.CanVoteOnPDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanVoteOnPDAOProposalResponse{}, fmt.Errorf("Could not decode protocol DAO can-override-vote response: %w", err)
	}
	if response.Error != "" {
		return api.CanVoteOnPDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-override-vote: %s", response.Error)
	}
	return response, nil
}

// Override the delegate's vote on a proposal
func (c *Client) PDAOOverrideVote(proposalID uint64, voteDirection types.VoteDirection) (api.VoteOnPDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao override-vote %d %s", proposalID, getVoteDirectionString(voteDirection)))
	if err != nil {
		return api.VoteOnPDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO override-vote: %w", err)
	}
	var response api.VoteOnPDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.VoteOnPDAOProposalResponse{}, fmt.Errorf("Could not decode protocol DAO override-vote response: %w", err)
	}
	if response.Error != "" {
		return api.VoteOnPDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO override-vote: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can execute a proposal
func (c *Client) PDAOCanExecuteProposal(proposalID uint64) (api.CanExecutePDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-execute-proposal %d", proposalID))
	if err != nil {
		return api.CanExecutePDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-execute-proposal: %w", err)
	}
	var response api.CanExecutePDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanExecutePDAOProposalResponse{}, fmt.Errorf("Could not decode protocol DAO can-execute-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.CanExecutePDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-execute-proposal: %s", response.Error)
	}
	return response, nil
}

// Execute a proposal
func (c *Client) PDAOExecuteProposal(proposalID uint64) (api.ExecutePDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao execute-proposal %d", proposalID))
	if err != nil {
		return api.ExecutePDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO execute-proposal: %w", err)
	}
	var response api.ExecutePDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ExecutePDAOProposalResponse{}, fmt.Errorf("Could not decode protocol DAO execute-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.ExecutePDAOProposalResponse{}, fmt.Errorf("Could not get protocol DAO execute-proposal: %s", response.Error)
	}
	return response, nil
}

// Get protocol DAO settings
func (c *Client) PDAOGetSettings() (api.GetPDAOSettingsResponse, error) {
	responseBytes, err := c.callAPI("pdao get-settings")
	if err != nil {
		return api.GetPDAOSettingsResponse{}, fmt.Errorf("Could not get protocol DAO get-settings: %w", err)
	}
	var response api.GetPDAOSettingsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetPDAOSettingsResponse{}, fmt.Errorf("Could not decode protocol DAO get-settings response: %w", err)
	}
	if response.Error != "" {
		return api.GetPDAOSettingsResponse{}, fmt.Errorf("Could not get protocol DAO get-settings: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose updating a PDAO setting
func (c *Client) PDAOCanProposeSetting(contract string, setting string, value string) (api.CanProposePDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-propose-setting %s %s %s", contract, setting, value))
	if err != nil {
		return api.CanProposePDAOSettingResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-setting: %w", err)
	}
	var response api.CanProposePDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposePDAOSettingResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-setting response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposePDAOSettingResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-setting: %s", response.Error)
	}
	return response, nil
}

// Propose updating a PDAO setting (use can-propose-setting to get the pollard)
func (c *Client) PDAOProposeSetting(contract string, setting string, value string, blockNumber uint32) (api.ProposePDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao propose-setting %s %s %s %d", contract, setting, value, blockNumber))
	if err != nil {
		return api.ProposePDAOSettingResponse{}, fmt.Errorf("Could not get protocol DAO propose-setting: %w", err)
	}
	var response api.ProposePDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposePDAOSettingResponse{}, fmt.Errorf("Could not decode protocol DAO propose-setting response: %w", err)
	}
	if response.Error != "" {
		return api.ProposePDAOSettingResponse{}, fmt.Errorf("Could not get protocol DAO propose-setting: %s", response.Error)
	}
	return response, nil
}

// Get the allocation percentages of RPL rewards for the Oracle DAO, the Protocol DAO, and the node operators
func (c *Client) PDAOGetRewardsPercentages() (api.PDAOGetRewardsPercentagesResponse, error) {
	responseBytes, err := c.callAPI("pdao get-rewards-percentages")
	if err != nil {
		return api.PDAOGetRewardsPercentagesResponse{}, fmt.Errorf("Could not get protocol DAO get-rewards-percentages: %w", err)
	}
	var response api.PDAOGetRewardsPercentagesResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOGetRewardsPercentagesResponse{}, fmt.Errorf("Could not decode protocol DAO get-rewards-percentages response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOGetRewardsPercentagesResponse{}, fmt.Errorf("Could not get protocol DAO get-rewards-percentages: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose new RPL rewards allocation percentages for the Oracle DAO, the Protocol DAO, and the node operators
func (c *Client) PDAOCanProposeRewardsPercentages(node *big.Int, odao *big.Int, pdao *big.Int) (api.PDAOCanProposeRewardsPercentagesResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-propose-rewards-percentages %s %s %s", node.String(), odao.String(), pdao.String()))
	if err != nil {
		return api.PDAOCanProposeRewardsPercentagesResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-rewards-percentages: %w", err)
	}
	var response api.PDAOCanProposeRewardsPercentagesResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanProposeRewardsPercentagesResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-rewards-percentages response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanProposeRewardsPercentagesResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-rewards-percentages: %s", response.Error)
	}
	return response, nil
}

// Propose new RPL rewards allocation percentages for the Oracle DAO, the Protocol DAO, and the node operators
func (c *Client) PDAOProposeRewardsPercentages(node *big.Int, odao *big.Int, pdao *big.Int, blockNumber uint32) (api.ProposePDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao propose-rewards-percentages %s %s %s %d", node, odao, pdao, blockNumber))
	if err != nil {
		return api.ProposePDAOSettingResponse{}, fmt.Errorf("Could not get protocol DAO propose-rewards-percentages: %w", err)
	}
	var response api.ProposePDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposePDAOSettingResponse{}, fmt.Errorf("Could not decode protocol DAO propose-rewards-percentages response: %w", err)
	}
	if response.Error != "" {
		return api.ProposePDAOSettingResponse{}, fmt.Errorf("Could not get protocol DAO propose-rewards-percentages: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose a one-time spend of the Protocol DAO's treasury
func (c *Client) PDAOCanProposeOneTimeSpend(invoiceID string, recipient common.Address, amount *big.Int) (api.PDAOCanProposeOneTimeSpendResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-propose-one-time-spend %s %s %s", invoiceID, recipient.Hex(), amount.String()))
	if err != nil {
		return api.PDAOCanProposeOneTimeSpendResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-one-time-spend: %w", err)
	}
	var response api.PDAOCanProposeOneTimeSpendResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanProposeOneTimeSpendResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-one-time-spend response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanProposeOneTimeSpendResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-one-time-spend: %s", response.Error)
	}
	return response, nil
}

// Propose a one-time spend of the Protocol DAO's treasury
func (c *Client) PDAOProposeOneTimeSpend(invoiceID string, recipient common.Address, amount *big.Int, blockNumber uint32) (api.PDAOProposeOneTimeSpendResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao propose-one-time-spend %s %s %s %d", invoiceID, recipient.Hex(), amount.String(), blockNumber))
	if err != nil {
		return api.PDAOProposeOneTimeSpendResponse{}, fmt.Errorf("Could not get protocol DAO propose-one-time-spend: %w", err)
	}
	var response api.PDAOProposeOneTimeSpendResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposeOneTimeSpendResponse{}, fmt.Errorf("Could not decode protocol DAO propose-one-time-spend response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposeOneTimeSpendResponse{}, fmt.Errorf("Could not get protocol DAO propose-one-time-spend: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose a recurring spend of the Protocol DAO's treasury
func (c *Client) PDAOCanProposeRecurringSpend(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64) (api.PDAOCanProposeRecurringSpendResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-propose-recurring-spend %s %s %s %s %d %d", contractName, recipient.Hex(), amountPerPeriod.String(), periodLength.String(), startTime.Unix(), numberOfPeriods))
	if err != nil {
		return api.PDAOCanProposeRecurringSpendResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-recurring-spend: %w", err)
	}
	var response api.PDAOCanProposeRecurringSpendResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanProposeRecurringSpendResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-recurring-spend response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanProposeRecurringSpendResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-recurring-spend: %s", response.Error)
	}
	return response, nil
}

// Propose a recurring spend of the Protocol DAO's treasury
func (c *Client) PDAOProposeRecurringSpend(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64, blockNumber uint32) (api.PDAOProposeRecurringSpendResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao propose-recurring-spend %s %s %s %s %d %d %d", contractName, recipient.Hex(), amountPerPeriod.String(), periodLength.String(), startTime.Unix(), numberOfPeriods, blockNumber))
	if err != nil {
		return api.PDAOProposeRecurringSpendResponse{}, fmt.Errorf("Could not get protocol DAO propose-recurring-spend: %w", err)
	}
	var response api.PDAOProposeRecurringSpendResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposeRecurringSpendResponse{}, fmt.Errorf("Could not decode protocol DAO propose-recurring-spend response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposeRecurringSpendResponse{}, fmt.Errorf("Could not get protocol DAO propose-recurring-spend: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose an update to an existing recurring spend plan
func (c *Client) PDAOCanProposeRecurringSpendUpdate(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64) (api.PDAOCanProposeRecurringSpendUpdateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-propose-recurring-spend-update %s %s %s %s %d", contractName, recipient.Hex(), amountPerPeriod.String(), periodLength.String(), numberOfPeriods))
	if err != nil {
		return api.PDAOCanProposeRecurringSpendUpdateResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-recurring-spend-update: %w", err)
	}
	var response api.PDAOCanProposeRecurringSpendUpdateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanProposeRecurringSpendUpdateResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-recurring-spend-update response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanProposeRecurringSpendUpdateResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-recurring-spend-update: %s", response.Error)
	}
	return response, nil
}

// Propose an update to an existing recurring spend plan
func (c *Client) PDAOProposeRecurringSpendUpdate(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, blockNumber uint32) (api.PDAOProposeRecurringSpendUpdateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao propose-recurring-spend-update %s %s %s %s %d %d", contractName, recipient.Hex(), amountPerPeriod.String(), periodLength.String(), numberOfPeriods, blockNumber))
	if err != nil {
		return api.PDAOProposeRecurringSpendUpdateResponse{}, fmt.Errorf("Could not get protocol DAO propose-recurring-spend-update: %w", err)
	}
	var response api.PDAOProposeRecurringSpendUpdateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposeRecurringSpendUpdateResponse{}, fmt.Errorf("Could not decode protocol DAO propose-recurring-spend-update response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposeRecurringSpendUpdateResponse{}, fmt.Errorf("Could not get protocol DAO propose-recurring-spend-update: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can invite someone to the security council
func (c *Client) PDAOCanProposeInviteToSecurityCouncil(id string, address common.Address) (api.PDAOCanProposeInviteToSecurityCouncilResponse, error) {
	responseBytes, err := c.callAPI("pdao can-propose-invite-to-security-council", id, address.Hex())
	if err != nil {
		return api.PDAOCanProposeInviteToSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-invite-to-security-council: %w", err)
	}
	var response api.PDAOCanProposeInviteToSecurityCouncilResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanProposeInviteToSecurityCouncilResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-invite-to-security-council response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanProposeInviteToSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-invite-to-security-council: %s", response.Error)
	}
	return response, nil
}

// Propose inviting someone to the security council
func (c *Client) PDAOProposeInviteToSecurityCouncil(id string, address common.Address, blockNumber uint32) (api.PDAOProposeInviteToSecurityCouncilResponse, error) {
	responseBytes, err := c.callAPI("pdao propose-invite-to-security-council", id, address.Hex(), fmt.Sprint(blockNumber))
	if err != nil {
		return api.PDAOProposeInviteToSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO propose-invite-to-security-council: %w", err)
	}
	var response api.PDAOProposeInviteToSecurityCouncilResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposeInviteToSecurityCouncilResponse{}, fmt.Errorf("Could not decode protocol DAO propose-invite-to-security-council response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposeInviteToSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO propose-invite-to-security-council: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can kick someone from the security council
func (c *Client) PDAOCanProposeKickFromSecurityCouncil(address common.Address) (api.PDAOCanProposeKickFromSecurityCouncilResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-propose-kick-from-security-council %s", address.Hex()))
	if err != nil {
		return api.PDAOCanProposeKickFromSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-kick-from-security-council: %w", err)
	}
	var response api.PDAOCanProposeKickFromSecurityCouncilResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanProposeKickFromSecurityCouncilResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-kick-from-security-council response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanProposeKickFromSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-kick-from-security-council: %s", response.Error)
	}
	return response, nil
}

// Propose kicking someone from the security council
func (c *Client) PDAOProposeKickFromSecurityCouncil(address common.Address, blockNumber uint32) (api.PDAOProposeKickFromSecurityCouncilResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao propose-kick-from-security-council %s %d", address.Hex(), blockNumber))
	if err != nil {
		return api.PDAOProposeKickFromSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO propose-kick-from-security-council: %w", err)
	}
	var response api.PDAOProposeKickFromSecurityCouncilResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposeKickFromSecurityCouncilResponse{}, fmt.Errorf("Could not decode protocol DAO propose-kick-from-security-council response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposeKickFromSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO propose-kick-from-security-council: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can kick multiple members from the security council
func (c *Client) PDAOCanProposeKickMultiFromSecurityCouncil(addresses []common.Address) (api.PDAOCanProposeKickMultiFromSecurityCouncilResponse, error) {
	addressStrings := make([]string, len(addresses))
	for i, address := range addresses {
		addressStrings[i] = address.Hex()
	}

	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-propose-kick-multi-from-security-council %s", strings.Join(addressStrings, ",")))
	if err != nil {
		return api.PDAOCanProposeKickMultiFromSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-kick-multi-from-security-council: %w", err)
	}
	var response api.PDAOCanProposeKickMultiFromSecurityCouncilResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanProposeKickMultiFromSecurityCouncilResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-kick-multi-from-security-council response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanProposeKickMultiFromSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-kick-multi-from-security-council: %s", response.Error)
	}
	return response, nil
}

// Propose kicking multiple members from the security council
func (c *Client) PDAOProposeKickMultiFromSecurityCouncil(addresses []common.Address, blockNumber uint32) (api.PDAOProposeKickMultiFromSecurityCouncilResponse, error) {
	addressStrings := make([]string, len(addresses))
	for i, address := range addresses {
		addressStrings[i] = address.Hex()
	}

	responseBytes, err := c.callAPI(fmt.Sprintf("pdao propose-kick-multi-from-security-council %s %d", strings.Join(addressStrings, ","), blockNumber))
	if err != nil {
		return api.PDAOProposeKickMultiFromSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO propose-kick-multi-from-security-council: %w", err)
	}
	var response api.PDAOProposeKickMultiFromSecurityCouncilResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposeKickMultiFromSecurityCouncilResponse{}, fmt.Errorf("Could not decode protocol DAO propose-kick-multi-from-security-council response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposeKickMultiFromSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO propose-kick-multi-from-security-council: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose replacing someone on the security council with another member
func (c *Client) PDAOCanProposeReplaceMemberOfSecurityCouncil(existingAddress common.Address, newID string, newAddress common.Address) (api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-propose-replace-member-of-security-council %s", existingAddress.Hex()), newID, newAddress.Hex())
	if err != nil {
		return api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-replace-member-of-security-council: %w", err)
	}
	var response api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse{}, fmt.Errorf("Could not decode protocol DAO can-propose-replace-member-of-security-council response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO can-propose-replace-member-of-security-council: %s", response.Error)
	}
	return response, nil
}

// Propose replacing someone on the security council with another member
func (c *Client) PDAOProposeReplaceMemberOfSecurityCouncil(existingAddress common.Address, newID string, newAddress common.Address, blockNumber uint32) (api.PDAOProposeReplaceMemberOfSecurityCouncilResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao propose-replace-member-of-security-council %s", existingAddress.Hex()), newID, newAddress.Hex(), fmt.Sprint(blockNumber))
	if err != nil {
		return api.PDAOProposeReplaceMemberOfSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO propose-replace-member-of-security-council: %w", err)
	}
	var response api.PDAOProposeReplaceMemberOfSecurityCouncilResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOProposeReplaceMemberOfSecurityCouncilResponse{}, fmt.Errorf("Could not decode protocol DAO propose-replace-member-of-security-council response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOProposeReplaceMemberOfSecurityCouncilResponse{}, fmt.Errorf("Could not get protocol DAO propose-replace-member-of-security-council: %s", response.Error)
	}
	return response, nil
}

// Get the list of proposals with claimable / rewardable bonds, and the relevant indices for each one
func (c *Client) PDAOGetClaimableBonds() (api.PDAOGetClaimableBondsResponse, error) {
	responseBytes, err := c.callAPI("pdao get-claimable-bonds")
	if err != nil {
		return api.PDAOGetClaimableBondsResponse{}, fmt.Errorf("Could not get protocol DAO get-claimable-bonds: %w", err)
	}
	var response api.PDAOGetClaimableBondsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOGetClaimableBondsResponse{}, fmt.Errorf("Could not decode protocol DAO get-claimable-bonds response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOGetClaimableBondsResponse{}, fmt.Errorf("Could not get protocol DAO get-claimable-bonds: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can claim / unlock bonds from a proposal
func (c *Client) PDAOCanClaimBonds(proposalID uint64, indices []uint64) (api.PDAOCanClaimBondsResponse, error) {
	indicesStrings := make([]string, len(indices))
	for i, index := range indices {
		indicesStrings[i] = fmt.Sprint(index)
	}

	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-claim-bonds %d %s", proposalID, strings.Join(indicesStrings, ",")))
	if err != nil {
		return api.PDAOCanClaimBondsResponse{}, fmt.Errorf("Could not get protocol DAO can-claim-bonds: %w", err)
	}
	var response api.PDAOCanClaimBondsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanClaimBondsResponse{}, fmt.Errorf("Could not decode protocol DAO can-claim-bonds response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanClaimBondsResponse{}, fmt.Errorf("Could not get protocol DAO can-claim-bonds: %s", response.Error)
	}
	return response, nil
}

// Claim / unlock bonds from a proposal
func (c *Client) PDAOClaimBonds(isProposer bool, proposalID uint64, indices []uint64) (api.PDAOClaimBondsResponse, error) {
	indicesStrings := make([]string, len(indices))
	for i, index := range indices {
		indicesStrings[i] = fmt.Sprint(index)
	}

	responseBytes, err := c.callAPI(fmt.Sprintf("pdao claim-bonds %t %d %s", isProposer, proposalID, strings.Join(indicesStrings, ",")))
	if err != nil {
		return api.PDAOClaimBondsResponse{}, fmt.Errorf("Could not get protocol DAO claim-bonds: %w", err)
	}
	var response api.PDAOClaimBondsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOClaimBondsResponse{}, fmt.Errorf("Could not decode protocol DAO claim-bonds response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOClaimBondsResponse{}, fmt.Errorf("Could not get protocol DAO claim-bonds: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can defeat a proposal
func (c *Client) PDAOCanDefeatProposal(proposalID uint64, index uint64) (api.PDAOCanDefeatProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-defeat-proposal %d %d", proposalID, index))
	if err != nil {
		return api.PDAOCanDefeatProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-defeat-proposal: %w", err)
	}
	var response api.PDAOCanDefeatProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanDefeatProposalResponse{}, fmt.Errorf("Could not decode protocol DAO can-defeat-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanDefeatProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-defeat-proposal: %s", response.Error)
	}
	return response, nil
}

// Defeat a proposal
func (c *Client) PDAODefeatProposal(proposalID uint64, index uint64) (api.PDAODefeatProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao defeat-proposal %d %d", proposalID, index))
	if err != nil {
		return api.PDAODefeatProposalResponse{}, fmt.Errorf("Could not get protocol DAO defeat-proposal: %w", err)
	}
	var response api.PDAODefeatProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAODefeatProposalResponse{}, fmt.Errorf("Could not decode protocol DAO defeat-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.PDAODefeatProposalResponse{}, fmt.Errorf("Could not get protocol DAO defeat-proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can finalize a proposal
func (c *Client) PDAOCanFinalizeProposal(proposalID uint64) (api.PDAOCanFinalizeProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-finalize-proposal %d", proposalID))
	if err != nil {
		return api.PDAOCanFinalizeProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-finalize-proposal: %w", err)
	}
	var response api.PDAOCanFinalizeProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanFinalizeProposalResponse{}, fmt.Errorf("Could not decode protocol DAO can-finalize-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanFinalizeProposalResponse{}, fmt.Errorf("Could not get protocol DAO can-finalize-proposal: %s", response.Error)
	}
	return response, nil
}

// Finalize a proposal
func (c *Client) PDAOFinalizeProposal(proposalID uint64) (api.PDAOFinalizeProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao finalize-proposal %d", proposalID))
	if err != nil {
		return api.PDAOFinalizeProposalResponse{}, fmt.Errorf("Could not get protocol DAO finalize-proposal: %w", err)
	}
	var response api.PDAOFinalizeProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOFinalizeProposalResponse{}, fmt.Errorf("Could not decode protocol DAO finalize-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOFinalizeProposalResponse{}, fmt.Errorf("Could not get protocol DAO finalize-proposal: %s", response.Error)
	}
	return response, nil
}

// CanSetVotingDelegate estimates the gas required to set an on-chain voting delegate
func (c *Client) EstimateSetVotingDelegateGas(address common.Address) (api.PDAOCanSetVotingDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao estimate-set-voting-delegate-gas %s", address.Hex()))
	if err != nil {
		return api.PDAOCanSetVotingDelegateResponse{}, fmt.Errorf("could not call estimate-set-voting-delegate-gas: %w", err)
	}
	var response api.PDAOCanSetVotingDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanSetVotingDelegateResponse{}, fmt.Errorf("could not decode estimate-set-voting-delegate-gas response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanSetVotingDelegateResponse{}, fmt.Errorf("error after requesting estimate-set-voting-delegate-gas: %s", response.Error)
	}
	return response, nil
}

// SetVotingDelegate set an on-chain voting delegate for the node
func (c *Client) SetVotingDelegate(address common.Address) (api.PDAOSetVotingDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao set-voting-delegate %s", address.Hex()))
	if err != nil {
		return api.PDAOSetVotingDelegateResponse{}, fmt.Errorf("could not call set-voting-delegate: %w", err)
	}
	var response api.PDAOSetVotingDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOSetVotingDelegateResponse{}, fmt.Errorf("could not decode set-voting-delegate response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOSetVotingDelegateResponse{}, fmt.Errorf("error after requesting set-voting-delegate: %s", response.Error)
	}
	return response, nil
}

// GetCurrentVotingDelegate gets the node current on-chain voting delegate
func (c *Client) GetCurrentVotingDelegate() (api.PDAOCurrentVotingDelegateResponse, error) {
	responseBytes, err := c.callAPI("pdao get-current-voting-delegate")
	if err != nil {
		return api.PDAOCurrentVotingDelegateResponse{}, fmt.Errorf("could not request get-current-voting-delegate: %w", err)
	}
	var response api.PDAOCurrentVotingDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCurrentVotingDelegateResponse{}, fmt.Errorf("could not decode get-current-voting-delegate: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCurrentVotingDelegateResponse{}, fmt.Errorf("error after requesting get-current-voting-delegate: %s", response.Error)
	}
	return response, nil
}

// CanInitializeVoting fetches whether the node's is initialized for on-chain voting
func (c *Client) CanInitializeVoting(delegateAddress common.Address) (api.PDAOCanInitializeVotingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-initialize-voting %s", delegateAddress.Hex()))
	if err != nil {
		return api.PDAOCanInitializeVotingResponse{}, fmt.Errorf("could not call can-initialize-voting: %w", err)
	}
	var response api.PDAOCanInitializeVotingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanInitializeVotingResponse{}, fmt.Errorf("could not decode can-initialize-voting response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanInitializeVotingResponse{}, fmt.Errorf("error after requesting can-initialize-voting: %s", response.Error)
	}
	return response, nil
}

// InitializeVoting unlocks the node's voting power
func (c *Client) InitializeVoting(delegateAddress common.Address) (api.PDAOInitializeVotingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao initialize-voting %s", delegateAddress.Hex()))
	if err != nil {
		return api.PDAOInitializeVotingResponse{}, fmt.Errorf("could not call initialize-voting: %w", err)
	}
	var response api.PDAOInitializeVotingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOInitializeVotingResponse{}, fmt.Errorf("could not decode initialize-voting response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOInitializeVotingResponse{}, fmt.Errorf("error after requesting initialize-voting: %s", response.Error)
	}
	return response, nil
}

// CanSetSignallingAddress fetches gas info and if a node can set the signalling address
func (c *Client) CanSetSignallingAddress(signallingAddress common.Address, signature string) (api.PDAOCanSetSignallingAddressResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao can-set-signalling-address %s %s", signallingAddress.Hex(), signature))
	if err != nil {
		return api.PDAOCanSetSignallingAddressResponse{}, fmt.Errorf("could not call can-set-signalling-address: %w", err)
	}
	var response api.PDAOCanSetSignallingAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanSetSignallingAddressResponse{}, fmt.Errorf("could not decode can-set-signalling-address response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanSetSignallingAddressResponse{}, fmt.Errorf("error after requesting can-set-signalling-address: %s", response.Error)
	}
	return response, nil
}

// SetSignallingAddress sets the node's signalling address
func (c *Client) SetSignallingAddress(signallingAddress common.Address, signature string) (api.PDAOSetSignallingAddressResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("pdao set-signalling-address %s %s", signallingAddress.Hex(), signature))
	if err != nil {
		return api.PDAOSetSignallingAddressResponse{}, fmt.Errorf("could not call set-signalling-address: %w", err)
	}
	var response api.PDAOSetSignallingAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOSetSignallingAddressResponse{}, fmt.Errorf("could not decode set-signalling-address response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOSetSignallingAddressResponse{}, fmt.Errorf("error after requesting set-signalling-address: %s", response.Error)
	}
	return response, nil
}

// CanClearSignallingAddress fetches gas info and if a node can clear a signalling address
func (c *Client) CanClearSignallingAddress() (api.PDAOCanClearSignallingAddressResponse, error) {
	responseBytes, err := c.callAPI("pdao can-clear-signalling-address")
	if err != nil {
		return api.PDAOCanClearSignallingAddressResponse{}, fmt.Errorf("could not call can-clear-signalling-address: %w", err)
	}
	var response api.PDAOCanClearSignallingAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOCanClearSignallingAddressResponse{}, fmt.Errorf("could not decode can-clear-signalling-address response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOCanClearSignallingAddressResponse{}, fmt.Errorf("error after requesting can-clear-signalling-address: %s", response.Error)
	}
	return response, nil
}

// ClearSignallingAddress sets the node's signalling address
func (c *Client) ClearSignallingAddress() (api.PDAOSetSignallingAddressResponse, error) {
	responseBytes, err := c.callAPI("pdao clear-signalling-address")
	if err != nil {
		return api.PDAOSetSignallingAddressResponse{}, fmt.Errorf("could not call clear-signalling-address: %w", err)
	}
	var response api.PDAOSetSignallingAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOSetSignallingAddressResponse{}, fmt.Errorf("could not decode clear-signalling-address response: %w", err)
	}
	if response.Error != "" {
		return api.PDAOSetSignallingAddressResponse{}, fmt.Errorf("error after requesting clear-signalling-address: %s", response.Error)
	}
	return response, nil
}

// Get PDAO Status
func (c *Client) PDAOStatus() (api.PDAOStatusResponse, error) {
	responseBytes, err := c.callAPI("pdao status")
	if err != nil {
		return api.PDAOStatusResponse{}, fmt.Errorf("could not call get pdao status: %w", err)
	}
	var response api.PDAOStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PDAOStatusResponse{}, fmt.Errorf("could not decode get-voting-power: %w", err)
	}
	if response.Error != "" {
		return api.PDAOStatusResponse{}, fmt.Errorf("error after requesting get-voting-power: %s", response.Error)
	}
	return response, nil
}
