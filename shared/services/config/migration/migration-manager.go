package migration

type ConfigMigrationManager struct {
}

func (m *ConfigMigrationManager) UpdateConfig(serializedConfig map[string]map[string]string) (map[string]map[string]string, error) {

}

// TODO: array of config managers, find the one with the correct version, cascade updates from it to the latest
