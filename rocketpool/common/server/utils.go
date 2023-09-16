package server

import (
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/gorilla/mux"
)

// Handles a Node daemon response
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

func RegisterSingleStageRoute[ContextType ISingleStageCallContext[DataType, CommonContextType], DataType any, CommonContextType any](
	router *mux.Router,
	functionName string,
	factory IContextFactory[ContextType, DataType, CommonContextType],
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		// Create the handler
		vars := mux.Vars(r)
		context, err := factory.Create(vars)
		if err != nil {
			HandleInputError(w, err)
			return
		}

		// Run the body
		response, err := factory.Run(context)
		HandleResponse(w, response, err)
	})
}
