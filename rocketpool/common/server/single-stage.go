package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/goccy/go-json"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Wrapper for callbacks used by call runners that follow a common single-stage pattern:
// Create bindings, query the chain, and then do whatever else they want.
// Structs implementing this will handle the caller-specific functionality.
type ISingleStageCallContext[DataType any] interface {
	// Initialize the context with any bootstrapping, requirements checks, or bindings it needs to set up
	Initialize() error

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(mc *batch.MultiCaller)

	// Prepare the response data in whatever way the context needs to do
	PrepareData(data *DataType, opts *bind.TransactOpts) error
}

// Interface for single-stage call context factories - these will be invoked during route handling to create the
// unique context for the route
type ISingleStageCallContextFactory[ContextType ISingleStageCallContext[DataType], DataType any] interface {
	// Create the context for the route
	Create(args url.Values) (ContextType, error)
}

// Registers a new route with the router, which will invoke the provided factory to create and execute the context
// for the route when it's called; use this for typical general-purpose calls
func RegisterSingleStageRoute[ContextType ISingleStageCallContext[DataType], DataType any](
	router *mux.Router,
	functionName string,
	factory ISingleStageCallContextFactory[ContextType, DataType],
	serviceProvider *services.ServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			handleInvalidMethod(w)
			return
		}

		// Create the handler and deal with any input validation errors
		context, err := factory.Create(r.URL.Query())
		if err != nil {
			handleInputError(w, err)
			return
		}

		// Run the context's processing routine
		response, err := runSingleStageRoute[DataType](context, serviceProvider)
		handleResponse(w, response, err)
	})
}

// Registers a new route with the router, which will invoke the provided factory to create and execute the context
// for the route when it's called via POST; use this for typical general-purpose calls
func RegisterSingleStagePost[ContextType ISingleStageCallContext[DataType], BodyType any, DataType any](
	router *mux.Router,
	functionName string,
	factory ISingleStageCallContextFactory[ContextType, BodyType, DataType],
	serviceProvider *services.ServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handleInvalidMethod(w)
			return
		}

		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			handleInputError(w, fmt.Errorf("error reading request body: %w", err))
			return
		}

		// Deserialize the body
		var body BodyType
		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			handleInputError(w, fmt.Errorf("error deserializing request body: %w", err))
			return
		}

		// Create the handler and deal with any input validation errors
		context, err := factory.Create(body)
		if err != nil {
			handleInputError(w, err)
			return
		}

		// Run the context's processing routine
		response, err := runSingleStageRoute[DataType](context, serviceProvider)
		handleResponse(w, response, err)
	})
}

// Run a route registered with the common single-stage querying pattern
func runSingleStageRoute[DataType any](ctx ISingleStageCallContext[DataType], serviceProvider *services.ServiceProvider) (*api.ApiResponse[DataType], error) {
	// Get the services
	w := serviceProvider.GetWallet()
	rp := serviceProvider.GetRocketPool()

	// Initialize the context with any bootstrapping, requirements checks, or bindings it needs to set up
	err := ctx.Initialize()
	if err != nil {
		return nil, err
	}

	// Get the context-specific contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		ctx.GetState(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get the transact opts if this node is ready for transaction
	var opts *bind.TransactOpts
	walletStatus := w.GetStatus()
	if walletStatus == sharedtypes.WalletStatus_Ready {
		var err error
		opts, err = w.GetTransactor()
		if err != nil {
			return nil, fmt.Errorf("error getting node account transactor: %w", err)
		}
	}

	// Create the response and data
	data := new(DataType)
	response := &api.ApiResponse[DataType]{
		WalletStatus: walletStatus,
		Data:         data,
	}

	// Prep the data with the context-specific behavior
	err = ctx.PrepareData(data, opts)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}
