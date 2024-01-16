package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

// An option that can be selected from a list of choices in the CLI
type SelectionOption[DataType any] struct {
	// The underlying element this option represents
	Element *DataType

	// The human-readable ID of the option, used for non-interactive selection
	ID string

	// The text to display for this option when listing the available options
	Display string
}

// Get the list of elements the user wants to use for a multi-select operation
func GetMultiselectIndices[DataType any](c *cli.Context, flagName string, options []SelectionOption[DataType], prompt string) ([]*DataType, error) {
	flag := strings.TrimSpace(c.String(flagName))

	// Handle all
	if flag == "all" {
		selectedElements := make([]*DataType, len(options))
		for i, option := range options {
			selectedElements[i] = option.Element
		}
		return selectedElements, nil
	}

	// Handle one or more
	if flag != "" {
		return parseOptionIDs(flag, options)
	}

	// No headless flag, so prompt interactively
	if c.Bool(YesFlag.Name) {
		return nil, fmt.Errorf("the '%s' flag (non-interactive mode) is specified but the '%s' flag (selection) is missing", YesFlag.Name, flagName)
	}
	indexSelection := Prompt("%s\nUse a comma separated list (such as '1,2,3') or leave it blank to select all options.", "^$|^\\d+(,\\d+)*$", "Invalid index selection")
	return parseIndexSelection(indexSelection, options)
}

// Parse a comma-separated list of indices to select in a multi-index operation
func parseIndexSelection[DataType any](selectionString string, options []SelectionOption[DataType]) ([]*DataType, error) {
	// Select all
	if selectionString == "" {
		selectedElements := make([]*DataType, len(options))
		for i, option := range options {
			selectedElements[i] = option.Element
		}
		return selectedElements, nil
	}

	// Trim spaces
	elements := strings.Split(selectionString, ",")
	trimmedElements := make([]string, len(elements))
	for i, element := range elements {
		trimmedElements[i] = strings.TrimSpace(element)
	}

	// Remove duplicates
	uniqueElements := make([]string, 0, len(elements))
	seenIndices := map[string]bool{}
	for _, element := range trimmedElements {
		_, exists := seenIndices[element]
		if !exists {
			uniqueElements = append(uniqueElements, element)
			seenIndices[element] = true
		}
	}

	// Validate
	optionLength := uint64(len(options))
	selectedElements := make([]*DataType, len(uniqueElements))
	for i, element := range uniqueElements {
		// Parse it
		index, err := strconv.ParseUint(element, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing index '%s': %w", element, err)
		}

		// Make sure it's in the list of options
		if index >= optionLength {
			return nil, fmt.Errorf("selection '%s' is not a valid option", element)
		}
		selectedElements[i] = options[index].Element
	}
	return selectedElements, nil
}

// Parse a comma-separated list of option IDs to select in a multi-index operation
func parseOptionIDs[DataType any](selectionString string, options []SelectionOption[DataType]) ([]*DataType, error) {
	elements := strings.Split(selectionString, ",")

	// Trim spaces
	trimmedElements := make([]string, len(elements))
	for i, element := range elements {
		trimmedElements[i] = strings.TrimSpace(element)
	}

	// Remove duplicates
	uniqueElements := make([]string, 0, len(elements))
	seenIndices := map[string]bool{}
	for _, element := range trimmedElements {
		_, exists := seenIndices[element]
		if !exists {
			uniqueElements = append(uniqueElements, element)
			seenIndices[element] = true
		}
	}

	// Validate
	selectedElements := make([]*DataType, len(uniqueElements))
	for i, element := range uniqueElements {
		// Make sure it's in the list of options
		found := false
		for _, option := range options {
			if option.ID == element {
				found = true
				selectedElements[i] = option.Element
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("element '%s' is not a valid option", element)
		}
	}
	return selectedElements, nil
}
