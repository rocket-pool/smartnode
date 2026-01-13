package migration

func upgradeFromV1189(serializedConfig map[string]map[string]string) error {

	// If the previous priority fee is 2, set it to 0.01
	priorityFee, exists := serializedConfig["smartnode"]["priorityFee"]
	if exists && priorityFee == "2" {
		serializedConfig["smartnode"]["priorityFee"] = "0.01"
	}

	// If the previous auto tx gas threshold is 150, set it to 20
	autoTxGasThreshold, exists := serializedConfig["smartnode"]["minipoolStakeGasThreshold"]
	if exists && autoTxGasThreshold == "150" {
		serializedConfig["smartnode"]["minipoolStakeGasThreshold"] = "20"
	}
	return nil
}
