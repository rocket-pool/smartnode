package api


type AccountStatusResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
    PasswordExists bool     `json:"passwordExists"`
    AccountExists bool      `json:"accountExists"`
    AccountAddress string   `json:"accountAddress"`
}


type ExportAccountResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
    Password string         `json:"password"`
    KeystorePath string     `json:"keystorePath"`
    KeystoreFile string     `json:"keystoreFile"`
}


type InitAccountResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
    AccountAddress string   `json:"accountAddress"`
}

