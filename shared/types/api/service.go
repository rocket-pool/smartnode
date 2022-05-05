package api

import "github.com/ethereum/go-ethereum/common"

type TerminateDataFolderResponse struct {
	Status        string `json:"status"`
	Error         string `json:"error"`
	FolderExisted bool   `json:"folderExisted"`
}

type CreateFeeRecipientFileResponse struct {
	Status      string         `json:"status"`
	Error       string         `json:"error"`
	Distributor common.Address `json:"distributor"`
}

// This is a wrapper for the EC status report
type ExecutionClientStatus struct {
	IsWorking    bool    `json:"isWorking"`
	IsSynced     bool    `json:"isSynced"`
	SyncProgress float64 `json:"syncProgress"`
	Error        string  `json:"error"`
}

// This is a wrapper for the manager's overall status report
type ExecutionClientManagerStatus struct {
	PrimaryEcStatus  ExecutionClientStatus `json:"primaryEcStatus"`
	FallbackEnabled  bool                  `json:"fallbackEnabled"`
	FallbackEcStatus ExecutionClientStatus `json:"fallbackEcStatus"`
}

type ExecutionClientStatusResponse struct {
	Status        string                       `json:"status"`
	Error         string                       `json:"error"`
	ManagerStatus ExecutionClientManagerStatus `json:"managerStatus"`
}
