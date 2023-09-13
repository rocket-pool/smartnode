package faucet

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/cli"

	types "github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Access the legacy RPL faucet",
		Subcommands: []cli.Command{
			// Status
			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the faucet's status",
				UsageText: "rocketpool api faucet status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runFaucetCall[types.FaucetStatusResponse](c, &faucetStatusHandler{})
					api.PrintResponse(response, err)
					return nil

				},
			},

			// Withdraw RPL
			{
				Name:      "withdraw-rpl",
				Aliases:   []string{"w"},
				Usage:     "Withdraw legacy RPL from the faucet",
				UsageText: "rocketpool api faucet withdraw-rpl",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runFaucetCall[types.FaucetWithdrawRplResponse](c, &faucetWithdrawHandler{})
					api.PrintResponse(response, err)
					return nil

				},
			},
		},
	})
}

func RegisterRoutes(router *mux.Router, name string, handler ResponseHandler) {
	route := "faucet"

	// Status
	router.HandleFunc(fmt.Sprintf("/%s/status", route), func(w http.ResponseWriter, r *http.Request) {
		response, err := runFaucetCall[types.FaucetStatusResponse](c, &faucetStatusHandler{})
		handler(w, response, err)
	})
}

type ResponseHandler func(w http.ResponseWriter, response any, err error)

func HandleResponse(w http.ResponseWriter, response any, err error) {
	// Write out any errors
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	// Write the serialized response
	bytes, err := json.Marshal(response)
	if err != nil {
		err = fmt.Errorf("error serializing response: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}
