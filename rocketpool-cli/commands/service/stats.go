package service

import (
	"fmt"
)

// View the Rocket Pool service stats
func serviceStats() error {
	fmt.Println("No longer supported - please run `docker stats -a` instead.")
	return nil
}
