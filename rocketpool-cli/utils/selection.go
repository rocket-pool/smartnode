package utils

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/node-manager-core/cli/utils"
	"github.com/urfave/cli/v2"
)

// Get the list of elements the user wants to use for a multi-select operation
func GetMultiselectIndices[DataType any](c *cli.Context, flagName string, options []utils.SelectionOption[DataType], prompt string) ([]*DataType, error) {
	flag := ""
	if c.IsSet(flagName) {
		flag = strings.TrimSpace(fmt.Sprintf("%v", c.Value(flagName)))
	}

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
		return utils.ParseOptionIDs(flag, options)
	}

	// No headless flag, so prompt interactively
	if c.Bool(YesFlag.Name) {
		return nil, fmt.Errorf("the '%s' flag (non-interactive mode) is specified but the '%s' flag (selection) is missing", YesFlag.Name, flagName)
	}
	fmt.Println(prompt)
	fmt.Println()
	for i, option := range options {
		fmt.Printf("%d: %s\n", i+1, option.Display)
	}
	fmt.Println()
	indexSelection := Prompt("Use a comma separated list (such as '1,2,3' or '1-4,6-8,10') or leave it blank to select all options.", ".*", "Invalid index selection")
	return utils.ParseIndexSelection(indexSelection, options)
}
