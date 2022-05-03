package api

type TerminateDataFolderResponse struct {
	Status        string `json:"status"`
	Error         string `json:"error"`
	FolderExisted bool   `json:"folderExisted"`
}

type ExecutionClientStatusResponse struct {
	Status     string `json:"status"`
	Error      string `json:"error"`
	UsePrimary bool   `json:"usePrimary"`
	Log        string `json:"log"`
}
