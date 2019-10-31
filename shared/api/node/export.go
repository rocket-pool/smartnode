package node

import (
    "io/ioutil"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Node export response type
type NodeExportResponse struct {
    Password string         `json:"password"`
    KeystorePath string     `json:"keystorePath"`
    KeystoreFile string     `json:"keystoreFile"`
}


// Export node account
func ExportNodeAccount(p *services.Provider) (*NodeExportResponse, error) {

    // Get passphrase
    passphrase, err := p.PM.GetPassphrase()
    if err != nil { return nil, err }

    // Get node account
    nodeAccount, err := p.AM.GetNodeAccount()
    if err != nil { return nil, err }

    // Get node account keystore file
    keystoreFile, err := ioutil.ReadFile(nodeAccount.URL.Path)
    if err != nil { return nil, err }

    // Return response
    return &NodeExportResponse{
        Password: passphrase,
        KeystorePath: nodeAccount.URL.Path,
        KeystoreFile: string(keystoreFile),
    }, nil

}

