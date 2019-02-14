package database

import (
    "errors"

    "github.com/rapidloop/skv"
)


// Account manager
type Database struct {
    store *skv.KVStore
}


/**
 * Create database
 */
func NewDatabase(databasePath string) (*Database, error) {

    // Initialise storage
    store, err := skv.Open(databasePath)
    if err != nil {
        return nil, errors.New("Error initialising database: " + err.Error())
    }

    // Create & return database
    return &Database{
        store: store,
    }, nil

}


/**
 * Close database
 */
func (db *Database) Close() error {
    err := db.store.Close()
    if err != nil {
        return errors.New("Error closing database: " + err.Error())
    }
    return nil
}


/**
 * Read a value from the database
 */
func (db *Database) Get(key string, value interface{}) error {
    err := db.store.Get(key, value)
    if err != nil {
        return errors.New("Error reading value from database: " + err.Error())
    }
    return nil
}


/**
 * Write a value to the database
 */
func (db *Database) Put(key string, value interface{}) error {
    err := db.store.Put(key, value)
    if err != nil {
        return errors.New("Error writing value to database: " + err.Error())
    }
    return nil
}


/**
 * Remove a value from the database
 */
func (db *Database) Delete(key string) error {
    err := db.store.Delete(key)
    if err != nil {
        return errors.New("Error removing value from database: " + err.Error())
    }
    return nil
}

