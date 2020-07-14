package api


type AccountStatusResponse struct {
    Status string           `json:"status"`
    Error string            `json:"error"`
    PasswordExists bool     `json:"passwordExists"`
    AccountExists bool      `json:"accountExists"`
    AccountAddress string   `json:"accountAddress"`
}

