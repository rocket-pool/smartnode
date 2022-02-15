package config

import (
	"fmt"
	"reflect"
	"strconv"
)

// A parameter that can be configured by the user
type Parameter struct {
	ID                   string                  `yaml:"id,omitempty"`
	Name                 string                  `yaml:"name,omitempty"`
	Description          string                  `yaml:"description,omitempty"`
	Type                 ParameterType           `yaml:"type,omitempty"`
	Default              map[Network]interface{} `yaml:"default,omitempty"`
	Advanced             bool                    `yaml:"advanced,omitempty"`
	AffectsContainers    []ContainerID           `yaml:"affectsContainers,omitempty"`
	EnvironmentVariables []string                `yaml:"environmentVariables,omitempty"`
	CanBeBlank           bool                    `yaml:"canBeBlank,omitempty"`
	OverwriteOnUpgrade   bool                    `yaml:"overwriteOnUpgrade,omitempty"`
	Options              []ParameterOption       `yaml:"options,omitempty"`
	Value                interface{}             `yaml:"-"`
}

// A single option in a choice parameter
type ParameterOption struct {
	ID          string      `yaml:"id,omitempty"`
	Name        string      `yaml:"name,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Value       interface{} `yaml:"value,omitempty"`
}

// Apply a network change to a parameter
func (param *Parameter) changeNetwork(oldNetwork Network, newNetwork Network) {

	// Get the current value and the defaults per-network
	currentValue := param.Value
	oldDefault, exists := param.Default[oldNetwork]
	if !exists {
		oldDefault = param.Default[Network_All]
	}
	newDefault, exists := param.Default[newNetwork]
	if !exists {
		newDefault = param.Default[Network_All]
	}

	// If the old value matches the old default, replace it with the new default
	if currentValue == oldDefault {
		param.Value = newDefault
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

// Set the value to the default for the provided config's network
func (param *Parameter) setToDefault(config *MasterConfig) error {
	defaultSetting, err := param.GetDefault(config)
	if err != nil {
		return err
	}
	param.Value = defaultSetting
	return nil
}

// Get the default value based on the provided configuration's current network
func (param *Parameter) GetDefault(config *MasterConfig) (interface{}, error) {
	currentNetwork := config.Smartnode.Network.Value.(Network)
	defaultSetting, exists := param.Default[currentNetwork]
	if !exists {
		defaultSetting, exists = param.Default[Network_All]
		if !exists {
			return nil, fmt.Errorf("%s doesn't have a default for network %s or all networks", param.ID, currentNetwork)
		}
	}

	return defaultSetting, nil
}
