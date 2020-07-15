package cli

import (
    "fmt"

    "github.com/urfave/cli"
)


// Check command argument count
func CheckArgCount(c *cli.Context, count int) error {
    if len(c.Args()) != count {
        return cli.NewExitError(fmt.Sprintf("USAGE:\n   %s", c.Command.UsageText), 1)
    }
    return nil
}


// Check API command argument count
func CheckAPIArgCount(c *cli.Context, count int) error {
    if len(c.Args()) != count {
        return fmt.Errorf("Incorrect argument count; usage: %s", c.Command.UsageText)
    }
    return nil
}

