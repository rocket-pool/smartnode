package server

import (
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
)

// Registers a new route with the router, which will invoke the provided factory to create and execute the context
// for the route when it's called; use this for typical general-purpose calls
func RegisterSingleStageRoute[ContextType ISingleStageCallContext[DataType], DataType any](
	router *mux.Router,
	functionName string,
	factory ISingleStageCallContextFactory[ContextType, DataType],
	serviceProvider *services.ServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		// Create the handler and deal with any input validation errors
		vars := mux.Vars(r)
		context, err := factory.Create(vars)
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
// for the route when it's called; use this for complex calls that will iterate over and query each minipool in the node
func RegisterMinipoolRoute[ContextType IMinipoolCallContext[DataType], DataType any](
	router *mux.Router,
	functionName string,
	factory IMinipoolCallContextFactory[ContextType, DataType],
	serviceProvider *services.ServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		// Create the handler and deal with any input validation errors
		vars := mux.Vars(r)
		context, err := factory.Create(vars)
		if err != nil {
			handleInputError(w, err)
			return
		}

		// Run the context's processing routine
		response, err := runMinipoolRoute[DataType](context, serviceProvider)
		handleResponse(w, response, err)
	})
}

// Handles an error related to parsing the input parameters of a request
func handleInputError(w http.ResponseWriter, err error) {
	// Write out any errors
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
}

// Handles a Node daemon response
func handleResponse(w http.ResponseWriter, response any, err error) {
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
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}
