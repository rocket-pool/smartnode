package exchange

import (
    "github.com/rocket-pool/smartnode/shared/services"
)


// Get token price response type
type GetTokenPriceResponse struct {

}


// Get token price
func GetTokenPrice(p *services.Provider, token string) (*GetTokenPriceResponse, error) {

    // Response
    response := &GetTokenPriceResponse{}



    // Return response
    return response, nil

}

