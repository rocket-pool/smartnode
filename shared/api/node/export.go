package node

import (
    "io/ioutil"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Node export response type
type NodeExportResponse struct {
    Passphrase string   `json:"passphrase"`
    KeyFile string      `json:"keyFile"`
}


// Export node account
func ExportNodeAccount(p *services.Provider) (*NodeExportResponse, error) {

    // Get passphrase
    passphrase, err := p.PM.GetPassphrase()
    if err != nil { return nil, err }

    // Get node account
    nodeAccount, err := p.AM.GetNodeAccount()
    if err != nil { return nil, err }

    // Get node account key file
    keyFile, err := ioutil.ReadFile(nodeAccount.URL.Path)
    if err != nil { return nil, err }

    // Return response
    return &NodeExportResponse{
        Passphrase: passphrase,
        KeyFile: string(keyFile),
    }, nil

}

