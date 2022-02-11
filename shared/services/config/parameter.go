package config

import (
	"fmt"
	"reflect"
	"strconv"
)

// A parameter that can be configured by the user
type Parameter struct {
	ID                   string
	Name                 string
	Description          string
	Type                 ParameterType
	Default              map[Network]interface{}
	Advanced             bool
	AffectsContainers    []ContainerID
	EnvironmentVariables []string
	CanBeBlank           bool
	OverwriteOnUpgrade   bool
	Options              []ParameterOption
	Value                interface{}
}

// A single option in a choice parameter
type ParameterOption struct {
	ID          string
	Name        string
	Description string
	Value       interface{}
}

// Apply a network change to a parameter
func changeNetworkForParameter(parameter *Parameter, oldNetwork Network, newNetwork Network) {

	// Get the current value and the defaults per-network
	currentValue := parameter.Value
	oldDefault, exists := parameter.Default[oldNetwork]
	if !exists {
		oldDefault = parameter.Default[Network_All]
	}
	newDefault, exists := parameter.Default[newNetwork]
	if !exists {
		newDefault = parameter.Default[Network_All]
	}

	// If the old value matches the old default, replace it with the new default
	if currentValue == oldDefault {
		parameter.Value = newDefault
	}

}

// Serializes the parameter's value into a string
func (param *Parameter) serialize(serializedParams map[string]string) {
	var value string
	if param.Value == nil {
		value = ""
	} else {
		value = fmt.Sprint(param.Value)
	}

	serializedParams[param.ID] = value
}

// Deserializes a map of settings into this parameter
func (param *Parameter) deserialize(serializedParams map[string]string) error {
	value, exists := serializedParams[param.ID]
	if !exists {
		value = ""
	}

	var err error
	switch param.Type {
	case ParameterType_Int:
		param.Value, err = strconv.ParseInt(value, 0, 0)
	case ParameterType_Uint:
		param.Value, err = strconv.ParseUint(value, 0, 0)
	case ParameterType_Uint16:
		param.Value, err = strconv.ParseUint(value, 0, 16)
	case ParameterType_Bool:
		param.Value, err = strconv.ParseBool(value)
	case ParameterType_String:
		param.Value = value
	case ParameterType_Choice:
		// The more complicated one since Go doesn't have generics
		// Get the value of the first option, get its type, and convert the value to that
		if len(param.Options) < 1 {
			err = fmt.Errorf("this parameter is marked as a choice but does not have any options")
		} else {
			valueType := reflect.TypeOf(value)
			paramType := reflect.TypeOf(param.Options[0].Value)
			if !valueType.ConvertibleTo(paramType) {
				err = fmt.Errorf("value type %s cannot be converted to parameter type %s", valueType.Name(), paramType.Name())
			} else {
				param.Value = reflect.ValueOf(value).Convert(paramType)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("cannot deserialize parameter [%s]: %w", param.ID, err)
	}
	return nil
}
