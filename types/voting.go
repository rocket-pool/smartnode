package types

// Challenge states
type ChallengeState uint8

const (
	ChallengeState_Unchallenged ChallengeState = iota
	ChallengeState_Challenged
	ChallengeState_Responded
	ChallengeState_Paid
)
