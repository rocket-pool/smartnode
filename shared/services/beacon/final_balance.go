package beacon

import "fmt"

func HasFinalBalanceWithdrawal(status ValidatorStatus) bool {
	return status.Exists && status.Status == ValidatorState_WithdrawalDone
}

func FinalBalanceSweepNote(status ValidatorStatus) string {
	switch status.Status {
	case ValidatorState_WithdrawalPossible:
		return "withdrawable; waiting for the beacon withdrawal sweep (full balance)"
	case ValidatorState_ExitedUnslashed, ValidatorState_ExitedSlashed:
		return "exited; waiting to become withdrawal_possible, then for the sweep"
	default:
		return fmt.Sprintf("withdrawable epoch reached; waiting for full withdrawal (status %s)", status.Status)
	}
}
