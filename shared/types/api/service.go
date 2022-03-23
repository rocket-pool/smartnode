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
