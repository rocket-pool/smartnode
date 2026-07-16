#!/bin/sh

# Generates optimized SSZ methods for the eth2 beacon types with dynssz-gen.
# The generated methods are picked up automatically by the shared
# generic.SSZ (dynamic-ssz) instance; reflection is used for any type
# without generated code.
# NOTE: shared/services/rewards/ssz_types still uses fastssz (see its gen.sh).
set -e

# dynssz-gen is a tool dependency pinned in go.mod (see the tool directive)
DYNSSZ_GEN="go tool dynssz-gen"

find ./shared/types/eth2 -name "*_generated.go" -delete

$DYNSSZ_GEN -package ./shared/types/eth2/generic -with-streaming \
	-types "Fork,BeaconBlockHeader,Eth1Data,Validator,Checkpoint,SyncCommittee,HistoricalSummary,HistoricalSummaryLists,ExecutionPayloadHeader,PendingDeposit,PendingPartialWithdrawal,PendingConsolidation,ProposerSlashing,SignedBeaconBlockHeader,AttesterSlashing,IndexedAttestation,AttestationData,Attestation,Deposit,SignedVoluntaryExit,SyncAggregate,ExecutionPayload,Withdrawal,BLSToExecutionChange,SignedBLSToExecutionChange,DepositDataNoSignature,DepositData,SigningRoot,VoluntaryExit,WithdrawalCredentialsChange,Uint256" \
	-output ./shared/types/eth2/generic/generic_generated.go

$DYNSSZ_GEN -package ./shared/types/eth2/fork/deneb -with-streaming \
	-types "BeaconState,SignedBeaconBlock,BeaconBlock,BeaconBlockBody,ExecutionPayload" \
	-output ./shared/types/eth2/fork/deneb/deneb_generated.go

$DYNSSZ_GEN -package ./shared/types/eth2/fork/electra -with-streaming \
	-types "BeaconState,SignedBeaconBlock,BeaconBlock,BeaconBlockBody,Attestation,ExecutionRequests,DepositRequest,WithdrawalRequest,ConsolidationRequest,AttesterSlashing,IndexedAttestation" \
	-output ./shared/types/eth2/fork/electra/electra_generated.go

$DYNSSZ_GEN -package ./shared/types/eth2/fork/fulu -with-streaming \
	-types "BeaconState,SignedBeaconBlock,BeaconBlock,BeaconBlockBody,Attestation,ExecutionRequests,DepositRequest,WithdrawalRequest,ConsolidationRequest,AttesterSlashing,IndexedAttestation" \
	-output ./shared/types/eth2/fork/fulu/fulu_generated.go

$DYNSSZ_GEN -package ./shared/types/eth2/fork/gloas -with-streaming \
	-types "BeaconState,SignedBeaconBlock,BeaconBlock,BeaconBlockBody,Builder,BuilderPendingWithdrawal,BuilderPendingPayment,Attestation,ExecutionPayloadBid,SignedExecutionPayloadBid,PayloadAttestationData,PayloadAttestation,ExecutionRequests,DepositRequest,WithdrawalRequest,ConsolidationRequest,BuilderDepositRequest,BuilderExitRequest,AttesterSlashing,IndexedAttestation" \
	-output ./shared/types/eth2/fork/gloas/gloas_generated.go