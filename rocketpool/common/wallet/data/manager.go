package data

import (
	"fmt"
	"io/fs"
	"os"
)

// File manager
type DataManager[DataType any] struct {
	name          string
	path          string
	fileMode      fs.FileMode
	serializer    dataSerializer[DataType]
	isInitialized bool
	data          DataType
	hasValue      bool
}

// Create a new general data manager that can load from and save to a file
func NewDataManager[DataType any](name string, path string, fileMode fs.FileMode, serializer dataSerializer[DataType]) *DataManager[DataType] {
	var data DataType
	return &DataManager[DataType]{
		name:          name,
		path:          path,
		fileMode:      fileMode,
		serializer:    serializer,
		data:          data,
		isInitialized: false,
	}
}

// Checks whether or not a value has been set
func (m *DataManager[DataType]) HasValue() bool {
	return m.hasValue
}

// Get the data - if it isn't loaded yet, read it from disk
func (m *DataManager[DataType]) InitializeData() (DataType, bool, error) {
	// Done if it's already initialized
	if m.isInitialized {
		return m.data, true, nil
	}

	// Check if the file exists on disk
	_, err := os.Stat(m.path)
	if os.IsNotExist(err) {
		m.hasValue = false
		m.isInitialized = true
		return m.data, false, nil
	} else if err != nil {
		return m.data, false, fmt.Errorf("error checking %s file path: %w", m.name, err)
	}

	// Load the file if it exists
	value, err := os.ReadFile(m.path)
	if err != nil {
		return m.data, false, fmt.Errorf("error reading %s file: %w", m.name, err)
	}

	// Deserialize
	m.data, err = m.serializer.deserialize(value)
	if err != nil {
		return m.data, false, fmt.Errorf("error deserializing %s file: %w", m.name, err)
	}
	m.isInitialized = true
	m.hasValue = true
	return m.data, true, err
}

// Gets the data and whether or not it's been set
func (m *DataManager[DataType]) Get() (DataType, bool) {
	return m.data, m.hasValue
}

// Gets the data as a string, and whether or not it's been set
func (m *DataManager[DataType]) String() (string, bool, error) {
	if m.hasValue {
		// Serialize the data
		bytes, err := m.serializer.serialize(m.data)
		if err != nil {
			return "", true, fmt.Errorf("error serializing %s data: %w", m.name, err)
		}

		return string(bytes), true, nil
	}
	return "", false, nil
}

// Sets the data value
func (m *DataManager[DataType]) Set(data DataType) {
	m.data = data
	m.hasValue = true
	m.isInitialized = true
}

// Clears the data value, setting it to the data type's default value
func (m *DataManager[DataType]) Clear() {
	var data DataType
	m.data = data
	m.hasValue = false
	m.isInitialized = true
}

// Stores the data to disk
func (m *DataManager[DataType]) Save() error {
	if !m.hasValue {
		return fmt.Errorf("data has not been set")
	}

	// Serialize the data
	bytes, err := m.serializer.serialize(m.data)
	if err != nil {
		return fmt.Errorf("error serializing %s data: %w", m.name, err)
	}

	// Write it to disk
	if m.path != "" {
		err := os.WriteFile(m.path, bytes, m.fileMode)
		if err != nil {
			return fmt.Errorf("error writing %s to disk: %w", m.name, err)
		}
	}
	return nil
}

// Delete the file from disk
func (m *DataManager[DataType]) Delete() error {
	if m.path == "" {
		return nil
	}

	// Check if it exists
	_, err := os.Stat(m.path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking %s file path: %w", m.name, err)
	}

	// Delete it
	err = os.Remove(m.path)
	if err != nil {
		return fmt.Errorf("error deleting %s file: %w", m.name, err)
	}
	return nil
}
