package megapool

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canDelegateUpgrade(c *cli.Command, megapoolAddress common.Address) (*api.MegapoolCanDelegateUpgradeResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolCanDelegateUpgradeResponse{}

	// Create megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := mega.EstimateDelegateUpgradeGas(opts)
	if err == nil {
		response.GasLimits = gasLimits
	}

	// Return response
	return &response, nil
}

func delegateUpgrade(c *cli.Command, megapoolAddress common.Address, opts *bind.TransactOpts) (*api.MegapoolDelegateUpgradeResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolDelegateUpgradeResponse{}

	// Create megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Upgrade
	hash, err := mega.DelegateUpgrade(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}

func getUseLatestDelegate(c *cli.Command) (*api.MegapoolGetUseLatestDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolGetUseLatestDelegateResponse{}

	// Get node account and derive the megapool address
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Create megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	setting, err := mega.GetUseLatestDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting use latest delegate: %w", err)
	}

	//Return response
	response.Setting = setting
	return &response, nil

}

func canSetUseLatestDelegate(c *cli.Command, useLatest bool) (*api.MegapoolCanSetUseLatestDelegateResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolCanSetUseLatestDelegateResponse{}

	// Create Megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Return if requested setting change is the same as current setting
	currentSetting, err := mega.GetUseLatestDelegate(nil)
	if err != nil {
		return nil, err
	}
	if currentSetting == useLatest {
		response.MatchesCurrentSetting = true
		return &response, nil
	}
	response.MatchesCurrentSetting = false

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	gasLimits, err := mega.EstimateSetUseLatestDelegateGas(useLatest, opts)
	if err == nil {
		response.GasLimits = gasLimits
	}

	// Return response
	return &response, nil

}

func setUseLatestDelegate(c *cli.Command, useLatest bool, opts *bind.TransactOpts) (*api.MegapoolSetUseLatestDelegateResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Get node account and derive the megapool address
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolSetUseLatestDelegateResponse{}

	// Create megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Set the new setting
	hash, err := mega.SetUseLatestDelegate(useLatest, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func getDelegate(c *cli.Command) (*api.MegapoolGetDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolGetDelegateResponse{}

	// Get node account and derive the megapool address
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Create megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mega.GetDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}
	response.Address = address

	// Return response
	return &response, nil

}

func getEffectiveDelegate(c *cli.Command) (*api.MegapoolGetEffectiveDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolGetEffectiveDelegateResponse{}

	// Get node account and derive the megapool address
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Create Megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mega.GetEffectiveDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}

	// Return response
	response.Address = address
	return &response, nil
}

func getUseLatestDelegateHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := getUseLatestDelegate(c)
		response.WriteResponse(w, resp, err)
	}
}

func canDelegateUpgradeHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.URL.Query().Get("address"))
		resp, err := canDelegateUpgrade(c, address)
		response.WriteResponse(w, resp, err)
	}
}

func delegateUpgradeHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.FormValue("address"))
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := delegateUpgrade(c, address, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canSetUseLatestDelegateHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setLatest, err := parseBool(r, "setLatest")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canSetUseLatestDelegate(c, setLatest)
		response.WriteResponse(w, resp, err)
	}
}

func setUseLatestDelegateHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setting, err := parseBool(r, "setting")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := setUseLatestDelegate(c, setting, opts)
		response.WriteResponse(w, resp, err)
	}
}

func getDelegateHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := getDelegate(c)
		response.WriteResponse(w, resp, err)
	}
}

func getEffectiveDelegateHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := getEffectiveDelegate(c)
		response.WriteResponse(w, resp, err)
	}
}
