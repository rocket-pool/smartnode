package server

import (
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool/api/handlers"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Handles a Node daemon response
func HandleResponse[DataType any](w http.ResponseWriter, response *api.ApiResponse[DataType], err error) {
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

// Handles an error related to parsing the input parameters of a request
func HandleInputError(w http.ResponseWriter, err error) {
	// Write out any errors
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
}

func RegisterSingleStageHandler[DataType any, ContextType any, HandlerType handlers.ISingleStageCallHandler[DataType, ContextType]](
	router *mux.Router,
	packageName string,
	functionName string,
	constructor func(vars map[string]string) (HandlerType, error),
	runner func(handlers.ISingleStageCallHandler[DataType, ContextType]) (*api.ApiResponse[DataType], error),
) {
	router.HandleFunc(fmt.Sprintf("/%s/%s", packageName, functionName), func(w http.ResponseWriter, r *http.Request) {
		// Create the handler
		vars := mux.Vars(r)
		handler, err := constructor(vars)
		if err != nil {
			HandleInputError(w, err)
			return
		}

		// Run the body
		response, err := runner(handler)
		HandleResponse(w, response, err)
	})
}
