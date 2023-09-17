package server

import (
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
)

// Function for validating an argument (wraps the old CLI validators)
type ArgValidator[ArgType any] func(string, string) (ArgType, error)

// Validates an argument, ensuring it exists and can be converted to the required type
func ValidateArg[ArgType any](name string, args map[string]string, impl ArgValidator[ArgType], result_Out *ArgType) error {
	// Make sure it exists
	arg, exists := args[name]
	if !exists {
		return fmt.Errorf("missing argument '%s'", name)
	}

	// Run the parser
	result, err := impl(name, arg)
	if err != nil {
		return err
	}

	// Set the result
	*result_Out = result
	return nil
}

// Gets a string argument, ensuring that it exists in the provided vars list
func GetStringFromVars(name string, args map[string]string, result_Out *string) error {
	// Make sure it exists
	arg, exists := args[name]
	if !exists {
		return fmt.Errorf("missing argument '%s'", name)
	}

	// Set the result
	*result_Out = arg
	return nil
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
