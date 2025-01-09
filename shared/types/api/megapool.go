package api

import "github.com/ethereum/go-ethereum/common"

type MegapoolStatusResponse struct {
	Status                      string          `json:"status"`
	Error                       string          `json:"error"`
	Megapool                    MegapoolDetails `json:"megapoolDetails"`
	NodeAccount                 common.Address  `json:"nodeAccount"`
	NodeAccountAddressFormatted string          `json:"nodeAccountAddressFormatted"`
}

type MegapoolDetails struct {
	MegapoolAddress          common.Address `json:"address"`
	MegapoolAddressFormatted string         `json:"megapoolAddressFormatted"`
	MegapoolDeployed         bool           `json:"megapoolDeployed"`
	DelegateExpiry           uint64         `json:"DelegateExpiry"`
}
