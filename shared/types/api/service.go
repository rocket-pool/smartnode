package api

import "github.com/rocket-pool/smartnode/shared/services"

type TerminateDataFolderResponse struct {
	Status        string `json:"status"`
	Error         string `json:"error"`
	FolderExisted bool   `json:"folderExisted"`
}

type ExecutionClientStatusResponse struct {
	Status        string                                `json:"status"`
	Error         string                                `json:"error"`
	ManagerStatus services.ExecutionClientManagerStatus `json:"managerStatus"`
}
