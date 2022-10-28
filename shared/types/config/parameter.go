package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// A parameter that can be configured by the user
type Parameter struct {
	ID                    string                  `yaml:"id,omitempty"`
	Name                  string                  `yaml:"name,omitempty"`
	Description           string                  `yaml:"description,omitempty"`
	Type                  ParameterType           `yaml:"type,omitempty"`
	Default               map[Network]interface{} `yaml:"default,omitempty"`
	MaxLength             int                     `yaml:"maxLength,omitempty"`
	Regex                 string                  `yaml:"regex,omitempty"`
	Advanced              bool                    `yaml:"advanced,omitempty"`
	AffectsContainers     []ContainerID           `yaml:"affectsContainers,omitempty"`
	EnvironmentVariables  []string                `yaml:"environmentVariables,omitempty"`
	CanBeBlank            bool                    `yaml:"canBeBlank,omitempty"`
	OverwriteOnUpgrade    bool                    `yaml:"overwriteOnUpgrade,omitempty"`
	Options               []ParameterOption       `yaml:"options,omitempty"`
	Value                 interface{}             `yaml:"-"`
	DescriptionsByNetwork map[Network]string      `yaml:"-"`
}

// A single option in a choice parameter
type ParameterOption struct {
	Name        string      `yaml:"name,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Value       interface{} `yaml:"value,omitempty"`
}

// Apply a network change to a parameter
func (param *Parameter) ChangeNetwork(oldNetwork Network, newNetwork Network) {

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

	// Update the description, if applicable
	param.UpdateDescription(newNetwork)
}

// Serializes the parameter's value into a string
func (param *Parameter) Serialize(serializedParams map[string]string) {
	var value string
	if param.Value == nil {
		value = ""
	} else {
		value = fmt.Sprint(param.Value)
	}

	serializedParams[param.ID] = value
}

// Deserializes a map of settings into this parameter
func (param *Parameter) Deserialize(serializedParams map[string]string, network Network) error {
	// Update the description, if applicable
	param.UpdateDescription(network)

	value, exists := serializedParams[param.ID]
	if !exists {
		return param.SetToDefault(network)
	}

	var err error
	switch param.Type {
	case ParameterType_Int:
		param.Value, err = strconv.ParseInt(value, 0, 0)
	case ParameterType_Uint:
		param.Value, err = strconv.ParseUint(value, 0, 0)
	case ParameterType_Uint16:
		var result uint64
		result, err = strconv.ParseUint(value, 0, 16)
		param.Value = uint16(result)
	case ParameterType_Bool:
		param.Value, err = strconv.ParseBool(value)
	case ParameterType_String:
		if param.Regex != "" {
			regex := regexp.MustCompile(param.Regex)
			if param.Value != "" && !regex.MatchString(value) {
				return fmt.Errorf("cannot deserialize parameter [%s]: value [%s] did not match the expected format", param.ID, value)
			}
		}
		if param.MaxLength > 0 {
			if len(value) > param.MaxLength {
				return fmt.Errorf("cannot deserialize parameter [%s]: value [%s] is longer than the max length of [%d]", param.ID, value, param.MaxLength)
			}
		}
		if !param.CanBeBlank && value == "" {
			return param.SetToDefault(network)
		}
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
				param.Value = reflect.ValueOf(value).Convert(paramType).Interface()
			}
		}
	case ParameterType_Float:
		param.Value, err = strconv.ParseFloat(value, 64)
	}

	if err != nil {
		return fmt.Errorf("cannot deserialize parameter [%s]: %w", param.ID, err)
	}

	return nil
}

// Set the value to the default for the provided config's network
func (param *Parameter) SetToDefault(network Network) error {
	defaultSetting, err := param.GetDefault(network)
	if err != nil {
		return err
	}
	param.Value = defaultSetting
	return nil
}

// Get the default value for the provided network
func (param *Parameter) GetDefault(network Network) (interface{}, error) {
	defaultSetting, exists := param.Default[network]
	if !exists {
		defaultSetting, exists = param.Default[Network_All]
		if !exists {
			return nil, fmt.Errorf("%s doesn't have a default for network %s or all networks", param.ID, network)
		}
	}

	return defaultSetting, nil
}

// Set the network-specific description of the parameter
func (param *Parameter) UpdateDescription(network Network) {
	if param.DescriptionsByNetwork != nil {
		newDesc, exists := param.DescriptionsByNetwork[network]
		if exists {
			param.Description = newDesc
		}
	}
}

// Add the parameters to the collection of environment variabes
func AddParametersToEnvVars(params []*Parameter, envVars map[string]string) {
	for _, param := range params {
		for _, envVar := range param.EnvironmentVariables {
			if envVar != "" {
				envVars[envVar] = fmt.Sprint(param.Value)
			}
		}
	}
}
