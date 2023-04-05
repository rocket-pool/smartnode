package client

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// Request types
type VoluntaryExitMessage struct {
	Epoch          uinteger `json:"epoch"`
	ValidatorIndex uinteger `json:"validator_index"`
}
type VoluntaryExitRequest struct {
	Message   VoluntaryExitMessage `json:"message"`
	Signature byteArray            `json:"signature"`
}
type BLSToExecutionChangeMessage struct {
	ValidatorIndex     uinteger  `json:"validator_index"`
	FromBLSPubkey      byteArray `json:"from_bls_pubkey"`
	ToExecutionAddress byteArray `json:"to_execution_address"`
}
type BLSToExecutionChangeRequest struct {
	Message   BLSToExecutionChangeMessage `json:"message"`
	Signature byteArray                   `json:"signature"`
}

// Response types
type SyncStatusResponse struct {
	Data struct {
		IsSyncing    bool     `json:"is_syncing"`
		HeadSlot     uinteger `json:"head_slot"`
		SyncDistance uinteger `json:"sync_distance"`
	} `json:"data"`
}
type Eth2ConfigResponse struct {
	Data struct {
		SecondsPerSlot               uinteger `json:"SECONDS_PER_SLOT"`
		SlotsPerEpoch                uinteger `json:"SLOTS_PER_EPOCH"`
		EpochsPerSyncCommitteePeriod uinteger `json:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD"`
	} `json:"data"`
}
type Eth2DepositContractResponse struct {
	Data struct {
		ChainID uinteger       `json:"chain_id"`
		Address common.Address `json:"address"`
	} `json:"data"`
}
type GenesisResponse struct {
	Data struct {
		GenesisTime           uinteger  `json:"genesis_time"`
		GenesisForkVersion    byteArray `json:"genesis_fork_version"`
		GenesisValidatorsRoot byteArray `json:"genesis_validators_root"`
	} `json:"data"`
}
type FinalityCheckpointsResponse struct {
	Data struct {
		PreviousJustified struct {
			Epoch uinteger `json:"epoch"`
		} `json:"previous_justified"`
		CurrentJustified struct {
			Epoch uinteger `json:"epoch"`
		} `json:"current_justified"`
		Finalized struct {
			Epoch uinteger `json:"epoch"`
		} `json:"finalized"`
	} `json:"data"`
}
type ForkResponse struct {
	Data struct {
		PreviousVersion byteArray `json:"previous_version"`
		CurrentVersion  byteArray `json:"current_version"`
		Epoch           uinteger  `json:"epoch"`
	} `json:"data"`
}
type AttestationsResponse struct {
	Data []Attestation `json:"data"`
}
type BeaconBlockResponse struct {
	Data struct {
		Message struct {
			Slot          uinteger `json:"slot"`
			ProposerIndex uinteger `json:"proposer_index"`
			Body          struct {
				Eth1Data struct {
					DepositRoot  byteArray `json:"deposit_root"`
					DepositCount uinteger  `json:"deposit_count"`
					BlockHash    byteArray `json:"block_hash"`
				} `json:"eth1_data"`
				Attestations     []Attestation `json:"attestations"`
				ExecutionPayload *struct {
					FeeRecipient byteArray `json:"fee_recipient"`
					BlockNumber  uinteger  `json:"block_number"`
				} `json:"execution_payload"`
			} `json:"body"`
		} `json:"message"`
	} `json:"data"`
}
type ValidatorsResponse struct {
	Data []Validator `json:"data"`
}
type Validator struct {
	Index     uinteger `json:"index"`
	Balance   uinteger `json:"balance"`
	Status    string   `json:"status"`
	Validator struct {
		Pubkey                     byteArray `json:"pubkey"`
		WithdrawalCredentials      byteArray `json:"withdrawal_credentials"`
		EffectiveBalance           uinteger  `json:"effective_balance"`
		Slashed                    bool      `json:"slashed"`
		ActivationEligibilityEpoch uinteger  `json:"activation_eligibility_epoch"`
		ActivationEpoch            uinteger  `json:"activation_epoch"`
		ExitEpoch                  uinteger  `json:"exit_epoch"`
		WithdrawableEpoch          uinteger  `json:"withdrawable_epoch"`
	} `json:"validator"`
}
type SyncDutiesResponse struct {
	Data []SyncDuty `json:"data"`
}
type SyncDuty struct {
	Pubkey               byteArray  `json:"pubkey"`
	ValidatorIndex       uinteger   `json:"validator_index"`
	SyncCommitteeIndices []uinteger `json:"validator_sync_committee_indices"`
}
type ProposerDutiesResponse struct {
	Data []ProposerDuty `json:"data"`
}
type ProposerDuty struct {
	ValidatorIndex uinteger `json:"validator_index"`
}

type CommitteesResponse struct {
	Data []Committee `json:"data"`
}

type Committee struct {
	Index      uinteger   `json:"index"`
	Slot       uinteger   `json:"slot"`
	Validators []uinteger `json:"validators"`
}

type Attestation struct {
	AggregationBits string `json:"aggregation_bits"`
	Data            struct {
		Slot  uinteger `json:"slot"`
		Index uinteger `json:"index"`
	} `json:"data"`
}

// Don't mind meeeee
const intSize = 32 << (^uint(0) >> 63)

// Unsigned integer type
type uinteger uint64

func (i uinteger) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(i)))
}
func (i *uinteger) UnmarshalJSON(data []byte) error {

	// Unmarshal string
	var dataStr string
	/*if err := json.Unmarshal(data, &dataStr); err != nil {
		return err
	}*/
	dataLen := len(data)
	if dataLen <= 2 {
		return fmt.Errorf("Invalid json uinteger '%s'", string(data))
	}
	dataStr = string(data[1 : dataLen-1])
	sLen := dataLen - 2

	// Check fast paths
	if intSize == 32 && (0 < sLen && sLen < 10) ||
		intSize == 64 && (0 < sLen && sLen < 19) {

		if sLen > 0 && dataStr[0] != '-' {
			signed, err := strconv.Atoi(dataStr)
			if err == nil {
				*i = uinteger(signed)
				return nil
			}
		}

		// If fast path failed just fall through
	}

	// Parse integer value
	value, err := strconv.ParseUint(dataStr, 10, 64)
	if err != nil {
		return err
	}

	// Set value and return
	*i = uinteger(value)
	return nil

}

// Byte array type
type byteArray []byte

func (b byteArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(hexutil.AddPrefix(hex.EncodeToString(b)))
}
func (b *byteArray) UnmarshalJSON(data []byte) error {

	// Unmarshal string
	var dataStr string
	if err := json.Unmarshal(data, &dataStr); err != nil {
		return err
	}

	// Decode hex
	value, err := hex.DecodeString(hexutil.RemovePrefix(dataStr))
	if err != nil {
		return err
	}

	// Set value and return
	*b = value
	return nil

}
