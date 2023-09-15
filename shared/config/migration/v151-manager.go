package migration

import "fmt"

func upgradeFromV151(serializedConfig map[string]map[string]string) error {
	// v1.5.1 had the Nimbus BN additional flags named differently
	nimbusSettings, exists := serializedConfig["nimbus"]
	if !exists {
		return fmt.Errorf("expected a section called `nimbus` but it didn't exist")
	}
	additionalFlags, exists := nimbusSettings["additionalFlags"]
	if !exists {
		return fmt.Errorf("expected a Nimbus setting named `additionalFlags` but it didn't exist")
	}

	// Update the config
	nimbusSettings["additionalBnFlags"] = additionalFlags
	serializedConfig["nimbus"] = nimbusSettings

	return nil
}
