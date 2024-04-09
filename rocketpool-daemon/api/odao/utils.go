package odao

import (
	"time"
)

// Check if we are currently within a proposal's actionability window
func isProposalActionable(actionWindow time.Duration, executedTime time.Time, currentTime time.Time) bool {
	return currentTime.Before(executedTime.Add(actionWindow))
}

// Check if the node's proposal cooldown is still active, so it can't make new proposals yet
func isProposalCooldownActive(cooldownTime time.Duration, lastProposalTime time.Time, currentTime time.Time) bool {
	return cooldownTime > currentTime.Sub(lastProposalTime)
}
