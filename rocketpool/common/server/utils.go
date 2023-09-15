package server

import (
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
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
