package api

import "github.com/ethereum/go-ethereum/common"

type ServiceTerminateDataFolderData struct {
	FolderExisted bool `json:"folderExisted"`
}

type ServiceCreateFeeRecipientFileData struct {
	Distributor common.Address `json:"distributor"`
}

// This is a wrapper for the EC status report
type ClientStatus struct {
	IsWorking    bool    `json:"isWorking"`
	IsSynced     bool    `json:"isSynced"`
	SyncProgress float64 `json:"syncProgress"`
	NetworkId    uint    `json:"networkId"`
	Error        string  `json:"error"`
}

// This is a wrapper for the manager's overall status report
type ClientManagerStatus struct {
	PrimaryClientStatus  ClientStatus `json:"primaryEcStatus"`
	FallbackEnabled      bool         `json:"fallbackEnabled"`
	FallbackClientStatus ClientStatus `json:"fallbackEcStatus"`
}

type ServiceClientStatusData struct {
	EcManagerStatus ClientManagerStatus `json:"ecManagerStatus"`
	BcManagerStatus ClientManagerStatus `json:"bcManagerStatus"`
}

type ServiceVersionData struct {
	Version string `json:"version"`
}
