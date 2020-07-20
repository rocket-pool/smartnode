package api

import (
    "github.com/ethereum/go-ethereum/common"
)


type AccountStatusResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    PasswordExists bool             `json:"passwordExists"`
    AccountExists bool              `json:"accountExists"`
    AccountAddress common.Address   `json:"accountAddress"`
}


type InitPasswordResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
}


type InitAccountResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    AccountAddress common.Address   `json:"accountAddress"`
}


type ExportAccountResponse struct {
    Status string                   `json:"status"`
    Error string                    `json:"error"`
    Password string                 `json:"password"`
    KeystorePath string             `json:"keystorePath"`
    KeystoreFile string             `json:"keystoreFile"`
}

