package migration

func upgradeFromV1210(serializedConfig map[string]map[string]string) error {
	smartnode, exists := serializedConfig["smartnode"]
	if !exists {
		return nil
	}

	port, exists := smartnode["apiPort"]
	if !exists || port == "" {
		return nil
	}

	api, exists := serializedConfig["api"]
	if !exists {
		api = map[string]string{}
		serializedConfig["api"] = api
	}
	if api["apiPort"] == "" {
		api["apiPort"] = port
	}
	delete(smartnode, "apiPort")
	return nil
}
