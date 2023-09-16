package types

type ValidatorState string

const (
	ValidatorState_PendingInitialized ValidatorState = "pending_initialized"
	ValidatorState_PendingQueued      ValidatorState = "pending_queued"
	ValidatorState_ActiveOngoing      ValidatorState = "active_ongoing"
	ValidatorState_ActiveExiting      ValidatorState = "active_exiting"
	ValidatorState_ActiveSlashed      ValidatorState = "active_slashed"
	ValidatorState_ExitedUnslashed    ValidatorState = "exited_unslashed"
	ValidatorState_ExitedSlashed      ValidatorState = "exited_slashed"
	ValidatorState_WithdrawalPossible ValidatorState = "withdrawal_possible"
	ValidatorState_WithdrawalDone     ValidatorState = "withdrawal_done"
)
