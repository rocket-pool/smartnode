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

func RegisterSingleStageHandler[DataType any, ContextType any, ImplType any, HandlerType handlers.ISingleStageCallHandler[DataType, ContextType, ImplType]](
	router *mux.Router,
	packageName string,
	functionName string,
	inputParsers []func(h HandlerType, vars map[string]string) error,
	runner func(HandlerType) (*api.ApiResponse[DataType], error),
) {

	router.HandleFunc(fmt.Sprintf("/%s/%s", packageName, functionName), func(w http.ResponseWriter, r *http.Request) {
		// Create the handler
		var handlerImpl ImplType
		handler := HandlerType(&handlerImpl)

		// Parse the input
		vars := mux.Vars(r)
		for _, parser := range inputParsers {
			err := parser(handler, vars)
			if err != nil {
				HandleInputError(w, err)
				return
			}
		}

		// Run the body
		response, err := runner(handler)
		HandleResponse(w, response, err)
	})

}
