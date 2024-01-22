package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/utils/sys"
)

// Get the list of features required for modern client containers but not supported by the CPU
func checkCpuFeatures() error {
	unsupportedFeatures := sys.GetMissingModernCpuFeatures()
	if len(unsupportedFeatures) > 0 {
		fmt.Println("Your CPU is missing support for the following features:")
		for _, name := range unsupportedFeatures {
			fmt.Printf("  - %s\n", name)
		}

		fmt.Println("\nYou must use the 'portable' image.")
		return nil
	}

	fmt.Println("Your CPU supports all required features for 'modern' images.")
	return nil
}
