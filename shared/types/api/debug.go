package api

type RewardsEventResponse struct {
	Status            string   `json:"status"`
	Error             string   `json:"error"`
	Found             bool     `json:"found"`
	Index             string   `json:"index"`
	ExecutionBlock    string   `json:"executionBlock"`
	ConsensusBlock    string   `json:"consensusBlock"`
	MerkleRoot        string   `json:"merkleRoot"`
	IntervalsPassed   string   `json:"intervalsPassed"`
	TreasuryRPL       string   `json:"treasuryRPL"`
	TrustedNodeRPL    []string `json:"trustedNodeRPL"`
	NodeRPL           []string `json:"nodeRPL"`
	NodeETH           []string `json:"nodeETH"`
	UserETH           string   `json:"userETH"`
	IntervalStartTime int64    `json:"intervalStartTime"`
	IntervalEndTime   int64    `json:"intervalEndTime"`
	SubmissionTime    int64    `json:"submissionTime"`
}
